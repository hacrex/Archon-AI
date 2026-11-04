# Archon AI Developer Guide

This guide explains how to run Archon locally with `archon dev`, create a development identity, and build Python agents using the Archon SDK. It covers the local lifecycle, environment variables, vector bindings, OpenAI-compatible model calls, streaming, testing, and security boundaries.

## 1. What `archon dev` provides

`archon dev` starts the single-node development stack defined in `deploy/docker-compose.yml`. The stack includes:

- Archon API
- PostgreSQL
- Qdrant
- NATS

The command is intentionally local-first. It binds the API to `127.0.0.1:8080` on the host and stores only local lifecycle metadata in `.archon/state.json`. It does not claim to provide production isolation, multi-node orchestration, or a sandboxed agent runtime.

The normal local workflow is:

```text
archon dev
  ├─ docker compose up -d --build
  ├─ wait for /livez
  ├─ wait for /readyz
  └─ write .archon/state.json

archon bootstrap-user
  └─ create a local owner account after PostgreSQL is ready

build or run an agent with the SDK

archon dev down
  └─ docker compose down and remove local state
```

## 2. Prerequisites

Install the following tools:

- Go 1.22 or newer for building the CLI and API.
- Docker Engine with the Compose v2 plugin.
- Python 3.10 or newer for agents.
- Git.

From the repository root, verify the main tools:

```bash
go version
docker version
docker compose version
python3 --version
```

The local stack downloads container images the first time it starts. Allow additional time for the initial build and image pulls.

## 3. Start and stop the local stack

Start the stack from the repository root:

```bash
archon dev
```

If the CLI has not been installed yet, use one of these equivalent commands:

```bash
go run ./cmd/archon dev
go build -o ./bin/archon ./cmd/archon
./bin/archon dev
```

A successful start prints the API URL and the commands for creating the first local user. The CLI checks both endpoints:

```bash
curl -fsS http://127.0.0.1:8080/livez
curl -fsS http://127.0.0.1:8080/readyz
```

`/livez` means the API process is alive. `/readyz` also checks required dependencies, so the command does not report success while PostgreSQL or another required dependency is unavailable.

Stop services without deleting persistent development data:

```bash
archon dev down
```

Stop services and delete the local Compose volumes, including PostgreSQL and Qdrant data:

```bash
archon dev down --volumes
```

Use `--volumes` only when you intentionally want a clean local environment. It permanently removes local database and vector data.

### 3.1 CLI options

The startup command supports:

```bash
archon dev \
  --compose-file deploy/docker-compose.yml \
  --api-url http://127.0.0.1:8080 \
  --timeout 2m
```

The corresponding environment variables are:

| Variable | Default | Purpose |
|---|---|---|
| `ARCHON_COMPOSE_FILE` | `deploy/docker-compose.yml` | Compose file to run |
| `ARCHON_API_URL` | `http://127.0.0.1:8080` | API URL used for readiness checks |

`archon dev down` remembers the Compose file used by the last successful start in `.archon/state.json`. You can override it explicitly:

```bash
archon dev down --compose-file ./deploy/docker-compose.yml
```

The state directory is mode `0700` and the state file is mode `0600`. It is ignored by Git and should not be copied between machines.

### 3.2 Troubleshooting startup

If startup fails, inspect the services:

```bash
docker compose -f deploy/docker-compose.yml ps
docker compose -f deploy/docker-compose.yml logs api
docker compose -f deploy/docker-compose.yml logs postgres qdrant
```

The CLI automatically attempts `docker compose down` when the readiness timeout expires. If Docker itself fails before the readiness phase, stop any partially started services manually:

```bash
archon dev down
```

If port `8080` is already in use, either stop the conflicting process or use a separate Compose file with a different host port and pass a matching `--api-url`.

## 4. Create the first local user

After `archon dev` succeeds, create an organization owner. Prefer stdin so the password is not exposed in shell history:

```bash
printf '%s\n' 'replace-with-a-long-development-password' \
  | archon bootstrap-user \
      --email owner@example.com \
      --display-name 'Local Owner' \
      --password-stdin
```

The local password must contain at least 12 characters. The command reads `ARCHON_DB_URL` from the environment. The development Compose file maps PostgreSQL to loopback port `5432`, so run the CLI against the host-mapped service by exporting the equivalent local URL first:

```bash
export ARCHON_DB_URL='postgres://archon:archon@127.0.0.1:5432/archon?sslmode=disable'
```

For API calls from a program, log in and retain the returned bearer token only for the current development session:

```bash
curl -sS http://127.0.0.1:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"owner@example.com","password":"replace-with-a-long-development-password"}'
```

## 5. Install and import the Python SDK

The SDK is dependency-free and requires Python 3.10 or newer:

```bash
python3 -m venv .venv
. .venv/bin/activate
pip install -e sdk/python
```

For a simple agent, add the SDK source directory to `PYTHONPATH` instead:

```bash
export PYTHONPATH="$PWD/sdk/python:$PYTHONPATH"
```

The public entry points are:

```python
from archon import agent, llm, vector
```

`@agent.handler` marks the function that the future agent runner should invoke. The decorator does not itself create a process or HTTP server.

## 6. Configure SDK bindings

The SDK reads connection details from environment variables. These values should be injected by the agent runtime rather than hardcoded in source code.

### 6.1 Common variables

| Variable | Purpose |
|---|---|
| `ARCHON_API_URL` | Archon control-plane or runtime gateway base URL |
| `ARCHON_API_TOKEN` | Bearer token sent to Archon HTTP endpoints |
| `ARCHON_PROJECT` | Project name for project-scoped API vector bindings |
| `ARCHON_VECTOR_URL` | Text-capable vector binding gateway |
| `ARCHON_QDRANT_URL` | Direct Qdrant HTTP URL for vector-based searches |
| `ARCHON_LLM_URL` | OpenAI-compatible model gateway base URL |
| `ARCHON_LLM_API_KEY` | API key for the model gateway |
| `OPENAI_BASE_URL` | Fallback OpenAI-compatible base URL |
| `OPENAI_API_KEY` | Fallback model API key |

The SDK does not silently guess credentials. If no vector binding endpoint is configured, vector calls fail with a useful `SDKError`.

### 6.2 Text-capable vector gateway

Set `ARCHON_VECTOR_URL` when the runtime gateway accepts text and is responsible for embedding the query:

```bash
export ARCHON_VECTOR_URL='http://127.0.0.1:8090'
export ARCHON_API_TOKEN='development-token'
```

The SDK sends:

```http
POST /v1/vector/query
Authorization: Bearer development-token
Content-Type: application/json

{
  "instance": "docs-index",
  "text": "How do I reset a password?",
  "topK": 5
}
```

The gateway should return either a list or an object containing `items` or `result`. Each hit should contain `id`, `score`, and `payload`.

### 6.3 Direct Qdrant binding

Use `ARCHON_QDRANT_URL` when the agent already has an embedding vector and should call Qdrant directly:

```bash
export ARCHON_QDRANT_URL='http://127.0.0.1:6333'
```

The SDK calls:

```http
POST /collections/docs-index/points/search
Content-Type: application/json

{
  "vector": [0.12, 0.08, 0.44],
  "limit": 5,
  "with_payload": true
}
```

Direct Qdrant mode requires `query(vector=[...])`. Passing only text is rejected because Qdrant does not create embeddings by itself.

### 6.4 Archon project-scoped vector binding

A runtime can expose a project-scoped Archon query endpoint using:

```bash
export ARCHON_API_URL='http://127.0.0.1:8080'
export ARCHON_PROJECT='support-bot'
export ARCHON_API_TOKEN='development-token'
```

The SDK posts text, `topK`, and an optional vector to:

```text
/v1/projects/{project}/databases/{instance}/query
```

This endpoint must be provided by the runtime or gateway layer. The current core control-plane CRUD API manages DatabaseInstances but does not itself provide semantic embedding generation.

## 7. Build an agent

A minimal text-capable support agent looks like this:

```python
from archon import agent, llm, vector


@agent.handler
async def handle(req):
    question = req.json["message"]
    hits = await vector("docs-index").query(text=question, top_k=5)
    context = "\n".join(hit.payload.get("text", "") for hit in hits)

    return await llm("chat-llm").chat(
        [
            {
                "role": "system",
                "content": "Answer using only the supplied support documentation.\n\n" + context,
            },
            {"role": "user", "content": question},
        ]
    )
```

The example in `examples/support-agent/app.py` follows this pattern. A handler receives a request object with a `json` mapping supplied by the eventual runtime. The SDK clients are asynchronous, so the handler should remain `async def`.

### 7.1 Inspecting vector hits

Each result is a `VectorHit`:

```python
hits = await vector("docs-index").query(text="billing", top_k=3)
for hit in hits:
    print(hit.id, hit.score, hit.payload)
```

`payload` is a mapping and may not contain a `text` key. Production handlers should use `.get()` and handle empty results.

### 7.2 Model calls

The model client sends an OpenAI-compatible request to `/chat/completions`:

```python
response = await llm("chat-llm").chat(
    [
        {"role": "system", "content": "You are concise."},
        {"role": "user", "content": "Summarize this incident."},
    ]
)

answer = response["choices"][0]["message"]["content"]
```

The returned value is the provider JSON response. Keeping the provider envelope allows access to usage, finish reasons, tool calls, and other compatible fields.

### 7.3 Streaming model responses

Set `stream=True` to receive an asynchronous iterator over server-sent events:

```python
stream = await llm("chat-llm").chat(
    [{"role": "user", "content": "Explain vector search."}],
    stream=True,
)

async for event in stream:
    for choice in event.get("choices", []):
        delta = choice.get("delta", {})
        token = delta.get("content")
        if token:
            print(token, end="", flush=True)
```

The SDK stops when it receives `data: [DONE]`. Provider errors and malformed stream JSON are raised as `SDKError`.

## 8. Agent manifest and runtime boundary

The example manifest is at `examples/support-agent/agent.yaml`. It documents the intended runtime contract:

- runtime and entrypoint
- HTTP and event triggers
- CPU, memory, timeout, and concurrency limits
- vector, storage, model, and tool bindings
- secrets and egress policy
- isolation mode

The manifest is documentation for the planned agent runtime at this stage. The current local SDK provides HTTP bindings, but the repository does not yet provide a trusted subprocess runner, gVisor isolation, microVM isolation, or a complete deploy/invoke CLI. Do not describe this development setup as a security sandbox.

## 9. Testing agents and bindings

Run the Go CLI integration tests:

```bash
go test ./cmd/archon -count=1
```

Run the Python SDK integration tests:

```bash
python3 -m unittest discover -s sdk/python/tests -p 'test_*.py' -v
```

Or use the Make target:

```bash
make test-sdk
```

The Python tests start a local HTTP server and verify:

- text-capable vector gateway requests
- bearer authorization headers
- typed `VectorHit` responses
- direct Qdrant vector search
- rejection of text-only queries in direct Qdrant mode
- OpenAI-compatible non-streaming chat
- SSE streaming chat
- local argument validation before network calls

The Go tests use a fake `docker` executable and an `httptest` API server. They verify that `archon dev` starts Compose, waits for both readiness endpoints, writes protected state, and that `archon dev down` stops Compose and removes the state. They also verify readiness-timeout cleanup.

## 10. Security and operational best practices

- Use development credentials only with the local Compose stack.
- Never commit `.archon/state.json`, API tokens, model keys, passwords, or signed URLs.
- Prefer environment injection or a secret manager over source-code constants.
- Keep the API bound to loopback for local work.
- Do not expose Qdrant or PostgreSQL publicly.
- Use least-privilege project and vector bindings.
- Validate and limit user input before sending it to a model.
- Treat retrieved vector payloads as untrusted content; defend against prompt injection.
- Set model and HTTP timeouts in the eventual runtime.
- Log request IDs and operational errors, but never log bearer tokens, secrets, full prompts, or sensitive retrieved payloads.
- Delete local volumes only when data loss is acceptable.
- Do not claim production isolation until a hardened runtime boundary has been implemented and tested.

## 11. Useful commands

```bash
# Start the stack.
archon dev

# Check health.
curl -fsS http://127.0.0.1:8080/livez
curl -fsS http://127.0.0.1:8080/readyz

# Inspect services.
docker compose -f deploy/docker-compose.yml ps
docker compose -f deploy/docker-compose.yml logs -f api

# Run all Go checks.
go test ./... -count=1
go vet ./...
go build ./...

# Run SDK checks.
make test-sdk
python3 -m compileall -q sdk/python examples/support-agent

# Stop services while preserving data.
archon dev down

# Stop services and delete local data.
archon dev down --volumes
```
