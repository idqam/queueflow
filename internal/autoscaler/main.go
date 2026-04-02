package autoscaler

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"owen/queueflow/internal/config"
	"owen/queueflow/internal/kafka"
	"owen/queueflow/internal/telemetry"

	"github.com/joho/godotenv"
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

	k8sClient, err := NewK8sClient("", cfg.WorkerNamespace, cfg.WorkerDeployment)
	if err != nil {
		slog.Error("failed to initialize kubernetes client", "err", err)
		os.Exit(1)
	}

	kafkaLagFetcher, err := NewKafkaFetcher(cfg.KafkaBrokers, cfg.KafkaGroupID, kafka.AllJobTopics)
	if err != nil {
		slog.Error("failed to initialize kafka lag fetcher", "err", err)
		os.Exit(1)
	}

	policy := NewScalingPolicy(cfg.JobsPerWorkerTarget, cfg.MinReplicas, cfg.MaxReplicas)
	controller := NewController(kafkaLagFetcher, k8sClient, policy, cfg, nil)

	go func() {
		if err := controller.Run(ctx); err != nil && err != context.Canceled {
			slog.Error("autoscaler stopped with error", "err", err)
			os.Exit(1)
		}
	}()

	slog.Info("autoscaler started")

	<-ctx.Done()
	slog.Info("shutdown signal received for autoscaler")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	<-shutdownCtx.Done()

	slog.Info("autoscaler shutdown complete")
	fmt.Println("autoscaler exiting")
}
