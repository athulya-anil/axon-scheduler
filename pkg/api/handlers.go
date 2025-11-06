package api

import (
	"net/http"
	"time"

	"github.com/athulya-anil/axon-scheduler/pkg/models"
	"github.com/athulya-anil/axon-scheduler/pkg/queue"
	"github.com/athulya-anil/axon-scheduler/pkg/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// API wraps the scheduler and provides HTTP handlers
type API struct {
	scheduler *scheduler.Scheduler
}

// NewAPI creates a new API instance
func NewAPI(s *scheduler.Scheduler) *API {
	return &API{
		scheduler: s,
	}
}

// SetupRoutes configures all API routes
func (a *API) SetupRoutes(router *gin.Engine) {
	// Job endpoints
	router.POST("/jobs", a.submitJob)
	router.GET("/jobs", a.listJobs)
	router.GET("/jobs/:id", a.getJob)

	// Worker endpoints
	router.POST("/workers/register", a.registerWorker)
	router.GET("/workers", a.listWorkers)

	// Status endpoints
	router.GET("/status", a.getStatus)
	router.GET("/health", a.healthCheck)
}

// JobSubmitRequest represents the payload for submitting a job
type JobSubmitRequest struct {
	Type       string            `json:"type" binding:"required"`
	Payload    map[string]string `json:"payload"`
	Priority   int               `json:"priority"`
	MaxRetries int               `json:"max_retries"`
	DependsOn  []string          `json:"depends_on"`
}

// WorkerRegisterRequest represents the payload for worker registration
type WorkerRegisterRequest struct {
	ID       string `json:"id" binding:"required"`
	Address  string `json:"address" binding:"required"`
	Capacity int    `json:"capacity" binding:"required,min=1"`
}

// submitJob handles POST /jobs
func (a *API) submitJob(c *gin.Context) {
	var req JobSubmitRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create job
	job := &queue.Job{
		ID:         uuid.New().String(),
		Type:       req.Type,
		Payload:    []byte(jsonMarshal(req.Payload)),
		Priority:   req.Priority,
		Status:     queue.PENDING,
		CreatedAt:  time.Now(),
		MaxRetries: req.MaxRetries,
		DependsOn:  req.DependsOn,
	}

	// Set default max retries if not specified
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}

	// Submit to scheduler
	if err := a.scheduler.AddJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"job_id":  job.ID,
		"status":  string(job.Status),
		"message": "job submitted successfully",
	})
}

// listJobs handles GET /jobs
func (a *API) listJobs(c *gin.Context) {
	jobs := a.scheduler.ListJobs()

	c.JSON(http.StatusOK, gin.H{
		"count": len(jobs),
		"jobs":  jobs,
	})
}

// getJob handles GET /jobs/:id
func (a *API) getJob(c *gin.Context) {
	jobID := c.Param("id")

	job := a.scheduler.GetJob(jobID)
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// registerWorker handles POST /workers/register
func (a *API) registerWorker(c *gin.Context) {
	var req WorkerRegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create worker info
	workerInfo := &models.WorkerInfo{
		ID:            req.ID,
		Address:       req.Address,
		Capacity:      req.Capacity,
		LastHeartbeat: time.Now(),
		CurrentJobs:   []string{},
		Healthy:       true,
	}

	// Register with scheduler
	if err := a.scheduler.RegisterWorker(workerInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"worker_id": workerInfo.ID,
		"message":   "worker registered successfully",
	})
}

// listWorkers handles GET /workers
func (a *API) listWorkers(c *gin.Context) {
	workers := a.scheduler.ListWorkers()

	c.JSON(http.StatusOK, gin.H{
		"count":   len(workers),
		"workers": workers,
	})
}

// getStatus handles GET /status
func (a *API) getStatus(c *gin.Context) {
	workers := a.scheduler.ListWorkers()
	healthyCount := 0
	for _, w := range workers {
		if w.Healthy {
			healthyCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"node_id":         a.scheduler.NodeID,
		"is_leader":       a.scheduler.IsLeader,
		"queue_length":    a.scheduler.GetQueueLength(),
		"workers_total":   len(workers),
		"workers_healthy": healthyCount,
		"timestamp":       time.Now(),
	})
}

// healthCheck handles GET /health
func (a *API) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"node":   a.scheduler.NodeID,
	})
}

// Helper function to marshal map to JSON string
func jsonMarshal(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}

	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ","
		}
		result += `"` + k + `":"` + v + `"`
		first = false
	}
	result += "}"

	return result
}
