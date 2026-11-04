# Archon AI Capability Status Registry

This registry is the source of truth for the capability labels used in the Public Developer Preview. The console, CLI help, API documentation, roadmap, and release notes should use the same status and limitations.

## Status meanings

| Status | Meaning |
|---|---|
| Available | Implemented, tested, and supported within the current preview contract. |
| Preview | Usable for evaluation, but limited or subject to breaking changes. |
| Planned | Not available; the interface must not imply successful execution. |
| Unsupported | Explicitly rejected or outside the release boundary. |
| Degraded | Implemented but currently unable to satisfy a dependency or desired state. |

## Current registry

| Capability | Status | User-visible limitation |
|---|---|---|
| Local Compose lifecycle | Available | Single-node development deployment only. |
| `archon doctor` | Available | Resource checks are advisory; required tool and workspace checks are enforced. |
| Projects API | Available | Authenticated local control plane; production tenancy is not implied. |
| DatabaseInstance API | Available | Qdrant is the only supported engine in this preview. |
| Qdrant provisioning | Preview | Reconciliation and provider recovery remain bounded preview behavior. |
| Authentication | Available | Local email/password and bearer sessions; enterprise identity is not included. |
| Project authorization | Preview | Supported routes are protected; full enterprise role management is not included. |
| Python SDK vector binding | Available | Requires a configured text gateway, Archon runtime endpoint, or direct Qdrant vector. |
| Python SDK model binding | Available | Requires an OpenAI-compatible gateway; provider routing is outside the preview. |
| Trusted local agent execution | Preview | Developer-controlled code only; this is not a security sandbox. |
| Agent deployment/invocation | Planned | Do not show a successful deployment until the runtime endpoint exists. |
| Web console | Preview | Some administrative and operational surfaces may remain unavailable. |
| Audit events | Preview | Coverage expands as lifecycle and membership mutations are implemented. |
| Additional vector engines | Unsupported | Chroma, Weaviate, Milvus, pgvector, and ScyllaDB are not available in this release. |
| Hosted multi-tenant service | Unsupported | This repository provides a self-hosted preview, not a managed service. |
| Kubernetes production controller | Planned | Helm and K3s practice assets do not constitute production controller support. |
| Agent sandbox isolation | Planned | gVisor or microVM isolation is required before making a sandbox claim. |

## Change process

A capability may move from Planned or Preview to Available only when its tests, documentation, security boundary, and release-gate evidence are complete. A capability moves to Degraded at runtime when its dependency or observed state prevents normal operation; the UI and API must preserve the reason and recovery action.
