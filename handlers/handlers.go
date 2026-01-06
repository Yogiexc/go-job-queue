package handlers

import (
	"encoding/json"
	"go-job-queue/models"
	"go-job-queue/queue"
	"net/http"
	"strings"
)

// Handler adalah HTTP handler untuk job queue
type Handler struct {
	queue *queue.JobQueue
}

// NewHandler membuat handler baru
func NewHandler(q *queue.JobQueue) *Handler {
	return &Handler{
		queue: q,
	}
}

// CreateJobRequest adalah request untuk membuat job
type CreateJobRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// CreateJobResponse adalah response setelah membuat job
type CreateJobResponse struct {
	JobID   string            `json:"job_id"`
	Status  models.JobStatus  `json:"status"`
	Message string            `json:"message"`
}

// ErrorResponse adalah response untuk error
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateJob handler untuk POST /jobs
func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Type == "" || req.Payload == "" {
		h.sendError(w, "Type and payload are required", http.StatusBadRequest)
		return
	}

	job := models.NewJob(req.Type, req.Payload)

	if err := h.queue.AddJob(job); err != nil {
		h.sendError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	resp := CreateJobResponse{
		JobID:   job.ID,
		Status:  job.Status,
		Message: "Job berhasil ditambahkan ke queue",
	}

	h.sendJSON(w, resp, http.StatusCreated)
}

// GetAllJobs handler untuk GET /jobs
func (h *Handler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs := h.queue.GetAllJobs()
	h.sendJSON(w, jobs, http.StatusOK)
}

// GetJob handler untuk GET /jobs/{id}
func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/jobs/")
	jobID := path

	if jobID == "" {
		h.sendError(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	job, exists := h.queue.GetJob(jobID)
	if !exists {
		h.sendError(w, "Job not found", http.StatusNotFound)
		return
	}

	h.sendJSON(w, job, http.StatusOK)
}

// GetLogs handler untuk GET /logs
func (h *Handler) GetLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logs := h.queue.GetLogs()
	h.sendJSON(w, logs, http.StatusOK)
}

// HealthCheck handler untuk GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "OK",
		"service": "go-job-queue",
	}
	h.sendJSON(w, response, http.StatusOK)
}

// sendJSON helper untuk mengirim JSON response
func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// sendError helper untuk mengirim error response
func (h *Handler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// ExportJobs handler untuk GET /export
func (h *Handler) ExportJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs := h.queue.GetAllJobs()
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=jobs-export.json")
	
	json.NewEncoder(w).Encode(jobs)
}

// GetJobsByStatus handler untuk GET /jobs/status/{status}
func (h *Handler) GetJobsByStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/jobs/status/")
	status := strings.ToUpper(path)

	if status == "" {
		h.sendError(w, "Status is required", http.StatusBadRequest)
		return
	}

	allJobs := h.queue.GetAllJobs()
	filteredJobs := make([]*models.Job, 0)

	for _, job := range allJobs {
		if string(job.Status) == status {
			filteredJobs = append(filteredJobs, job)
		}
	}

	h.sendJSON(w, filteredJobs, http.StatusOK)
}

// GetStats handler untuk GET /stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs := h.queue.GetAllJobs()
	
	stats := map[string]interface{}{
		"total_jobs": len(jobs),
		"pending":    0,
		"processing": 0,
		"done":       0,
		"failed":     0,
	}

	for _, job := range jobs {
		switch job.Status {
		case models.StatusPending:
			stats["pending"] = stats["pending"].(int) + 1
		case models.StatusProcessing:
			stats["processing"] = stats["processing"].(int) + 1
		case models.StatusDone:
			stats["done"] = stats["done"].(int) + 1
		case models.StatusFailed:
			stats["failed"] = stats["failed"].(int) + 1
		}
	}

	h.sendJSON(w, stats, http.StatusOK)
}

// DeleteJob handler untuk DELETE /jobs/{id}
func (h *Handler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/jobs/")
	jobID := path

	if jobID == "" {
		h.sendError(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	deleted := h.queue.DeleteJob(jobID)
	if !deleted {
		h.sendError(w, "Job not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}