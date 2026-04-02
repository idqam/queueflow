package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"owen/queueflow/internal/config"
	"owen/queueflow/internal/db"
	"owen/queueflow/internal/kafka"
	"owen/queueflow/internal/telemetry"
	"owen/queueflow/internal/worker"
	"owen/queueflow/internal/worker/jobs"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	kafkago "github.com/segmentio/kafka-go"
)

func main() {
	godotenv.Load(".env")
	telemetry.InitLogger(os.Getenv("LOG_LEVEL"))

	if err := run(); err != nil {
		slog.Error("worker exited with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	slog.Info("postgres connected")

	queries := db.NewQueries(pool)

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     cfg.KafkaBrokers,
		GroupID:     cfg.KafkaGroupID,
		GroupTopics: kafka.AllJobTopics,
		MinBytes:    1,
		MaxBytes:    10e6,
		CommitInterval: 0,
		StartOffset: kafkago.FirstOffset,
	})
	defer reader.Close()
	slog.Info("kafka consumer initialised", "topics", kafka.AllJobTopics, "group", cfg.KafkaGroupID)

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	slog.Info("redis connected")

	workerID := worker.WorkerID()

	registry := jobs.NewRegistry()
	jobs.RegisterAll(registry)

	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	retryHandler := worker.NewRetryHandler(queries, producer)

	heartbeat := worker.NewHeartbeat(rdb, queries, workerID, cfg.WorkerHrbeatTTL, cfg.WorkerHrBeatInterval)
	go heartbeat.Start(ctx)

	runner := worker.NewRunner(worker.RunnerConfig{
		WorkerID: workerID,
		Reader:   reader,
		DB:       queries,
		Registry: registry,
		Retry:    retryHandler,
	})

	slog.Info("worker started, waiting for jobs", "worker_id", workerID)

	if err := runner.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("runner: %w", err)
	}

	slog.Info("shutdown signal received, draining in-flight job")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := runner.Drain(shutdownCtx); err != nil {
		slog.Error("drain did not complete cleanly", "err", err)
	} else {
		slog.Info("drain complete")
	}

	return nil
}
