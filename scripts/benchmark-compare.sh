#!/bin/bash
# Script to compare benchmark results and detect performance regressions
# Usage: ./scripts/benchmark-compare.sh [old_results.txt] [new_results.txt]

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

OLD_RESULTS="${1:-benchmarks-baseline.txt}"
NEW_RESULTS="${2:-benchmarks-new.txt}"

# Thresholds for regression detection
TIME_THRESHOLD=1.20    # 20% slower
MEMORY_THRESHOLD=1.30  # 30% more memory
ALLOCS_THRESHOLD=1.40  # 40% more allocations

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Performance Benchmark Comparison"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  Old results: ${OLD_RESULTS}"
echo "  New results: ${NEW_RESULTS}"
echo ""
echo "  Regression thresholds:"
echo "    Time:        ${TIME_THRESHOLD}x (20% slower)"
echo "    Memory:      ${MEMORY_THRESHOLD}x (30% increase)"
echo "    Allocations: ${ALLOCS_THRESHOLD}x (40% increase)"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if files exist
if [[ ! -f "$OLD_RESULTS" ]]; then
    echo -e "${YELLOW}Warning: Baseline file not found: ${OLD_RESULTS}${NC}"
    echo "Creating baseline from new results..."
    cp "$NEW_RESULTS" "$OLD_RESULTS"
    echo -e "${GREEN}✓ Baseline created${NC}"
    exit 0
fi

if [[ ! -f "$NEW_RESULTS" ]]; then
    echo -e "${RED}Error: New results file not found: ${NEW_RESULTS}${NC}"
    exit 1
fi

# Check if benchstat is available
if ! command -v benchstat &> /dev/null; then
    echo -e "${YELLOW}Warning: benchstat not found, installing...${NC}"
    go install golang.org/x/perf/cmd/benchstat@latest
    if ! command -v benchstat &> /dev/null; then
        echo -e "${RED}Error: Failed to install benchstat${NC}"
        echo "Please install manually: go install golang.org/x/perf/cmd/benchstat@latest"
        exit 1
    fi
fi

# Run benchstat comparison
echo "Running benchstat analysis..."
echo ""

BENCHSTAT_OUTPUT=$(benchstat "$OLD_RESULTS" "$NEW_RESULTS" 2>&1)
echo "$BENCHSTAT_OUTPUT"
echo ""

# Parse benchstat output for regressions
# This is a simplified check - benchstat output format may vary
REGRESSIONS_FOUND=0

# Extract delta percentages and check for regressions
# Look for lines with significant negative deltas (regressions)
while IFS= read -r line; do
    # Check for time regression (lines containing "time/op")
    if [[ "$line" =~ time/op.*\+([0-9]+\.[0-9]+)% ]]; then
        PERCENT="${BASH_REMATCH[1]}"
        RATIO=$(echo "scale=2; 1 + $PERCENT / 100" | bc)
        if (( $(echo "$RATIO > $TIME_THRESHOLD" | bc -l) )); then
            echo -e "${RED}⚠ Time regression detected: +${PERCENT}%${NC}"
            REGRESSIONS_FOUND=$((REGRESSIONS_FOUND + 1))
        fi
    fi
    
    # Check for memory regression (lines containing "alloc/op")
    if [[ "$line" =~ alloc/op.*\+([0-9]+\.[0-9]+)% ]]; then
        PERCENT="${BASH_REMATCH[1]}"
        RATIO=$(echo "scale=2; 1 + $PERCENT / 100" | bc)
        if (( $(echo "$RATIO > $MEMORY_THRESHOLD" | bc -l) )); then
            echo -e "${RED}⚠ Memory regression detected: +${PERCENT}%${NC}"
            REGRESSIONS_FOUND=$((REGRESSIONS_FOUND + 1))
        fi
    fi
    
    # Check for allocation regression (lines containing "allocs/op")
    if [[ "$line" =~ allocs/op.*\+([0-9]+\.[0-9]+)% ]]; then
        PERCENT="${BASH_REMATCH[1]}"
        RATIO=$(echo "scale=2; 1 + $PERCENT / 100" | bc)
        if (( $(echo "$RATIO > $ALLOCS_THRESHOLD" | bc -l) )); then
            echo -e "${RED}⚠ Allocation regression detected: +${PERCENT}%${NC}"
            REGRESSIONS_FOUND=$((REGRESSIONS_FOUND + 1))
        fi
    fi
done <<< "$BENCHSTAT_OUTPUT"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ $REGRESSIONS_FOUND -eq 0 ]]; then
    echo -e "${GREEN}✓ No significant performance regressions detected${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo -e "${RED}✗ ${REGRESSIONS_FOUND} performance regression(s) detected${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Please review the performance changes above."
    echo "If the regressions are expected, update the baseline:"
    echo "  cp ${NEW_RESULTS} ${OLD_RESULTS}"
    echo ""
    exit 1
fi
