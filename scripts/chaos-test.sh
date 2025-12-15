#!/bin/bash

# Axon Chaos Engineering Test Script
# Simulates failures and tests system resilience

set -e

NAMESPACE="${NAMESPACE:-axon-scheduler}"
SCHEDULER_URL="${SCHEDULER_URL:-http://localhost:8080}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "======================================"
echo "  Axon Chaos Engineering Tests"
echo "======================================"
echo "Namespace: $NAMESPACE"
echo "======================================"
echo ""

# Test 1: Leader Election Failover
echo -e "${YELLOW}Test 1: Leader Election Failover${NC}"
echo "Finding current leader..."

leader_pod=$(kubectl get pods -n $NAMESPACE -l app=axon-scheduler --no-headers | head -1 | awk '{print $1}')

if [ -z "$leader_pod" ]; then
    echo -e "${RED}❌ No scheduler pods found${NC}"
    exit 1
fi

echo "Current leader candidate: $leader_pod"
echo "Deleting leader pod to trigger failover..."

kubectl delete pod -n $NAMESPACE $leader_pod --wait=false

echo "Waiting for new leader election (30s)..."
sleep 30

new_pods=$(kubectl get pods -n $NAMESPACE -l app=axon-scheduler --no-headers | wc -l)
echo -e "${GREEN}✓ Scheduler pods after failover: $new_pods${NC}"
echo ""

# Test 2: Worker Failure and Job Reassignment
echo -e "${YELLOW}Test 2: Worker Failure and Job Reassignment${NC}"
echo "Finding a worker pod..."

worker_pod=$(kubectl get pods -n $NAMESPACE -l app=axon-worker --no-headers | head -1 | awk '{print $1}')

if [ -z "$worker_pod" ]; then
    echo -e "${YELLOW}⚠️ No worker pods found, skipping test${NC}"
else
    echo "Worker to kill: $worker_pod"

    # Submit some jobs first
    echo "Submitting 10 test jobs..."
    for i in {1..10}; do
        curl -s -X POST "$SCHEDULER_URL/jobs" \
            -H "Content-Type: application/json" \
            -d "{\"type\": \"chaos_test\", \"payload\": {\"test\": \"$i\"}, \"priority\": 5}" > /dev/null
    done

    echo "Deleting worker pod..."
    kubectl delete pod -n $NAMESPACE $worker_pod --wait=false

    echo "Waiting for job reassignment (20s)..."
    sleep 20

    worker_count=$(kubectl get pods -n $NAMESPACE -l app=axon-worker --no-headers | wc -l)
    echo -e "${GREEN}✓ Worker pods after failure: $worker_count${NC}"
fi
echo ""

# Test 3: Cache Service Restart
echo -e "${YELLOW}Test 3: Cache Service Resilience${NC}"
echo "Finding cache pod..."

cache_pod=$(kubectl get pods -n $NAMESPACE -l app=axon-cache --no-headers | head -1 | awk '{print $1}')

if [ -z "$cache_pod" ]; then
    echo -e "${YELLOW}⚠️ No cache pods found, skipping test${NC}"
else
    echo "Cache pod: $cache_pod"
    echo "Restarting cache pod..."

    kubectl delete pod -n $NAMESPACE $cache_pod --wait=false

    echo "Waiting for cache to restart (15s)..."
    sleep 15

    cache_ready=$(kubectl get pods -n $NAMESPACE -l app=axon-cache --no-headers | grep Running | wc -l)
    echo -e "${GREEN}✓ Cache pods running: $cache_ready${NC}"
fi
echo ""

# Test 4: Network Partition Simulation (requires NetworkPolicy support)
echo -e "${YELLOW}Test 4: Network Partition Simulation${NC}"
echo "This test requires NetworkPolicy support in your cluster"
echo -e "${YELLOW}⚠️ Skipping network partition test (manual intervention required)${NC}"
echo ""

# Test 5: Load During Chaos
echo -e "${YELLOW}Test 5: Load Test During Chaos${NC}"
echo "Submitting jobs while inducing random pod failures..."

# Submit jobs in background
(
    for i in {1..50}; do
        curl -s -X POST "$SCHEDULER_URL/jobs" \
            -H "Content-Type: application/json" \
            -d "{\"type\": \"chaos_load\", \"payload\": {\"test\": \"$i\"}, \"priority\": 5}" > /dev/null &
        sleep 0.2
    done
    wait
) &

# Randomly delete a pod
sleep 5
random_pod=$(kubectl get pods -n $NAMESPACE -l 'app in (axon-scheduler,axon-worker)' --no-headers | shuf -n 1 | awk '{print $1}')

if [ -n "$random_pod" ]; then
    echo "Randomly deleting pod: $random_pod"
    kubectl delete pod -n $NAMESPACE $random_pod --wait=false
fi

wait
echo -e "${GREEN}✓ Load test during chaos completed${NC}"
echo ""

# Final System Health Check
echo -e "${YELLOW}Final System Health Check:${NC}"
echo ""

echo "Scheduler Pods:"
kubectl get pods -n $NAMESPACE -l app=axon-scheduler

echo ""
echo "Worker Pods:"
kubectl get pods -n $NAMESPACE -l app=axon-worker

echo ""
echo "Cache Pods:"
kubectl get pods -n $NAMESPACE -l app=axon-cache

echo ""
echo -e "${GREEN}✓ Chaos engineering tests completed!${NC}"
echo "Review the system state above to verify resilience."
