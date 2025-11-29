# Axon Scheduler Deployment Guide

This directory contains deployment configurations for Axon Scheduler in production environments.

## Table of Contents

- [Quick Start](#quick-start)
- [Deployment Options](#deployment-options)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Helm Deployment](#helm-deployment)
- [Docker Compose](#docker-compose)
- [Configuration](#configuration)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

- **Docker Compose**: Docker 20.10+ and Docker Compose 1.29+
- **Kubernetes**: kubectl 1.24+ and a K8s cluster (1.24+)
- **Helm**: Helm 3.8+ (for Helm deployment)
- **Storage**: Persistent volume provisioner (for production)

### Local Development (Docker Compose)

```bash
# Build images
make build-images

# Start all services
make docker-up

# Check status
docker-compose ps

# Access dashboard
open http://localhost:8080/dashboard

# Stop services
make docker-down
```

## Deployment Options

### 1. Docker Compose (Development/Testing)

Best for: Local development, testing, demos

```bash
cd deployments
docker-compose up -d
```

**Components:**
- 3 Scheduler replicas (ports 8080-8082)
- 5 Worker replicas (ports 50051-50055)
- 3 etcd nodes (Raft cluster)
- Semantic cache service
- All services networked together

### 2. Kubernetes with kubectl/kustomize

Best for: Production, full control, GitOps workflows

```bash
# Deploy everything
make deploy-k8s

# Or manually
kubectl apply -k k8s/

# Check status
make k8s-status
```

### 3. Helm (Recommended for Production)

Best for: Production, easy configuration, upgrades

```bash
# Deploy with default values
make deploy-helm

# Or with custom values
helm install axon helm/axon-scheduler/ \
  --namespace axon-scheduler \
  --create-namespace \
  --values custom-values.yaml

# Upgrade
make upgrade-helm

# Status
make helm-status
```

## Kubernetes Deployment

### Architecture

```
┌─────────────────────────────────────────┐
│         Ingress (nginx)                 │
│         axon.yourdomain.com             │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│   Scheduler Service (LoadBalancer)      │
│   - HTTP: 8080 (REST API + Dashboard)   │
│   - gRPC: 9090 (Worker communication)   │
└────────────────┬────────────────────────┘
                 │
         ┌───────┴────────┐
         ▼                ▼
┌──────────────┐   ┌──────────────┐
│  Scheduler   │   │    Workers   │
│  (3 pods)    │   │   (5 pods)   │
└──────┬───────┘   └──────┬───────┘
       │                  │
       ▼                  ▼
┌──────────────┐   ┌──────────────┐
│ etcd Cluster │   │ Cache Service│
│ StatefulSet  │   │  + Redis     │
│  (3 nodes)   │   │  (2 pods)    │
└──────────────┘   └──────────────┘
```

### Step-by-Step Deployment

#### 1. Build and Push Images

```bash
# Build images
cd deployments
make build-images

# Tag for your registry
docker tag axon-scheduler:latest your-registry/axon-scheduler:v1.0.0
docker tag axon-worker:latest your-registry/axon-worker:v1.0.0
docker tag axon-cache:latest your-registry/axon-cache:v1.0.0

# Push to registry
docker push your-registry/axon-scheduler:v1.0.0
docker push your-registry/axon-worker:v1.0.0
docker push your-registry/axon-cache:v1.0.0
```

#### 2. Configure Storage

Ensure your cluster has a default storage class or create one:

```bash
kubectl get storageclass

# For local testing (minikube/kind)
kubectl apply -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
EOF
```

#### 3. Deploy

**Using kubectl:**

```bash
# Update image tags in k8s/*.yaml if needed
kubectl apply -k k8s/

# Wait for pods
kubectl wait --for=condition=ready pod -l app=axon-scheduler -n axon-scheduler --timeout=300s
```

**Using Helm:**

```bash
helm install axon helm/axon-scheduler/ \
  --namespace axon-scheduler \
  --create-namespace \
  --set scheduler.image.repository=your-registry/axon-scheduler \
  --set scheduler.image.tag=v1.0.0 \
  --set worker.image.repository=your-registry/axon-worker \
  --set worker.image.tag=v1.0.0 \
  --set cache.image.repository=your-registry/axon-cache \
  --set cache.image.tag=v1.0.0
```

#### 4. Verify Deployment

```bash
# Check all resources
kubectl get all -n axon-scheduler

# Check etcd cluster
make etcd-health

# View scheduler logs
make logs

# Port forward for local access
make port-forward
```

## Helm Deployment

### Chart Structure

```
helm/axon-scheduler/
├── Chart.yaml              # Chart metadata
├── values.yaml             # Default configuration
└── templates/
    ├── _helpers.tpl        # Template helpers
    ├── namespace.yaml
    ├── configmap.yaml
    ├── etcd-statefulset.yaml
    ├── scheduler-deployment.yaml
    ├── worker-deployment.yaml
    ├── cache-deployment.yaml
    └── ingress.yaml
```

### Custom Configuration

Create a `custom-values.yaml`:

```yaml
# Scale workers based on load
worker:
  replicaCount: 10
  autoscaling:
    enabled: true
    minReplicas: 5
    maxReplicas: 50
    targetCPUUtilizationPercentage: 70

# Enable TLS
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: axon.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: axon-tls-secret
      hosts:
        - axon.yourdomain.com
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"

# Production storage
etcd:
  persistence:
    enabled: true
    storageClass: "ssd"
    size: 20Gi

# Resource limits
scheduler:
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 2000m
      memory: 1Gi
```

Deploy with custom values:

```bash
helm install axon helm/axon-scheduler/ \
  --namespace axon-scheduler \
  --create-namespace \
  --values custom-values.yaml
```

### Helm Operations

```bash
# List releases
helm list -n axon-scheduler

# Upgrade
helm upgrade axon helm/axon-scheduler/ \
  --namespace axon-scheduler \
  --values custom-values.yaml

# Rollback
helm rollback axon -n axon-scheduler

# Uninstall
helm uninstall axon -n axon-scheduler
```

## Configuration

### Environment Variables

Configured via ConfigMap (k8s/configmap.yaml):

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | 8080 | Scheduler HTTP/REST port |
| `GRPC_PORT` | 9090 | Scheduler gRPC port |
| `ETCD_ENDPOINTS` | etcd cluster | etcd connection string |
| `WORKER_PORT` | 50051 | Worker gRPC port |
| `WORKER_CAPACITY` | 5 | Concurrent jobs per worker |
| `CACHE_ADDR` | http://axon-cache:8000 | Cache service URL |
| `CACHE_TTL` | 3600 | Cache TTL in seconds |
| `SIMILARITY_THRESHOLD` | 0.85 | Semantic similarity threshold |

### Scaling

```bash
# Scale schedulers
kubectl scale deployment axon-scheduler -n axon-scheduler --replicas=5

# Scale workers (or use Makefile)
make scale-workers
# Enter: 20

# Enable HPA for workers
kubectl autoscale deployment axon-worker \
  -n axon-scheduler \
  --cpu-percent=70 \
  --min=5 \
  --max=50
```

## Monitoring

### Prometheus Setup

The deployment includes ServiceMonitors for Prometheus Operator:

```bash
# Install Prometheus Operator (if not already installed)
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace

# ServiceMonitors are automatically detected
kubectl get servicemonitor -n axon-scheduler
```

### Metrics Exposed

**Scheduler metrics:**
- `axon_scheduler_is_leader` - Leader election status
- `axon_scheduler_queue_length` - Current queue size
- `axon_scheduler_jobs_total` - Total jobs processed
- `axon_scheduler_workers_healthy` - Healthy worker count

**Worker metrics:**
- `axon_worker_jobs_running` - Currently executing jobs
- `axon_worker_jobs_completed` - Completed job count
- `axon_worker_capacity` - Worker capacity

**Cache metrics:**
- `axon_cache_hits` - Cache hit count
- `axon_cache_misses` - Cache miss count
- `axon_cache_size` - Number of cached items

### Accessing Dashboards

```bash
# Port forward Prometheus
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090

# Port forward Grafana
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80

# Access Axon dashboard
make port-forward
```

## Troubleshooting

### Common Issues

#### Pods not starting

```bash
# Check pod status
kubectl get pods -n axon-scheduler

# Describe pod
kubectl describe pod <pod-name> -n axon-scheduler

# Check logs
kubectl logs <pod-name> -n axon-scheduler

# Check events
kubectl get events -n axon-scheduler --sort-by='.lastTimestamp'
```

#### etcd cluster issues

```bash
# Check etcd health
make etcd-health

# Check etcd logs
kubectl logs axon-etcd-0 -n axon-scheduler

# Exec into etcd pod
kubectl exec -it axon-etcd-0 -n axon-scheduler -- /bin/sh

# Inside pod, check cluster
etcdctl member list
etcdctl endpoint status --cluster
```

#### No leader elected

```bash
# Check scheduler logs for election errors
kubectl logs -l app=axon-scheduler -n axon-scheduler | grep -i leader

# Verify etcd connectivity
kubectl exec -it <scheduler-pod> -n axon-scheduler -- \
  nc -zv axon-etcd-0.axon-etcd 2379
```

#### Workers not connecting

```bash
# Check worker logs
kubectl logs -l app=axon-worker -n axon-scheduler

# Verify scheduler service
kubectl get svc axon-scheduler -n axon-scheduler

# Test connectivity
kubectl exec -it <worker-pod> -n axon-scheduler -- \
  nc -zv axon-scheduler 9090
```

#### Cache service issues

```bash
# Check cache logs
make cache-logs

# Check Redis connectivity
kubectl exec -it <cache-pod> -n axon-scheduler -- \
  redis-cli -h axon-redis ping
```

### Performance Tuning

#### Increase worker capacity

Edit `k8s/configmap.yaml` or Helm values:

```yaml
worker:
  config:
    capacity: 10  # Increase from 5
```

#### Adjust resource limits

```yaml
worker:
  resources:
    requests:
      cpu: 1000m
      memory: 1Gi
    limits:
      cpu: 4000m
      memory: 4Gi
```

#### Scale horizontally

```bash
# Add more workers
kubectl scale deployment axon-worker -n axon-scheduler --replicas=20

# Add more schedulers (max 3 for leader election)
kubectl scale deployment axon-scheduler -n axon-scheduler --replicas=3
```

## Production Checklist

- [ ] Use persistent volumes for etcd
- [ ] Configure resource requests/limits
- [ ] Set up Ingress with TLS
- [ ] Enable Prometheus monitoring
- [ ] Configure alerts for critical metrics
- [ ] Set up log aggregation (e.g., ELK, Loki)
- [ ] Use private container registry
- [ ] Enable network policies
- [ ] Configure pod security policies
- [ ] Set up backup for etcd data
- [ ] Test disaster recovery procedures
- [ ] Configure horizontal pod autoscaling
- [ ] Set up health check monitoring
- [ ] Document runbooks for common issues

## Next Steps

1. **Set up CI/CD**: Automate image builds and deployments
2. **Configure observability**: Add tracing with Jaeger/Tempo
3. **Implement GitOps**: Use ArgoCD or Flux for deployment
4. **Multi-region**: Deploy across multiple regions for HA
5. **Security hardening**: Add OPA policies, network policies

## Support

For issues and questions:
- GitHub Issues: https://github.com/athulya-anil/axon-scheduler/issues
- Documentation: See main README.md
