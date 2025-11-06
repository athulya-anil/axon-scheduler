package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/athulya-anil/axon-scheduler/pkg/models"
	"github.com/athulya-anil/axon-scheduler/pkg/queue"
	"github.com/athulya-anil/axon-scheduler/proto/workerpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Scheduler manages job distribution and worker coordination
type Scheduler struct {
	workerpb.UnimplementedWorkerServiceServer

	NodeID   string
	IsLeader bool

	queue    *queue.PriorityQueue
	workers  *WorkerRegistry
	etcdCli  *clientv3.Client
	session  *concurrency.Session
	election *concurrency.Election

	// gRPC connections to workers
	workerConns   map[string]*grpc.ClientConn
	workerClients map[string]workerpb.WorkerServiceClient
	connMu        sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

// NewScheduler creates a new scheduler instance
func NewScheduler(nodeID string, etcdEndpoints []string) (*Scheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Connect to etcd
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to etcd: %w", err)
	}

	// Create session for leader election
	session, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
	if err != nil {
		cli.Close()
		cancel()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	election := concurrency.NewElection(session, "/axon-leader-election")

	s := &Scheduler{
		NodeID:        nodeID,
		queue:         queue.NewPriorityQueue(),
		workers:       NewWorkerRegistry(),
		etcdCli:       cli,
		session:       session,
		election:      election,
		workerConns:   make(map[string]*grpc.ClientConn),
		workerClients: make(map[string]workerpb.WorkerServiceClient),
		ctx:           ctx,
		cancel:        cancel,
	}

	return s, nil
}

// Start begins the scheduler's operation
func (s *Scheduler) Start() error {
	log.Printf("🚀 Starting scheduler %s...", s.NodeID)

	// Start leader election
	go s.runLeaderElection()

	// Start worker health monitoring
	go s.workers.MonitorWorkers(10*time.Second, 15*time.Second)

	// Wait for context cancellation
	<-s.ctx.Done()
	return nil
}

// Stop gracefully shuts down the scheduler
func (s *Scheduler) Stop() {
	log.Printf("🛑 Stopping scheduler %s...", s.NodeID)

	s.cancel()

	// Close worker connections
	s.connMu.Lock()
	for _, conn := range s.workerConns {
		conn.Close()
	}
	s.connMu.Unlock()

	// Resign from leadership
	if s.IsLeader {
		s.election.Resign(context.Background())
	}

	s.session.Close()
	s.etcdCli.Close()
}

// runLeaderElection participates in leader election
func (s *Scheduler) runLeaderElection() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// Try to become leader
			if err := s.election.Campaign(s.ctx, s.NodeID); err != nil {
				log.Printf("⚠️ Election campaign failed: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			log.Printf("🏆 Node %s has become the LEADER!", s.NodeID)
			s.IsLeader = true

			// Run as leader
			s.runAsLeader()

			// If we get here, we lost leadership
			s.IsLeader = false
			log.Printf("⚠️ Node %s lost leadership", s.NodeID)
		}
	}
}

// runAsLeader executes leader responsibilities
func (s *Scheduler) runAsLeader() {
	// Start job assignment loop
	assignCtx, assignCancel := context.WithCancel(s.ctx)
	defer assignCancel()

	go s.assignJobsLoop(assignCtx)

	// Wait for session to expire or context to be cancelled
	select {
	case <-s.session.Done():
		log.Println("⚠️ Leadership session expired")
	case <-s.ctx.Done():
		log.Println("🛑 Scheduler stopping")
	}
}

// assignJobsLoop continuously assigns jobs to workers
func (s *Scheduler) assignJobsLoop(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tryAssignJob()
		}
	}
}

// tryAssignJob attempts to assign one job from the queue
func (s *Scheduler) tryAssignJob() {
	// Peek at next job
	job := s.queue.Peek()
	if job == nil {
		return // Queue empty
	}

	// Check dependencies
	if !s.dependenciesSatisfied(job) {
		return // Dependencies not ready
	}

	// Find available worker
	worker := s.findLeastLoadedWorker()
	if worker == nil {
		return // No workers available
	}

	// Pop job from queue
	job = s.queue.Pop()
	if job == nil {
		return // Race condition, another goroutine took it
	}

	// Assign to worker
	if err := s.assignJobToWorker(job, worker); err != nil {
		log.Printf("❌ Failed to assign job %s to worker %s: %v", job.ID, worker.ID, err)

		// Put job back in queue
		job.Status = queue.PENDING
		s.queue.Push(job)
		return
	}

	// Update job status
	now := time.Now()
	job.Status = queue.RUNNING
	job.StartedAt = &now
	job.WorkerID = worker.ID

	log.Printf("✅ Assigned job %s to worker %s", job.ID, worker.ID)
}

// dependenciesSatisfied checks if all job dependencies are completed
func (s *Scheduler) dependenciesSatisfied(job *queue.Job) bool {
	for _, depID := range job.DependsOn {
		depJob := s.queue.Get(depID)
		if depJob == nil || depJob.Status != queue.COMPLETED {
			return false
		}
	}
	return true
}

// RegisterWorker adds a new worker to the pool (called via REST API)
func (s *Scheduler) RegisterWorker(workerInfo *models.WorkerInfo) error {
	log.Printf("📝 Registering worker %s at %s", workerInfo.ID, workerInfo.Address)

	// Add to registry
	s.workers.Register(workerInfo)

	// Create gRPC connection to worker
	// Note: Address should be in format "host:port"
	conn, err := grpc.NewClient(
		workerInfo.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to worker: %w", err)
	}

	s.connMu.Lock()
	s.workerConns[workerInfo.ID] = conn
	s.workerClients[workerInfo.ID] = workerpb.NewWorkerServiceClient(conn)
	s.connMu.Unlock()

	return nil
}

// AddJob adds a job to the queue (called via REST API)
func (s *Scheduler) AddJob(job *queue.Job) error {
	log.Printf("📥 Received job %s (priority: %d, type: %s)", job.ID, job.Priority, job.Type)

	job.Status = queue.PENDING
	job.CreatedAt = time.Now()

	s.queue.Push(job)

	return nil
}

// SubmitJob handles job submission via gRPC (not used in Phase 1 - workers receive jobs, not submit them)
func (s *Scheduler) SubmitJob(ctx context.Context, req *workerpb.JobRequest) (*workerpb.JobResponse, error) {
	// This is here to satisfy the gRPC interface
	// In Phase 1, we don't use this - the scheduler pushes jobs to workers
	return &workerpb.JobResponse{
		Success: false,
		Message: "not implemented",
	}, nil
}

// GetJob retrieves a job by ID
func (s *Scheduler) GetJob(jobID string) *queue.Job {
	return s.queue.Get(jobID)
}

// ListJobs returns all jobs (for API/dashboard)
func (s *Scheduler) ListJobs() []*queue.Job {
	// Get all jobs from the queue's index
	return s.queue.GetAll()
}

// ListWorkers returns all registered workers
func (s *Scheduler) ListWorkers() []*models.WorkerInfo {
	s.workers.mu.RLock()
	defer s.workers.mu.RUnlock()

	workers := make([]*models.WorkerInfo, 0, len(s.workers.workers))
	for _, w := range s.workers.workers {
		workers = append(workers, w)
	}
	return workers
}

// GetQueueLength returns the number of pending jobs
func (s *Scheduler) GetQueueLength() int {
	return s.queue.Len()
}

// SendHeartbeat handles heartbeat from workers (gRPC method)
func (s *Scheduler) SendHeartbeat(ctx context.Context, req *workerpb.HeartbeatRequest) (*workerpb.HeartbeatResponse, error) {
	workerID := req.GetWorkerId()

	// Update worker heartbeat with active jobs from the request
	activeJobs := req.GetActiveJobIds()
	s.workers.Heartbeat(workerID, activeJobs)

	return &workerpb.HeartbeatResponse{
		Acknowledged: true,
	}, nil
}
