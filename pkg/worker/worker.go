package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/athulya-anil/axon-scheduler/proto/workerpb"
	"google.golang.org/grpc"
)

// Worker represents a worker node that executes jobs
type Worker struct {
	workerpb.UnimplementedWorkerServiceServer

	ID            string
	Capacity      int // Max concurrent jobs
	schedulerAddr string
	grpcServer    *grpc.Server

	mu         sync.RWMutex
	activeJobs map[string]bool // Job IDs currently being executed
	jobCount   int

	ctx    context.Context
	cancel context.CancelFunc
}

// NewWorker creates a new worker instance
func NewWorker(id string, capacity int, schedulerAddr string) *Worker {
	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		ID:            id,
		Capacity:      capacity,
		schedulerAddr: schedulerAddr,
		activeJobs:    make(map[string]bool),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins the worker's operation
func (w *Worker) Start(port string) error {
	// Start gRPC server in a goroutine
	go func() {
		if err := w.startGRPCServer(port); err != nil {
			log.Printf("[ERROR] gRPC server failed: %v", err)
		}
	}()

	// Start heartbeat sender
	go w.sendHeartbeats()

	log.Printf("[WORKER] Worker %s started on port %s (capacity: %d)", w.ID, port, w.Capacity)

	// Keep worker running
	<-w.ctx.Done()
	return nil
}

// Stop gracefully shuts down the worker
func (w *Worker) Stop() {
	log.Printf("[STOP] Stopping worker %s...", w.ID)
	w.cancel()

	if w.grpcServer != nil {
		w.grpcServer.GracefulStop()
	}
}

// startGRPCServer starts the gRPC server for receiving jobs
// Note: This is not used in Phase 1, actual server started in cmd/worker/main.go
func (w *Worker) startGRPCServer(port string) error {
	// Placeholder - actual implementation in cmd/worker/main.go
	return nil
}

// SubmitJob handles incoming job assignments from the scheduler (gRPC method)
func (w *Worker) SubmitJob(ctx context.Context, req *workerpb.JobRequest) (*workerpb.JobResponse, error) {
	jobID := req.GetJobId()

	log.Printf("[JOB] Worker %s received job %s", w.ID, jobID)

	// Check if we have capacity
	w.mu.Lock()
	if w.jobCount >= w.Capacity {
		w.mu.Unlock()
		return &workerpb.JobResponse{
			Success: false,
			Message: "worker at capacity",
		}, nil
	}

	// Mark job as active
	w.activeJobs[jobID] = true
	w.jobCount++
	w.mu.Unlock()

	// Execute job asynchronously
	go func() {
		if err := w.executeJob(jobID, req.GetPayload()); err != nil {
			log.Printf("[ERROR] Job %s failed: %v", jobID, err)
		}

		// Remove from active jobs
		w.mu.Lock()
		delete(w.activeJobs, jobID)
		w.jobCount--
		w.mu.Unlock()
	}()

	return &workerpb.JobResponse{
		Success: true,
		Message: "job accepted",
	}, nil
}

// SendHeartbeat handles heartbeat requests (gRPC method - implemented by scheduler calling us)
func (w *Worker) SendHeartbeat(ctx context.Context, req *workerpb.HeartbeatRequest) (*workerpb.HeartbeatResponse, error) {
	return &workerpb.HeartbeatResponse{
		Acknowledged: true,
	}, nil
}

// GetActiveJobs returns the list of currently running job IDs
func (w *Worker) GetActiveJobs() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	jobs := make([]string, 0, len(w.activeJobs))
	for jobID := range w.activeJobs {
		jobs = append(jobs, jobID)
	}
	return jobs
}

// GetJobCount returns the current number of active jobs
func (w *Worker) GetJobCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.jobCount
}

// sendHeartbeats periodically sends heartbeat to scheduler
func (w *Worker) sendHeartbeats() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			// TODO: Send heartbeat to scheduler via gRPC
			// For now, just log
			activeJobs := w.GetActiveJobs()
			log.Printf("[HB] Worker %s heartbeat: %d/%d jobs active %v",
				w.ID, w.GetJobCount(), w.Capacity, activeJobs)
		}
	}
}
