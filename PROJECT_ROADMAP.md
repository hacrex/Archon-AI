# Archon AI — Step-by-Step Completion Roadmap

## Vision

**Supabase for AI applications.**

Developers building AI products today stitch together a dozen moving parts: a database, a vector store, an embedding pipeline, agent code, model API keys, secrets, queues, and observability — none of which speak to each other. Archon AI collapses all of that into **one coherent platform** with the same feel Supabase gave to the Postgres ecosystem:

```bash
archon init my-ai-app     # scaffold project (schema, agents, config)
cd my-ai-app
archon dev                # full stack on localhost in <5 minutes
```

What the developer gets, locally and (later) hosted:

| Layer | Archon equivalent | Supabase analogue |
|-------|-------------------|-------------------|
| Relational data | Per-project Postgres + pgvector, migrations, SQL API, table editor | Postgres + Table Editor |
| Vector search | Qdrant-backed collections with automatic sync from Postgres | *(new — AI-specific)* |
| Agents | Python SDK + sandboxed runtime, HTTP/cron/event triggers | Edge Functions |
| Models | OpenAI-compatible gateway with per-key metering & budgets | *(new — AI-specific)* |
| Secrets & keys | Project-scoped API keys (anon / service-role), encrypted secrets vault | API Keys & Vault |
| Local workflow | `archon dev`, `db push/pull/reset`, seeds | `supabase start`, `db push` |
| Dashboard | Archon Studio: schema, vectors, agent runs, usage | Studio |
| Scale-out path | Same manifests → Docker Compose → Kubernetes/Helm | Self-host → Cloud |

**The pitch in one sentence:** *"If you can build a CRUD app on Supabase, you can build an agentic AI app on Archon — same ergonomics, plus vectors, agents, and models that stay in sync with your data."*

Success looks like: a founder prototypes a RAG chatbot on a laptop Tuesday morning, deploys the identical manifests to their own VPS or our hosted cloud by Friday, and never rewrites application code between those two steps.

---

## Guiding Principles

Five rules govern every phase below. When a design decision is contested, these settle it — in order of priority.

### 1. Local-first DX
The laptop *is* the primary environment; the cloud is just another deployment target.

- `archon init && archon dev` must bring up the entire stack (Postgres+pgvector, Qdrant, control-plane API, dashboard, agent runtime) in **under 5 minutes**, offline-capable after first image pull.
- Every feature that ships in the hosted/cloud form must first be usable locally through the exact same CLI commands and API surface. No "cloud-only" APIs until Phase 8 gates explicitly say otherwise.
- Error messages in the CLI follow a fixed format: **what failed → why → the exact next command to run**. Developers should never need to read source code to debug their setup.
- Seeding and reset are first-class: `archon db reset --seed` restores a deterministic state so demos, tests, and onboarding are reproducible.

**Litmus test:** a new contributor clones the repo and has a working demo without asking a single question in chat.

### 2. Single Binary
Everything the platform does ships as one Go binary (`archon`) plus published container images.

- `archon serve`, `archon dev`, `archon db`, `archon agent`, `archon migrate` — all subcommands of one artifact. No sidecar daemons, no separate "studio server," no manual service wiring ever demanded of the user.
- Distribution: static binaries (linux/darwin × amd64/arm64) via `curl -fsSL https://get.archon.dev | sh`, Homebrew tap, and versioned Docker images. One `Makefile` target builds all three.
- The binary embeds what it needs at runtime: default configs, the web dashboard assets (`go:embed`), and migration SQL. Version stamping via `-ldflags` so `archon version` always reports the exact build.
- Multi-service orchestration (docker-compose generation, later Kubernetes manifests) is something the binary *produces*, never something the user hand-maintains.

**Litmus test:** uninstalling Archon = deleting one binary and running `archon dev down`. Nothing else lingers.

### 3. Postgres Is Home; Vectors Are a Projection
Source of truth lives in Postgres. Qdrant is a derived, disposable index.

- Any embedding-relevant row change writes an **outbox record in the same transaction**. An async reconciler pushes changes to Qdrant — never the other way around.
- Qdrant collections can be wiped and rebuilt from Postgres at any time with zero downtime (build new collection → atomic alias swap). Losing the vector store is an availability blip, **never** data loss.
- Embedding inputs, metadata filters, and ACL columns are declared once in schema/config; the projection pipeline derives everything from that declaration. No ad-hoc `upsert(vector=...)` paths that can drift from the relational state.
- Consistency guarantees are explicit and documented: eventual (bounded lag, exposed as a metric + `archon db vectors status`), with an opt-in synchronous mode for small tables/tests.

**Litmus test:** `docker compose stop qdrant && delete all collections && restart` → reconciliation rebuilds everything and search results match Postgres exactly.

### 4. Gated Phases
Progress moves through Definition-of-Done gates, not calendar dates.

- Each phase (0–9 below) lists concrete deliverables and verification criteria. **The next phase does not start until the current gate passes** in CI, not by opinion.
- Gates are machine-checkable wherever possible: integration suites against real Postgres/Qdrant (testcontainers), e2e smoke scripts, performance budgets (e.g., `dev up` < 5 min, p95 search < 100 ms), and coverage floors for core packages.
- If a feature can't pass its gate, it stays behind a flag or is cut — it does not block the roadmap. Scope discipline beats heroics.
- Later-phase temptations (NATS, gVisor, billing, K8s CRDs) are explicitly fenced until Phase 8; nothing jumps the queue because it's fun.

**Litmus test:** at any moment, `todo.md` checkboxes + CI badges tell any stranger exactly which gate we're at — no tribal knowledge required.

### 5. Security Is a Feature
Every phase carries its share of the security baseline; none of it is deferred to a mythical "hardening sprint."

- **Phase 1:** hashed, scoped API keys (anon vs service-role), rotation, encrypted secrets at rest, redaction in logs/responses.
- **Phase 2–3:** per-project tenancy boundaries enforced at the adapter layer — a request can never touch another project's schema or collection, verified by automated IDOR tests.
- **Phase 4:** agents execute in sandboxed subprocesses with CPU/memory/network caps, least-privilege tokens scoped to a single run, and full audit trails of every tool/model call.
- **Phase 6+:** RBAC, rate limits, audit log, signed images + SBOM, SOC2-track controls before hosted GA.
- Defaults are secure: bind to localhost in dev, no wildcard CORS, no `DEBUG` in release builds, dependency vulnerability scanning in CI on every PR.

**Litmus test:** a pen-test-style checklist runs as part of CI; shipping a feature that fails it is as broken as failing a unit test.

---

## How the Principles Map to Phases

| Principle | P0 | P1 | P2 | P3 | P4 | P5 | P6 | P7 | P8 | P9 |
|-----------|----|----|----|----|----|----|----|----|----|----|
| Local-first DX | ● | ○ | ○ | ★ | ★ | ★ | ○ | ○ | ○ | ★ |
| Single binary | ★ | ● | ● | ★ | ● | ● | ○ | ○ | ● | ● |
| Postgres home / vectors projection |   | ● | ★ | ● | ● | ○ | ○ | ● | ● | ○ |
| Gated phases | ★ | ★ | ★ | ★ | ★ | ★ | ★ | ★ | ★ | ★ |
| Security is a feature | ● | ★ | ● | ● | ★ | ○ | ★ | ● | ★ | ● |

★ = defining milestone for that principle · ● = substantial work · ○ = maintained, not extended

---

## Phase 0 — Foundation Hardening *(~1 week)*

Goal: make the existing repo trustworthy to build on. Current code builds and unit-tests pass; this phase closes the small gaps.

| Step | Task | Verification |
|------|------|--------------|
| 0.1 | Pin all Go module + npm/pnpm dependencies; commit lockfiles; add Renovate/Dependabot config | `go mod verify`, fresh `pnpm install --frozen-lockfile` succeeds |
| 0.2 | Add project-level `.editorconfig`, lint configs (golangci-lint, eslint, prettier) to CI | CI runs linters, zero new violations |
| 0.3 | Document repo layout + contribution guide in `docs/` | New contributor can build both stacks from docs alone |
| 0.4 | Fix known gaps: add `qdrant-config/config.schema.json`; remove dead placeholder code paths flagged by vet/staticcheck | Build + tests green |
| 0.5 | Version-stamp binaries (`-ldflags` git SHA/tag); `archon version` prints it | Reproducible release artifacts |

**Gate 0:** CI green on main; anyone cloning gets a working dev loop with one Make target.

---

## Phase 1 — Control Plane MVP: Projects & Postgres Data Plane *(~4 weeks)*

Goal: the "database branch" equivalent — create projects, manage Postgres schema/migrations through the API. This is the supabase-like core.

### 1.1 Project lifecycle API
- `POST /v1/projects` — create project (name, region, plan).
- `GET /v1/projects` — list with cursor pagination + filtering.
- `GET /v1/projects/{id}` / `PATCH` / `DELETE` — optimistic concurrency via `If-Match` ETag.
- Soft-delete + retention window (like Supabase's pause/delete flow).

### 1.2 Postgres provisioning & connection strings
- Provision per-project schema (or per-project database in docker-compose mode).
- Issue scoped DB credentials; return Supabase-style connection string (`postgres://...`) + pooler URL.
- Store secrets encrypted at rest; redact in logs/API responses.

### 1.3 Schema & migration management
- `POST /v1/projects/{id}/migrations` — apply SQL migrations transactionally; record history table.
- Migration diff/drift detection (compare declared vs actual schema).
- Seed data endpoint; "SQL editor" API (`POST /v1/sql`) with read-only guard option.

### 1.4 Outbox pattern (the bridge to vectors)
- Generic outbox table written in the same transaction as any embedding-relevant row change.
- `OutboxPublisher` worker drains outbox → emits events (in-process now, NATS later).
- Idempotency keys on all mutating API endpoints.

### 1.5 Auth & API keys (baseline slice)
- Service-role key + anon key per project; key rotation; hashed storage.
- Middleware enforcing authn/authz on every `/v1/*` route.

**Verification:** integration test suite against real Postgres (testcontainers): full create→migrate→insert→outbox-drain cycle; concurrent PATCH returns 409 on ETag mismatch.

**Gate 1:** `archon api up` + curl walkthrough creates a project, applies a migration, inserts a row, and the outbox shows the pending event. Dashboard can list/create projects against the live API.

---

## Phase 2 — Vector Search Plane: Qdrant Adapter & Reconciliation *(~3 weeks)*

Goal: the "pgvector/semantic search" layer — Supabase Vector equivalent, but production-grade reconciliation.

### 2.1 Collections API
- `POST /v1/projects/{id}/collections` — create collection (size, distance metric, payload indexes).
- List/get/delete collections; alias support for blue/green re-indexing.

### 2.2 Qdrant adapter hardening
- Typed client wrapper: retries with exponential backoff + jitter, circuit breaker, per-call timeouts.
- Binding operations: ensure-collection, upsert points, delete points, payload-index create.
- Conformance test suite runnable against real Qdrant (docker) and a fake.

### 2.3 Embedding pipeline
- Ingest job: reads outbox events → calls pluggable embedder (OpenAI/Ollama/local interface) → upserts to Qdrant with payload `{project_id, table, pk, updated_at}`.
- Dead-letter queue + replay endpoint for failed embeddings.

### 2.4 Search API
- `POST /v1/projects/{id}/collections/{c}/search` — vector top-k, filters, hybrid (BM25 placeholder → real later).
- Optional query embedding: accept raw text, embed server-side, then search.

### 2.5 Rebuild & consistency
- `POST .../collections/{c}/rebuild` — stream full table scan → new collection → atomic alias swap (zero-downtime reindex).
- Nightly reconciliation job: compare Postgres row count/checksum vs Qdrant point count; alert/report.

**Verification:** kill Qdrant mid-stream, restart, confirm convergence within N seconds; chaos test: delete collection out-of-band, reconcile recreates it; 10k-row rebuild completes with no downtime to reads.

**Gate 2:** insert a row via API → within ~2s the semantic search endpoint finds it; wipe Qdrant volume → `rebuild` restores exact parity.

---

## Phase 3 — Local Dev Runtime: `archon dev` *(~3 weeks)*

Goal: the moment that sells the product — Supabase Studio/CLI local experience for AI apps.

### 3.1 Orchestration engine
- `archon dev` starts/steps through: Postgres → API → ingest worker → Qdrant → dashboard (docker compose or embedded process manager).
- Service registry with health checks, readiness gating, ordered startup, graceful shutdown.
- State file (`~/.archon/state.json`): ports, credentials, container IDs.

### 3.2 Companion commands
- `archon dev down` / `restart <svc>` / `status` / `logs [-f] [svc]`.
- `archon db pull` / `db push` / `db reset` (schema sync between local and remote, like `supabase db`).
- `archon init` scaffolds a starter app (Next.js + Python agent example) with pinned versions.

### 3.3 Seeding & fixtures
- Auto-seed demo tables + sample embeddings on first boot so search works instantly.
- `archon seed --dataset tiny` for offline demos.

### 3.4 Error UX
- Human-readable failures ("Qdrant image not found → run docker login?" style hints), doctor command: `archon doctor` checks Docker, ports, disk, versions.

**Verification:** on a clean machine (macOS/Linux CI runner): `curl -fsSL .../install.sh | sh && archon init myapp && cd myapp && archon dev` reaches "all services healthy" in < 5 min; e2e smoke script hits health, creates project, inserts doc, searches.

**Gate 3:** the demo video is possible: laptop, one terminal, 5 minutes, working semantic-search app.

---

## Phase 4 — Agent Runtime & Python SDK *(~4 weeks)*

Goal: the differentiator — hosted agent execution, like "Edge Functions" but for LLM agents.

### 4.1 Agent manifest & packaging
- `archon.toml` per agent: name, entrypoint, model requirements, tools, env, memory/collection bindings, limits (CPU/mem/timeout).
- Validation at `archon agent deploy`; reproducible lockfile for Python deps.

### 4.2 Python SDK v1
- Clients: `archon.db` (typed queries), `archon.vectors.search()`, `archon.secrets.get()`, `archon.llm.complete()` (provider-agnostic gateway).
- Streaming responses; automatic retry/idempotency; structured logging hooks.

### 4.3 Execution plane (local first)
- Subprocess-based runner with resource caps (rlimits), timeout kill, crash isolation, stdout/stderr capture.
- Run lifecycle: `queued → running → succeeded/failed/cancelled`; persisted run records + step events.

### 4.4 Triggers & tooling
- HTTP trigger (`POST /v1/agents/{name}/runs`), schedule trigger (cron), event trigger (outbox/webhook).
- Built-in tool registry: web fetch, sql query (scoped), vector search — declared in manifest, enforced by capability tokens.

### 4.5 Model gateway
- Unified OpenAI-compatible endpoint backed by pluggable providers (OpenAI, Anthropic, Ollama).
- Per-project rate limits, token metering, request/response audit log (with secret redaction).

**Verification:** example RAG agent deployed via `archon agent deploy`, triggered by HTTP, answers a question using vector search + LLM; kill -9 mid-run → runner marks failed, no zombie; SDK contract tests against recorded API fixtures.

**Gate 4:** cookbook app "docs-Q&A agent" runs entirely on Archon local stack with < 100 lines of user code.

---

## Phase 5 — Web Dashboard (Archon Studio) *(~3 weeks, overlaps Phase 4)*

Goal: visual control plane — Supabase Studio equivalent.

### 5.1 Core pages wired to real API
- Projects: list/create/settings/connection-string reveal.
- Table editor (browse/edit rows, SQL runner with results grid).
- Vector explorer: collections, dimension/metric config, semantic search playground with live top-k preview.
- Agents: list, deploy status, run history, log viewer (streaming).
- API Keys: create/rotate/revoke with one-time reveal.

### 5.2 Developer aids
- Docs/examples panel per page (copy-paste snippets in JS + Python).
- Dark mode, keyboard palette (cmd-K), empty states that teach.

### 5.3 Ops views
- Service health widget (from `dev status`), outbox lag, ingest DLQ browser with replay button.

**Verification:** Playwright e2e covers golden path: create project → seed → search in playground → deploy sample agent → see run logs.

**Gate 5:** a user can complete the entire Phase-3 demo without leaving the browser except for `archon dev`.

---

## Phase 6 — Security, Tenancy & Compliance Baseline *(~3 weeks)*

Goal: safe to expose beyond localhost.

### 6.1 Authorization
- Role model: `owner / developer / service / readonly`; policy engine checked at every route; project-scoped tokens cannot cross tenants (automated tenancy tests).
### 6.2 Platform hardening
- Rate limiting (per key, per IP), request size caps, CORS allow-lists, SSRF guards on outbound fetches (agent web tool).
- Secret handling: encryption at rest (KMS-ready interface), redaction middleware for logs/traces/errors.
- Audit event stream: who did what to which project, immutable append-only store.
### 6.3 Supply chain
- Signed container images + SBOM generation in release CI; reproducible builds.
- Dependency vulnerability scanning (govulncheck, pnpm audit, pip-audit) blocking CI.
### 6.4 Self-host story
- Single-binary server mode + hardened docker-compose (no privileged containers, least-privilege DB roles).

**Verification:** pen-test checklist passes (auth bypass attempts, IDOR sweep across all routes, injection probes); threat-model doc reviewed; SOC2-prep evidence folder populated.

**Gate 6:** public beta users allowed on hosted instance; self-hosters get a security README they can trust.

---

## Phase 7 — Observability & Operability *(~2 weeks)*

- Metrics (Prometheus): RED per endpoint, outbox lag, ingest throughput, agent run duration/tokens.
- Structured logs + trace correlation (OTel) across API → ingest → Qdrant → agent runner.
- Grafana dashboards shipped in `deploy/grafana/`; alert rules (DLQ growth, reconciliation drift, error budget).
- Runbooks: `docs/runbooks/*.md` for every alert (e.g., "collection drift detected").
- `archon status --remote` and admin diagnostics endpoint.

**Gate 7:** inject faults in staging; on-call can diagnose each incident from dashboards alone within 10 minutes.

---

## Phase 8 — Cloud / Kubernetes Multi-Tenancy *(~5 weeks)*

Goal: the hosted product (the "supabase.com" side).

### 8.1 Architecture
- CRDs: `Project`, `Collection`, `Agent`, `ModelGateway`. Controllers reconcile them (operator pattern already prototyped).
- Per-tenant isolation: schema-per-project (small) → database-per-project (large); namespace-per-tenant on k8s.
### 8.2 Event bus upgrade
- Replace in-process pub/sub with NATS JetStream; durable consumers for ingest; exactly-once semantics via idempotency keys.
### 8.3 Autoscaling agent execution
- Move agent runs to pooled gVisor/Kata sandboxes; scale-to-zero; cold-start budget < 1s for cached images.
### 8.4 Platform ops
- Helm charts (HA Postgres via Patroni, Qdrant cluster mode), backup/restore (PITR drills), blue/green upgrades with migration pre-checks.
- Billing hooks: usage metering from model gateway + storage; plan enforcement.

**Gate 8:** 50-tenant load test on staging at agreed SLO (p95 search < 150 ms, agent scheduling < 2 s); restore-from-backup drill passes.

---

## Phase 9 — Product Polish & Growth *(ongoing)*

- **Integrations:** Next.js/Remix starters, LangChain/LlamaIndex adapters, MCP server exposing Archon tools.
- **Realtime:** subscribe to row/vector changes (WS/SSE) — "Realtime" parity feature.
- **Edge delivery:** CDN-backed function invocations; regional inference routing.
- **Docs site:** tutorials, API reference generated from OpenAPI, comparison pages ("vs Supabase", "vs LangGraph platform").
- **Community:** public roadmap repo mirror, RFC process, template gallery of agents.
- **Enterprise:** SSO/SAML, BYO-VPC deployment, compliance attestations.

---

## Cross-Cutting Rules (apply every phase)

1. **Tests are part of the feature** — every endpoint ships with integration coverage; every PR runs the full matrix (unit, race, e2e-smoke).
2. **Feature flags for risky paths** — reconciliation, new triggers, sandbox runtimes ship dark, ramp behind flags.
3. **Versioned APIs from day one** — `/v1` frozen at Gate 1; breaking changes require deprecation window + migration guide.
4. **Weekly demo** — working software shown every Friday; blocked items escalate, not stall silently.
5. **Scope discipline** — anything marked "later" stays later until its phase gate opens.

---

## Timeline Summary

| Phase | Duration | Cumulative | Milestone |
|-------|----------|------------|-----------|
| 0 Foundation | 1 wk | W1 | Clean CI, pinned deps |
| 1 Control Plane | 4 wks | W5 | Projects + Postgres + outbox live |
| 2 Vector Plane | 3 wks | W8 | Reliable search + rebuild |
| 3 Local Dev | 3 wks | W11 | `archon dev` 5-min demo ⭐ |
| 4 Agents/SDK | 4 wks | W15 | Deployed RAG agent on laptop |
| 5 Dashboard | 3 wks | W18* | Studio golden path (*overlaps W13–18) |
| 6 Security | 3 wks | W21 | Public-beta ready |
| 7 Observability | 2 wks | W23 | Staging operable |
| 8 Kubernetes | 5 wks | W28 | Hosted multi-tenant GA 🚀 |
| 9 Growth | ongoing | — | Integrations, realtime, enterprise |

**North-star metrics:** weekly `archon dev` activations; time-to-first-search < 5 min; agent runs/day; % projects with rebuilt-and-converged vector plane.
