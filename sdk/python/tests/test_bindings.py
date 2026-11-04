import asyncio
import json
import os
import sys
import threading
import unittest
from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib.parse import urlparse

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from archon import SDKError, llm, vector


class BindingHandler(BaseHTTPRequestHandler):
    requests = []

    def do_POST(self):  # noqa: N802 - stdlib handler API
        size = int(self.headers.get("Content-Length", "0"))
        payload = json.loads(self.rfile.read(size) or "{}")
        self.__class__.requests.append((self.path, payload, self.headers.get("Authorization")))
        if self.path == "/v1/vector/query":
            self.respond_json({"items": [{"id": "doc-1", "score": 0.91, "payload": {"text": payload.get("text", "")}}]})
        elif self.path == "/collections/docs-index/points/search":
            self.respond_json({"result": [{"id": "doc-2", "score": 0.82, "payload": {"text": "direct"}}]})
        elif self.path == "/chat/completions":
            if payload.get("stream"):
                body = b'data: {"choices":[{"delta":{"content":"hello"}}]}\n\ndata: [DONE]\n\n'
                self.send_response(200)
                self.send_header("Content-Type", "text/event-stream")
                self.send_header("Content-Length", str(len(body)))
                self.end_headers()
                self.wfile.write(body)
            else:
                self.respond_json({"choices": [{"message": {"role": "assistant", "content": "hello"}}]})
        else:
            self.send_error(404)

    def respond_json(self, value):
        body = json.dumps(value).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *_):
        pass


class SDKBindingIntegrationTests(unittest.IsolatedAsyncioTestCase):
    @classmethod
    def setUpClass(cls):
        BindingHandler.requests = []
        cls.server = HTTPServer(("127.0.0.1", 0), BindingHandler)
        cls.thread = threading.Thread(target=cls.server.serve_forever, daemon=True)
        cls.thread.start()
        cls.base_url = f"http://127.0.0.1:{cls.server.server_port}"

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()
        cls.thread.join(timeout=2)

    def setUp(self):
        self.previous = os.environ.copy()
        for name in ("ARCHON_VECTOR_URL", "ARCHON_QDRANT_URL", "ARCHON_API_URL", "ARCHON_PROJECT", "ARCHON_API_TOKEN", "ARCHON_LLM_URL", "ARCHON_LLM_API_KEY"):
            os.environ.pop(name, None)

    def tearDown(self):
        os.environ.clear()
        os.environ.update(self.previous)

    async def test_text_vector_gateway_returns_typed_hits_and_auth_header(self):
        os.environ["ARCHON_VECTOR_URL"] = self.base_url
        os.environ["ARCHON_API_TOKEN"] = "agent-token"
        hits = await vector("docs-index").query("reset password", top_k=3)
        self.assertEqual(hits[0].id, "doc-1")
        self.assertEqual(hits[0].payload["text"], "reset password")
        path, payload, authorization = BindingHandler.requests[-1]
        self.assertEqual(path, "/v1/vector/query")
        self.assertEqual(payload["topK"], 3)
        self.assertEqual(authorization, "Bearer agent-token")

    async def test_direct_qdrant_search_requires_vector_and_returns_results(self):
        os.environ["ARCHON_QDRANT_URL"] = self.base_url
        with self.assertRaises(SDKError):
            await vector("docs-index").query("text")
        hits = await vector("docs-index").query(vector=[0.1, 0.2], top_k=2)
        self.assertEqual(hits[0].id, "doc-2")
        self.assertEqual(BindingHandler.requests[-1][0], "/collections/docs-index/points/search")
        self.assertEqual(BindingHandler.requests[-1][1]["limit"], 2)

    async def test_openai_compatible_chat_and_sse_stream(self):
        os.environ["ARCHON_LLM_URL"] = self.base_url
        os.environ["ARCHON_LLM_API_KEY"] = "model-token"
        result = await llm("chat-llm").chat([{"role": "user", "content": "hello"}])
        self.assertEqual(result["choices"][0]["message"]["content"], "hello")
        stream = await llm("chat-llm").chat([{"role": "user", "content": "hello"}], stream=True)
        events = [event async for event in stream]
        self.assertEqual(events[0]["choices"][0]["delta"]["content"], "hello")
        self.assertEqual(BindingHandler.requests[-1][2], "Bearer model-token")

    async def test_invalid_top_k_is_rejected_before_http(self):
        os.environ["ARCHON_VECTOR_URL"] = self.base_url
        with self.assertRaises(ValueError):
            await vector("docs-index").query("hello", top_k=0)


if __name__ == "__main__":
    unittest.main()
