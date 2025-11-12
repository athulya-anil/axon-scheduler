#!/bin/bash
# Test script for Axon Scheduler Dashboard

set -e

echo "🧪 Testing Axon Scheduler Dashboard"
echo "===================================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SCHEDULER_URL="http://localhost:8080"

# Wait for scheduler to be ready
echo -e "${BLUE}⏳ Waiting for scheduler to be ready...${NC}"
for i in {1..30}; do
    if curl -s "$SCHEDULER_URL/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Scheduler is ready${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ Scheduler did not start in time"
        exit 1
    fi
    sleep 1
done

echo ""
echo -e "${BLUE}📊 Testing Dashboard Endpoints${NC}"
echo "--------------------------------"

# Test dashboard pages
echo -n "1. Testing main dashboard page... "
if curl -s -o /dev/null -w "%{http_code}" "$SCHEDULER_URL/dashboard" | grep -q "200"; then
    echo -e "${GREEN}✅ OK${NC}"
else
    echo -e "❌ FAILED"
    exit 1
fi

echo -n "2. Testing jobs page... "
if curl -s -o /dev/null -w "%{http_code}" "$SCHEDULER_URL/dashboard/jobs" | grep -q "200"; then
    echo -e "${GREEN}✅ OK${NC}"
else
    echo -e "❌ FAILED"
    exit 1
fi

echo -n "3. Testing workers page... "
if curl -s -o /dev/null -w "%{http_code}" "$SCHEDULER_URL/dashboard/workers" | grep -q "200"; then
    echo -e "${GREEN}✅ OK${NC}"
else
    echo -e "❌ FAILED"
    exit 1
fi

# Test API endpoints used by HTMX
echo -n "4. Testing status API... "
if curl -s -o /dev/null -w "%{http_code}" "$SCHEDULER_URL/api/dashboard/status" | grep -q "200"; then
    echo -e "${GREEN}✅ OK${NC}"
else
    echo -e "❌ FAILED"
    exit 1
fi

echo -n "5. Testing jobs API... "
if curl -s -o /dev/null -w "%{http_code}" "$SCHEDULER_URL/api/dashboard/jobs" | grep -q "200"; then
    echo -e "${GREEN}✅ OK${NC}"
else
    echo -e "❌ FAILED"
    exit 1
fi

echo -n "6. Testing workers API... "
if curl -s -o /dev/null -w "%{http_code}" "$SCHEDULER_URL/api/dashboard/workers" | grep -q "200"; then
    echo -e "${GREEN}✅ OK${NC}"
else
    echo -e "❌ FAILED"
    exit 1
fi

echo ""
echo -e "${BLUE}🔧 Testing Job Submission via API${NC}"
echo "--------------------------------"

# Submit a test job
echo -n "7. Submitting test job... "
RESPONSE=$(curl -s -X POST "$SCHEDULER_URL/jobs" \
    -H "Content-Type: application/json" \
    -d '{
        "type": "test-job",
        "payload": {"message": "Hello from dashboard test"},
        "priority": 5,
        "max_retries": 3
    }')

if echo "$RESPONSE" | grep -q "job_id"; then
    echo -e "${GREEN}✅ OK${NC}"
    JOB_ID=$(echo "$RESPONSE" | grep -o '"job_id":"[^"]*"' | cut -d'"' -f4)
    echo "   Job ID: $JOB_ID"
else
    echo -e "❌ FAILED"
    echo "   Response: $RESPONSE"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ All dashboard tests passed!${NC}"
echo ""
echo -e "${YELLOW}📊 Dashboard is available at:${NC}"
echo -e "   ${BLUE}http://localhost:8080/dashboard${NC}"
echo ""
echo -e "${YELLOW}Available pages:${NC}"
echo "   • Overview: http://localhost:8080/dashboard"
echo "   • Jobs:     http://localhost:8080/dashboard/jobs"
echo "   • Workers:  http://localhost:8080/dashboard/workers"
echo ""
echo -e "${YELLOW}💡 Tip: Open the dashboard in your browser to see real-time updates!${NC}"
