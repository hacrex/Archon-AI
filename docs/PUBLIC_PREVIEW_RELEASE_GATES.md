# Public Developer Preview Release Gates

A release candidate may be published only after the gates below are reviewed and recorded against a versioned commit.

## Gate 1 — Scope and claims

- [ ] Public Developer Preview contract approved.
- [ ] Available, Preview, Planned, Unsupported, and Degraded labels are consistent across UI, CLI, API documentation, and release notes.
- [ ] Documentation states that trusted local agent execution is not a sandbox.
- [ ] Documentation states that the release is not for sensitive production workloads.
- [ ] Unsupported engines and incomplete features are not presented as operational.

## Gate 2 — Installation and lifecycle

- [ ] `archon doctor` passes required checks on supported profiles.
- [ ] `archon dev` starts from a clean workspace.
- [ ] `/livez` and `/readyz` become healthy within the documented timeout.
- [ ] A readiness timeout cleans up partially started services.
- [ ] `archon dev down` stops services and removes state.
- [ ] `archon dev down --volumes` has an explicit data-loss warning in the documentation.
- [ ] Compose ports bind to loopback by default.
- [ ] Clean-machine installation has been tested outside the author’s workstation.

## Gate 3 — API safety

- [ ] Authentication and session revocation tests pass.
- [ ] Cross-project and cross-organization authorization tests pass.
- [ ] Idempotency keys are implemented for supported mutations.
- [ ] Optimistic concurrency prevents silent overwrites.
- [ ] Request body, pagination, and timeout limits are enforced.
- [ ] Login and sensitive-mutation rate limiting is enabled.
- [ ] Migration startup is safe with concurrent API instances.
- [ ] API errors contain stable codes, field paths, and request IDs.

## Gate 4 — Resource lifecycle

- [ ] Qdrant is the only engine marked Available.
- [ ] Create, update, delete, retry, and drift behavior is covered by adapter tests.
- [ ] Dependency outages produce bounded retries and visible Degraded/Failed state.
- [ ] Resource status includes observed generation and conditions.
- [ ] Outbox writes and publish failures are observable.
- [ ] PostgreSQL and Qdrant backup/reset procedures have been tested.

## Gate 5 — Agent workflow

- [ ] Manifest validation rejects unsupported runtime and isolation claims.
- [ ] Trusted subprocess execution enforces timeout and captures stdout/stderr.
- [ ] Secrets are explicitly allowlisted and redacted.
- [ ] Agent crash, timeout, malformed response, vector error, and model error paths are tested.
- [ ] Deploy, invoke, status, and logs workflows work through the supported interface.
- [ ] Vector and model bindings have integration coverage.
- [ ] Streaming response behavior is documented and tested.

## Gate 6 — Observability and support

- [ ] API, reconciliation, Qdrant, outbox, and invocation metrics are available.
- [ ] Structured logs include request IDs and omit secrets, tokens, prompts, and signed URLs.
- [ ] `archon diagnostics` produces a safe-to-share bundle.
- [ ] Issue templates request versions, commands, request IDs, and sanitized diagnostics.
- [ ] Incident and security reporting instructions are public.

## Gate 7 — Supply chain and release

- [ ] Go and Python dependencies are pinned or reviewed by policy.
- [ ] Container and dependency vulnerability scans pass the approved threshold.
- [ ] SBOM is generated and stored with the release.
- [ ] Images and release artifacts identify source commit and versions.
- [ ] Rollback and migration notes are published.
- [ ] Full CI is green: tests, vet, build, formatting, SDK tests, frontend checks, YAML/JSON validation, Helm, Compose, and security scans.
- [ ] Product, engineering, and security approvers have signed off.

## Evidence package

Each release candidate should retain:

- Commit SHA and version.
- CI run URL and artifact links.
- Clean-machine installation output.
- End-to-end test report.
- Vulnerability scan and SBOM.
- Known limitations and rollback instructions.
- Approval record with named reviewers.
