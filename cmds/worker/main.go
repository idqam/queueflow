package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"owen/queueflow/internal/config"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sys/windows/registry"
)

const (
	shutdownTimeOut = 60 * time.Second
	jobsTopic = "jobsTemp"
)

func main() {
	
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("worker exited with error", "err", err)
		os.Exit(1)
	}
}


func run() error {

	config, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open postgres pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	slog.Info("postgres connected")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        config.KafkaBrokers,
		Topic:          jobsTopic,
		GroupID:        config.KafkaGroupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0, // required for at-least-once
		StartOffset:    kafka.FirstOffset,
		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			slog.Debug(fmt.Sprintf(msg, args...))
		}),
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			slog.Error(fmt.Sprintf(msg, args...))
		}),
	})
	defer func() {
		slog.Info("closing kafka reader")
		if err := reader.Close(); err != nil {
			slog.Error("kafka reader close", "err", err)
		}
	}()
	slog.Info("kafka consumer initialised", "topic", jobsTopic, "group", config.KafkaTopic)

	redisOpts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(redisOpts)
	defer func() {
		if err := rdb.Close(); err != nil {
			slog.Error("redis close", "err", err)
		}
	}()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	slog.Info("redis connected")

	workerID, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("get hostname for worker id: %w", err)
	}

	reg := registry.New()
	jobs.Register(reg) //TO BE IMPLEMENTED

	go runHeartbeat(ctx, rdb, pool, workerID, config.WorkerHrbeatTTL, config.WorkerHrBeatInterval)

	runner := worker.NewRunner(worker.RunnerConfig{
		WorkerID: workerID,
		Reader:   reader,
		Pool:     pool,
		Registry: reg,
		Logger:   slog.Default(),
	})

	slog.Info("worker started, waiting for jobs", "worker_id", workerID)

	if err := runner.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("runner: %w", err)
	}

slog.Info("shutdown signal received, draining in-flight job…")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 60 * time.Second) //change timeout
	defer cancel()

	if err := runner.Drain(shutdownCtx); err != nil {
		slog.Error("drain did not complete cleanly", "err", err)
	} else {
		slog.Info("drain complete, offsets committed")
	}

	return nil
}

func runHeartbeat(
	ctx context.Context,
	rdb *redis.Client,
	pool *pgxpool.Pool,
	workerID string,
	ttl, interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	key := fmt.Sprintf("worker:heartbeat:%s", workerID)

	beat := func() {
		now := time.Now().UTC()

		if err := rdb.Set(ctx, key, now.Unix(), ttl).Err(); err != nil {
			slog.Warn("heartbeat redis set failed", "err", err)
		}

		_, err := pool.Exec(ctx, `
			INSERT INTO workers (id, last_heartbeat_at, started_at)
			VALUES ($1, $2, $2)
			ON CONFLICT (id) DO UPDATE
			  SET last_heartbeat_at = EXCLUDED.last_heartbeat_at
		`, workerID, now)
		if err != nil {
			slog.Warn("heartbeat postgres upsert failed", "err", err)
		}
	}

	beat() 

	for {
		select {
		case <-ctx.Done():
			slog.Info("heartbeat stopped")
			return
		case <-ticker.C:
			beat()
		}
	}
}