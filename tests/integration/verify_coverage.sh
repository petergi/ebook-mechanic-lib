#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=== Test Coverage Verification ==="
echo

cd "$PROJECT_ROOT"

echo "Step 1: Checking all error codes are tested..."
echo

EPUB_ERROR_CODES=()
    "EPUB-CONTAINER-001"
    "EPUB-CONTAINER-002"
    "EPUB-CONTAINER-003"
    "EPUB-CONTAINER-004"
    "EPUB-CONTAINER-005"
    "EPUB-OPF-001"
    "EPUB-OPF-002"
    "EPUB-OPF-003"
    "EPUB-OPF-004"
    "EPUB-OPF-005"
    "EPUB-OPF-009"
    "EPUB-NAV-002"
    "EPUB-NAV-006"
    "EPUB-CONTENT-002"
    "EPUB-CONTENT-007"
)

PDF_ERROR_CODES=()
    "PDF-HEADER-001"
    "PDF-HEADER-002"
    "PDF-TRAILER-001"
    "PDF-TRAILER-003"
    "PDF-XREF-001"
    "PDF-CATALOG-003"
    "PDF-STRUCTURE-012"
)

missing_tests=0

for code in "${EPUB_ERROR_CODES[@]}"; do
    if ! grep -r "$code" tests/integration/*.go > /dev/null 2>&1; then
        echo "  ✗ Missing test for $code"
        missing_tests=$((missing_tests + 1))
    else
        echo "  ✓ $code"
    fi
done

for code in "${PDF_ERROR_CODES[@]}"; do
    if ! grep -r "$code" tests/integration/*.go > /dev/null 2>&1; then
        echo "  ✗ Missing test for $code"
        missing_tests=$((missing_tests + 1))
    else
        echo "  ✓ $code"
    fi
done

echo
if [ $missing_tests -eq 0 ]; then
    echo "✓ All error codes have tests"
else
    echo "✗ $missing_tests error codes missing tests"
    exit 1
fi

echo
echo "Step 2: Checking test fixtures exist..."
echo

missing_fixtures=0

REQUIRED_FIXTURES=()
    "testdata/epub/valid/minimal.epub"
    "testdata/epub/invalid/not_zip.epub"
    "testdata/epub/invalid/wrong_mimetype.epub"
    "testdata/pdf/valid/minimal.pdf"
    "testdata/pdf/invalid/not_pdf.pdf"
    "testdata/pdf/invalid/no_header.pdf"
)

for fixture in "${REQUIRED_FIXTURES[@]}"; do
    if [ ! -f "$fixture" ]; then
        echo "  ✗ Missing fixture: $fixture"
        missing_fixtures=$((missing_fixtures + 1))
    else
        echo "  ✓ $fixture"
    fi
done

echo
if [ $missing_fixtures -eq 0 ]; then
    echo "✓ All required fixtures exist"
else
    echo "✗ $missing_fixtures fixtures missing"
    echo "Run: make generate-fixtures"
    exit 1
fi

echo
echo "Step 3: Running coverage analysis..."
echo

go test -coverprofile=coverage.out -covermode=atomic ./... > /dev/null 2>&1

total_coverage=$(go tool cover -func=coverage.out | tail -n 1 | awk '{print $3}' | sed 's/%//')

echo "Total coverage: ${total_coverage}%"

if (( $(echo "$total_coverage >= 80.0" | bc -l) )); then
    echo "✓ Coverage target met (≥80%)"
else
    echo "✗ Coverage target not met (${total_coverage}% < 80%)"
    echo
    echo "Packages with low coverage:"
    go tool cover -func=coverage.out | awk '$3 < 80.0 {print "  " $1 ": " $3}'
fi

echo
echo "Step 4: Checking benchmark tests..."
echo

if grep -r "^func Benchmark" tests/integration/*.go > /dev/null 2>&1; then
    bench_count=$(grep -r "^func Benchmark" tests/integration/*.go | wc -l)
    echo "✓ Found $bench_count benchmark tests"
else
    echo "✗ No benchmark tests found"
    exit 1
fi

echo
echo "=== Verification Complete ==="

if [ $missing_tests -eq 0 ] && [ $missing_fixtures -eq 0 ]; then
    echo "✓ Test suite is complete and ready"
    exit 0
else
    echo "✗ Test suite has issues that need to be addressed"
    exit 1
fi
