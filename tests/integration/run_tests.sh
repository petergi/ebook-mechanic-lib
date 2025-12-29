#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=== EBMLib Integration Test Suite ==="
echo "Project root: $PROJECT_ROOT"
echo

cd "$PROJECT_ROOT"

echo "Step 1: Checking if test fixtures exist..."
if [ ! -f "testdata/epub/valid/minimal.epub" ]; then
    echo "  Generating EPUB fixtures..."
    cd testdata/epub
    go run generate_fixtures.go .
    cd "$PROJECT_ROOT"
else
    echo "  EPUB fixtures found"
fi

if [ ! -f "testdata/pdf/valid/minimal.pdf" ]; then
    echo "  Generating PDF fixtures..."
    cd testdata/pdf
    go run generate_fixtures.go .
    cd "$PROJECT_ROOT"
else
    echo "  PDF fixtures found"
fi

echo
echo "Step 2: Running integration tests..."
cd tests/integration
go test -v -coverprofile=coverage.out

echo
echo "Step 3: Generating coverage report..."
go tool cover -func=coverage.out
echo
echo "Detailed HTML coverage report: coverage.html"
go tool cover -html=coverage.out -o coverage.html

echo
echo "Step 4: Running benchmarks..."
go test -v -bench=. -benchmem -run=^$

echo
echo "=== Test Suite Complete ==="
echo "Coverage report: $SCRIPT_DIR/coverage.html"
