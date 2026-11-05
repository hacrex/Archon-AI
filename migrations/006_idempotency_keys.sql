-- Archon AI idempotent mutations, migration 006.
-- Idempotency keys let clients safely retry mutating requests. The scope is
-- the authenticated principal (or "anonymous" when no session is present), so
-- keys never collide across users. Replays return the stored response without
-- re-executing the request.

CREATE TABLE idempotency_keys (
    id                 bigserial PRIMARY KEY,
    scope              text NOT NULL,
    key                text NOT NULL,
    request_hash       text NOT NULL,
    response_status    integer,
    response_body      jsonb,
    created_at         timestamptz NOT NULL DEFAULT now(),
    completed_at       timestamptz,
    CONSTRAINT idempotency_request_hash_check CHECK (length(request_hash) = 64),
    CONSTRAINT idempotency_completed_pair_check CHECK (
        (completed_at IS NULL AND response_status IS NULL AND response_body IS NULL)
        OR (completed_at IS NOT NULL AND response_status IS NOT NULL AND response_body IS NOT NULL)
    )
);

CREATE UNIQUE INDEX idempotency_scope_key_idx ON idempotency_keys (scope, key);

CREATE INDEX idempotency_stale_idx ON idempotency_keys (created_at)
    WHERE completed_at IS NULL;
