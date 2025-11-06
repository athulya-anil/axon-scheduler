package tests

import (
	"testing"
	"time"

	"github.com/athulya-anil/axon-scheduler/pkg/models"
	"github.com/athulya-anil/axon-scheduler/pkg/queue"
	"github.com/athulya-anil/axon-scheduler/pkg/scheduler"
)

// TestJobLifecycle tests the complete job lifecycle from submission to completion
func TestJobLifecycle(t *testing.T) {
	// Create a scheduler instance
	s, err := scheduler.NewScheduler("test-scheduler", []string{"localhost:2379"})
	if err != nil {
		t.Skipf("Skipping test - etcd not available: %v", err)
	}
	defer s.Stop()

	// Create a test job
	job := &queue.Job{
		ID:         "test-job-1",
		Type:       "test",
		Payload:    []byte(`{"data": "test"}`),
		Priority:   1,
		Status:     queue.PENDING,
		CreatedAt:  time.Now(),
		MaxRetries: 3,
		Timeout:    30 * time.Second,
	}

	// Test 1: Add job to scheduler
	err = s.AddJob(job)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Test 2: Verify job is in queue
	retrievedJob := s.GetJob(job.ID)
	if retrievedJob == nil {
		t.Fatal("Job not found in queue after adding")
	}

	if retrievedJob.Status != queue.PENDING {
		t.Errorf("Expected job status PENDING, got %s", retrievedJob.Status)
	}

	// Test 3: List jobs
	jobs := s.ListJobs()
	if len(jobs) == 0 {
		t.Fatal("ListJobs returned empty list")
	}

	found := false
	for _, j := range jobs {
		if j.ID == job.ID {
			found = true
			break
		}
	}

	if !found {
		t.Error("Job not found in ListJobs result")
	}
}

// TestWorkerRegistration tests worker registration and heartbeat
func TestWorkerRegistration(t *testing.T) {
	// Create a scheduler instance
	s, err := scheduler.NewScheduler("test-scheduler", []string{"localhost:2379"})
	if err != nil {
		t.Skipf("Skipping test - etcd not available: %v", err)
	}
	defer s.Stop()

	// Create a test worker
	worker := &models.WorkerInfo{
		ID:            "test-worker-1",
		Address:       "localhost:50051",
		Capacity:      5,
		LastHeartbeat: time.Now(),
		CurrentJobs:   []string{},
		Healthy:       true,
	}

	// Test 1: Register worker
	err = s.RegisterWorker(worker)
	if err != nil {
		t.Fatalf("Failed to register worker: %v", err)
	}

	// Test 2: List workers
	workers := s.ListWorkers()
	if len(workers) == 0 {
		t.Fatal("ListWorkers returned empty list")
	}

	found := false
	for _, w := range workers {
		if w.ID == worker.ID {
			found = true
			if !w.Healthy {
				t.Error("Worker should be healthy after registration")
			}
			break
		}
	}

	if !found {
		t.Error("Worker not found in ListWorkers result")
	}

	// Test 3: Simulate heartbeat
	time.Sleep(100 * time.Millisecond)

	// Worker should still be healthy
	workers = s.ListWorkers()
	for _, w := range workers {
		if w.ID == worker.ID && !w.Healthy {
			t.Error("Worker became unhealthy unexpectedly")
		}
	}
}

// TestQueuePriorityOrdering tests that jobs are dequeued in priority order
func TestQueuePriorityOrdering(t *testing.T) {
	pq := queue.NewPriorityQueue()

	// Add jobs with different priorities
	highPriorityJob := &queue.Job{
		ID:        "job-high",
		Priority:  1,
		CreatedAt: time.Now(),
		Status:    queue.PENDING,
	}

	mediumPriorityJob := &queue.Job{
		ID:        "job-medium",
		Priority:  5,
		CreatedAt: time.Now(),
		Status:    queue.PENDING,
	}

	lowPriorityJob := &queue.Job{
		ID:        "job-low",
		Priority:  10,
		CreatedAt: time.Now(),
		Status:    queue.PENDING,
	}

	// Push in reverse order
	pq.Push(lowPriorityJob)
	pq.Push(mediumPriorityJob)
	pq.Push(highPriorityJob)

	// Verify they come out in priority order (lower number = higher priority)
	job1 := pq.Pop()
	if job1.ID != "job-high" {
		t.Errorf("Expected job-high first, got %s", job1.ID)
	}

	job2 := pq.Pop()
	if job2.ID != "job-medium" {
		t.Errorf("Expected job-medium second, got %s", job2.ID)
	}

	job3 := pq.Pop()
	if job3.ID != "job-low" {
		t.Errorf("Expected job-low third, got %s", job3.ID)
	}
}

// TestJobStatusTransitions tests job status changes
func TestJobStatusTransitions(t *testing.T) {
	pq := queue.NewPriorityQueue()

	job := &queue.Job{
		ID:        "status-test",
		Priority:  1,
		CreatedAt: time.Now(),
		Status:    queue.PENDING,
	}

	pq.Push(job)

	// Test 1: Initial status should be PENDING
	retrievedJob := pq.Get(job.ID)
	if retrievedJob.Status != queue.PENDING {
		t.Errorf("Expected PENDING, got %s", retrievedJob.Status)
	}

	// Test 2: Update to RUNNING
	pq.UpdateStatus(job.ID, queue.RUNNING)
	retrievedJob = pq.Get(job.ID)
	if retrievedJob.Status != queue.RUNNING {
		t.Errorf("Expected RUNNING, got %s", retrievedJob.Status)
	}

	// Test 3: Update to COMPLETED
	pq.UpdateStatus(job.ID, queue.COMPLETED)
	retrievedJob = pq.Get(job.ID)
	if retrievedJob.Status != queue.COMPLETED {
		t.Errorf("Expected COMPLETED, got %s", retrievedJob.Status)
	}
}

// BenchmarkJobSubmission benchmarks job submission performance
func BenchmarkJobSubmission(b *testing.B) {
	s, err := scheduler.NewScheduler("bench-scheduler", []string{"localhost:2379"})
	if err != nil {
		b.Skipf("Skipping benchmark - etcd not available: %v", err)
	}
	defer s.Stop()

	// Note: Start() is blocking and handles its own ports, so we skip it for benchmarks

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := &queue.Job{
			ID:         string(rune(i)),
			Type:       "benchmark",
			Payload:    []byte("test"),
			Priority:   i % 10,
			Status:     queue.PENDING,
			CreatedAt:  time.Now(),
			MaxRetries: 3,
		}
		s.AddJob(job)
	}
}
