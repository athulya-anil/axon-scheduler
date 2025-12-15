#!/bin/bash

# Axon Scheduler Load Test Script
# Tests the distributed scheduler under various load conditions

set -e

# Configuration
SCHEDULER_URL="${SCHEDULER_URL:-http://localhost:8080}"
NUM_JOBS="${NUM_JOBS:-1000}"
CONCURRENT_REQUESTS="${CONCURRENT_REQUESTS:-10}"
JOB_TYPES=("data_processing" "ml_training" "image_processing" "batch_job" "analytics")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "======================================"
echo "  Axon Scheduler Load Test"
echo "======================================"
echo "Scheduler URL: $SCHEDULER_URL"
echo "Total Jobs: $NUM_JOBS"
echo "Concurrent Requests: $CONCURRENT_REQUESTS"
echo "======================================"
echo ""

# Check if scheduler is healthy
echo -e "${YELLOW}Checking scheduler health...${NC}"
if ! curl -sf "$SCHEDULER_URL/health" > /dev/null; then
    echo -e "${RED}❌ Scheduler is not healthy!${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Scheduler is healthy${NC}"
echo ""

# Function to submit a job
submit_job() {
    local job_num=$1
    local job_type=${JOB_TYPES[$RANDOM % ${#JOB_TYPES[@]}]}
    local priority=$((RANDOM % 10 + 1))

    curl -s -X POST "$SCHEDULER_URL/jobs" \
        -H "Content-Type: application/json" \
        -d "{
            \"type\": \"$job_type\",
            \"payload\": {\"job_number\": \"$job_num\", \"data\": \"test_data_$job_num\"},
            \"priority\": $priority,
            \"max_retries\": 3,
            \"timeout_seconds\": 30
        }" > /dev/null

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}.${NC}"
    else
        echo -e "${RED}F${NC}"
    fi
}

# Phase 1: Submit jobs with concurrency control
echo -e "${YELLOW}Phase 1: Submitting $NUM_JOBS jobs...${NC}"
start_time=$(date +%s)

# Use xargs for parallel execution
seq 1 $NUM_JOBS | xargs -P $CONCURRENT_REQUESTS -I {} bash -c "$(declare -f submit_job); submit_job {}"

end_time=$(date +%s)
duration=$((end_time - start_time))

echo ""
echo -e "${GREEN}✓ Submitted $NUM_JOBS jobs in ${duration}s${NC}"
echo -e "  Throughput: $((NUM_JOBS / duration)) jobs/sec"
echo ""

# Phase 2: Monitor queue and job execution
echo -e "${YELLOW}Phase 2: Monitoring job execution...${NC}"

monitor_interval=5
max_wait=300  # 5 minutes max
waited=0

while [ $waited -lt $max_wait ]; do
    status=$(curl -s "$SCHEDULER_URL/status")
    queue_length=$(echo $status | grep -o '"queue_length":[0-9]*' | grep -o '[0-9]*')
    workers_healthy=$(echo $status | grep -o '"workers_healthy":[0-9]*' | grep -o '[0-9]*')

    echo -e "  Queue: $queue_length | Healthy Workers: $workers_healthy"

    # Exit if queue is empty
    if [ "$queue_length" -eq 0 ]; then
        echo -e "${GREEN}✓ All jobs processed!${NC}"
        break
    fi

    sleep $monitor_interval
    waited=$((waited + monitor_interval))
done

if [ $waited -ge $max_wait ]; then
    echo -e "${RED}⚠️ Timeout waiting for jobs to complete${NC}"
fi

echo ""

# Phase 3: Collect metrics and statistics
echo -e "${YELLOW}Phase 3: Collecting metrics...${NC}"

jobs_response=$(curl -s "$SCHEDULER_URL/jobs")
total_jobs=$(echo $jobs_response | grep -o '"count":[0-9]*' | head -1 | grep -o '[0-9]*')

echo ""
echo "======================================"
echo "  Load Test Results"
echo "======================================"
echo "Total Jobs Submitted: $NUM_JOBS"
echo "Total Jobs in System: $total_jobs"
echo "Submission Duration: ${duration}s"
echo "Average Throughput: $((NUM_JOBS / duration)) jobs/sec"
echo "======================================"
echo ""

# Phase 4: Stress test with burst traffic
echo -e "${YELLOW}Phase 4: Burst load test (100 jobs in 1 second)...${NC}"
burst_start=$(date +%s)

seq 1 100 | xargs -P 100 -I {} bash -c "$(declare -f submit_job); submit_job burst_{}"

burst_end=$(date +%s)
burst_duration=$((burst_end - burst_start))

echo ""
echo -e "${GREEN}✓ Burst test completed in ${burst_duration}s${NC}"
echo ""

# Final status check
echo -e "${YELLOW}Final System Status:${NC}"
curl -s "$SCHEDULER_URL/status" | jq '.' || curl -s "$SCHEDULER_URL/status"

echo ""
echo -e "${GREEN}✓ Load test completed successfully!${NC}"
