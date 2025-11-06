package scheduler

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/athulya-anil/axon-scheduler/pkg/models"
	"github.com/athulya-anil/axon-scheduler/pkg/queue"
	"github.com/athulya-anil/axon-scheduler/proto/workerpb"
)

// findLeastLoadedWorker selects the worker with the most available capacity
func (s *Scheduler) findLeastLoadedWorker() *models.WorkerInfo {
	healthyWorkers := s.workers.GetHealthy()

	if len(healthyWorkers) == 0 {
		return nil
	}

	var bestWorker *models.WorkerInfo
	minLoad := math.MaxInt

	for _, worker := range healthyWorkers {
		// Calculate load (current jobs vs capacity)
		load := len(worker.CurrentJobs)

		// Check if worker has capacity
		if load >= worker.Capacity {
			continue
		}

		// Select worker with least load
		if load < minLoad {
			minLoad = load
			bestWorker = worker
		}
	}

	return bestWorker
}

// assignJobToWorker sends a job to a worker via gRPC
func (s *Scheduler) assignJobToWorker(job *queue.Job, worker *models.WorkerInfo) error {
	// Get worker client
	s.connMu.RLock()
	client, exists := s.workerClients[worker.ID]
	s.connMu.RUnlock()

	if !exists {
		return fmt.Errorf("no gRPC client for worker %s", worker.ID)
	}

	// Create job request
	req := &workerpb.JobRequest{
		JobId:    job.ID,
		WorkerId: worker.ID,
		Payload:  string(job.Payload),
	}

	// Send job to worker with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.SubmitJob(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC call failed: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("worker rejected job: %s", resp.GetMessage())
	}

	log.Printf("✅ Successfully assigned job %s to worker %s", job.ID, worker.ID)

	// Update worker's active jobs in registry
	s.workers.mu.Lock()
	if w, ok := s.workers.workers[worker.ID]; ok {
		w.CurrentJobs = append(w.CurrentJobs, job.ID)
	}
	s.workers.mu.Unlock()

	return nil
}

// reassignJobsFromDeadWorker reassigns all jobs from a failed worker
func (s *Scheduler) reassignJobsFromDeadWorker(workerID string, jobIDs []string) {
	log.Printf("♻️ Reassigning %d jobs from dead worker %s", len(jobIDs), workerID)

	for _, jobID := range jobIDs {
		job := s.queue.Get(jobID)
		if job == nil {
			log.Printf("⚠️ Job %s not found in queue", jobID)
			continue
		}

		// Reset job status
		job.Status = queue.PENDING
		job.StartedAt = nil
		job.WorkerID = ""
		job.Retries++

		// Check if max retries exceeded
		if job.Retries > job.MaxRetries {
			log.Printf("❌ Job %s exceeded max retries (%d)", jobID, job.MaxRetries)
			job.Status = queue.FAILED
			continue
		}

		// Re-add to queue
		s.queue.Push(job)
		log.Printf("♻️ Job %s re-queued (retry %d/%d)", jobID, job.Retries, job.MaxRetries)
	}
}
