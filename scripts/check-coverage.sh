#!/bin/bash

set -e

THRESHOLD=${1:-80}
COVERAGE_FILE="coverage.out"

echo "🧪 Running tests with coverage..."
go test -race -coverprofile=${COVERAGE_FILE} -covermode=atomic ./...

echo ""
echo "📊 Coverage report:"
go tool cover -func=${COVERAGE_FILE}

echo ""
COVERAGE=$(go tool cover -func=${COVERAGE_FILE} | grep total | awk '{print $3}' | sed 's/%//')

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Total coverage: ${COVERAGE}%"
echo "Threshold: ${THRESHOLD}%"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
    echo "❌ Coverage ${COVERAGE}% is below threshold ${THRESHOLD}%"
    echo ""
    echo "To see detailed coverage, run:"
    echo "  go tool cover -html=${COVERAGE_FILE}"
    exit 1
else
    echo "✅ Coverage ${COVERAGE}% meets threshold ${THRESHOLD}%"
fi

echo ""
echo "To view HTML coverage report, run:"
echo "  go tool cover -html=${COVERAGE_FILE}"
