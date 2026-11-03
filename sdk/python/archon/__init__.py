"""Archon AI Python SDK for local agent functions.

The SDK deliberately uses only the Python standard library. Runtime bindings
are configured through environment variables injected by the agent runner:

- ``ARCHON_API_URL`` and ``ARCHON_API_TOKEN`` for an Archon HTTP gateway.
- ``ARCHON_PROJECT`` for project-scoped vector bindings.
- ``ARCHON_VECTOR_URL`` for a text-capable vector binding gateway.
- ``ARCHON_QDRANT_URL`` for direct Qdrant vector searches.
- ``ARCHON_LLM_URL`` or ``OPENAI_BASE_URL`` for an OpenAI-compatible gateway.
"""

from __future__ import annotations

import asyncio
import json
import os
from dataclasses import dataclass
from typing import Any, AsyncIterator, Awaitable, Callable, Mapping, Sequence
from urllib.error import HTTPError, URLError
from urllib.parse import quote
from urllib.request import Request, urlopen


class SDKError(RuntimeError):
    """Raised when a runtime binding cannot complete an HTTP request."""


@dataclass
class VectorHit:
    id: Any
    score: float | None
    payload: Mapping[str, Any]

    @classmethod
    def from_json(cls, value: Mapping[str, Any]) -> "VectorHit":
        return cls(
            id=value.get("id"),
            score=value.get("score"),
            payload=value.get("payload") or {},
        )


class _Agent:
    def handler(self, fn: Callable[..., Awaitable[Any]]):
        fn.__archon_handler__ = True
        return fn


agent = _Agent()


def _env(name: str, fallback: str | None = None) -> str | None:
    value = os.environ.get(name, fallback)
    return value.rstrip("/") if value else value


def _headers(token: str | None = None) -> dict[str, str]:
    result = {"Accept": "application/json", "Content-Type": "application/json"}
    if token:
        result["Authorization"] = f"Bearer {token}"
    return result


def _request_json(
    url: str,
    payload: Mapping[str, Any] | None = None,
    *,
    token: str | None = None,
    timeout: float = 30.0,
) -> Any:
    body = None if payload is None else json.dumps(payload).encode("utf-8")
    request = Request(url, data=body, headers=_headers(token), method="POST" if body else "GET")
    try:
        with urlopen(request, timeout=timeout) as response:
            raw = response.read()
    except HTTPError as error:
        detail = error.read(4096).decode("utf-8", "replace")
        raise SDKError(f"HTTP {error.code} from {url}: {detail}") from error
    except URLError as error:
        raise SDKError(f"request to {url} failed: {error.reason}") from error
    except TimeoutError as error:
        raise SDKError(f"request to {url} timed out") from error
    if not raw:
        return {}
    try:
        return json.loads(raw.decode("utf-8"))
    except json.JSONDecodeError as error:
        raise SDKError(f"invalid JSON response from {url}") from error


class _VectorClient:
    def __init__(self, instance: str):
        if not instance.strip():
            raise ValueError("vector instance is required")
        self.instance = instance
        self.api_url = _env("ARCHON_API_URL")
        self.project = os.environ.get("ARCHON_PROJECT", "")
        self.gateway_url = _env("ARCHON_VECTOR_URL")
        self.qdrant_url = _env("ARCHON_QDRANT_URL")
        self.token = os.environ.get("ARCHON_API_TOKEN")

    async def query(
        self,
        text: str | None = None,
        top_k: int = 5,
        *,
        vector: Sequence[float] | None = None,
    ) -> list[VectorHit]:
        """Search a project binding or Qdrant collection.

        A configured ``ARCHON_VECTOR_URL`` receives ``instance``, ``text``,
        and ``topK`` and is intended for an embedding-aware runtime gateway.
        A configured ``ARCHON_QDRANT_URL`` uses Qdrant's native search API and
        requires the caller to provide an already embedded ``vector``.
        """
        if top_k < 1 or top_k > 1000:
            raise ValueError("top_k must be between 1 and 1000")
        if text is None and vector is None:
            raise ValueError("text or vector is required")
        if self.gateway_url:
            url = f"{self.gateway_url}/v1/vector/query"
            payload: dict[str, Any] = {"instance": self.instance, "topK": top_k}
            if text is not None:
                payload["text"] = text
            if vector is not None:
                payload["vector"] = list(vector)
            response = await asyncio.to_thread(_request_json, url, payload, token=self.token)
        elif self.qdrant_url:
            if vector is None:
                raise SDKError("ARCHON_QDRANT_URL requires query(vector=[...]); configure ARCHON_VECTOR_URL for text search")
            url = f"{self.qdrant_url}/collections/{quote(self.instance, safe='')}/points/search"
            response = await asyncio.to_thread(
                _request_json,
                url,
                {"vector": list(vector), "limit": top_k, "with_payload": True},
            )
        elif self.api_url and self.project:
            url = f"{self.api_url}/v1/projects/{quote(self.project, safe='')}/databases/{quote(self.instance, safe='')}/query"
            payload = {"text": text, "topK": top_k}
            if vector is not None:
                payload["vector"] = list(vector)
            response = await asyncio.to_thread(_request_json, url, payload, token=self.token)
        else:
            raise SDKError("configure ARCHON_VECTOR_URL, ARCHON_QDRANT_URL, or ARCHON_API_URL with ARCHON_PROJECT")
        values = response.get("result", response.get("items", response)) if isinstance(response, dict) else response
        if not isinstance(values, list):
            raise SDKError("vector binding returned an unexpected response")
        return [VectorHit.from_json(item) for item in values if isinstance(item, Mapping)]


class _LLMClient:
    def __init__(self, alias: str):
        if not alias.strip():
            raise ValueError("model alias is required")
        self.alias = alias
        self.base_url = _env("ARCHON_LLM_URL", _env("OPENAI_BASE_URL", "http://127.0.0.1:11434/v1"))
        self.token = os.environ.get("ARCHON_LLM_API_KEY", os.environ.get("OPENAI_API_KEY"))

    async def chat(self, messages: Sequence[Mapping[str, Any]], stream: bool = False) -> Any:
        if not messages:
            raise ValueError("messages are required")
        payload = {"model": self.alias, "messages": list(messages), "stream": stream}
        url = f"{self.base_url}/chat/completions"
        if stream:
            return self._stream(url, payload)
        return await asyncio.to_thread(_request_json, url, payload, token=self.token)

    async def _stream(self, url: str, payload: Mapping[str, Any]) -> AsyncIterator[dict[str, Any]]:
        body = json.dumps(payload).encode("utf-8")
        request = Request(url, data=body, headers=_headers(self.token), method="POST")
        try:
            response = await asyncio.to_thread(urlopen, request, timeout=60.0)
        except (HTTPError, URLError, TimeoutError) as error:
            raise SDKError(f"streaming request to {url} failed: {error}") from error
        try:
            while True:
                line = await asyncio.to_thread(response.readline)
                if not line:
                    break
                decoded = line.decode("utf-8", "replace").strip()
                if not decoded or not decoded.startswith("data:"):
                    continue
                value = decoded[5:].strip()
                if value == "[DONE]":
                    break
                try:
                    parsed = json.loads(value)
                except json.JSONDecodeError as error:
                    raise SDKError("invalid JSON in model stream") from error
                if isinstance(parsed, dict):
                    yield parsed
        finally:
            response.close()


def vector(instance: str) -> _VectorClient:
    return _VectorClient(instance)


def llm(alias: str) -> _LLMClient:
    return _LLMClient(alias)
