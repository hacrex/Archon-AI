# Public Developer Preview Threat Model

This threat model defines the security boundary for the Archon AI Public Developer Preview. It applies to the local Compose deployment and trusted local Python agent workflow.

## Assumptions

- The developer controls the host running Docker and Archon.
- The local host may already contain unrelated processes and credentials.
- Agent source is developer-controlled and trusted for this preview.
- PostgreSQL, Qdrant, and NATS are private local dependencies, not public services.
- Preview users may expose services accidentally, copy logs, or use real credentials unless the product warns them clearly.
- The preview is not a strong sandbox and must not execute untrusted agent source.

## Assets and threats

| Asset | Threat | Impact | Required mitigation | Verification |
|---|---|---|---|---|
| Session tokens | Token appears in logs, diagnostics, shell history, or browser storage | Account takeover | Redact tokens; prefer stdin/env injection; provide logout/revocation | Log fixture and diagnostics tests |
| Passwords and API keys | Credential stored in source, manifest, state, or logs | Credential compromise | Hash passwords; persist only hashes; inject secrets explicitly; show keys once | Secret redaction and storage tests |
| Project data | Cross-project API or SDK access | Confidentiality breach | Enforce project membership on every project-scoped route and binding | Cross-project authorization suite |
| Qdrant collections | Public port exposure or collection deletion | Data exposure or loss | Loopback port mapping; project-scoped mediation; reconciliation and deletion confirmation | Compose config and lifecycle tests |
| Agent process | Malicious or compromised code escapes intended boundary | Host compromise | Trusted-code-only warning; subprocess limits; no sandbox claim; future gVisor/microVM work | Runtime failure and documentation tests |
| Retrieved documents | Prompt injection or malicious content influences model | Incorrect or unsafe agent output | Treat payloads as untrusted; preserve source metadata; document prompt-injection risk | Agent examples and security review |
| API | Brute force, oversized requests, or destructive mutation | Availability or data loss | Rate limits, body limits, idempotency, optimistic concurrency, audit events | API abuse and mutation tests |
| Outbox/reconciliation | Provider outage or lost event leaves desired/observed state divergent | Stale or missing resources | Transactional writes, retry/backoff, drift detection, visible degraded state | Failure-injection tests |
| Images/dependencies | Vulnerable or tampered artifact | Supply-chain compromise | Pin versions, generate SBOM, scan images, publish provenance | Release CI artifacts |
| Local volumes | Accidental reset or disk failure | Data loss | Explicit `--volumes` warning, backup/reset runbook, recovery documentation | Restore and teardown test |

## Security requirements before public preview

1. The API, PostgreSQL, Qdrant, and NATS ports bind to loopback by default.
2. The UI and CLI label trusted subprocess execution as non-sandboxed.
3. Project authorization is applied consistently to reads, writes, logs, bindings, and audit data.
4. Sensitive values never appear in structured logs, error bodies, diagnostics, or state files.
5. Mutation retries are idempotent and concurrent updates do not silently overwrite one another.
6. Dependency failures produce bounded retries and explicit degraded status.
7. Public issues do not request credentials, tokens, or unredacted diagnostics.
8. Releases identify exact images, source commit, dependencies, and known limitations.

## Non-goals

This preview does not defend against a malicious local administrator, a malicious agent with host-level access, a compromised Docker daemon, a public internet deployment, or a determined tenant attempting to escape a production sandbox. Those require a production threat model and a hardened isolation architecture.
