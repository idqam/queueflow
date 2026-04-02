package autoscaler

import (
	"context"
	"log/slog"
	"owen/queueflow/internal/config"
	"owen/queueflow/internal/telemetry"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type OptionalControllerMetrics struct {
	WorkerDesiredGauge  prometheus.Gauge
	WorkerCurrentGauge  prometheus.Gauge
	QueueLagGauge       *prometheus.GaugeVec
	AutoscaleEvents     *prometheus.CounterVec
}


type Controller struct {
	KafkaLagFetcher           *KafkaLagFetcher
	K8sClient                 *K8sClient
	ScalingPolicy             *ScalingPolicy
	Config                    *config.Config
	ScaleUp                   time.Time
	ScaleDown                 time.Time
	OptionalControllerMetrics *OptionalControllerMetrics
}

func NewController(kafkaLag *KafkaLagFetcher, k8sClient *K8sClient, policy *ScalingPolicy, cfg *config.Config, metrics *OptionalControllerMetrics) *Controller {
	return &Controller{
		KafkaLagFetcher:           kafkaLag,
		K8sClient:                 k8sClient,
		ScalingPolicy:             policy,
		Config:                    cfg,
		ScaleUp:                   time.Time{},
		ScaleDown:                 time.Time{},
		OptionalControllerMetrics: metrics,
	}
}

func (c *Controller) Run(ctx context.Context) error {
	interval := c.Config.AutoscalerInterval
	if interval <= 0 {
		interval = 15 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := c.tick(ctx); err != nil {
				// continue on error, do not stop autoscaler
				continue
			}
		}
	}
}

func (c *Controller) metrics() *OptionalControllerMetrics {
	if c.OptionalControllerMetrics != nil {
		return c.OptionalControllerMetrics
	}

	return &OptionalControllerMetrics{
		WorkerDesiredGauge:  telemetry.WorkersDesired,
		WorkerCurrentGauge:  telemetry.WorkersCurrent,
		QueueLagGauge:       telemetry.QueueLag,
		AutoscaleEvents:     telemetry.ScaleEvents,
	}
}

// tick executes one control iteration: lag->desired->compare->patch (with cooldown) + metrics.
func (c *Controller) tick(ctx context.Context) error {
	lagByTopic, totalLag, err := c.KafkaLagFetcher.FetchLag(ctx)
	if err != nil {
		return err
	}

	if totalLag <= 0 {
		for _, v := range lagByTopic {
			totalLag += v
		}
	}

	desired := c.ScalingPolicy.ComputeDesired(totalLag)
	current, err := c.K8sClient.GetCurrentReplicas(ctx)
	if err != nil {
		return err
	}

	metrics := c.metrics()
	metrics.WorkerDesiredGauge.Set(float64(desired))
	metrics.WorkerCurrentGauge.Set(float64(current))
	metrics.QueueLagGauge.WithLabelValues("total").Set(float64(totalLag))
	for topic, topicLag := range lagByTopic {
		metrics.QueueLagGauge.WithLabelValues(topic).Set(float64(topicLag))
	}

	action := "none"
	now := time.Now()
	if desired > current {
		if now.Sub(c.ScaleUp) >= c.Config.ScaleUpCooldown {
			if err := c.K8sClient.PatchReplicas(ctx, desired); err != nil {
				return err
			}
			c.ScaleUp = now
			action = "scale_up"
		}
	} else if desired < current {
		if now.Sub(c.ScaleDown) >= c.Config.ScaleDownCooldown {
			if err := c.K8sClient.PatchReplicas(ctx, desired); err != nil {
				return err
			}
			c.ScaleDown = now
			action = "scale_down"
		}
	}

	if action != "none" {
		metrics.AutoscaleEvents.WithLabelValues(action).Inc()
	}
	slog.Info("autoscaler tick", "lag", totalLag, "current", current, "desired", desired, "action", action)

	return nil
}


