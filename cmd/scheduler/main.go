package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/athulya-anil/axon-scheduler/pkg/api"
	"github.com/athulya-anil/axon-scheduler/pkg/dashboard"
	"github.com/athulya-anil/axon-scheduler/pkg/scheduler"
	"github.com/athulya-anil/axon-scheduler/proto/workerpb"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	// Get configuration from environment or use defaults
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		hostname, _ := os.Hostname()
		nodeID = "scheduler-" + hostname
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	etcdEndpoints := os.Getenv("ETCD_ENDPOINTS")
	if etcdEndpoints == "" {
		etcdEndpoints = "127.0.0.1:2379"
	}

	endpoints := strings.Split(etcdEndpoints, ",")

	log.Printf("🚀 Starting Axon Scheduler %s...", nodeID)
	log.Printf("📡 HTTP API: :%s, gRPC: :%s", httpPort, grpcPort)
	log.Printf("🗄️  etcd endpoints: %v", endpoints)

	// Create scheduler instance
	sched, err := scheduler.NewScheduler(nodeID, endpoints)
	if err != nil {
		log.Fatalf("❌ Failed to create scheduler: %v", err)
	}
	defer sched.Stop()

	// Start scheduler in background
	go func() {
		if err := sched.Start(); err != nil {
			log.Fatalf("❌ Scheduler failed: %v", err)
		}
	}()

	// Start gRPC server for worker heartbeats
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("❌ Failed to listen on gRPC port: %v", err)
		}

		grpcServer := grpc.NewServer()
		workerpb.RegisterWorkerServiceServer(grpcServer, sched)

		log.Printf("🎧 gRPC server listening on port %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("❌ gRPC server failed: %v", err)
		}
	}()

	// Set up HTTP API with Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS middleware for testing
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Register API routes
	apiHandler := api.NewAPI(sched)
	apiHandler.SetupRoutes(router)

	// Register Dashboard routes
	dashboardHandler, err := dashboard.NewDashboard(sched)
	if err != nil {
		log.Fatalf("❌ Failed to create dashboard: %v", err)
	}
	dashboardHandler.SetupRoutes(router)

	log.Printf("📊 Dashboard available at http://localhost:%s/dashboard", httpPort)

	// Start HTTP server in background
	go func() {
		log.Printf("🌐 HTTP API server listening on port %s", httpPort)
		if err := router.Run(":" + httpPort); err != nil {
			log.Fatalf("❌ HTTP server failed: %v", err)
		}
	}()

	log.Printf("✅ Scheduler %s ready and running", nodeID)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Printf("🛑 Shutting down scheduler %s...", nodeID)
	log.Println("👋 Scheduler stopped")
}

