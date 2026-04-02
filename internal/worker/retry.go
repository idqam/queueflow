package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"time"

	"owen/queueflow/internal/db"
	"owen/queueflow/internal/kafka"
	"owen/queueflow/internal/models"
)

type RetryHandler struct {
	db       *db.Queries
	producer *kafka.Producer
}

func NewRetryHandler(db *db.Queries, producer *kafka.Producer) *RetryHandler {
	return &RetryHandler{db: db, producer: producer}
}

func (h *RetryHandler) Handle(ctx context.Context, msg models.KafkaMessage, topic string, jobErr error) {
	job, err := h.db.GetJob(ctx, msg.JobID)
	if err != nil {
		slog.Error("retry: failed to get job", "job_id", msg.JobID, "err", err)
		return
	}

	if job.Attempts < job.MaxAttempts {
		backoff := time.Duration(math.Pow(2, float64(job.Attempts-1))) * 2 * time.Second
		slog.Info("retrying job", "job_id", msg.JobID, "attempt", job.Attempts, "backoff", backoff)

		h.db.UpdateJobFailed(ctx, msg.JobID, jobErr.Error())
		h.db.InsertJobEvent(ctx, &models.JobEvent{
			JobID:     msg.JobID,
			EventType: "retry",
			OldStatus: models.JobStatusProcessing,
			NewStatus: models.JobStatusQueued,
		})

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		}

		h.db.UpdateJobStatus(ctx, msg.JobID, models.JobStatusQueued)

		retryMsg := models.KafkaMessage{
			JobID:      msg.JobID,
			Type:       msg.Type,
			Attempt:    job.Attempts,
			EnqueuedAt: time.Now().UTC(),
			Payload:    msg.Payload,
		}
		msgBytes, _ := json.Marshal(retryMsg)
		if err := h.producer.Produce(ctx, topic, msg.JobID, msgBytes); err != nil {
			slog.Error("retry: produce failed", "job_id", msg.JobID, "err", err)
		}
		return
	}

	slog.Warn("job exhausted retries, sending to DLQ", "job_id", msg.JobID, "attempts", job.Attempts)

	h.db.UpdateJobDead(ctx, msg.JobID, jobErr.Error())
	h.db.InsertJobEvent(ctx, &models.JobEvent{
		JobID:     msg.JobID,
		EventType: "dlq_enqueue",
		OldStatus: models.JobStatusProcessing,
		NewStatus: models.JobStatusDead,
	})

	dlqMsg := models.KafkaMessage{
		JobID:      msg.JobID,
		Type:       msg.Type,
		Attempt:    job.Attempts,
		EnqueuedAt: time.Now().UTC(),
		Payload:    msg.Payload,
	}
	msgBytes, _ := json.Marshal(dlqMsg)
	if err := h.producer.Produce(ctx, kafka.TopicJobsDLQ, msg.JobID, msgBytes); err != nil {
		slog.Error("dlq produce failed", "job_id", msg.JobID, "err", err)
	}
}
