DROP INDEX IF EXISTS outbox_due_idx;
ALTER TABLE outbox_events DROP COLUMN IF EXISTS next_retry_at;
