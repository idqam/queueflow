package env

import "os"

type EnvKey string

func (key EnvKey) GetValue() string {
 return os.Getenv(string(key))
}

const (
 ServerPort             EnvKey = "SERVER_PORT"
 APIKey                 EnvKey = "API_KEY"
 DatabaseURL            EnvKey = "DATABASE_URL"
 KafkaBrokers           EnvKey = "KAFKA_BROKERS"
 KafkaGroupID           EnvKey = "KAFKA_GROUP_ID"
 KafkaTopic 			EnvKey = "DUMMY"
 RedisURL               EnvKey = "REDIS_URL"
 WorkerHeartbeatTTL     EnvKey = "WORKER_HEARTBEAT_TTL"
 WorkerHeartbeatInterval EnvKey = "WORKER_HEARTBEAT_INTERVAL"
 JobsPerWorkerTarget    EnvKey = "JOBS_PER_WORKER_TARGET"
 MinReplicas            EnvKey = "MIN_REPLICAS"
 MaxReplicas            EnvKey = "MAX_REPLICAS"
 ScaleUpCooldown        EnvKey = "SCALE_UP_COOLDOWN"
 ScaleDownCooldown      EnvKey = "SCALE_DOWN_COOLDOWN"
 AutoscalerInterval     EnvKey = "AUTOSCALER_INTERVAL"
 WorkerDeployment       EnvKey = "WORKER_DEPLOYMENT"
 WorkerNamespace        EnvKey = "WORKER_NAMESPACE"
 MetricsPort            EnvKey = "METRICS_PORT"
)