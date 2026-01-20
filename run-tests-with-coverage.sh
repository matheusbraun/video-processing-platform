#!/bin/bash
set -e

echo "Running tests with coverage for all services..."
echo ""

# Array to store coverage percentages
declare -a coverages

# Test each service
for service in services/auth services/api-gateway services/storage services/notification services/processing-worker; do
    echo "=== Testing $service ==="
    cd "$service"
    
    # Run tests with coverage
    go test -race -coverprofile=coverage.out -covermode=atomic ./... > /dev/null 2>&1
    
    # Extract coverage percentage
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    coverages+=("$service: $coverage")
    
    cd - > /dev/null
    echo "✓ $service: $coverage coverage"
done

echo ""
echo "=== Coverage Summary ==="
for cov in "${coverages[@]}"; do
    echo "  $cov"
done
