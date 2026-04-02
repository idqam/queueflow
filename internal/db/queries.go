package db

import (
	"context"
	"encoding/json"
	"fmt"

	"owen/queueflow/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Queries struct {
	pool *pgxpool.Pool
}

func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

func (q *Queries) InsertJob(ctx context.Context, job *models.Job) error {
	_, err := q.pool.Exec(ctx, `
		INSERT INTO jobs (id, type, priority, status, payload, max_attempts, callback_url, namespace)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		job.ID, job.Type, job.Priority, job.Status, job.Payload, job.MaxAttempts, job.CallbackURL, job.Namespace,
	)
	return err
}

func (q *Queries) GetJob(ctx context.Context, id string) (*models.Job, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT id, type, priority, status, payload, result, error_message,
		       attempts, max_attempts, worker_id, callback_url, namespace,
		       created_at, queued_at, started_at, completed_at
		FROM jobs WHERE id = $1`, id)

	var j models.Job
	var result, payload []byte
	var errMsg, workerID, callbackURL *string

	err := row.Scan(
		&j.ID, &j.Type, &j.Priority, &j.Status, &payload, &result, &errMsg,
		&j.Attempts, &j.MaxAttempts, &workerID, &callbackURL, &j.Namespace,
		&j.CreatedAt, &j.QueuedAt, &j.StartedAt, &j.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	j.Payload = payload
	j.Result = result
	if errMsg != nil {
		j.ErrorMessage = *errMsg
	}
	if workerID != nil {
		j.WorkerID = *workerID
	}
	if callbackURL != nil {
		j.CallbackURL = *callbackURL
	}
	return &j, nil
}

type ListJobsParams struct {
	Status    string
	Priority  string
	Namespace string
	JobType   string
	Limit     int
	Cursor    string
}

func (q *Queries) ListJobs(ctx context.Context, params ListJobsParams) ([]models.Job, error) {
	query := `SELECT id, type, priority, status, payload, result, error_message,
	                 attempts, max_attempts, worker_id, callback_url, namespace,
	                 created_at, queued_at, started_at, completed_at
	          FROM jobs WHERE 1=1`
	args := []any{}
	argIdx := 1

	if params.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, params.Status)
		argIdx++
	}
	if params.Priority != "" {
		query += fmt.Sprintf(" AND priority = $%d", argIdx)
		args = append(args, params.Priority)
		argIdx++
	}
	if params.Namespace != "" {
		query += fmt.Sprintf(" AND namespace = $%d", argIdx)
		args = append(args, params.Namespace)
		argIdx++
	}
	if params.JobType != "" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, params.JobType)
		argIdx++
	}
	if params.Cursor != "" {
		query += fmt.Sprintf(" AND created_at < (SELECT created_at FROM jobs WHERE id = $%d)", argIdx)
		args = append(args, params.Cursor)
		argIdx++
	}

	if params.Limit <= 0 || params.Limit > 200 {
		params.Limit = 50
	}
	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, params.Limit)

	rows, err := q.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanJobs(rows)
}

func scanJobs(rows pgx.Rows) ([]models.Job, error) {
	var jobs []models.Job
	for rows.Next() {
		var j models.Job
		var result, payload []byte
		var errMsg, workerID, callbackURL *string

		if err := rows.Scan(
			&j.ID, &j.Type, &j.Priority, &j.Status, &payload, &result, &errMsg,
			&j.Attempts, &j.MaxAttempts, &workerID, &callbackURL, &j.Namespace,
			&j.CreatedAt, &j.QueuedAt, &j.StartedAt, &j.CompletedAt,
		); err != nil {
			return nil, err
		}

		j.Payload = payload
		j.Result = result
		if errMsg != nil {
			j.ErrorMessage = *errMsg
		}
		if workerID != nil {
			j.WorkerID = *workerID
		}
		if callbackURL != nil {
			j.CallbackURL = *callbackURL
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (q *Queries) UpdateJobStatus(ctx context.Context, id string, status models.JobStatus) error {
	_, err := q.pool.Exec(ctx, `UPDATE jobs SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (q *Queries) UpdateJobProcessing(ctx context.Context, id, workerID string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE jobs SET status = 'PROCESSING', worker_id = $1, started_at = NOW(), attempts = attempts + 1
		WHERE id = $2`, workerID, id)
	return err
}

func (q *Queries) UpdateJobCompleted(ctx context.Context, id string, result json.RawMessage) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE jobs SET status = 'COMPLETED', result = $1, completed_at = NOW()
		WHERE id = $2`, result, id)
	return err
}

func (q *Queries) UpdateJobFailed(ctx context.Context, id, errMessage string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE jobs SET status = 'FAILED', error_message = $1
		WHERE id = $2`, errMessage, id)
	return err
}

func (q *Queries) UpdateJobDead(ctx context.Context, id, errMessage string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE jobs SET status = 'DEAD', error_message = $1
		WHERE id = $2`, errMessage, id)
	return err
}

func (q *Queries) ResetJobForRetry(ctx context.Context, id string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE jobs SET status = 'QUEUED', attempts = 0, error_message = NULL,
		       result = NULL, worker_id = NULL, started_at = NULL, completed_at = NULL, queued_at = NOW()
		WHERE id = $1`, id)
	return err
}

func (q *Queries) InsertJobEvent(ctx context.Context, event *models.JobEvent) error {
	_, err := q.pool.Exec(ctx, `
		INSERT INTO job_events (job_id, event_type, old_status, new_status, worker_id)
		VALUES ($1, $2, $3, $4, $5)`,
		event.JobID, event.EventType, event.OldStatus, event.NewStatus, event.WorkerID,
	)
	return err
}

func (q *Queries) GetJobEvents(ctx context.Context, jobID string) ([]models.JobEvent, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, job_id, event_type, old_status, new_status, worker_id, occurred_at
		FROM job_events WHERE job_id = $1 ORDER BY occurred_at ASC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.JobEvent
	for rows.Next() {
		var e models.JobEvent
		var oldStatus, newStatus, workerID *string
		if err := rows.Scan(&e.ID, &e.JobID, &e.EventType, &oldStatus, &newStatus, &workerID, &e.OccuredAt); err != nil {
			return nil, err
		}
		if oldStatus != nil {
			e.OldStatus = models.JobStatus(*oldStatus)
		}
		if newStatus != nil {
			e.NewStatus = models.JobStatus(*newStatus)
		}
		if workerID != nil {
			e.WorkerID = *workerID
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (q *Queries) UpsertWorker(ctx context.Context, workerID string) error {
	_, err := q.pool.Exec(ctx, `
		INSERT INTO workers (id, last_heartbeat, started_at)
		VALUES ($1, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE
		SET last_heartbeat = NOW()`, workerID)
	return err
}

func (q *Queries) IncrementWorkerJobsProcessed(ctx context.Context, workerID string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE workers SET jobs_processed = jobs_processed + 1 WHERE id = $1`, workerID)
	return err
}

func (q *Queries) ListWorkers(ctx context.Context) ([]models.Worker, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, status, jobs_processed, last_heartbeat, started_at FROM workers ORDER BY started_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workers []models.Worker
	for rows.Next() {
		var w models.Worker
		if err := rows.Scan(&w.ID, &w.Status, &w.JobsProcessed, &w.LastHeartbeat, &w.StartedAt); err != nil {
			return nil, err
		}
		workers = append(workers, w)
	}
	return workers, rows.Err()
}
