package models

import (
	"time"
)

// JobStatus adalah status dari job
type JobStatus string

const (
	StatusPending    JobStatus = "PENDING"
	StatusProcessing JobStatus = "PROCESSING"
	StatusDone       JobStatus = "DONE"
	StatusFailed     JobStatus = "FAILED"
)

// Job adalah struktur data job
type Job struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Payload     string    `json:"payload"`
	Status      JobStatus `json:"status"`
	RetryCount  int       `json:"retry_count"`
	MaxRetries  int       `json:"max_retries"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
}

// NewJob membuat job baru
func NewJob(jobType, payload string) *Job {
	now := time.Now()
	return &Job{
		ID:         generateID(),
		Type:       jobType,
		Payload:    payload,
		Status:     StatusPending,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// generateID membuat ID unik sederhana
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}