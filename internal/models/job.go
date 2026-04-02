package models

import (
	"encoding/json"
	"time"
)

type JobStatus string

const (
	JobStatusQueued     JobStatus = "QUEUED"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
	JobStatusDead       JobStatus = "DEAD"
)

type JobPriority string

const (
	PriorityHigh   JobPriority = "high"
	PriorityNormal JobPriority = "normal"
	PriorityLow    JobPriority = "low"
)

var ValidPriorities = map[JobPriority]bool{
	PriorityHigh:   true,
	PriorityNormal: true,
	PriorityLow:    true,
}

type Job struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Priority     JobPriority     `json:"priority"`
	Status       JobStatus       `json:"status"`
	Payload      json.RawMessage `json:"payload"`
	Result       json.RawMessage `json:"result,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
	Attempts     int             `json:"attempts"`
	MaxAttempts  int             `json:"max_attempts"`
	WorkerID     string          `json:"worker_id,omitempty"`
	CallbackURL  string          `json:"callback_url,omitempty"`
	Namespace    string          `json:"namespace"`
	CreatedAt    time.Time       `json:"created_at"`
	QueuedAt     time.Time       `json:"queued_at"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
}

type KafkaMessage struct {
	JobID      string          `json:"job_id"`
	Type       string          `json:"type"`
	Attempt    int             `json:"attempt"`
	EnqueuedAt time.Time       `json:"enqueued_at"`
	Payload    json.RawMessage `json:"payload"`
}

type JobEvent struct {
	ID        int64     `json:"id"`
	JobID     string    `json:"job_id"`
	EventType string    `json:"event_type"`
	OldStatus JobStatus `json:"old_status,omitempty"`
	NewStatus JobStatus `json:"new_status,omitempty"`
	WorkerID  string    `json:"worker_id,omitempty"`
	OccuredAt time.Time `json:"occurred_at"`
}
