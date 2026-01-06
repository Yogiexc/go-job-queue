package workers

import (
	"fmt"
	"go-job-queue/models"
	"go-job-queue/queue"
	"math/rand"
	"sync"
	"time"
)

// WorkerPool adalah pool dari worker
type WorkerPool struct {
	workerCount int
	queue       *queue.JobQueue
	wg          sync.WaitGroup
	quit        chan bool
}

// NewWorkerPool membuat worker pool baru
func NewWorkerPool(workerCount int, q *queue.JobQueue) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		queue:       q,
		quit:        make(chan bool),
	}
}

// Start memulai semua worker
func (wp *WorkerPool) Start() {
	fmt.Printf("🚀 Memulai %d worker...\n", wp.workerCount)
	
	for i := 1; i <= wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker adalah goroutine yang memproses job
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	
	fmt.Printf("✅ Worker #%d siap!\n", id)

	for {
		select {
		case job, ok := <-wp.queue.GetJobChannel():
			if !ok {
				fmt.Printf("🛑 Worker #%d berhenti (channel closed)\n", id)
				return
			}
			
			wp.processJob(id, job)

		case <-wp.quit:
			fmt.Printf("🛑 Worker #%d berhenti (graceful shutdown)\n", id)
			return
		}
	}
}

// processJob memproses job
func (wp *WorkerPool) processJob(workerID int, job *models.Job) {
	fmt.Printf("⚙️  Worker #%d memproses Job %s (%s)\n", workerID, job.ID, job.Type)
	
	wp.queue.UpdateJobStatus(job.ID, models.StatusProcessing, "")

	success := wp.simulateJobExecution(job)

	if success {
		wp.queue.UpdateJobStatus(job.ID, models.StatusDone, "")
		fmt.Printf("✅ Worker #%d selesai Job %s\n", workerID, job.ID)
	} else {
		fmt.Printf("❌ Worker #%d gagal Job %s\n", workerID, job.ID)
		
		retried := wp.queue.RetryJob(job.ID)
		if !retried {
			fmt.Printf("💀 Job %s gagal permanen (max retries exceeded)\n", job.ID)
		}
	}
}

// simulateJobExecution simulasi eksekusi job
func (wp *WorkerPool) simulateJobExecution(job *models.Job) bool {
	switch job.Type {
	case "email":
		duration := time.Duration(2+rand.Intn(3)) * time.Second
		time.Sleep(duration)
		return rand.Float32() > 0.1

	case "notification":
		duration := time.Duration(1+rand.Intn(2)) * time.Second
		time.Sleep(duration)
		return rand.Float32() > 0.05

	case "report":
		duration := time.Duration(3+rand.Intn(3)) * time.Second
		time.Sleep(duration)
		return rand.Float32() > 0.15

	default:
		duration := time.Duration(1+rand.Intn(3)) * time.Second
		time.Sleep(duration)
		return rand.Float32() > 0.1
	}
}

// Stop menghentikan semua worker dengan graceful shutdown
func (wp *WorkerPool) Stop() {
	fmt.Println("🛑 Menghentikan worker pool...")
	
	close(wp.quit)
	
	wp.wg.Wait()
	
	fmt.Println("✅ Semua worker telah berhenti")
}