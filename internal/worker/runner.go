package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"owen/queueflow/internal/db"
	"owen/queueflow/internal/models"
	"owen/queueflow/internal/worker/jobs"

	kafkago "github.com/segmentio/kafka-go"
)

type Runner struct {
	workerID string
	consumer *kafkago.Reader
	db       *db.Queries
	registry *jobs.Registry
	retry    *RetryHandler

	mu         sync.Mutex
	processing bool
}

type RunnerConfig struct {
	WorkerID string
	Reader   *kafkago.Reader
	DB       *db.Queries
	Registry *jobs.Registry
	Retry    *RetryHandler
}

func NewRunner(cfg RunnerConfig) *Runner {
	return &Runner{
		workerID: cfg.WorkerID,
		consumer: cfg.Reader,
		db:       cfg.DB,
		registry: cfg.Registry,
		retry:    cfg.Retry,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	for {
		msg, err := r.consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("fetch message failed", "err", err)
			continue
		}

		r.mu.Lock()
		r.processing = true
		r.mu.Unlock()

		r.handleMessage(ctx, msg)

		r.mu.Lock()
		r.processing = false
		r.mu.Unlock()

		if err := r.consumer.CommitMessages(ctx, msg); err != nil {
			slog.Error("commit offset failed", "err", err)
		}
	}
}

func (r *Runner) handleMessage(ctx context.Context, msg kafkago.Message) {
	var km models.KafkaMessage
	if err := json.Unmarshal(msg.Value, &km); err != nil {
		slog.Error("unmarshal kafka message failed", "err", err)
		return
	}

	handler, ok := r.registry.Get(km.Type)
	if !ok {
		slog.Error("unknown job type", "type", km.Type, "job_id", km.JobID)
		return
	}

	slog.Info("processing job", "job_id", km.JobID, "type", km.Type, "attempt", km.Attempt)

	if err := r.db.UpdateJobProcessing(ctx, km.JobID, r.workerID); err != nil {
		slog.Error("update job processing failed", "job_id", km.JobID, "err", err)
		return
	}

	r.db.InsertJobEvent(ctx, &models.JobEvent{
		JobID:     km.JobID,
		EventType: "status_change",
		OldStatus: models.JobStatusQueued,
		NewStatus: models.JobStatusProcessing,
		WorkerID:  r.workerID,
	})

	result, err := r.safeCall(ctx, handler, km.Payload)
	if err != nil {
		slog.Warn("job failed", "job_id", km.JobID, "err", err)
		r.retry.Handle(ctx, km, msg.Topic, err)
		return
	}

	if err := r.db.UpdateJobCompleted(ctx, km.JobID, result); err != nil {
		slog.Error("update job completed failed", "job_id", km.JobID, "err", err)
		return
	}

	r.db.InsertJobEvent(ctx, &models.JobEvent{
		JobID:     km.JobID,
		EventType: "status_change",
		OldStatus: models.JobStatusProcessing,
		NewStatus: models.JobStatusCompleted,
		WorkerID:  r.workerID,
	})

	r.db.IncrementWorkerJobsProcessed(ctx, r.workerID)
	slog.Info("job completed", "job_id", km.JobID, "type", km.Type)
}

func (r *Runner) safeCall(ctx context.Context, handler jobs.Handler, payload json.RawMessage) (result json.RawMessage, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("handler panicked: %v", rec)
		}
	}()
	return handler(ctx, payload)
}

func (r *Runner) Drain(ctx context.Context) error {
	r.mu.Lock()
	isProcessing := r.processing
	r.mu.Unlock()

	if !isProcessing {
		return nil
	}

	<-ctx.Done()
	return ctx.Err()
}
