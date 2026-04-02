package handlers

import (
	"encoding/json"
	"net/http"

	"owen/queueflow/internal/db"
	"owen/queueflow/internal/kafka"
	"owen/queueflow/internal/models"
	"owen/queueflow/internal/worker/jobs"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type JobsHandler struct {
	db       *db.Queries
	producer *kafka.Producer
	registry *jobs.Registry
}

func NewJobsHandler(db *db.Queries, producer *kafka.Producer, registry *jobs.Registry) *JobsHandler {
	return &JobsHandler{db: db, producer: producer, registry: registry}
}

type submitJobRequest struct {
	Type        string          `json:"type"`
	Priority    string          `json:"priority"`
	Payload     json.RawMessage `json:"payload"`
	MaxAttempts int             `json:"max_attempts"`
	CallbackURL string          `json:"callback_url"`
	Namespace   string          `json:"namespace"`
}

func (h *JobsHandler) SubmitJob(w http.ResponseWriter, r *http.Request) {
	var req submitJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}

	if _, ok := h.registry.Get(req.Type); !ok {
		writeError(w, http.StatusBadRequest, "unknown job type: "+req.Type)
		return
	}

	if req.Priority == "" {
		req.Priority = string(models.PriorityNormal)
	}
	if !models.ValidPriorities[models.JobPriority(req.Priority)] {
		writeError(w, http.StatusBadRequest, "invalid priority, must be high/normal/low")
		return
	}

	if req.MaxAttempts <= 0 {
		req.MaxAttempts = 3
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if req.Payload == nil {
		req.Payload = json.RawMessage("{}")
	}

	jobID := uuid.New().String()

	job := &models.Job{
		ID:          jobID,
		Type:        req.Type,
		Priority:    models.JobPriority(req.Priority),
		Status:      models.JobStatusQueued,
		Payload:     req.Payload,
		MaxAttempts: req.MaxAttempts,
		CallbackURL: req.CallbackURL,
		Namespace:   req.Namespace,
	}

	if err := h.db.InsertJob(r.Context(), job); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to insert job")
		return
	}

	msg := models.KafkaMessage{
		JobID:   jobID,
		Type:    req.Type,
		Attempt: 0,
		Payload: req.Payload,
	}
	msgBytes, _ := json.Marshal(msg)

	topic := kafka.TopicForPriority(req.Priority)
	if err := h.producer.Produce(r.Context(), topic, jobID, msgBytes); err != nil {
		h.db.UpdateJobFailed(r.Context(), jobID, "kafka produce failed: "+err.Error())
		writeError(w, http.StatusInternalServerError, "failed to enqueue job")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"job_id": jobID,
		"status": models.JobStatusQueued,
	})
}

func (h *JobsHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	job, err := h.db.GetJob(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	events, _ := h.db.GetJobEvents(r.Context(), id)

	writeJSON(w, http.StatusOK, map[string]any{
		"id":            job.ID,
		"type":          job.Type,
		"status":        job.Status,
		"priority":      job.Priority,
		"namespace":     job.Namespace,
		"attempts":      job.Attempts,
		"max_attempts":  job.MaxAttempts,
		"worker_id":     job.WorkerID,
		"payload":       job.Payload,
		"result":        job.Result,
		"error_message": job.ErrorMessage,
		"events":        events,
		"created_at":    job.CreatedAt,
		"started_at":    job.StartedAt,
		"completed_at":  job.CompletedAt,
	})
}

func (h *JobsHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 50
	if l := q.Get("limit"); l != "" {
		if n := parseIntDefault(l, 50); n > 0 && n <= 200 {
			limit = n
		}
	}

	params := db.ListJobsParams{
		Status:    q.Get("status"),
		Priority:  q.Get("priority"),
		Namespace: q.Get("namespace"),
		JobType:   q.Get("type"),
		Limit:     limit,
		Cursor:    q.Get("cursor"),
	}

	jobsList, err := h.db.ListJobs(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list jobs")
		return
	}

	if jobsList == nil {
		jobsList = []models.Job{}
	}

	var nextCursor string
	if len(jobsList) == limit {
		nextCursor = jobsList[len(jobsList)-1].ID
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"jobs":        jobsList,
		"next_cursor": nextCursor,
	})
}

func (h *JobsHandler) RetryJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	job, err := h.db.GetJob(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	if job.Status != models.JobStatusFailed && job.Status != models.JobStatusDead {
		writeError(w, http.StatusBadRequest, "job must be in FAILED or DEAD status to retry")
		return
	}

	if err := h.db.ResetJobForRetry(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reset job")
		return
	}

	h.db.InsertJobEvent(r.Context(), &models.JobEvent{
		JobID:     id,
		EventType: "retry",
		OldStatus: job.Status,
		NewStatus: models.JobStatusQueued,
	})

	msg := models.KafkaMessage{
		JobID:   id,
		Type:    job.Type,
		Attempt: 0,
		Payload: job.Payload,
	}
	msgBytes, _ := json.Marshal(msg)
	topic := kafka.TopicForPriority(string(job.Priority))
	h.producer.Produce(r.Context(), topic, id, msgBytes)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"job_id":         id,
		"status":         models.JobStatusQueued,
		"attempts_reset": true,
	})
}

func parseIntDefault(s string, def int) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return def
		}
	}
	return n
}
