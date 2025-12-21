#!/bin/bash
# Walgo Deployment Benchmark Script
# Measures deployment performance with and without caching

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           Walgo Deployment Benchmark Suite                 ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if we're in a walgo project
if [ ! -f "walgo.yaml" ]; then
    echo -e "${RED}Error: walgo.yaml not found. Run this from a walgo project directory.${NC}"
    exit 1
fi

# Check if walgo is installed
if ! command -v walgo &> /dev/null; then
    echo -e "${RED}Error: walgo command not found. Please install walgo first.${NC}"
    exit 1
fi

# Configuration
ITERATIONS=${1:-3}
RESULTS_FILE="benchmark_results.json"

echo -e "${YELLOW}Configuration:${NC}"
echo "  Iterations per test: $ITERATIONS"
echo "  Results file: $RESULTS_FILE"
echo ""

# Function to measure time in milliseconds
measure_time() {
    local start=$(date +%s%N)
    "$@"
    local end=$(date +%s%N)
    echo $(( (end - start) / 1000000 ))
}

# Function to get directory size in bytes
get_dir_size() {
    du -sb "$1" 2>/dev/null | cut -f1
}

# Function to count files
count_files() {
    find "$1" -type f 2>/dev/null | wc -l
}

# Initialize results
echo "{" > $RESULTS_FILE
echo "  \"timestamp\": \"$(date -Iseconds)\"," >> $RESULTS_FILE
echo "  \"walgo_version\": \"$(walgo version --short 2>/dev/null || echo 'unknown')\"," >> $RESULTS_FILE

# Get project info
PUBLISH_DIR=$(grep -oP 'publishDir:\s*"\K[^"]+' walgo.yaml 2>/dev/null || echo "public")
echo "  \"publish_dir\": \"$PUBLISH_DIR\"," >> $RESULTS_FILE

echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Test 1: Initial Build (No Cache)${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"

# Clear cache
rm -rf .walgo/cache.db 2>/dev/null || true
rm -rf "$PUBLISH_DIR" 2>/dev/null || true

build_times_cold=()
for i in $(seq 1 $ITERATIONS); do
    echo -e "  ${YELLOW}Iteration $i/$ITERATIONS...${NC}"
    rm -rf .walgo/cache.db 2>/dev/null || true
    rm -rf "$PUBLISH_DIR" 2>/dev/null || true

    time_ms=$(measure_time walgo build 2>/dev/null)
    build_times_cold+=($time_ms)
    echo -e "    Time: ${BLUE}${time_ms}ms${NC}"
done

# Calculate average
avg_cold=0
for t in "${build_times_cold[@]}"; do
    avg_cold=$((avg_cold + t))
done
avg_cold=$((avg_cold / ${#build_times_cold[@]}))

echo -e "  ${GREEN}Average cold build time: ${avg_cold}ms${NC}"
echo ""

# Get build output stats
if [ -d "$PUBLISH_DIR" ]; then
    file_count=$(count_files "$PUBLISH_DIR")
    dir_size=$(get_dir_size "$PUBLISH_DIR")
    echo "  Files generated: $file_count"
    echo "  Total size: $(numfmt --to=iec $dir_size 2>/dev/null || echo "${dir_size} bytes")"
fi

echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Test 2: Incremental Build (With Cache)${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"

# First build to populate cache
walgo build 2>/dev/null

build_times_warm=()
for i in $(seq 1 $ITERATIONS); do
    echo -e "  ${YELLOW}Iteration $i/$ITERATIONS...${NC}"

    time_ms=$(measure_time walgo build 2>/dev/null)
    build_times_warm+=($time_ms)
    echo -e "    Time: ${BLUE}${time_ms}ms${NC}"
done

# Calculate average
avg_warm=0
for t in "${build_times_warm[@]}"; do
    avg_warm=$((avg_warm + t))
done
avg_warm=$((avg_warm / ${#build_times_warm[@]}))

echo -e "  ${GREEN}Average warm build time: ${avg_warm}ms${NC}"
echo ""

# Calculate improvement
if [ $avg_cold -gt 0 ]; then
    improvement=$(( (avg_cold - avg_warm) * 100 / avg_cold ))
    echo -e "  ${GREEN}Performance improvement: ${improvement}% faster with cache${NC}"
fi

echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Test 3: Deployment Dry-Run (Change Detection)${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"

# Measure deploy --dry-run times
deploy_times=()
for i in $(seq 1 $ITERATIONS); do
    echo -e "  ${YELLOW}Iteration $i/$ITERATIONS...${NC}"

    time_ms=$(measure_time walgo deploy --dry-run 2>/dev/null || echo "0")
    deploy_times+=($time_ms)
    echo -e "    Time: ${BLUE}${time_ms}ms${NC}"
done

# Calculate average
avg_deploy=0
for t in "${deploy_times[@]}"; do
    avg_deploy=$((avg_deploy + t))
done
if [ ${#deploy_times[@]} -gt 0 ]; then
    avg_deploy=$((avg_deploy / ${#deploy_times[@]}))
fi

echo -e "  ${GREEN}Average deploy analysis time: ${avg_deploy}ms${NC}"
echo ""

# Write results to JSON
echo "  \"results\": {" >> $RESULTS_FILE
echo "    \"cold_build\": {" >> $RESULTS_FILE
echo "      \"iterations\": [${build_times_cold[*]}]," >> $RESULTS_FILE
echo "      \"average_ms\": $avg_cold" >> $RESULTS_FILE
echo "    }," >> $RESULTS_FILE
echo "    \"warm_build\": {" >> $RESULTS_FILE
echo "      \"iterations\": [${build_times_warm[*]}]," >> $RESULTS_FILE
echo "      \"average_ms\": $avg_warm" >> $RESULTS_FILE
echo "    }," >> $RESULTS_FILE
echo "    \"deploy_analysis\": {" >> $RESULTS_FILE
echo "      \"iterations\": [${deploy_times[*]}]," >> $RESULTS_FILE
echo "      \"average_ms\": $avg_deploy" >> $RESULTS_FILE
echo "    }," >> $RESULTS_FILE
echo "    \"improvement_percent\": $improvement," >> $RESULTS_FILE
echo "    \"file_count\": $file_count," >> $RESULTS_FILE
echo "    \"total_size_bytes\": $dir_size" >> $RESULTS_FILE
echo "  }" >> $RESULTS_FILE
echo "}" >> $RESULTS_FILE

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}                     BENCHMARK SUMMARY                      ${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "  Cold Build (no cache):    ${YELLOW}${avg_cold}ms${NC}"
echo -e "  Warm Build (with cache):  ${GREEN}${avg_warm}ms${NC}"
echo -e "  Deploy Analysis:          ${BLUE}${avg_deploy}ms${NC}"
echo ""
echo -e "  ${GREEN}Cache Performance Gain: ${improvement}%${NC}"
echo ""
echo -e "  Results saved to: ${BLUE}$RESULTS_FILE${NC}"
echo ""

# Verify the claimed 50%+ improvement
if [ $improvement -ge 50 ]; then
    echo -e "  ${GREEN}✓ Verified: Cache provides 50%+ performance improvement${NC}"
else
    echo -e "  ${YELLOW}⚠ Note: Cache improvement is ${improvement}% (target: 50%+)${NC}"
    echo -e "  ${YELLOW}  This may vary based on site size and complexity${NC}"
fi
