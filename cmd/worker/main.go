package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/athulya-anil/axon-scheduler/pkg/worker"
	"github.com/athulya-anil/axon-scheduler/proto/workerpb"
	"google.golang.org/grpc"
)

func main() {
	// Get configuration from environment or use defaults
	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		hostname, _ := os.Hostname()
		workerID = "worker-" + hostname
	}

	port := os.Getenv("WORKER_PORT")
	if port == "" {
		port = "50051"
	}

	schedulerAddr := os.Getenv("SCHEDULER_ADDR")
	if schedulerAddr == "" {
		schedulerAddr = "localhost:8080"
	}

	capacity := 5 // Max concurrent jobs - can be made configurable

	log.Printf("🚀 Starting Axon Worker %s on port %s...", workerID, port)
	log.Printf("📡 Scheduler address: %s", schedulerAddr)

	// Create worker instance
	w := worker.NewWorker(workerID, capacity, schedulerAddr)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	workerpb.RegisterWorkerServiceServer(grpcServer, w)

	// Start server in goroutine
	go func() {
		log.Printf("🎧 Worker %s gRPC server listening on port %s", workerID, port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("❌ Failed to serve: %v", err)
		}
	}()

	// Start heartbeat sender
	heartbeat := worker.NewHeartbeatSender(workerID, schedulerAddr, 5*time.Second)
	if err := heartbeat.Start(); err != nil {
		log.Printf("⚠️ Failed to start heartbeat sender: %v", err)
	}
	defer heartbeat.Stop()

	log.Printf("✅ Worker %s ready (capacity: %d)", workerID, capacity)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Printf("🛑 Shutting down worker %s...", workerID)
	grpcServer.GracefulStop()
	w.Stop()

	log.Println("👋 Worker stopped")
}
