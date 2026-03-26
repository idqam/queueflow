package config

import (
	"log"
	"owen/queueflow/internal/env"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	
	ServerPort int
	APIKey string
	DatabaseURL string
	KafkaBrokers []string 
	KafkaGroupID string
	KafkaTopic string
	RedisURL string
	WorkerHrbeatTTL time.Duration
	WorkerHrBeatInterval time.Duration
	JobsPerWorkerTarget int32
	MinReplicas int32
	MaxReplicas int32
	ScaleUpCooldown time.Duration
	ScaleDownCooldown time.Duration
	AutoscalerInterval time.Duration
	WorkerDeployment string
	WorkerNamespace string

	MetricsPort int
}

func Load() (*Config, error) {
	config := &Config{}

	if val := env.ServerPort.GetValue(); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.ServerPort = port
		} else {
			config.ServerPort = 8080
		}
	} else {
		log.Printf("[NON-BLOCK ERROR] SERVER_PORT not set, using default: 8080")
		config.ServerPort = 8080
	}

	config.APIKey = env.APIKey.GetValue()
	if config.APIKey == "" {
		config.APIKey = ""
	}

	config.DatabaseURL = env.DatabaseURL.GetValue()
	if config.DatabaseURL == "" {
		log.Printf("[NON-BLOCK ERROR] DATABASE_URL not set, using default: postgres://user:pass@localhost:5432/queueflow")
		config.DatabaseURL = "postgres://user:pass@localhost:5432/queueflow"
	}

	if val := env.KafkaBrokers.GetValue(); val != "" {
		config.KafkaBrokers = strings.Split(val, ",")
	} else {
		log.Printf("[NON-BLOCK ERROR] KAFKA_BROKERS not set, using default: [localhost:9092]")
		config.KafkaBrokers = []string{"localhost:9092"}
	}

	config.KafkaGroupID = env.KafkaGroupID.GetValue()
	if config.KafkaGroupID == "" {
		log.Printf("[NON-BLOCK ERROR] KAFKA_GROUP_ID not set, using default: queueflow-group")
		config.KafkaGroupID = "queueflow-group"
	}

	config.RedisURL = env.RedisURL.GetValue()
	if config.RedisURL == "" {
		log.Printf("[NON-BLOCK ERROR] REDIS_URL not set, using default: redis://localhost:6379")
		config.RedisURL = "redis://localhost:6379"
	}

	if val := env.WorkerHeartbeatTTL.GetValue(); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.WorkerHrbeatTTL = d
		} else {
			config.WorkerHrbeatTTL = 30 * time.Second
		}
	} else {
		log.Printf("[NON-BLOCK ERROR] WORKER_HEARTBEAT_TTL not set, using default: 30s")
		config.WorkerHrbeatTTL = 30 * time.Second
	}

	if val := env.WorkerHeartbeatInterval.GetValue(); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.WorkerHrBeatInterval = d
		} else {
			config.WorkerHrBeatInterval = 10 * time.Second
		}
	} else {
		log.Printf("[NON-BLOCK ERROR] WORKER_HEARTBEAT_INTERVAL not set, using default: 10s")
		config.WorkerHrBeatInterval = 10 * time.Second
	}

	if val := env.JobsPerWorkerTarget.GetValue(); val != "" {
		if i, err := strconv.ParseInt(val, 10, 32); err == nil {
			config.JobsPerWorkerTarget = int32(i)
		} else {
			config.JobsPerWorkerTarget = 10
		}
	} else {
		log.Printf("[NON-BLOCK ERROR] JOBS_PER_WORKER_TARGET not set, using default: 10")
		config.JobsPerWorkerTarget = 10
	}

	if val := env.MinReplicas.GetValue(); val != "" {
		if i, err := strconv.ParseInt(val, 10, 32); err == nil {
			config.MinReplicas = int32(i)
		} else {
			config.MinReplicas = 1
		}
	} else {
		log.Printf("[NON-BLOCK ERROR] MIN_REPLICAS not set, using default: 1")
		config.MinReplicas = 1
	}

	if val := env.MaxReplicas.GetValue(); val != "" {
		if i, err := strconv.ParseInt(val, 10, 32); err == nil {
			config.MaxReplicas = int32(i)
		} else {
			config.MaxReplicas = 10
		}
	} else {
		log.Printf("[NON-BLOCK ERROR] MAX_REPLICAS not set, using default: 10")
		config.MaxReplicas = 10
	}

	if val := env.ScaleUpCooldown.GetValue(); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.ScaleUpCooldown = d
		} else {
			config.ScaleUpCooldown = 60 * time.Second
		}
	} else {
		log.Printf("[NON-BLOCK ERROR] SCALE_UP_COOLDOWN not set, using default: 60s")
		config.ScaleUpCooldown = 60 * time.Second
	}

	if val := env.ScaleDownCooldown.GetValue(); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.ScaleDownCooldown = d
		} else {
			config.ScaleDownCooldown = 120 * time.Second
		}
	} else {
		log.Printf("[NON-BLOCK ERROR] SCALE_DOWN_COOLDOWN not set, using default: 120s")
		config.ScaleDownCooldown = 120 * time.Second
	}

	if val := env.AutoscalerInterval.GetValue(); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.AutoscalerInterval = d
		} else {
			config.AutoscalerInterval = 30 * time.Second
		}
	} else {
		log.Printf("[NON-BLOCK ERROR] AUTOSCALER_INTERVAL not set, using default: 30s")
		config.AutoscalerInterval = 30 * time.Second
	}

	config.WorkerDeployment = env.WorkerDeployment.GetValue()
	if config.WorkerDeployment == "" {
		log.Printf("[NON-BLOCK ERROR] WORKER_DEPLOYMENT not set, using default: queueflow-worker")
		config.WorkerDeployment = "queueflow-worker"
	}

	config.WorkerNamespace = env.WorkerNamespace.GetValue()
	if config.WorkerNamespace == "" {
		log.Printf("[NON-BLOCK ERROR] WORKER_NAMESPACE not set, using default: default")
		config.WorkerNamespace = "default"
	}

	if val := env.MetricsPort.GetValue(); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.MetricsPort = port
		} else {
			config.MetricsPort = 9090
		}
	} else {
		log.Printf("[NON-BLOCK ERROR] METRICS_PORT not set, using default: 9090")
		config.MetricsPort = 9090
	}

	config.KafkaTopic = env.KafkaTopic.GetValue()
	if config.KafkaTopic == "" {
		config.KafkaTopic = "DUMMY"
	}

	return config, nil
}