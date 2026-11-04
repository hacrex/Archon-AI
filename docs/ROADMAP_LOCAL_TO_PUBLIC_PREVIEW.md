# Archon AI Roadmap: Local Developer Preview to Public Developer Preview

## Execution status

The first release-governance slice is implemented: PP-001 has a preview contract, PP-002 has a dedicated threat model, PP-003 has release gates, PP-004 has a capability registry, and PP-010 has the `archon doctor` prerequisite command. PP-011 through PP-016 remain the next installation and first-run workstream.

## Purpose

This roadmap moves Archon AI from a **Local Developer Preview** to a **Public Developer Preview** that external developers can install, evaluate, and use for non-sensitive development workloads without being misled about production guarantees.

The target release is not a fully managed production platform. It is a bounded, self-hostable developer preview with clear capability labels, reproducible installation, authenticated APIs, basic agent execution, reliable local reconciliation, actionable diagnostics, and enough security hardening to avoid exposing obvious credentials or cross-project data leaks.

## Target release definition

At Public Developer Preview, a new developer should be able to:

1. Install or start Archon using documented instructions.
2. Run prerequisite checks and receive actionable failures.
3. Create an account and organization owner.
4. Create a project.
5. Create a supported Qdrant DatabaseInstance.
6. Verify resource status and conditions.
7. Deploy a clearly labeled trusted local Python agent.
8. Invoke the agent over HTTP.
9. Use a vector binding and an OpenAI-compatible model binding.
10. Stream a model response.
11. Inspect invocation logs and errors.
12. Restart a dependency and observe bounded recovery or an explicit degraded state.
13. Delete the agent, database, project, and local data safely.
14. Report a reproducible issue with request ID, version, diagnostics, and environment information.

Every feature in the preview must be labeled **Available**, **Preview**, **Planned**, **Unsupported**, or **Degraded**. The product must not imply HA, production sandboxing, managed hosting, compliance, or protection of sensitive workloads.

## Release boundary

### Included in Public Developer Preview

- Single-node local or enthusiast deployment.
- PostgreSQL control plane.
- Qdrant as the only supported vector engine.
- Authenticated API and local web console.
- One organization per local installation, with documented team-access limitations.
- Trusted local Python subprocess execution with explicit warnings.
- HTTP vector gateway and OpenAI-compatible model gateway bindings.
- Project-scoped authorization.
- Basic audit events for authentication and resource mutations.
- Bounded resource reconciliation and retry behavior.
- Diagnostics, health endpoints, structured logs, and basic metrics.
- Versioned release artifacts and rollback instructions.

### Explicitly excluded from this release

- Multi-tenant hosted production service.
- HA control plane or multi-node data plane.
- Strong sandbox isolation using gVisor or microVMs.
- Compliance claims, sensitive-data guarantees, or regulated workloads.
- Billing, usage metering, quotas beyond local validation, and paid plans.
- SSO/SCIM, passkeys, MFA, and enterprise identity lifecycle.
- Multiple vector engines beyond Qdrant.
- Durable production model gateway and provider failover.
- Kubernetes operator/controller production claims.
- Automatic public cluster provisioning.

## Current baseline

The repository already has:

- Go API and CLI foundation.
- Project and DatabaseInstance resource models.
- PostgreSQL migrations and repositories.
- Authentication and bearer sessions.
- Organization membership lookup.
- Project authorization foundation.
- Qdrant collection provisioning.
- `archon dev` and `archon dev down`.
- Python SDK HTTP, vector, and OpenAI-compatible bindings.
- Local developer guide.
- Go, Python, frontend, deployment, and integration checks.

The largest gaps are complete agent execution, reliable reconciliation, production-preview security controls, live console workflows, operational diagnostics, and release quality.

## Priority model

- **P0**: Release blocker. Public preview cannot ship without it.
- **P1**: Required for a credible preview; may be feature-gated only with explicit approval.
- **P2**: Important follow-up after public preview; do not expand scope until P0/P1 is stable.

Each task has an owner role, dependency, and acceptance criteria. Replace role names with actual people when the work is assigned.

---

# Phase 0 — Scope lock and release governance

**Estimated duration:** 2–3 days  
**Priority:** P0  
**Exit:** Product scope and unsupported claims are frozen.

## Tasks

### PP-001 — Freeze the Public Developer Preview contract

- **Owner:** Product + engineering lead
- **Priority:** P0
- **Dependencies:** None
- **Work:** Convert the release boundary above into a one-page product contract. Define supported operating systems, Docker/Compose requirements, Python versions, supported model gateway protocol, and Qdrant version.
- **Acceptance criteria:** A reviewer can determine whether a requested capability is Available, Preview, Planned, Unsupported, or Degraded without reading implementation details.

### PP-002 — Define the threat model for the preview

- **Owner:** Security engineering
- **Priority:** P0
- **Dependencies:** PP-001
- **Work:** Document trusted local code execution, token handling, network exposure, prompt injection, vector payload trust, dependency compromise, and cross-project authorization threats.
- **Acceptance criteria:** Every P0 security control maps to a documented threat; every non-goal is visible in the public documentation.

### PP-003 — Establish release quality gates

- **Owner:** Engineering lead
- **Priority:** P0
- **Dependencies:** PP-001
- **Work:** Define required CI checks, end-to-end test environments, supported version matrix, severity definitions, release approval, rollback ownership, and incident response.
- **Acceptance criteria:** A tagged release cannot be published unless every P0 gate is green or explicitly waived by named approvers.

### PP-004 — Add a capability-status registry

- **Owner:** API + frontend
- **Priority:** P1
- **Dependencies:** PP-001
- **Work:** Define a shared capability catalog consumed by documentation and UI. Include `available`, `preview`, `planned`, `unsupported`, and `degraded` states.
- **Acceptance criteria:** The console and documentation use the same names and descriptions for capability status.

---

# Phase 1 — Reproducible installation and first-run experience

**Estimated duration:** 1–2 weeks  
**Priority:** P0  
**Exit:** A clean developer machine can install and reach a verified local control plane using documented commands.

## Tasks

### PP-010 — Add `archon doctor`

- **Owner:** CLI
- **Priority:** P0
- **Dependencies:** PP-001
- **Work:** Check Go/CLI version, Docker availability, Compose plugin, available ports, disk space, memory, CPU, Python version, and writable state directories.
- **Acceptance criteria:** Failures include the exact check, detected value, required value, and remediation command. `archon doctor --output json` is machine-readable.

### PP-011 — Harden `archon dev`

- **Owner:** CLI
- **Priority:** P0
- **Dependencies:** PP-010
- **Work:** Add preflight checks, stale-state recovery, Compose project naming, bounded startup cleanup, explicit port conflict messages, and `archon dev status`.
- **Acceptance criteria:** Re-running `archon dev` is safe; interrupted startup does not leave orphaned containers or misleading state.

### PP-012 — Make Compose configuration deterministic

- **Owner:** DevOps
- **Priority:** P0
- **Dependencies:** PP-001
- **Work:** Pin images by reviewed version or digest, add all required health checks, use loopback host bindings, set resource defaults, and document volume lifecycle.
- **Acceptance criteria:** `docker compose config` passes; a clean machine can start the stack without editing YAML; no database, Qdrant, or NATS port is publicly exposed by default.

### PP-013 — Add first-run setup command

- **Owner:** CLI + identity
- **Priority:** P0
- **Dependencies:** PP-011, PP-012
- **Work:** Add `archon init` or extend `archon dev` with an idempotent setup flow for migrations, first owner, organization, and local context.
- **Acceptance criteria:** A new developer does not need to manually discover database URLs or migration commands. Re-running setup does not overwrite an existing account.

### PP-014 — Add local context management

- **Owner:** CLI
- **Priority:** P1
- **Dependencies:** PP-013
- **Work:** Store endpoint, organization, project, and token metadata in a protected local context. Support `archon context list`, `use`, and `remove` without printing secrets.
- **Acceptance criteria:** CLI commands use the selected context consistently; context files are mode `0600` and excluded from Git.

### PP-015 — Publish installation and recovery documentation

- **Owner:** Developer experience
- **Priority:** P0
- **Dependencies:** PP-010–PP-013
- **Work:** Document Linux/macOS/Windows assumptions, Docker setup, port conflicts, reset procedures, logs, volume deletion, and uninstall.
- **Acceptance criteria:** A developer unfamiliar with the repository completes first-run setup by following documentation only.

### PP-016 — Add clean-machine installation test

- **Owner:** QA + DevOps
- **Priority:** P0
- **Dependencies:** PP-012, PP-015
- **Work:** Run installation in a fresh CI VM or disposable environment. Verify startup, readiness, bootstrap, API access, and teardown.
- **Acceptance criteria:** Test is repeatable and records version, image, OS, and failure diagnostics.

---

# Phase 2 — API correctness, authorization, and lifecycle safety

**Estimated duration:** 2 weeks  
**Priority:** P0  
**Exit:** External developers can safely use the core API without silent overwrites, cross-project access, or ambiguous mutation results.

## Tasks

### PP-020 — Add idempotency keys

- **Owner:** API + PostgreSQL
- **Priority:** P0
- **Dependencies:** PP-003
- **Work:** Support `Idempotency-Key` on project and DatabaseInstance mutations. Persist request hash, result, status, and expiration.
- **Acceptance criteria:** Retrying the same request returns the original result without duplicate resources. Reusing a key with a different body returns a typed conflict.

### PP-021 — Add optimistic concurrency

- **Owner:** API + PostgreSQL
- **Priority:** P0
- **Dependencies:** PP-020
- **Work:** Expose resource version or ETag and require `If-Match` for updates/deletes where appropriate.
- **Acceptance criteria:** Concurrent updates cannot silently overwrite each other. Conflict responses contain stable error code and current version information.

### PP-022 — Complete project authorization

- **Owner:** Identity + API
- **Priority:** P0
- **Dependencies:** PP-001
- **Work:** Enforce project membership on every project-scoped read and mutation. Verify organization isolation and role permissions.
- **Acceptance criteria:** Cross-project and cross-organization access tests return authorization errors for every endpoint, including future routes.

### PP-023 — Add API key/service-account lifecycle for integrations

- **Owner:** Identity + security
- **Priority:** P1
- **Dependencies:** PP-022
- **Work:** Create, list metadata, rotate, revoke, and expire project-scoped keys. Store only hashes and show secret material once.
- **Acceptance criteria:** Revoked and expired keys fail immediately; keys cannot access projects outside their scope; secrets never appear in logs or list responses.

### PP-024 — Add durable audit events

- **Owner:** API + PostgreSQL
- **Priority:** P0
- **Dependencies:** PP-022
- **Work:** Record login/logout, failed authentication, resource mutations, authorization failures, member changes, key lifecycle, and privileged actions.
- **Acceptance criteria:** Events include actor, organization, project, resource, action, result, request ID, timestamp, and redacted before/after metadata.

### PP-025 — Add audit query and export API

- **Owner:** API + frontend
- **Priority:** P1
- **Dependencies:** PP-024
- **Work:** Add filtering by actor, project, resource, action, result, and time range. Add documented JSONL or CSV export.
- **Acceptance criteria:** Organization administrators can investigate a resource mutation without direct database access; secrets and prompts are redacted.

### PP-026 — Add request and input limits

- **Owner:** API
- **Priority:** P0
- **Dependencies:** PP-003
- **Work:** Set body-size limits, URL/header limits, pagination bounds, timeout budgets, and validation for large labels/annotations.
- **Acceptance criteria:** Oversized requests fail predictably with typed errors and do not consume unbounded memory.

### PP-027 — Add login rate limiting and abuse controls

- **Owner:** Security + API
- **Priority:** P0
- **Dependencies:** PP-022
- **Work:** Limit failed login attempts and sensitive mutation bursts. Define local-preview versus production-like defaults.
- **Acceptance criteria:** Brute-force attempts are slowed or blocked; legitimate local development remains usable; events are auditable.

### PP-028 — Make migrations concurrency-safe

- **Owner:** PostgreSQL
- **Priority:** P0
- **Dependencies:** PP-003
- **Work:** Add PostgreSQL advisory locking or equivalent transaction coordination to migration startup.
- **Acceptance criteria:** Two API instances starting simultaneously cannot apply the same migration twice or leave a partially applied schema.

---

# Phase 3 — Resource reconciliation and Qdrant reliability

**Estimated duration:** 2–3 weeks  
**Priority:** P0  
**Exit:** Supported Qdrant resources converge to desired state across restarts, transient failures, and provider drift.

## Tasks

### PP-030 — Restrict unsupported engines explicitly

- **Owner:** API + product
- **Priority:** P0
- **Dependencies:** PP-001
- **Work:** Accept only Qdrant in the public preview or return a typed `unsupported_engine` response for Chroma, Weaviate, Milvus, pgvector, and ScyllaDB.
- **Acceptance criteria:** UI, API validation, documentation, and CLI all agree that Qdrant is the only Available engine.

### PP-031 — Define the reconciliation state machine

- **Owner:** Platform runtime
- **Priority:** P0
- **Dependencies:** PP-030
- **Work:** Define Pending, Reconciling, Ready, Degraded, Failed, Deleting, and Deleted behavior, including observed generation and conditions.
- **Acceptance criteria:** Every transition has a reason, timestamp, request ID, retry state, and user-visible message.

### PP-032 — Add a durable reconciler loop

- **Owner:** Platform runtime
- **Priority:** P0
- **Dependencies:** PP-031
- **Work:** Watch or poll desired resources, reconcile Qdrant collections, update observed status, and resume after API restart.
- **Acceptance criteria:** A resource created while Qdrant is unavailable eventually becomes Ready after Qdrant recovers, within bounded retry policy.

### PP-033 — Add retry and backoff policy

- **Owner:** Platform runtime
- **Priority:** P0
- **Dependencies:** PP-032
- **Work:** Use capped exponential backoff with jitter, classify retryable versus terminal errors, and expose next retry time.
- **Acceptance criteria:** Transient provider errors do not fail permanently; invalid configuration does not cause an infinite hot loop.

### PP-034 — Add drift detection and repair

- **Owner:** Platform runtime
- **Priority:** P0
- **Dependencies:** PP-032
- **Work:** Detect missing collections, changed dimensions/configuration, and out-of-band deletion. Define safe repair versus degraded state.
- **Acceptance criteria:** Removing a managed collection produces a clear Degraded condition and either repairs it safely or provides an explicit recovery action.

### PP-035 — Complete outbox publishing

- **Owner:** Platform runtime
- **Priority:** P1
- **Dependencies:** PP-024, PP-032
- **Work:** Add publisher interface, retry state, dead-letter behavior, metrics, and replay tooling. Integrate NATS only when the contract is stable.
- **Acceptance criteria:** Every resource mutation has an observable delivery result; failed events can be retried without duplicating side effects.

### PP-036 — Add adapter conformance tests

- **Owner:** Platform runtime + QA
- **Priority:** P0
- **Dependencies:** PP-032–PP-034
- **Work:** Test fake adapter and real Qdrant behavior for create, update, health, drift, retry, delete, and restart.
- **Acceptance criteria:** Tests cover provider timeout, 404, 409, malformed response, and partial failure behavior.

### PP-037 — Add backup and reset procedures

- **Owner:** DevOps
- **Priority:** P1
- **Dependencies:** PP-032
- **Work:** Document PostgreSQL and Qdrant backup/restore for the preview. Add a safe reset command that warns before deleting data.
- **Acceptance criteria:** A developer can restore a test backup and verify project/resource status without manual SQL editing.

---

# Phase 4 — Agent runtime and SDK integration

**Estimated duration:** 2–3 weeks  
**Priority:** P0  
**Exit:** A developer can deploy and invoke a clearly labeled trusted local Python agent end to end.

## Tasks

### PP-040 — Finalize the agent manifest schema

- **Owner:** Runtime + SDK
- **Priority:** P0
- **Dependencies:** PP-001
- **Work:** Define schema and validation for runtime, entrypoint, source, triggers, resources, timeout, concurrency, bindings, secrets, egress, and isolation.
- **Acceptance criteria:** Invalid manifests fail with field paths and stable error codes; unsupported isolation values cannot be accepted.

### PP-041 — Implement trusted subprocess runner

- **Owner:** Runtime
- **Priority:** P0
- **Dependencies:** PP-040
- **Work:** Create a temporary working directory, launch the configured entrypoint, pass an allowlisted environment, propagate request ID, enforce timeout, capture stdout/stderr, and clean up.
- **Acceptance criteria:** A working example agent can be invoked; timeout, crash, malformed output, and missing entrypoint produce bounded typed errors.

### PP-042 — Implement invocation API

- **Owner:** Runtime + API
- **Priority:** P0
- **Dependencies:** PP-041
- **Work:** Add deploy/status/invoke/log endpoints or equivalent CLI-backed service contract.
- **Acceptance criteria:** Invocation returns request ID, status, duration, structured result/error, and a link or identifier for logs.

### PP-043 — Implement `archon agent` commands

- **Owner:** CLI
- **Priority:** P0
- **Dependencies:** PP-042
- **Work:** Add `agent validate`, `agent deploy`, `agent status`, `agent invoke`, and `agent logs`.
- **Acceptance criteria:** Commands use local context, return non-zero on failure, support `--output json`, and do not print secrets.

### PP-044 — Complete Python SDK contract

- **Owner:** SDK
- **Priority:** P0
- **Dependencies:** PP-041, PP-042
- **Work:** Document request object, response envelope, vector hit schema, error model, timeouts, streaming, and supported environment variables.
- **Acceptance criteria:** Example agent runs against a local test gateway and the SDK integration suite covers success and failure paths.

### PP-045 — Add agent secret handling

- **Owner:** Runtime + security
- **Priority:** P0
- **Dependencies:** PP-041
- **Work:** Support explicit secret allowlists, do not inject unrelated environment variables, redact secret names/values in logs, and never persist secret values in manifests.
- **Acceptance criteria:** A test proves an undeclared secret is unavailable and declared secrets never appear in invocation logs.

### PP-046 — Add runtime resource and egress controls

- **Owner:** Runtime + security
- **Priority:** P1
- **Dependencies:** PP-041
- **Work:** Enforce timeout, concurrency, CPU/memory hints where supported, and document that trusted subprocess mode is not a sandbox.
- **Acceptance criteria:** The product clearly displays the isolation limitation; resource exhaustion does not take down the API process.

### PP-047 — Add invocation failure matrix

- **Owner:** QA
- **Priority:** P0
- **Dependencies:** PP-041–PP-046
- **Work:** Test startup failure, crash, timeout, invalid JSON, empty response, vector gateway error, model gateway error, cancellation, and dependency restart.
- **Acceptance criteria:** Each failure has stable status, useful diagnostics, no leaked secrets, and a deterministic exit/error code.

---

# Phase 5 — Web console and developer experience

**Estimated duration:** 2 weeks  
**Priority:** P1  
**Exit:** The web console does not present static preview data as live product state and supports the primary preview workflow.

## Tasks

### PP-050 — Remove misleading hardcoded operational data

- **Owner:** Frontend
- **Priority:** P0
- **Dependencies:** PP-004
- **Work:** Replace fake capacity, health, activity, operation, and resource values with API data or explicit unavailable/preview states.
- **Acceptance criteria:** A user can distinguish live data, preview data, planned features, and unavailable dependencies on every route.

### PP-051 — Add project context switching

- **Owner:** Frontend + API
- **Priority:** P0
- **Dependencies:** PP-022
- **Work:** Add project selector and load resources for the selected project rather than only the first project.
- **Acceptance criteria:** Switching projects changes every project-scoped route and cannot display another project’s resources.

### PP-052 — Add DatabaseInstance lifecycle UI

- **Owner:** Frontend + API
- **Priority:** P0
- **Dependencies:** PP-020–PP-034
- **Work:** Add create, detail, conditions, update, delete, and degraded-state views for Qdrant.
- **Acceptance criteria:** UI mutations show request IDs, status transitions, validation errors, conflicts, and deletion consequences.

### PP-053 — Add agent deployment and invocation UI

- **Owner:** Frontend + runtime
- **Priority:** P1
- **Dependencies:** PP-042–PP-043
- **Work:** Add manifest validation, deploy status, invoke form, streaming output, and log viewer.
- **Acceptance criteria:** The UI can complete the same primary workflow as the CLI without fake success states.

### PP-054 — Add admin membership and audit UI

- **Owner:** Frontend + API
- **Priority:** P1
- **Dependencies:** PP-024–PP-025
- **Work:** Add member listing and supported role changes, audit filters, request ID links, and redacted event details.
- **Acceptance criteria:** Users see only actions allowed by role; audit values never reveal credentials, tokens, or full sensitive prompts.

### PP-055 — Add accessible error and empty states

- **Owner:** Frontend
- **Priority:** P1
- **Dependencies:** PP-050
- **Work:** Add semantic status text, recoverable errors, retry actions, loading states, mobile layout, keyboard focus, and reduced-motion behavior.
- **Acceptance criteria:** Core workflows are usable without color-only status interpretation or browser console inspection.

### PP-056 — Add browser smoke tests

- **Owner:** QA + frontend
- **Priority:** P1
- **Dependencies:** PP-050–PP-055
- **Work:** Test login, project selection, project creation, DatabaseInstance creation/status, logout, authorization failure, and session expiry.
- **Acceptance criteria:** Tests run against a disposable local API and fail on placeholder success messages.

---

# Phase 6 — Observability, supportability, and release engineering

**Estimated duration:** 1–2 weeks  
**Priority:** P0  
**Exit:** Preview users can understand failures and maintain a local installation without engineering assistance.

## Tasks

### PP-060 — Define operational metrics

- **Owner:** Observability
- **Priority:** P0
- **Dependencies:** PP-031, PP-042
- **Work:** Add API request count/latency/errors, readiness, reconciliation state, retry count, outbox backlog, Qdrant health, invocation duration, timeout, crash, and dependency errors.
- **Acceptance criteria:** Metrics have stable names, labels have bounded cardinality, and dashboards do not require parsing logs.

### PP-061 — Improve structured logging and redaction

- **Owner:** API + runtime
- **Priority:** P0
- **Dependencies:** PP-024, PP-041
- **Work:** Standardize JSON fields: timestamp, level, request ID, actor, organization, project, resource, operation, and error code. Redact tokens, passwords, secrets, prompts, and signed URLs.
- **Acceptance criteria:** Automated tests prove sensitive fixtures do not appear in logs.

### PP-062 — Add diagnostics bundle

- **Owner:** CLI + DevOps
- **Priority:** P1
- **Dependencies:** PP-060–PP-061
- **Work:** Add `archon diagnostics` to collect versions, health summaries, Compose status, sanitized logs, migration version, and configuration names without values.
- **Acceptance criteria:** Bundle is safe to attach to a public issue and has a documented schema.

### PP-063 — Add vulnerability, SBOM, and dependency checks

- **Owner:** Supply chain security
- **Priority:** P0
- **Dependencies:** PP-003
- **Work:** Pin dependencies, run Go/Python/container vulnerability scanning, generate SBOM, and define severity thresholds.
- **Acceptance criteria:** Release CI fails on unapproved critical/high findings and stores SBOM artifacts.

### PP-064 — Add image provenance and release artifacts

- **Owner:** Release engineering
- **Priority:** P1
- **Dependencies:** PP-063
- **Work:** Build versioned API/monitor images, publish checksums, sign or attest artifacts where available, and document verification.
- **Acceptance criteria:** A user can identify the exact source commit and dependency set for every preview release.

### PP-065 — Add backup/restore and rollback runbooks

- **Owner:** DevOps
- **Priority:** P0
- **Dependencies:** PP-037, PP-063
- **Work:** Document database backup, Qdrant snapshot, migration rollback limitations, Compose image rollback, and data-loss warnings.
- **Acceptance criteria:** Runbooks are tested by someone other than the author in a disposable environment.

### PP-066 — Add issue templates and support policy

- **Owner:** Developer experience
- **Priority:** P1
- **Dependencies:** PP-062
- **Work:** Add bug report, security report, feature request, and diagnostics instructions. Define supported preview versions and response expectations.
- **Acceptance criteria:** A new issue can include version, OS, command, logs, request ID, and diagnostics without exposing secrets.

---

# Phase 7 — End-to-end release certification

**Estimated duration:** 1 week  
**Priority:** P0  
**Exit:** Public Developer Preview is approved and reproducible.

## Certification scenarios

### E2E-001 — Clean install

- Start from a clean VM or disposable runner.
- Run `archon doctor`.
- Run `archon dev`.
- Verify `/livez` and `/readyz`.
- Run first-user setup.
- Confirm protected local state.

### E2E-002 — Auth and authorization

- Create two projects.
- Use a principal with access to only one project.
- Verify reads and mutations for the other project are denied.
- Verify audit events contain actor, scope, action, and request ID.

### E2E-003 — Qdrant lifecycle

- Create a Qdrant DatabaseInstance.
- Verify Pending → Reconciling → Ready.
- Restart Qdrant.
- Delete the backing collection out of band.
- Verify drift detection and repair/degraded behavior.
- Delete the DatabaseInstance and verify cleanup.

### E2E-004 — Agent invocation

- Validate and deploy the support agent.
- Invoke it with a request ID.
- Verify vector binding request and model request.
- Verify non-streaming and streaming responses.
- Inspect redacted logs.

### E2E-005 — Failure behavior

- Stop PostgreSQL.
- Stop Qdrant.
- Make the model gateway return 500.
- Force an agent timeout and crash.
- Verify bounded errors, recovery, retry behavior, and no secret leakage.

### E2E-006 — Reset and teardown

- Delete agent, DatabaseInstance, and project.
- Run `archon dev down`.
- Verify containers stop.
- Run `archon dev down --volumes`.
- Verify documented data deletion behavior.

## Release gates

Public Developer Preview cannot ship unless all of the following are true:

- [ ] P0 tasks PP-001 through PP-003 are complete.
- [ ] Clean install and teardown pass on every supported platform profile.
- [ ] No known critical vulnerability or cross-project authorization issue.
- [ ] Agent runtime has explicit trusted-code warning and failure handling.
- [ ] Qdrant is the only engine presented as Available.
- [ ] No UI route presents fabricated operational data as live.
- [ ] Request IDs appear in API errors and diagnostics.
- [ ] Secrets, tokens, passwords, prompts, and signed URLs are redacted from logs.
- [ ] Backup/reset and rollback procedures have been tested.
- [ ] CI runs Go tests, vet, build, Python integration tests, frontend checks, Compose validation, Helm checks, and security scans.
- [ ] Documentation states the non-goals and unsupported production guarantees.
- [ ] Security review and product approval are recorded for the release candidate.

## Suggested milestone schedule

| Milestone | Scope | Main exit signal |
|---|---|---|
| M0 | Scope and threat model | Release boundary approved |
| M1 | Reproducible local installation | Clean-machine install passes |
| M2 | Safe authenticated API | Concurrency, authorization, audit, and limits pass |
| M3 | Reliable Qdrant lifecycle | Restart, retry, and drift tests pass |
| M4 | Working local agent runtime | Deploy/invoke/stream/log workflow passes |
| M5 | Honest live console | No fabricated operational state remains |
| M6 | Supportable release | Diagnostics, SBOM, scans, and runbooks pass |
| M7 | Public Developer Preview | All release gates approved |

A realistic schedule is approximately **8–12 engineering weeks for a small focused team**, depending on how much of the agent runtime and frontend is built in parallel. A single developer should expect a longer schedule and should release M1–M3 as internal milestones rather than attempting to deliver the entire public preview at once.

## Post-preview backlog

After Public Developer Preview stabilizes, prioritize:

1. gVisor or microVM isolation before stronger agent security claims.
2. OIDC/OAuth, MFA, passkeys, and organization invitations.
3. Production controller/operator and Kubernetes reconciliation.
4. NATS JetStream outbox delivery.
5. Model gateway provider adapters, routing, fallback, and usage accounting.
6. Storage, functions, tools, knowledge pipelines, and governance modules.
7. Multi-organization hosted architecture.
8. Backup automation, HA, upgrades, and disaster recovery.
9. Additional vector engines through conformance-tested adapters.
10. Billing, quotas, SSO/SCIM, BYOK, compliance exports, and air-gapped bundles.

## Working rule

Do not expand the feature surface while any P0 item is incomplete. A smaller preview with reliable lifecycle behavior, honest status labeling, strong diagnostics, and explicit security boundaries is more valuable than a broad console that implies capabilities the backend cannot yet guarantee.
