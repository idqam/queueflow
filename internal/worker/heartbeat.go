package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"owen/queueflow/internal/db"

	"github.com/redis/go-redis/v9"
)

type Heartbeat struct {
	rdb      *redis.Client
	db       *db.Queries
	workerID string
	ttl      time.Duration
	interval time.Duration
}

func NewHeartbeat(rdb *redis.Client, db *db.Queries, workerID string, ttl, interval time.Duration) *Heartbeat {
	return &Heartbeat{
		rdb:      rdb,
		db:       db,
		workerID: workerID,
		ttl:      ttl,
		interval: interval,
	}
}

func (h *Heartbeat) Start(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	h.beat(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("heartbeat stopped", "worker_id", h.workerID)
			return
		case <-ticker.C:
			h.beat(ctx)
		}
	}
}

func (h *Heartbeat) beat(ctx context.Context) {
	key := fmt.Sprintf("worker:%s:alive", h.workerID)

	if err := h.rdb.Set(ctx, key, time.Now().Unix(), h.ttl).Err(); err != nil {
		slog.Warn("heartbeat redis set failed", "err", err)
	}

	if err := h.db.UpsertWorker(ctx, h.workerID); err != nil {
		slog.Warn("heartbeat postgres upsert failed", "err", err)
	}
}

func WorkerID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Sprintf("worker-%d", time.Now().UnixNano())
	}
	return hostname
}
