# Axon Observability Guide

This guide covers the comprehensive observability stack implemented in Phase 5 of the Axon distributed job scheduler.

## Overview

Phase 5 adds production-grade observability with:
- **Prometheus** for metrics collection
- **Grafana** for visualization and dashboards
- **Comprehensive metrics** across all services
- **Load testing** tools
- **Chaos engineering** tests

## Metrics Architecture

### Scheduler Metrics (`/metrics` on port 8080)

**Queue Metrics:**
- `axon_scheduler_queue_length` - Current number of pending jobs
- `axon_scheduler_jobs_submitted_total` - Total jobs submitted (by type)
- `axon_scheduler_jobs_assigned_total` - Total jobs assigned to workers (by worker_id)
- `axon_scheduler_jobs_completed_total` - Total jobs completed (by status: success/failed)

**Job Performance:**
- `axon_scheduler_job_assignment_duration_seconds` - Time to assign a job to a worker
- `axon_scheduler_job_execution_duration_seconds` - Job execution time (by type and status)

**Worker Metrics:**
- `axon_scheduler_workers_registered` - Number of registered workers
- `axon_scheduler_workers_healthy` - Number of healthy workers
- `axon_worker_healthy` - Individual worker health (by worker_id)
- `axon_worker_capacity` - Worker capacity (by worker_id)
- `axon_worker_active_jobs` - Active jobs on worker (by worker_id)
- `axon_worker_heartbeats_total` - Heartbeats received (by worker_id)

**Leader Election:**
- `axon_scheduler_is_leader` - Whether this node is the leader (1/0)
- `axon_leader_election_campaigns_total` - Total election campaigns
- `axon_leader_election_failures_total` - Failed election campaigns
- `axon_leadership_changes_total` - Total leadership changes

**Job Reliability:**
- `axon_job_retries_total` - Job retries (by job_type)
- `axon_job_max_retries_exceeded_total` - Jobs that failed after max retries

### Worker Metrics (`/metrics` on port 9091)

- `axon_worker_active_jobs` - Current active jobs on this worker
- `axon_scheduler_jobs_completed_total` - Jobs completed (by status)
- `axon_scheduler_job_execution_duration_seconds` - Job execution duration

### Cache Service Metrics (`/metrics` on port 8000)

**Cache Performance:**
- `axon_cache_hits` - Total cache hits
- `axon_cache_misses` - Total cache misses
- `axon_cache_lookups_total` - Total cache lookups
- `axon_cache_size` - Current number of cached entries
- `axon_cache_adds_total` - Total entries added to cache

**Cache Latency:**
- `axon_cache_search_duration_seconds` - Cache search latency
- `axon_cache_add_duration_seconds` - Cache add operation latency
- `axon_cache_embedding_duration_seconds` - Embedding generation time

## Deployment

### Local Development

1. **Start services with metrics:**
   ```bash
   # Scheduler exposes metrics on :8080/metrics
   ./bin/scheduler

   # Worker exposes metrics on :9091/metrics (configurable via METRICS_PORT)
   WORKER_ID=worker-1 METRICS_PORT=9091 ./bin/worker

   # Cache service exposes metrics on :8000/metrics
   cd cache-service && python main.py
   ```

2. **Access metrics:**
   ```bash
   # Scheduler metrics
   curl http://localhost:8080/metrics

   # Worker metrics
   curl http://localhost:9091/metrics

   # Cache metrics
   curl http://localhost:8000/metrics
   ```

### Kubernetes Deployment

1. **Deploy Prometheus:**
   ```bash
   kubectl apply -f deployments/k8s/prometheus.yaml
   ```

2. **Deploy Grafana:**
   ```bash
   kubectl apply -f deployments/k8s/grafana.yaml
   ```

3. **Access Grafana:**
   ```bash
   kubectl port-forward -n axon-scheduler svc/grafana 3000:3000
   ```
   - Open http://localhost:3000
   - Login: admin / admin
   - Prometheus datasource is pre-configured

4. **Access Prometheus:**
   ```bash
   kubectl port-forward -n axon-scheduler svc/prometheus 9090:9090
   ```
   - Open http://localhost:9090

## Grafana Dashboards

The system includes a comprehensive dashboard: **Axon Scheduler - Production Monitoring**

**Dashboard Panels:**

1. **Queue Length** - Real-time queue depth
2. **Leader Status** - Current leader election status
3. **Healthy Workers** - Number of healthy workers
4. **Job Submission Rate** - Jobs/sec submitted (by type)
5. **Job Completion Rate** - Jobs/sec completed (by status)
6. **Cache Hit Rate** - Percentage of cache hits (with thresholds)
7. **Cache Lookup Duration** - p50/p95 latency for cache lookups
8. **Cache Operations** - Hits vs misses over time
9. **Job Execution Duration** - p50/p95 job execution time
10. **Worker Active Jobs** - Jobs per worker
11. **Leadership Changes** - Total leadership transitions
12. **Job Retry Rate** - Retries/sec (by job type)
13. **Max Retries Exceeded** - Failed jobs counter

### Key Metrics to Monitor

**Alerts to Configure:**

- Queue length > 100 for 5+ minutes (capacity issue)
- No leader elected for 2+ minutes (election failure)
- Cache hit rate < 30% for 10+ minutes (cache effectiveness)
- Worker healthy count = 0 (all workers down)
- High job retry rate (system instability)

## Load Testing

### Running Load Tests

```bash
# Basic load test (1000 jobs, 10 concurrent)
./scripts/load-test.sh

# Custom load test
NUM_JOBS=5000 CONCURRENT_REQUESTS=50 ./scripts/load-test.sh

# Against remote scheduler
SCHEDULER_URL=http://your-scheduler:8080 ./scripts/load-test.sh
```

**Load Test Phases:**
1. Job submission with controlled concurrency
2. Queue monitoring and execution tracking
3. Metrics collection and analysis
4. Burst traffic test (100 jobs in 1 second)

**Expected Results:**
- Throughput: 50-200 jobs/sec (depends on worker count)
- Queue processing: All jobs processed within 5 minutes
- System remains stable under burst load

## Chaos Engineering

### Running Chaos Tests

```bash
# Run all chaos tests (requires kubectl access)
./scripts/chaos-test.sh

# Custom namespace
NAMESPACE=my-namespace ./scripts/chaos-test.sh
```

**Chaos Test Scenarios:**

1. **Leader Election Failover**
   - Kills current leader pod
   - Verifies new leader election
   - Ensures system continues operating

2. **Worker Failure and Job Reassignment**
   - Submits jobs to workers
   - Kills a worker pod
   - Verifies jobs are reassigned

3. **Cache Service Resilience**
   - Restarts cache pod
   - Verifies system handles cache unavailability

4. **Load During Chaos**
   - Submits jobs while randomly killing pods
   - Tests system stability under failure conditions

## Querying Metrics

### Useful Prometheus Queries

**Cache Hit Rate:**
```promql
rate(axon_cache_hits[5m]) / (rate(axon_cache_hits[5m]) + rate(axon_cache_misses[5m]))
```

**Job Throughput (jobs/sec):**
```promql
rate(axon_scheduler_jobs_completed_total[5m])
```

**p95 Job Execution Time:**
```promql
histogram_quantile(0.95, rate(axon_scheduler_job_execution_duration_seconds_bucket[5m]))
```

**Worker Utilization:**
```promql
axon_worker_active_jobs / axon_worker_capacity
```

**Average Queue Length (5min):**
```promql
avg_over_time(axon_scheduler_queue_length[5m])
```

## Performance Targets (SLOs)

**Availability:**
- System uptime: 99.9%
- Leader election time: < 10 seconds

**Performance:**
- Job assignment latency: p95 < 100ms
- Job execution: p95 < 5 seconds (varies by job type)
- Cache lookup: p95 < 5ms
- Queue processing: All jobs within 5 minutes at steady state

**Reliability:**
- Cache hit rate: > 60%
- Job success rate: > 95%
- Worker heartbeat success: > 99%

## Troubleshooting

### High Queue Length

1. Check worker health: `axon_scheduler_workers_healthy`
2. Check worker capacity: sum of `axon_worker_capacity`
3. Look for job failures: `axon_job_max_retries_exceeded_total`
4. Scale workers if needed

### Low Cache Hit Rate

1. Check cache size: `axon_cache_size`
2. Review cache TTL settings
3. Examine query similarity patterns
4. Consider adjusting similarity threshold

### Leadership Issues

1. Monitor `axon_leadership_changes_total`
2. Check etcd health
3. Review scheduler logs for election failures
4. Verify network connectivity between schedulers

### Worker Health Issues

1. Check `axon_worker_heartbeats_total` rate
2. Verify network connectivity
3. Review worker logs
4. Check resource constraints (CPU/memory)

## Best Practices

1. **Monitor continuously** - Set up alerting on key metrics
2. **Baseline performance** - Run load tests regularly to establish baseline
3. **Gradual rollouts** - Test changes in staging with chaos tests
4. **Capacity planning** - Monitor queue length and worker utilization trends
5. **Cache optimization** - Track hit rate and adjust TTL/similarity thresholds

## Next Steps

Future enhancements (Phase 6+):
- Distributed tracing with OpenTelemetry/Jaeger
- Structured logging with correlation IDs
- Custom SLO dashboards and alerting rules
- Advanced anomaly detection
- Cost/performance optimization tools
