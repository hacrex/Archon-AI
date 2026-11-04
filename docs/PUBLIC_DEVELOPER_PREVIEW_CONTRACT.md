# Archon AI Public Developer Preview Contract

**Status:** Draft for release approval  
**Release class:** Public Developer Preview  
**Audience:** Developers evaluating Archon AI for local, self-hosted, non-sensitive AI application development

## Promise

The Public Developer Preview will provide a reproducible, self-hostable developer experience for creating projects, provisioning Qdrant-backed vector resources, and running clearly labeled trusted local Python agents through HTTP and OpenAI-compatible bindings.

This release is intended for evaluation and development. It is not a managed production service and does not provide production-grade isolation, high availability, compliance, or protection for sensitive workloads.

## Available in this preview

| Capability | Status | Contract |
|---|---|---|
| Local Compose stack | Available | Start and stop the control plane with `archon dev` and `archon dev down`. |
| Prerequisite checks | Available | `archon doctor` reports required tools, resources, ports, and workspace readiness. |
| Projects | Available | Authenticated CRUD through the versioned API and supported console workflows. |
| Qdrant DatabaseInstance | Available | Qdrant is the only supported DatabaseInstance engine in this release. |
| Authentication | Available | Local email/password authentication and bearer sessions. |
| Project authorization | Available | Project-scoped access checks for supported API routes. |
| Python SDK HTTP bindings | Available | Vector gateway, direct Qdrant vector search, and OpenAI-compatible chat bindings. |
| Trusted local Python agent execution | Preview | Runs developer-controlled code locally with explicit non-sandboxed warnings. |
| Resource reconciliation | Preview | Bounded Qdrant reconciliation, retry, and drift handling for supported resources. |
| Web console | Preview | Live supported workflows with explicit unavailable and planned states. |
| Audit events | Preview | Authentication and resource mutation events with redaction. |

## Not included

The preview does not promise:

- Multi-tenant hosted service isolation.
- High availability or automatic failover.
- Production-grade agent sandboxing.
- gVisor or microVM isolation.
- Compliance certifications or regulated-workload support.
- Encryption, retention, or recovery guarantees beyond documented local deployment behavior.
- Billing, usage accounting, quotas, or service-level objectives.
- SSO, SCIM, MFA, passkeys, or enterprise identity lifecycle.
- Multiple vector engines beyond Qdrant.
- Production model-provider failover or token accounting.
- Kubernetes operator or multi-cluster production guarantees.

## Supported profile

The first public preview supports a single-node local or enthusiast deployment with:

- Docker Engine and Compose v2.
- Go 1.22 or newer for source builds.
- Python 3.10 or newer for agents.
- At least 2 logical CPUs.
- At least 8 GiB RAM recommended.
- At least 20 GiB free disk recommended.
- Loopback-only host access by default.

Lower-capacity machines may work for small tests but are outside the recommended profile. `archon doctor` reports resource warnings without pretending that a warning is a hard platform guarantee.

## Data and security boundary

Users are responsible for the code and data they run in the preview. The trusted local agent runner must be treated as equivalent to running local developer code. Do not run untrusted agent source, expose the local API or Qdrant publicly, or use real credentials and sensitive production data.

The preview must redact passwords, bearer tokens, API keys, secret values, signed URLs, and full sensitive prompts from logs and diagnostics. A security vulnerability or suspected cross-project data exposure must be reported privately through the repository security process rather than a public issue.

## Compatibility policy

Each preview release will publish:

- Archon version and source commit.
- Supported OS and architecture profiles.
- Docker image versions or digests.
- Database and Qdrant compatibility versions.
- Python SDK compatibility range.
- Known limitations and migration notes.

Breaking changes may occur between preview releases. The release notes must identify configuration, API, SDK, and data migration impact.

## Capability labels

The console, CLI help, API documentation, and release notes use these meanings:

- **Available:** Implemented, tested, and supported within this preview contract.
- **Preview:** Usable for evaluation but subject to limitations and breaking changes.
- **Planned:** Not available; UI must not imply that it is operational.
- **Unsupported:** Explicitly rejected or outside the release boundary.
- **Degraded:** Implemented but currently unable to satisfy its dependency or desired state.

## Exit criteria for this contract

The contract is ready for public release only when the release gates document is approved, clean-machine installation passes, the primary agent workflow passes end to end, project authorization tests pass, operational diagnostics are safe to share, and the release documentation clearly repeats these limitations.
