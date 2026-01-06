package queue

import (
	"fmt"
	"go-job-queue/models"
	"sync"
	"time"
)

// JobQueue adalah struktur untuk mengelola job queue
type JobQueue struct {
	jobs      map[string]*models.Job
	jobChan   chan *models.Job
	mu        sync.RWMutex
	logs      []string
	logsMu    sync.Mutex
}

// NewJobQueue membuat instance baru JobQueue
func NewJobQueue(bufferSize int) *JobQueue {
	return &JobQueue{
		jobs:    make(map[string]*models.Job),
		jobChan: make(chan *models.Job, bufferSize),
		logs:    make([]string, 0),
	}
}

// AddJob menambahkan job baru ke queue
func (q *JobQueue) AddJob(job *models.Job) error {
	q.mu.Lock()
	q.jobs[job.ID] = job
	q.mu.Unlock()

	q.addLog(fmt.Sprintf("[ENQUEUE] Job %s (%s) ditambahkan ke queue", job.ID, job.Type))

	select {
	case q.jobChan <- job:
		return nil
	default:
		return fmt.Errorf("queue penuh, job tidak bisa ditambahkan")
	}
}

// GetJobChannel mengembalikan channel untuk worker
func (q *JobQueue) GetJobChannel() <-chan *models.Job {
	return q.jobChan
}

// UpdateJobStatus mengupdate status job
func (q *JobQueue) UpdateJobStatus(jobID string, status models.JobStatus, errorMsg string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if job, exists := q.jobs[jobID]; exists {
		job.Status = status
		job.UpdatedAt = time.Now()
		if errorMsg != "" {
			job.ErrorMsg = errorMsg
		}

		logMsg := fmt.Sprintf("[UPDATE] Job %s status: %s", jobID, status)
		if errorMsg != "" {
			logMsg += fmt.Sprintf(" (error: %s)", errorMsg)
		}
		q.addLog(logMsg)
	}
}

// RetryJob mencoba ulang job yang gagal
func (q *JobQueue) RetryJob(jobID string) bool {
	q.mu.Lock()
	job, exists := q.jobs[jobID]
	q.mu.Unlock()

	if !exists {
		return false
	}

	job.RetryCount++
	if job.RetryCount > job.MaxRetries {
		q.addLog(fmt.Sprintf("[RETRY] Job %s gagal setelah %d kali retry", jobID, job.MaxRetries))
		q.UpdateJobStatus(jobID, models.StatusFailed, "Max retries exceeded")
		return false
	}

	q.addLog(fmt.Sprintf("[RETRY] Job %s retry ke-%d", jobID, job.RetryCount))
	
	job.Status = models.StatusPending
	job.UpdatedAt = time.Now()
	
	select {
	case q.jobChan <- job:
		return true
	default:
		return false
	}
}

// GetAllJobs mengembalikan semua job
func (q *JobQueue) GetAllJobs() []*models.Job {
	q.mu.RLock()
	defer q.mu.RUnlock()

	jobs := make([]*models.Job, 0, len(q.jobs))
	for _, job := range q.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// GetJob mengembalikan job berdasarkan ID
func (q *JobQueue) GetJob(jobID string) (*models.Job, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	job, exists := q.jobs[jobID]
	return job, exists
}

// GetLogs mengembalikan semua log
func (q *JobQueue) GetLogs() []string {
	q.logsMu.Lock()
	defer q.logsMu.Unlock()

	logsCopy := make([]string, len(q.logs))
	copy(logsCopy, q.logs)
	return logsCopy
}

// addLog menambahkan log baru
func (q *JobQueue) addLog(message string) {
	q.logsMu.Lock()
	defer q.logsMu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	q.logs = append(q.logs, logEntry)
	
	fmt.Println(logEntry)
}

// Close menutup channel job
func (q *JobQueue) Close() {
	close(q.jobChan)
	q.addLog("[SYSTEM] Queue ditutup")
}