package models

import "time"

type Worker struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"`
	JobsProcessed int       `json:"jobs_processed"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	StartedAt     time.Time `json:"started_at"`
}
