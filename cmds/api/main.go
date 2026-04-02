package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"owen/queueflow/internal/config"
	"owen/queueflow/internal/db"
	"owen/queueflow/internal/handlers"
	"owen/queueflow/internal/kafka"
	"owen/queueflow/internal/middleware"
	"owen/queueflow/internal/telemetry"
	"owen/queueflow/internal/worker/jobs"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	godotenv.Load(".env")
	telemetry.InitLogger(os.Getenv("LOG_LEVEL"))
	telemetry.Register()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("postgres connected")

	queries := db.NewQueries(pool)

	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	registry := jobs.NewRegistry()
	jobs.RegisterAll(registry)

	jobsHandler := handlers.NewJobsHandler(queries, producer, registry)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogging)

	r.Get("/healthz", HealthCheck)
	r.Get("/ping", Ping)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/v1", func(r chi.Router) {
		r.Use(middleware.APIKeyAuth(cfg.APIKey))
		r.Post("/jobs", jobsHandler.SubmitJob)
		r.Get("/jobs", jobsHandler.ListJobs)
		r.Get("/jobs/{id}", jobsHandler.GetJob)
		r.Post("/jobs/{id}/retry", jobsHandler.RetryJob)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: r,
	}

	go func() {
		slog.Info("api server starting", "port", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down api server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "err", err)
	}
	slog.Info("api server stopped")
}
