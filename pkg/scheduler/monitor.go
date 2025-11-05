package scheduler

import (
	"log"
	"sync"
	"time"

	"github.com/athulya-anil/axon-scheduler/pkg/models"
)

// WorkerRegistry keeps track of all known workers and their health status
type WorkerRegistry struct {
	mu      sync.RWMutex
	workers map[string]*models.WorkerInfo
}

// NewWorkerRegistry creates a new registry
func NewWorkerRegistry() *WorkerRegistry {
	return &WorkerRegistry{
		workers: make(map[string]*models.WorkerInfo),
	}
}

// Register adds a worker to the registry
func (r *WorkerRegistry) Register(worker *models.WorkerInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.workers[worker.ID] = worker
	log.Printf("🧩 Worker registered: %s (capacity: %d)", worker.ID, worker.Capacity)
}

// Heartbeat updates the last-seen time for a worker
func (r *WorkerRegistry) Heartbeat(workerID string, activeJobs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if w, ok := r.workers[workerID]; ok {
		w.LastHeartbeat = time.Now()
		w.CurrentJobs = activeJobs
		w.Healthy = true
	} else {
		log.Printf("⚠️ Heartbeat received from unregistered worker: %s", workerID)
	}
}

// GetHealthy returns all currently healthy workers
func (r *WorkerRegistry) GetHealthy() []*models.WorkerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var healthy []*models.WorkerInfo
	for _, w := range r.workers {
		if w.Healthy {
			healthy = append(healthy, w)
		}
	}
	return healthy
}

// MonitorWorkers continuously checks for dead workers
func (r *WorkerRegistry) MonitorWorkers(interval time.Duration, timeout time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		r.mu.Lock()
		for id, w := range r.workers {
			if now.Sub(w.LastHeartbeat) > timeout {
				if w.Healthy {
					w.Healthy = false
					log.Printf("💀 Worker %s marked DEAD (last heartbeat %v ago)", id, now.Sub(w.LastHeartbeat))
				}
			}
		}
		r.mu.Unlock()
	}
}
