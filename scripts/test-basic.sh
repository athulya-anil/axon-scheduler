#!/bin/bash

echo "========================================="
echo "Axon Scheduler - Phase 1 Basic Test"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

SCHEDULER_URL="http://localhost:8080"

echo -e "${YELLOW}Step 1: Check if scheduler is running...${NC}"
if curl -s "$SCHEDULER_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Scheduler is running${NC}"
else
    echo -e "${RED}✗ Scheduler is not running. Start it with: ./bin/scheduler${NC}"
    echo -e "${YELLOW}  Make sure etcd is running: brew services start etcd (or docker run -d -p 2379:2379 quay.io/coreos/etcd)${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 2: Check scheduler status...${NC}"
curl -s "$SCHEDULER_URL/status" | python3 -m json.tool
echo ""

echo ""
echo -e "${YELLOW}Step 3: Register a worker (simulated)...${NC}"
curl -s -X POST "$SCHEDULER_URL/workers/register" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "worker-test-1",
    "address": "localhost:50051",
    "capacity": 5
  }' | python3 -m json.tool
echo ""

echo ""
echo -e "${YELLOW}Step 4: Submit test jobs...${NC}"
for i in {1..5}; do
  echo "Submitting job $i..."
  curl -s -X POST "$SCHEDULER_URL/jobs" \
    -H "Content-Type: application/json" \
    -d '{
      "type": "test_job",
      "payload": {"data": "test-'$i'"},
      "priority": '$i',
      "max_retries": 3
    }' | python3 -m json.tool
  sleep 0.2
done
echo ""

echo ""
echo -e "${YELLOW}Step 5: Check queue status...${NC}"
curl -s "$SCHEDULER_URL/status" | python3 -m json.tool
echo ""

echo ""
echo -e "${YELLOW}Step 6: List workers...${NC}"
curl -s "$SCHEDULER_URL/workers" | python3 -m json.tool
echo ""

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Test Complete!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo -e "To start the full system:"
echo -e "  1. Start etcd:       ${YELLOW}etcd${NC}"
echo -e "  2. Start scheduler:  ${YELLOW}./bin/scheduler${NC}"
echo -e "  3. Start worker(s):  ${YELLOW}WORKER_ID=worker-1 WORKER_PORT=50051 ./bin/worker${NC}"
echo -e "  4. Submit jobs:      ${YELLOW}curl -X POST http://localhost:8080/jobs ...${NC}"
echo ""
