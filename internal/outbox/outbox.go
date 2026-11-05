// Package outbox drains the transactional outbox written by the repository
// layer and publishes events to a pluggable Sink (in-process handlers today,
// NATS JetStream later). Publishing is at-least-once: an event is marked
// published only after the sink accepts it. Failed deliveries are retried
// with exponential backoff plus jitter and moved to the dead-letter state
// after MaxAttempts, where they can be replayed through the store API.
package outbox

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// Event is one row of the outbox_events table.
type Event struct {
	ID            int64
	Aggregate     string
	AggregateID   string
	EventType     string
	Payload       []byte
	Attempts      int
	LastErrorCode string
}

// Sink delivers an event. Returning an error marks the attempt as failed and
// schedules a retry with backoff.
type Sink interface {
	Publish(context.Context, Event) error
}

// SinkFunc adapts a function to the Sink interface.
type SinkFunc func(context.Context, Event) error

func (f SinkFunc) Publish(ctx context.Context, event Event) error { return f(ctx, event) }

// LogSink is the default in-process sink: it records the delivery so local
// development has a visible event stream without a message broker.
var LogSink = SinkFunc(func(_ context.Context, event Event) error {
	log.Printf(`{"outbox_event":true,"id":%d,"aggregate_type":%q,"aggregate_id":%q,"event_type":%q}`,
		event.ID, event.Aggregate, event.AggregateID, event.EventType)
	return nil
})

// PublisherConfig controls drain behaviour. Zero values select the defaults.
type PublisherConfig struct {
	BatchSize     int           // events claimed per poll (default 25, max 100)
	PollInterval  time.Duration // idle wait between polls (default 2s)
	MaxAttempts   int           // attempts before dead-lettering (default 10)
	BaseBackoff   time.Duration // first retry delay (default 500ms)
	MaxBackoff    time.Duration // retry delay ceiling (default 2m)
	BackoffFactor float64       // exponential growth factor (default 2.0)
	RetentionTime time.Duration // completed events older than this are pruned (default 72h; <=0 disables pruning)
}

func (c PublisherConfig) withDefaults() PublisherConfig {
	if c.BatchSize <= 0 {
		c.BatchSize = 25
	}
	if c.BatchSize > 100 {
		c.BatchSize = 100
	}
	if c.PollInterval <= 0 {
		c.PollInterval = 2 * time.Second
	}
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 10
	}
	if c.BaseBackoff <= 0 {
		c.BaseBackoff = 500 * time.Millisecond
	}
	if c.MaxBackoff <= 0 {
		c.MaxBackoff = 2 * time.Minute
	}
	if c.BackoffFactor < 1 {
		c.BackoffFactor = 2.0
	}
	if c.RetentionTime <= 0 {
		c.RetentionTime = 72 * time.Hour
	}
	return c
}

// NextRetryDelay computes the exponential backoff for a completed attempt
// count (attempts >= 1), capped at MaxBackoff, with up to 25 % jitter.
func (c PublisherConfig) NextRetryDelay(attempts int) time.Duration {
	c = c.withDefaults()
	delay := float64(c.BaseBackoff)
	for i := 1; i < attempts && delay < float64(c.MaxBackoff); i++ {
		delay *= c.BackoffFactor
	}
	if delay > float64(c.MaxBackoff) {
		delay = float64(c.MaxBackoff)
	}
	jitter := 1.0 + rand.Float64()*0.25
	return time.Duration(delay * jitter)
}

// Publisher drains outbox_events with SELECT ... FOR UPDATE SKIP LOCKED so
// multiple replicas can run concurrently without double-delivering a batch.
type Publisher struct {
	db     *sql.DB
	sink   Sink
	config PublisherConfig
}

func NewPublisher(db *sql.DB, sink Sink, config PublisherConfig) (*Publisher, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}
	if sink == nil {
		sink = LogSink
	}
	return &Publisher{db: db, sink: sink, config: config.withDefaults()}, nil
}

// Run polls until ctx is cancelled. It returns when the context ends; the
// final DrainOnce error is logged rather than returned because transient
// dependency failures must not kill the worker loop.
func (p *Publisher) Run(ctx context.Context) {
	for {
		published, err := p.DrainOnce(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("outbox drain error: %v", err)
		} else if published > 0 {
			log.Printf("outbox drained %d event(s)", published)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(p.config.PollInterval):
		}
	}
}

// DrainOnce claims one batch of due events, publishes each through the sink,
// and records success or scheduled retries. It returns the number of events
// successfully published.
func (p *Publisher) DrainOnce(ctx context.Context) (int, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin outbox drain: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, aggregate_type, aggregate_id::text, event_type, payload, attempts
		FROM outbox_events
		WHERE published_at IS NULL
		  AND last_error IS DISTINCT FROM 'dead_letter'
		  AND (next_retry_at IS NULL OR next_retry_at <= now())
		ORDER BY id
		LIMIT $1
		FOR UPDATE SKIP LOCKED`, p.config.BatchSize)
	if err != nil {
		return 0, fmt.Errorf("claim outbox events: %w", err)
	}
	events := []Event{}
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.Aggregate, &event.AggregateID, &event.EventType, &event.Payload, &event.Attempts); err != nil {
			_ = rows.Close()
			return 0, fmt.Errorf("scan outbox event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return 0, fmt.Errorf("iterate outbox events: %w", err)
	}
	_ = rows.Close()

	published := 0
	for _, event := range events {
		attempts := event.Attempts + 1
		if err := p.sink.Publish(ctx, event); err != nil {
			if publishErr := p.recordFailure(ctx, tx, event.ID, attempts, err); publishErr != nil {
				return published, publishErr
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE outbox_events SET published_at = now(), attempts = $2, last_error = NULL, next_retry_at = NULL WHERE id = $1`, event.ID, attempts); err != nil {
			return published, fmt.Errorf("mark outbox event %d published: %w", event.ID, err)
		}
		published++
	}
	if err := tx.Commit(); err != nil {
		return published, fmt.Errorf("commit outbox drain: %w", err)
	}
	if n := len(events) - published; n > 0 {
		log.Printf("outbox retried %d event(s) with backoff", n)
	}
	return published, nil
}

func (p *Publisher) recordFailure(ctx context.Context, tx *sql.Tx, id int64, attempts int, cause error) error {
	if attempts >= p.config.MaxAttempts {
		if _, err := tx.ExecContext(ctx, `UPDATE outbox_events SET attempts = $2, last_error = 'dead_letter', next_retry_at = NULL WHERE id = $1`, id, attempts); err != nil {
			return fmt.Errorf("dead-letter outbox event %d: %w", id, err)
		}
		log.Printf("outbox event %d moved to dead-letter after %d attempts: %v", id, attempts, cause)
		return nil
	}
	delay := p.config.NextRetryDelay(attempts)
	if _, err := tx.ExecContext(ctx, `UPDATE outbox_events SET attempts = $2, last_error = $3, next_retry_at = now() + $4::interval WHERE id = $1`,
		id, attempts, trimError(cause), fmt.Sprintf("%.3f seconds", delay.Seconds())); err != nil {
		return fmt.Errorf("schedule outbox retry %d: %w", id, err)
	}
	return nil
}

func trimError(err error) string {
	message := err.Error()
	if len(message) > 512 {
		return message[:512]
	}
	return message
}
