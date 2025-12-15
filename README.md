# Axon - Distributed Job Scheduler

## Vision

Axon is a production-grade distributed job scheduler designed to handle concurrent workloads with intelligence and reliability. The goal is to build a system that not only distributes jobs across workers efficiently but also learns from past executions to optimize performance through semantic caching.

At its core, Axon solves the challenge of coordinating work across multiple machines while maintaining fault tolerance and high availability. By combining leader election, priority-based scheduling, and semantic understanding of job similarity, Axon aims to reduce redundant computation and improve overall system throughput.

This project demonstrates modern distributed systems principles: consensus algorithms, gRPC communication, intelligent caching strategies, and real-time monitoring—all built with a focus on production-readiness and operational simplicity.

## Tech Stack

**Backend**
- **Go 1.21+** - Core scheduler and worker services
- **gRPC** - High-performance inter-service communication
- **etcd** - Distributed consensus and leader election (Raft)
- **Gin** - REST API framework

**Semantic Caching Layer**
- **Python FastAPI** - Cache service
- **Sentence-Transformers** - Semantic embeddings
- **FAISS** - Vector similarity search
- **Redis** - TTL-based cache storage

**Monitoring & UI**
- **HTMX** - Real-time dashboard
- **TailwindCSS** - Styling
- **Server-Sent Events** - Live updates

**Deployment**
- **Docker & Docker Compose** - Containerization
- **Kubernetes** - Orchestration (planned)

## Key Features

- **Leader Election**: Automatic failover using etcd Raft consensus
- **Priority Scheduling**: Min-heap based priority queue with O(log n) operations
- **Fault Tolerance**: Worker health monitoring with automatic job reassignment
- **Semantic Caching**: AI-powered cache that understands job similarity
- **Real-time Monitoring**: Live dashboard showing system state and metrics
- **Horizontal Scaling**: Support for multiple scheduler nodes and worker pools

## Architecture

```
Client → Scheduler Cluster (Leader + Followers)
           ↓
        etcd (Raft Consensus)
           ↓
      Worker Pool ← Semantic Cache
```

**Core Components**:
1. **Scheduler** - Manages job queue, assigns work, monitors workers
2. **Worker** - Executes jobs with configurable concurrency
3. **Semantic Cache** - Reduces redundant computation using embeddings
4. **Dashboard** - Real-time visualization of system state

## Quick Start

### Prerequisites
```bash
# Install etcd
brew install etcd
brew services start etcd

# Verify Go installation
go version  # 1.21+
```

### Build & Run
```bash
# Build binaries
go build -o bin/scheduler ./cmd/scheduler
go build -o bin/worker ./cmd/worker

# Start scheduler
./bin/scheduler

# Start worker (in new terminal)
WORKER_ID=worker-1 WORKER_PORT=50051 ./bin/worker

# Register worker
curl -X POST http://localhost:8080/workers/register \
  -H "Content-Type: application/json" \
  -d '{"id": "worker-1", "address": "localhost:50051", "capacity": 5}'

# Submit a job
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type": "data_processing", "payload": {"file": "data.csv"}, "priority": 5}'
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/status` | GET | Scheduler status and metrics |
| `/jobs` | POST | Submit a new job |
| `/jobs` | GET | List all jobs |
| `/jobs/:id` | GET | Get job details |
| `/workers/register` | POST | Register a worker |
| `/workers` | GET | List all workers |

## Project Structure

```
axon-scheduler/
├── cmd/
│   ├── scheduler/        # Scheduler service
│   └── worker/           # Worker service
├── pkg/
│   ├── api/              # REST API handlers
│   ├── cache/            # Cache client
│   ├── dashboard/        # HTMX dashboard
│   ├── leader/           # Leader election
│   ├── models/           # Data models
│   ├── queue/            # Priority queue
│   ├── scheduler/        # Scheduler logic
│   └── worker/           # Worker executor
├── proto/                # gRPC definitions
├── cache-service/        # Python semantic cache
└── scripts/              # Testing scripts
```

## Current Status

**Phase 1**: ✅ Basic job execution with leader election
**Phase 2**: ✅ Semantic caching integration
**Phase 3**: ✅ Real-time HTMX dashboard
**Phase 4**: ✅ Production deployment (Docker + K8s + Helm)
**Phase 5**: ✅ Production observability & performance

## Performance Metrics

- **Worker Capacity**: 5 concurrent jobs per worker
- **Heartbeat Interval**: 5 seconds
- **Worker Timeout**: 15 seconds
- **Leader Election TTL**: 10 seconds
- **Cache Hit Rate Target**: 60-80%
- **Cache Lookup Latency**: <5ms

## Testing

```bash
# Run unit tests
go test ./pkg/queue/...
go test ./pkg/scheduler/...

# Integration test
./scripts/test-basic.sh
```

## License

MIT License
