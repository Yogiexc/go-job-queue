package main

import (
	"context"
	"fmt"
	"go-job-queue/handlers"
	"go-job-queue/queue"
	"go-job-queue/workers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	PORT         = ":8080"
	WORKER_COUNT = 3
	BUFFER_SIZE  = 100
)

func main() {
	fmt.Println("=================================")
	fmt.Println("🚀 GO JOB QUEUE SYSTEM")
	fmt.Println("=================================")

	jobQueue := queue.NewJobQueue(BUFFER_SIZE)
	fmt.Printf("✅ Job Queue initialized (buffer: %d)\n", BUFFER_SIZE)

	workerPool := workers.NewWorkerPool(WORKER_COUNT, jobQueue)
	workerPool.Start()

	handler := handlers.NewHandler(jobQueue)

	mux := http.NewServeMux()
	
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to Go Job Queue API\n")
		fmt.Fprintf(w, "\nAvailable endpoints:\n")
		fmt.Fprintf(w, "- POST   /jobs       - Create new job\n")
		fmt.Fprintf(w, "- GET    /jobs       - Get all jobs\n")
		fmt.Fprintf(w, "- GET    /jobs/{id}  - Get job by ID\n")
		fmt.Fprintf(w, "- GET    /logs       - Get system logs\n")
		fmt.Fprintf(w, "- GET    /health     - Health check\n")
	})
	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateJob(w, r)
		} else if r.Method == http.MethodGet {
			handler.GetAllJobs(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/jobs/", handler.GetJob)
	mux.HandleFunc("/logs", handler.GetLogs)
	mux.HandleFunc("/health", handler.HealthCheck)

	server := &http.Server{
		Addr:    PORT,
		Handler: mux,
	}

	go func() {
		fmt.Printf("🌐 Server running on http://localhost%s\n", PORT)
		fmt.Println("=================================")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	fmt.Println("\n=================================")
	fmt.Println("🛑 Shutting down gracefully...")
	fmt.Println("=================================")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	fmt.Println("✅ HTTP Server stopped")

	jobQueue.Close()
	fmt.Println("✅ Job Queue closed")

	workerPool.Stop()
	fmt.Println("✅ All workers stopped")

	fmt.Println("=================================")
	fmt.Println("👋 Server exited cleanly")
	fmt.Println("=================================")
}