package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/athulya-anil/axon-scheduler/pkg/leader"
)

func main() {
	nodeID, _ := os.Hostname()
	log.Printf("🚀 Axon Scheduler starting on node %s...\n", nodeID)

	// --- Connect to etcd ---
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:4001"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to etcd: %v", err)
	}
	defer cli.Close()

	// --- Create a lease manager for leadership heartbeat ---
	lease := &leader.LeaseManager{Client: cli}
	if err := lease.CreateLease(10); err != nil {
		log.Fatalf("❌ Failed to create lease: %v", err)
	}
	lease.KeepAlive()

	// --- Start a session for leader election ---
	session, err := concurrency.NewSession(cli,
		concurrency.WithTTL(10),
		concurrency.WithLease(lease.LeaseID))
	if err != nil {
		log.Fatalf("❌ Failed to create election session: %v", err)
	}
	defer session.Close()

	election := concurrency.NewElection(session, "/axon-leader-election")

	ctx := context.Background()

	// --- Attempt to become leader ---
	if err := election.Campaign(ctx, nodeID); err != nil {
		log.Fatalf("❌ Election campaign failed: %v", err)
	}
	log.Printf("🏆 Node %s has become the leader!\n", nodeID)

	// --- Graceful shutdown handling ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("🛑 Caught shutdown signal, resigning leadership...")
		lease.Revoke()
		if err := election.Resign(ctx); err != nil {
			log.Printf("⚠️  Failed to resign election: %v", err)
		}
		session.Close()
		cli.Close()
		os.Exit(0)
	}()

	// --- Simulate ongoing leader work ---
	for {
		log.Println("🧭 Leader is active — scheduling jobs...")
		time.Sleep(5 * time.Second)
	}
}

