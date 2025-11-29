# Axon Scheduler - Quick Deployment Guide

Complete deployment guide for Axon distributed job scheduler in production environments.

## 🚀 Quick Start Options

### Option 1: Docker Compose (Fastest - Local Dev)

```bash
cd deployments
make build-images
make docker-up

# Access dashboard at http://localhost:8080/dashboard
```

### Option 2: Kubernetes with kubectl

```bash
cd deployments
make build-images

# Deploy to k8s
make deploy-k8s

# Check status
make k8s-status

# Port forward for access
make port-forward
```

### Option 3: Helm (Production Recommended)

```bash
cd deployments
make build-images

# Deploy with Helm
make deploy-helm

# Check status
make helm-status
```

## 📋 Prerequisites

| Deployment Type | Requirements |
|----------------|--------------|
| Docker Compose | Docker 20.10+, Docker Compose 1.29+ |
| Kubernetes | kubectl 1.24+, K8s cluster 1.24+ |
| Helm | Helm 3.8+, kubectl 1.24+ |

## 🏗️ What Gets Deployed

All deployment methods create:

- **3 Scheduler instances** - Leader-elected for HA
- **5 Worker instances** - Scalable job executors
- **3 etcd nodes** - Raft consensus cluster
- **2 Cache instances** - Semantic cache with FAISS
- **1 Redis instance** - Cache backend
- **Monitoring** - Prometheus ServiceMonitors & alerts

## 📊 Deployment Comparison

| Feature | Docker Compose | Kubernetes | Helm |
|---------|---------------|------------|------|
| **Setup Time** | 2 minutes | 10 minutes | 5 minutes |
| **Production Ready** | ❌ No | ✅ Yes | ✅ Yes |
| **Auto-scaling** | ❌ No | ⚙️ Manual | ✅ Yes |
| **Monitoring** | ⚙️ Basic | ✅ Full | ✅ Full |
| **Persistence** | ❌ Local | ✅ PV | ✅ PV |
| **Upgrades** | Manual | Manual | ✅ Easy |
| **Rollback** | ❌ No | ⚙️ Manual | ✅ Easy |

## 🔧 Configuration

### Helm Values

Create `custom-values.yaml`:

```yaml
# Scale workers
worker:
  replicaCount: 10
  autoscaling:
    enabled: true
    maxReplicas: 50

# Enable ingress
ingress:
  enabled: true
  hosts:
    - host: axon.yourdomain.com

# Production storage
etcd:
  persistence:
    size: 20Gi
    storageClass: "fast-ssd"
```

Deploy:
```bash
helm install axon helm/axon-scheduler/ \
  --namespace axon-scheduler \
  --create-namespace \
  --values custom-values.yaml
```

### Environment Variables

Key configurations in `deployments/k8s/configmap.yaml`:

```yaml
WORKER_CAPACITY: "5"          # Jobs per worker
CACHE_TTL: "3600"             # Cache lifetime
SIMILARITY_THRESHOLD: "0.85"  # Semantic similarity
```

## 📈 Scaling

```bash
# Scale workers
make scale-workers
# Enter: 20

# Or directly
kubectl scale deployment axon-worker -n axon-scheduler --replicas=20

# Enable auto-scaling
kubectl autoscale deployment axon-worker \
  -n axon-scheduler \
  --cpu-percent=70 \
  --min=5 \
  --max=50
```

## 🔍 Monitoring & Observability

### Metrics

Access Prometheus metrics:
```bash
# Port forward Prometheus
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090

# Or Grafana
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
```

Key metrics:
- `axon_scheduler_queue_length` - Job queue size
- `axon_scheduler_is_leader` - Leader election status
- `axon_worker_jobs_running` - Active jobs
- `axon_cache_hits / axon_cache_misses` - Cache performance

### Logs

```bash
# Scheduler logs
make logs

# Worker logs
make worker-logs

# Cache logs
make cache-logs

# All logs
kubectl logs -n axon-scheduler --all-containers=true --tail=100
```

## 🚨 Alerts

Pre-configured Prometheus alerts in `deployments/k8s/monitoring.yaml`:

- `SchedulerDown` - Scheduler unavailable
- `NoLeaderElected` - No leader for 2+ minutes
- `HighQueueLength` - Queue > 100 jobs for 5+ minutes
- `WorkerUnhealthy` - Worker unhealthy for 3+ minutes
- `LowCacheHitRate` - Hit rate < 30% for 10+ minutes

## 🔐 Production Checklist

Before going to production:

- [ ] Use private container registry
- [ ] Enable persistent volumes for etcd
- [ ] Configure Ingress with TLS/SSL
- [ ] Set up monitoring & alerting
- [ ] Configure resource limits
- [ ] Enable network policies
- [ ] Set up backup for etcd
- [ ] Test disaster recovery
- [ ] Configure pod autoscaling
- [ ] Document runbooks

## 🛠️ Useful Commands

```bash
# Health checks
make etcd-health              # Check etcd cluster
make k8s-status              # Full k8s status
make helm-status             # Helm release status

# Operations
make port-forward            # Access dashboard locally
make scale-workers          # Scale worker pool
make test-deployment        # Submit test job

# Debugging
kubectl describe pod <pod-name> -n axon-scheduler
kubectl get events -n axon-scheduler --sort-by='.lastTimestamp'
kubectl exec -it <pod-name> -n axon-scheduler -- /bin/sh
```

## 📖 Full Documentation

See `deployments/README.md` for:
- Detailed architecture diagrams
- Step-by-step deployment guides
- Advanced configuration options
- Troubleshooting guides
- Performance tuning tips

## 🆘 Troubleshooting

### Pods not starting
```bash
kubectl get pods -n axon-scheduler
kubectl describe pod <pod-name> -n axon-scheduler
kubectl logs <pod-name> -n axon-scheduler
```

### No leader elected
```bash
kubectl logs -l app=axon-scheduler -n axon-scheduler | grep -i leader
make etcd-health
```

### Workers not connecting
```bash
kubectl logs -l app=axon-worker -n axon-scheduler
kubectl get svc axon-scheduler -n axon-scheduler
```

## 🔄 Upgrades

### Helm
```bash
# Update values.yaml
helm upgrade axon helm/axon-scheduler/ \
  --namespace axon-scheduler \
  --values custom-values.yaml

# Rollback if needed
helm rollback axon -n axon-scheduler
```

### Kubernetes
```bash
# Update manifests
kubectl apply -k k8s/

# Rolling update
kubectl set image deployment/axon-scheduler \
  scheduler=axon-scheduler:v2.0.0 \
  -n axon-scheduler
```

## 🌐 Access Points

After deployment:

| Service | URL | Port Forward |
|---------|-----|--------------|
| Dashboard | http://axon.local | `make port-forward` |
| REST API | http://axon.local:8080 | `kubectl port-forward svc/axon-scheduler 8080:8080` |
| Prometheus | Configure via ServiceMonitor | See monitoring section |
| Grafana | Install via kube-prometheus-stack | See monitoring section |

## 📞 Support

- **Issues**: https://github.com/athulya-anil/axon-scheduler/issues
- **Documentation**: See README.md and deployments/README.md
- **Examples**: See scripts/ directory for test scripts

---

**Ready to deploy?** Start with Docker Compose for testing, then move to Helm for production!
