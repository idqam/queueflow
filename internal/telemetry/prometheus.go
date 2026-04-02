package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	JobsSubmitted = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "queueflow_jobs_submitted_total",
		Help: "Total jobs submitted via API",
	}, []string{"type", "priority"})

	JobsCompleted = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "queueflow_jobs_completed_total",
		Help: "Total jobs successfully completed",
	}, []string{"type", "priority", "worker_id"})

	JobsFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "queueflow_jobs_failed_total",
		Help: "Total jobs failed",
	}, []string{"type", "priority"})

	JobsDLQ = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "queueflow_jobs_dlq_total",
		Help: "Total jobs sent to DLQ",
	}, []string{"type"})

	JobDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "queueflow_job_duration_seconds",
		Help:    "Job processing duration",
		Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60},
	}, []string{"type", "priority"})

	QueueLag = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "queueflow_queue_lag",
		Help: "Consumer group lag per topic",
	}, []string{"topic"})

	WorkersDesired = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "queueflow_worker_replicas_desired",
		Help: "Desired worker replica count",
	})

	WorkersCurrent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "queueflow_worker_replicas_current",
		Help: "Current worker replica count",
	})

	ScaleEvents = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "queueflow_autoscale_events_total",
		Help: "Autoscale events fired",
	}, []string{"direction"})
)

func Register() {
	prometheus.MustRegister(
		JobsSubmitted,
		JobsCompleted,
		JobsFailed,
		JobsDLQ,
		JobDuration,
		QueueLag,
		WorkersDesired,
		WorkersCurrent,
		ScaleEvents,
	)
}
