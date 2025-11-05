package scheduler

import (
	"testing"
	"time"

	"github.com/athulya-anil/axon-scheduler/pkg/models"
)

func TestWorkerRegistryLifecycle(t *testing.T) {
	reg := NewWorkerRegistry()

	// Register a worker
	worker := &models.WorkerInfo{
		ID:       "worker-1",
		Capacity: 3,
	}
	reg.Register(worker)

	// Send heartbeat
	reg.Heartbeat("worker-1", []string{"job-1"})
	if len(reg.GetHealthy()) != 1 {
		t.Fatalf("expected 1 healthy worker, got %d", len(reg.GetHealthy()))
	}

	// Simulate worker timeout
	time.Sleep(2 * time.Second)
	go reg.MonitorWorkers(500*time.Millisecond, 1*time.Second)
	time.Sleep(2 * time.Second)

	healthy := reg.GetHealthy()
	if len(healthy) != 0 {
		t.Fatalf("expected 0 healthy workers after timeout, got %d", len(healthy))
	}
}

