package worker

import (
	"context"
	"log"
	"time"

	"github.com/athulya-anil/axon-scheduler/proto/workerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// HeartbeatSender manages heartbeat communication with the scheduler
type HeartbeatSender struct {
	workerID      string
	schedulerAddr string
	interval      time.Duration
	client        workerpb.WorkerServiceClient
	conn          *grpc.ClientConn
	ctx           context.Context
	cancel        context.CancelFunc
	getActiveJobs func() []string // Function to get active job IDs from worker
}

// NewHeartbeatSender creates a new heartbeat sender
func NewHeartbeatSender(workerID, schedulerAddr string, interval time.Duration, getActiveJobs func() []string) *HeartbeatSender {
	ctx, cancel := context.WithCancel(context.Background())

	return &HeartbeatSender{
		workerID:      workerID,
		schedulerAddr: schedulerAddr,
		interval:      interval,
		ctx:           ctx,
		cancel:        cancel,
		getActiveJobs: getActiveJobs,
	}
}

// Start begins sending heartbeats to the scheduler
func (h *HeartbeatSender) Start() error {
	// Connect to scheduler
	conn, err := grpc.NewClient(
		h.schedulerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	h.conn = conn
	h.client = workerpb.NewWorkerServiceClient(conn)

	// Start heartbeat loop
	go h.heartbeatLoop()

	log.Printf("[HB] Heartbeat sender started for worker %s (target: %s, interval: %v)",
		h.workerID, h.schedulerAddr, h.interval)

	return nil
}

// Stop gracefully stops the heartbeat sender
func (h *HeartbeatSender) Stop() {
	h.cancel()
	if h.conn != nil {
		h.conn.Close()
	}
}

// heartbeatLoop sends periodic heartbeats
func (h *HeartbeatSender) heartbeatLoop() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			log.Printf("[HB] Heartbeat sender stopped for worker %s", h.workerID)
			return

		case <-ticker.C:
			if err := h.sendHeartbeat(); err != nil {
				log.Printf("[WARN] Failed to send heartbeat from worker %s: %v", h.workerID, err)
			}
		}
	}
}

// sendHeartbeat sends a single heartbeat message
func (h *HeartbeatSender) sendHeartbeat() error {
	ctx, cancel := context.WithTimeout(h.ctx, 3*time.Second)
	defer cancel()

	// Get active jobs from worker
	var activeJobs []string
	if h.getActiveJobs != nil {
		activeJobs = h.getActiveJobs()
	}

	req := &workerpb.HeartbeatRequest{
		WorkerId:     h.workerID,
		Timestamp:    time.Now().Unix(),
		ActiveJobIds: activeJobs,
	}

	resp, err := h.client.SendHeartbeat(ctx, req)
	if err != nil {
		return err
	}

	if resp.GetAcknowledged() {
		log.Printf("[HB] Worker %s heartbeat acknowledged by scheduler (active jobs: %d)", h.workerID, len(activeJobs))
	}

	return nil
}
