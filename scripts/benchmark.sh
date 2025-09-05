#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BENCHMARK_DIR="$PROJECT_ROOT/test/benchmark"
RESULTS_DIR="$PROJECT_ROOT/benchmark-results"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create results directory
mkdir -p "$RESULTS_DIR"

# Generate timestamp for results
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULTS_FILE="$RESULTS_DIR/benchmark_$TIMESTAMP.txt"
JSON_RESULTS_FILE="$RESULTS_DIR/benchmark_$TIMESTAMP.json"

echo -e "${BLUE}🚀 Agent Manager Query System Performance Benchmark${NC}"
echo -e "${BLUE}=================================================${NC}"
echo ""

# Function to run benchmark and capture results
run_benchmark() {
    local test_name=$1
    local benchmark_func=$2

    echo -e "${YELLOW}Running $test_name...${NC}"

    # Run benchmark with detailed output
    cd "$PROJECT_ROOT"
    go test -bench="$benchmark_func" -benchmem -count=5 -timeout=30m \
        ./test/benchmark/... 2>&1 | tee -a "$RESULTS_FILE"

    echo ""
}

# Function to check performance requirements
check_performance_requirements() {
    echo -e "${BLUE}Analyzing Performance Requirements...${NC}"
    echo -e "${BLUE}===================================${NC}"

    local max_query_time_ms=50
    local max_memory_mb=50
    local target_agents=1000

    echo "Performance Targets:"
    echo "- Query Response Time: <${max_query_time_ms}ms"
    echo "- Memory Usage: <${max_memory_mb}MB for ${target_agents} agents"
    echo "- Concurrent Query Support: ✓"
    echo "- Index Build Performance: Measured"
    echo "- Cache Performance: Measured"
    echo ""

    # Extract key metrics from results
    if [[ -f "$RESULTS_FILE" ]]; then
        echo "Results Summary:" | tee -a "$RESULTS_FILE"
        echo "================" | tee -a "$RESULTS_FILE"

        # Parse query latency results
        local query_times=$(grep -E "BenchmarkQueryPerformance.*-[0-9]+" "$RESULTS_FILE" | awk '{print $3}' | sed 's/ns\/op//')
        if [[ -n "$query_times" ]]; then
            local avg_ns=$(echo "$query_times" | awk '{sum+=$1; count++} END {if(count>0) print int(sum/count)}')
            local avg_ms=$(echo "scale=2; $avg_ns / 1000000" | bc -l 2>/dev/null || echo "N/A")

            echo "Average Query Time: ${avg_ms}ms" | tee -a "$RESULTS_FILE"

            if (( $(echo "$avg_ms < $max_query_time_ms" | bc -l 2>/dev/null || echo 0) )); then
                echo -e "${GREEN}✓ Query time requirement MET${NC}" | tee -a "$RESULTS_FILE"
            else
                echo -e "${RED}✗ Query time requirement FAILED${NC}" | tee -a "$RESULTS_FILE"
            fi
        fi

        # Parse memory usage
        local memory_usage=$(grep -E "MB$" "$RESULTS_FILE" | head -1 | awk '{print $NF}' | sed 's/MB//')
        if [[ -n "$memory_usage" ]]; then
            echo "Memory Usage: ${memory_usage}MB" | tee -a "$RESULTS_FILE"

            if (( $(echo "$memory_usage < $max_memory_mb" | bc -l 2>/dev/null || echo 0) )); then
                echo -e "${GREEN}✓ Memory usage requirement MET${NC}" | tee -a "$RESULTS_FILE"
            else
                echo -e "${RED}✗ Memory usage requirement FAILED${NC}" | tee -a "$RESULTS_FILE"
            fi
        fi

        echo "" | tee -a "$RESULTS_FILE"
    fi
}

# Function to generate JSON results for CI/CD
generate_json_results() {
    echo -e "${BLUE}Generating JSON Results...${NC}"

    cat > "$JSON_RESULTS_FILE" << EOF
{
  "timestamp": "$TIMESTAMP",
  "benchmark_results": {
    "query_performance": {
      "target_response_time_ms": 50,
      "target_memory_mb": 50,
      "target_agents": 1000
    },
    "results_file": "$RESULTS_FILE",
    "system_info": {
      "go_version": "$(go version)",
      "os": "$(uname -s)",
      "arch": "$(uname -m)"
    }
  }
}
EOF

    echo "JSON results saved to: $JSON_RESULTS_FILE"
}

# Function to run memory profiling
run_memory_profile() {
    echo -e "${BLUE}Running Memory Profile...${NC}"

    cd "$PROJECT_ROOT"
    go test -bench=BenchmarkMemoryUsage -memprofile="$RESULTS_DIR/mem_$TIMESTAMP.prof" \
        -cpuprofile="$RESULTS_DIR/cpu_$TIMESTAMP.prof" ./test/benchmark/... 2>&1 | tee -a "$RESULTS_FILE"

    echo "Memory profile saved to: $RESULTS_DIR/mem_$TIMESTAMP.prof"
    echo "CPU profile saved to: $RESULTS_DIR/cpu_$TIMESTAMP.prof"
    echo ""
}

# Main benchmark execution
main() {
    echo "Benchmark started at: $(date)" | tee "$RESULTS_FILE"
    echo "System: $(uname -a)" | tee -a "$RESULTS_FILE"
    echo "Go version: $(go version)" | tee -a "$RESULTS_FILE"
    echo "" | tee -a "$RESULTS_FILE"

    # Run all benchmarks
    run_benchmark "Query Performance Tests" "BenchmarkQueryPerformance"
    run_benchmark "Memory Usage Tests" "BenchmarkMemoryUsage"
    run_benchmark "Index Operation Tests" "BenchmarkIndexOperations"
    run_benchmark "Cache Performance Tests" "BenchmarkCachePerformance"
    run_benchmark "Fuzzy Matching Tests" "BenchmarkFuzzyMatching"

    # Run memory profiling
    run_memory_profile

    # Analyze results
    check_performance_requirements

    # Generate JSON results
    generate_json_results

    echo "Benchmark completed at: $(date)" | tee -a "$RESULTS_FILE"
    echo "" | tee -a "$RESULTS_FILE"

    echo -e "${GREEN}Benchmark complete!${NC}"
    echo -e "Results saved to: ${YELLOW}$RESULTS_FILE${NC}"
    echo -e "JSON results: ${YELLOW}$JSON_RESULTS_FILE${NC}"

    # Display summary
    if [[ -f "$RESULTS_FILE" ]]; then
        echo ""
        echo -e "${BLUE}Quick Summary:${NC}"
        tail -20 "$RESULTS_FILE" | grep -E "(✓|✗|Average|Memory|Results)"
    fi
}

# Check dependencies
check_dependencies() {
    local missing_deps=()

    if ! command -v go >/dev/null 2>&1; then
        missing_deps+=("go")
    fi

    if ! command -v bc >/dev/null 2>&1; then
        missing_deps+=("bc")
    fi

    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        echo -e "${RED}Missing dependencies: ${missing_deps[*]}${NC}"
        echo "Please install the missing dependencies and try again."
        exit 1
    fi
}

# Handle command line arguments
case "${1:-}" in
    "--help"|"-h")
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --quick        Run quick benchmark (fewer iterations)"
        echo "  --profile      Run with detailed profiling"
        echo "  --clean        Clean previous benchmark results"
        echo ""
        exit 0
        ;;
    "--clean")
        echo "Cleaning previous benchmark results..."
        rm -rf "$RESULTS_DIR"
        echo "Cleaned."
        exit 0
        ;;
    "--quick")
        echo "Running quick benchmark..."
        export QUICK_BENCHMARK=1
        ;;
    "--profile")
        echo "Running with detailed profiling..."
        export DETAILED_PROFILE=1
        ;;
esac

# Check dependencies and run
check_dependencies
main

echo ""
echo -e "${GREEN}To view detailed results:${NC}"
echo -e "  cat ${YELLOW}$RESULTS_FILE${NC}"
echo ""
echo -e "${GREEN}To analyze profiles (if generated):${NC}"
echo -e "  go tool pprof ${YELLOW}$RESULTS_DIR/cpu_$TIMESTAMP.prof${NC}"
echo -e "  go tool pprof ${YELLOW}$RESULTS_DIR/mem_$TIMESTAMP.prof${NC}"