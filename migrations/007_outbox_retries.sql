-- Archon AI outbox publisher support, migration 007.
-- The outbox publisher schedules retries per event. next_retry_at gates when
-- an event becomes due again after a failed delivery; dead-lettered events
-- carry last_error = 'dead_letter' and are excluded until replayed.

ALTER TABLE outbox_events ADD COLUMN next_retry_at timestamptz;

CREATE INDEX outbox_due_idx ON outbox_events (next_retry_at, id)
    WHERE published_at IS NULL;
