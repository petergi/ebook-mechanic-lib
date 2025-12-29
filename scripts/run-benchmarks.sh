#!/bin/bash
# Script to run performance benchmarks and save results
# Usage: ./scripts/run-benchmarks.sh [output_file] [run_count]

set -euo pipefail

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

OUTPUT_FILE="${1:-benchmarks-$(date +%Y%m%d-%H%M%S).txt}"
RUN_COUNT="${2:-5}"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Running Performance Benchmarks"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  Output file: ${OUTPUT_FILE}"
echo "  Run count:   ${RUN_COUNT}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Ensure test fixtures exist
echo -e "${BLUE}→ Ensuring test fixtures are generated...${NC}"
make generate-fixtures
echo -e "${GREEN}✓ Test fixtures ready${NC}"
echo ""

# Run benchmarks
echo -e "${BLUE}→ Running benchmarks (this may take several minutes)...${NC}"
echo ""

go test \
    -bench=. \
    -benchmem \
    -benchtime=1s \
    -count="$RUN_COUNT" \
    -timeout=30m \
    ./tests/integration/... \
    | tee "$OUTPUT_FILE"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ Benchmarks complete${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Results saved to: ${OUTPUT_FILE}"
echo ""
echo "To analyze results:"
echo "  cat ${OUTPUT_FILE}"
echo ""
echo "To compare with baseline:"
echo "  ./scripts/benchmark-compare.sh benchmarks-baseline.txt ${OUTPUT_FILE}"
echo ""
echo "To set this as new baseline:"
echo "  cp ${OUTPUT_FILE} benchmarks-baseline.txt"
echo ""
