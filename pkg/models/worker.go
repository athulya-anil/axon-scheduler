package models

import "time"

// WorkerInfo holds the state and health of a worker node in the cluster.
type WorkerInfo struct {
	ID            string     `json:"id"`              // Unique worker ID
	Address       string     `json:"address"`         // Worker gRPC address (host:port)
	Capacity      int        `json:"capacity"`        // Maximum concurrent jobs
	LastHeartbeat time.Time  `json:"last_heartbeat"`  // Last heartbeat timestamp
	CurrentJobs   []string   `json:"current_jobs"`    // IDs of currently running jobs
	Healthy       bool       `json:"healthy"`         // True if the worker responded recently
}

