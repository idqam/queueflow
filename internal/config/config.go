package config

import (
	"owen/queueflow/env"
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
		config.ServerPort = 8080
	}

	config.APIKey = env.APIKey.GetValue()
	if config.APIKey == "" {
		config.APIKey = ""
	}

	config.DatabaseURL = env.DatabaseURL.GetValue()
	if config.DatabaseURL == "" {
		config.DatabaseURL = "postgres://user:pass@localhost:5432/queueflow"
	}

	if val := env.KafkaBrokers.GetValue(); val != "" {
		config.KafkaBrokers = strings.Split(val, ",")
	} else {
		config.KafkaBrokers = []string{"localhost:9092"}
	}

	config.KafkaGroupID = env.KafkaGroupID.GetValue()
	if config.KafkaGroupID == "" {
		config.KafkaGroupID = "queueflow-group"
	}

	config.RedisURL = env.RedisURL.GetValue()
	if config.RedisURL == "" {
		config.RedisURL = "redis://localhost:6379"
	}

	if val := env.WorkerHeartbeatTTL.GetValue(); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.WorkerHrbeatTTL = d
		} else {
			config.WorkerHrbeatTTL = 30 * time.Second
		}
	} else {
		config.WorkerHrbeatTTL = 30 * time.Second
	}

	if val := env.WorkerHeartbeatInterval.GetValue(); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.WorkerHrBeatInterval = d
		} else {
			config.WorkerHrBeatInterval = 10 * time.Second
		}
	} else {
		config.WorkerHrBeatInterval = 10 * time.Second
	}

	if val := env.JobsPerWorkerTarget.GetValue(); val != "" {
		if i, err := strconv.ParseInt(val, 10, 32); err == nil {
			config.JobsPerWorkerTarget = int32(i)
		} else {
			config.JobsPerWorkerTarget = 10
		}
	} else {
		config.JobsPerWorkerTarget = 10
	}

	if val := env.MinReplicas.GetValue(); val != "" {
		if i, err := strconv.ParseInt(val, 10, 32); err == nil {
			config.MinReplicas = int32(i)
		} else {
			config.MinReplicas = 1
		}
	} else {
		config.MinReplicas = 1
	}

	if val := env.MaxReplicas.GetValue(); val != "" {
		if i, err := strconv.ParseInt(val, 10, 32); err == nil {
			config.MaxReplicas = int32(i)
		} else {
			config.MaxReplicas = 10
		}
	} else {
		config.MaxReplicas = 10
	}

	if val := env.ScaleUpCooldown.GetValue(); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.ScaleUpCooldown = d
		} else {
			config.ScaleUpCooldown = 60 * time.Second
		}
	} else {
		config.ScaleUpCooldown = 60 * time.Second
	}

	if val := env.ScaleDownCooldown.GetValue(); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.ScaleDownCooldown = d
		} else {
			config.ScaleDownCooldown = 120 * time.Second
		}
	} else {
		config.ScaleDownCooldown = 120 * time.Second
	}

	if val := env.AutoscalerInterval.GetValue(); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.AutoscalerInterval = d
		} else {
			config.AutoscalerInterval = 30 * time.Second
		}
	} else {
		config.AutoscalerInterval = 30 * time.Second
	}

	config.WorkerDeployment = env.WorkerDeployment.GetValue()
	if config.WorkerDeployment == "" {
		config.WorkerDeployment = "queueflow-worker"
	}

	config.WorkerNamespace = env.WorkerNamespace.GetValue()
	if config.WorkerNamespace == "" {
		config.WorkerNamespace = "default"
	}

	if val := env.MetricsPort.GetValue(); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.MetricsPort = port
		} else {
			config.MetricsPort = 9090
		}
	} else {
		config.MetricsPort = 9090
	}

	return config, nil
}