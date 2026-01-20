#!/bin/bash
set -e

echo "================================================"
echo "Running tests with race detection and coverage"
echo "================================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

total_coverage=0
count=0
failed=0

# Test each service
for service in services/auth services/api-gateway services/storage services/notification services/processing-worker; do
    echo "Testing $service..."

    if (cd "$service" && go test -race -coverprofile=coverage.out -covermode=atomic ./... > /dev/null 2>&1); then
        # Extract coverage percentage
        coverage=$(cd "$service" && go tool cover -func=coverage.out 2>/dev/null | grep total | awk '{print $3}' | sed 's/%//')

        if [ -n "$coverage" ]; then
            printf "${GREEN}✓${NC} $service: ${coverage}%% coverage\n"
            total_coverage=$(echo "$total_coverage + $coverage" | bc)
            count=$((count + 1))
        else
            printf "${GREEN}✓${NC} $service: tests passed\n"
        fi
    else
        printf "${RED}✗${NC} $service: tests failed\n"
        failed=$((failed + 1))
    fi
done

echo ""
echo "================================================"
if [ $count -gt 0 ]; then
    avg_coverage=$(echo "scale=1; $total_coverage / $count" | bc)
    echo "Average Coverage: ${avg_coverage}%"
fi

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}$failed service(s) failed${NC}"
    exit 1
fi
