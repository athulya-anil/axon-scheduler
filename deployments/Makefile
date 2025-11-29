.PHONY: help build-images deploy-k8s deploy-helm delete-k8s delete-helm logs status

# Default target
help:
	@echo "Axon Scheduler Deployment Commands"
	@echo "===================================="
	@echo ""
	@echo "Docker Commands:"
	@echo "  build-images       - Build all Docker images"
	@echo "  docker-up          - Start with Docker Compose"
	@echo "  docker-down        - Stop Docker Compose"
	@echo ""
	@echo "Kubernetes Commands:"
	@echo "  deploy-k8s         - Deploy using kubectl/kustomize"
	@echo "  delete-k8s         - Delete k8s deployment"
	@echo "  k8s-status         - Check k8s deployment status"
	@echo ""
	@echo "Helm Commands:"
	@echo "  deploy-helm        - Deploy using Helm"
	@echo "  upgrade-helm       - Upgrade Helm release"
	@echo "  delete-helm        - Delete Helm release"
	@echo "  helm-status        - Check Helm release status"
	@echo ""
	@echo "Development:"
	@echo "  logs               - View scheduler logs"
	@echo "  port-forward       - Forward ports for local access"

# Build all Docker images
build-images:
	@echo "Building Docker images..."
	docker build -t axon-scheduler:latest -f Dockerfile.scheduler ..
	docker build -t axon-worker:latest -f Dockerfile.worker ..
	docker build -t axon-cache:latest -f ../cache-service/Dockerfile ../cache-service
	@echo "Images built successfully!"

# Docker Compose commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Kubernetes deployment using kubectl/kustomize
deploy-k8s:
	@echo "Deploying to Kubernetes..."
	kubectl apply -k k8s/
	@echo "Deployment complete!"
	@echo "Run 'make k8s-status' to check status"

delete-k8s:
	@echo "Deleting Kubernetes resources..."
	kubectl delete -k k8s/
	@echo "Resources deleted"

k8s-status:
	@echo "=== Namespace ==="
	kubectl get ns axon-scheduler
	@echo ""
	@echo "=== Pods ==="
	kubectl get pods -n axon-scheduler
	@echo ""
	@echo "=== Services ==="
	kubectl get svc -n axon-scheduler
	@echo ""
	@echo "=== StatefulSets ==="
	kubectl get sts -n axon-scheduler
	@echo ""
	@echo "=== Deployments ==="
	kubectl get deploy -n axon-scheduler

# Helm deployment
deploy-helm:
	@echo "Deploying with Helm..."
	helm install axon helm/axon-scheduler/ --namespace axon-scheduler --create-namespace
	@echo "Deployment complete!"
	@echo "Run 'make helm-status' to check status"

upgrade-helm:
	@echo "Upgrading Helm release..."
	helm upgrade axon helm/axon-scheduler/ --namespace axon-scheduler

delete-helm:
	@echo "Deleting Helm release..."
	helm uninstall axon --namespace axon-scheduler
	kubectl delete namespace axon-scheduler

helm-status:
	helm status axon --namespace axon-scheduler
	@echo ""
	kubectl get all -n axon-scheduler

# View logs
logs:
	kubectl logs -n axon-scheduler -l app=axon-scheduler --tail=100 -f

worker-logs:
	kubectl logs -n axon-scheduler -l app=axon-worker --tail=100 -f

cache-logs:
	kubectl logs -n axon-scheduler -l app=axon-cache --tail=100 -f

# Port forwarding for local development
port-forward:
	@echo "Forwarding ports..."
	@echo "Dashboard: http://localhost:8080"
	kubectl port-forward -n axon-scheduler svc/axon-scheduler 8080:8080

# Check etcd cluster health
etcd-health:
	kubectl exec -n axon-scheduler axon-etcd-0 -- etcdctl endpoint health --cluster
	kubectl exec -n axon-scheduler axon-etcd-0 -- etcdctl member list

# Scale workers
scale-workers:
	@read -p "Enter number of workers: " count; \
	kubectl scale deployment axon-worker -n axon-scheduler --replicas=$$count

# Run tests
test-deployment:
	@echo "Testing deployment..."
	curl -X POST http://localhost:8080/jobs \
	  -H "Content-Type: application/json" \
	  -d '{"type": "test", "payload": {"message": "hello"}, "priority": 5}'
