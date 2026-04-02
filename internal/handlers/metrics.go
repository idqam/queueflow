package handlers

import (
	"context"
	"net/http"
	"time"

	"owen/queueflow/internal/autoscaler"
	"owen/queueflow/internal/db"
	"owen/queueflow/internal/kafka"

	"github.com/redis/go-redis/v9"
)

const workerDeadThreshold = 45 * time.Second

type MetricsHandler struct {
	db       *db.Queries
	lagFetch *autoscaler.KafkaLagFetcher
	redis    *redis.Client
}

func NewMetricsHandler(db *db.Queries, lagFetch *autoscaler.KafkaLagFetcher, rdb *redis.Client) *MetricsHandler {
	return &MetricsHandler{db: db, lagFetch: lagFetch, redis: rdb}
}

type queueMetricsResponse struct {
	TopicLag map[string]int64 `json:"topic_lag"`
	DLQDepth int64            `json:"dlq_depth"`
	TotalLag int64            `json:"total_lag"`
}

func (h *MetricsHandler) GetQueueMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	perTopic, totalLag, err := h.lagFetch.FetchLag(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch kafka lag: "+err.Error())
		return
	}

	topicLag := map[string]int64{
		kafka.TopicJobsHigh:   perTopic[kafka.TopicJobsHigh],
		kafka.TopicJobsNormal: perTopic[kafka.TopicJobsNormal],
		kafka.TopicJobsLow:    perTopic[kafka.TopicJobsLow],
	}

	dlqDepth, err := h.db.CountDeadJobs(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch dlq depth")
		return
	}

	writeJSON(w, http.StatusOK, queueMetricsResponse{
		TopicLag: topicLag,
		DLQDepth: dlqDepth,
		TotalLag: totalLag,
	})
}

type workerInfo struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"`
	JobsProcessed int       `json:"jobs_processed"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

type workersMetricsResponse struct {
	DesiredReplicas int          `json:"desired_replicas"`
	ReadyReplicas   int          `json:"ready_replicas"`
	Workers         []workerInfo `json:"workers"`
}

func (h *MetricsHandler) GetWorkerMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	workers, err := h.db.ListWorkers(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list workers")
		return
	}

	infos := make([]workerInfo, 0, len(workers))
	readyCount := 0

	for _, wr := range workers {
		status := resolveWorkerStatus(ctx, h.redis, wr.ID, wr.LastHeartbeat)
		if status == "active" {
			readyCount++
		}
		infos = append(infos, workerInfo{
			ID:            wr.ID,
			Status:        status,
			JobsProcessed: wr.JobsProcessed,
			LastHeartbeat: wr.LastHeartbeat,
		})
	}

	writeJSON(w, http.StatusOK, workersMetricsResponse{
		DesiredReplicas: len(workers),
		ReadyReplicas:   readyCount,
		Workers:         infos,
	})
}

func resolveWorkerStatus(ctx context.Context, rdb *redis.Client, workerID string, lastHeartbeat time.Time) string {
	key := "worker:" + workerID + ":alive"
	ttl, err := rdb.TTL(ctx, key).Result()
	if err == nil && ttl > 0 {
		return "active"
	}

	if time.Since(lastHeartbeat) <= workerDeadThreshold {
		return "active"
	}
	return "dead"
}
