package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Enqueuer interface {
	Enqueue(ctx context.Context, jobType string, payload any, runAt time.Time) (string, error)
}

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) Enqueue(ctx context.Context, jobType string, payload any, runAt time.Time) (string, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal job payload: %w", err)
	}

	var id string
	if err := s.db.QueryRow(ctx, `
		INSERT INTO jobs (type, payload, status, run_at)
		VALUES ($1, $2::jsonb, 'queued', $3)
		RETURNING id::text
	`, jobType, string(encoded), runAt.UTC()).Scan(&id); err != nil {
		return "", fmt.Errorf("insert job: %w", err)
	}

	return id, nil
}

type Job struct {
	ID          string
	Type        string
	PayloadJSON []byte
	Attempts    int
	MaxAttempts int
}

type Claimer struct {
	db        *pgxpool.Pool
	workerID  string
	lockTTL   time.Duration
	maxJitter time.Duration
}

type ClaimerConfig struct {
	WorkerID string
	LockTTL  time.Duration
}

func NewClaimer(db *pgxpool.Pool, cfg ClaimerConfig) *Claimer {
	lockTTL := cfg.LockTTL
	if lockTTL <= 0 {
		lockTTL = 5 * time.Minute
	}

	return &Claimer{
		db:       db,
		workerID: cfg.WorkerID,
		lockTTL:  lockTTL,
	}
}

func (c *Claimer) ClaimNext(ctx context.Context) (*Job, error) {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin claim tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var job Job
	err = tx.QueryRow(ctx, `
		WITH next_job AS (
			SELECT id
			FROM jobs
			WHERE status = 'queued'
			  AND run_at <= now()
			  AND (locked_until IS NULL OR locked_until < now())
			ORDER BY run_at ASC, created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE jobs
		SET status = 'processing',
		    attempts = attempts + 1,
		    locked_until = now() + ($1::int * interval '1 second'),
		    locked_by = $2,
		    updated_at = now()
		WHERE id IN (SELECT id FROM next_job)
		RETURNING id::text, type, payload::text, attempts, max_attempts
	`, int(c.lockTTL.Seconds()), c.workerID).Scan(&job.ID, &job.Type, &job.PayloadJSON, &job.Attempts, &job.MaxAttempts)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("claim job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit claim tx: %w", err)
	}

	return &job, nil
}

func (c *Claimer) Complete(ctx context.Context, jobID string) error {
	_, err := c.db.Exec(ctx, `
		UPDATE jobs
		SET status = 'done',
		    locked_until = NULL,
		    locked_by = NULL,
		    updated_at = now()
		WHERE id = $1
	`, jobID)
	if err != nil {
		return fmt.Errorf("complete job: %w", err)
	}
	return nil
}

type FailureInput struct {
	JobID       string
	Attempts    int
	MaxAttempts int
	Err         error
}

func (c *Claimer) Fail(ctx context.Context, input FailureInput) error {
	status := "queued"
	nextRunAt := time.Now().UTC().Add(backoff(input.Attempts))
	if input.Attempts >= input.MaxAttempts {
		status = "failed"
		nextRunAt = time.Now().UTC()
	}

	lastErr := ""
	if input.Err != nil {
		lastErr = input.Err.Error()
		if len(lastErr) > 2000 {
			lastErr = lastErr[:2000]
		}
	}

	_, err := c.db.Exec(ctx, `
		UPDATE jobs
		SET status = $1,
		    run_at = $2,
		    locked_until = NULL,
		    locked_by = NULL,
		    last_error = $3,
		    updated_at = now()
		WHERE id = $4
	`, status, nextRunAt, lastErr, input.JobID)
	if err != nil {
		return fmt.Errorf("fail job: %w", err)
	}
	return nil
}

func backoff(attempt int) time.Duration {
	// attempt is 1-based here, because we increment attempts on claim.
	if attempt <= 1 {
		return 5 * time.Second
	}

	delay := time.Duration(1<<min(attempt, 8)) * time.Second
	if delay > 10*time.Minute {
		delay = 10 * time.Minute
	}
	return delay
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
