#!/bin/bash
################################################################################
# Run All Chaos Engineering Experiments
#
# This script runs all chaos experiments sequentially with proper cleanup
# between each experiment.
#
# Duration: ~90 minutes
# Root Required: Yes (for most experiments)
################################################################################

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
SUMMARY_LOG="/tmp/chaos-all-experiments-$TIMESTAMP.log"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

echo "╔═══════════════════════════════════════════════════════════════════════════╗" | tee "$SUMMARY_LOG"
echo "║              RUN ALL CHAOS ENGINEERING EXPERIMENTS                        ║" | tee -a "$SUMMARY_LOG"
echo "╚═══════════════════════════════════════════════════════════════════════════╝" | tee -a "$SUMMARY_LOG"
echo "" | tee -a "$SUMMARY_LOG"

echo -e "${CYAN}Started at: $(date)${NC}" | tee -a "$SUMMARY_LOG"
echo -e "${CYAN}Summary log: $SUMMARY_LOG${NC}" | tee -a "$SUMMARY_LOG"
echo "" | tee -a "$SUMMARY_LOG"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    echo -e "${YELLOW}⚠️  Warning: Not running as root${NC}" | tee -a "$SUMMARY_LOG"
    echo -e "${YELLOW}Some experiments may fail or require sudo${NC}" | tee -a "$SUMMARY_LOG"
    echo "" | tee -a "$SUMMARY_LOG"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted." | tee -a "$SUMMARY_LOG"
        exit 1
    fi
fi

# List of experiments
EXPERIMENTS=(
    "00-preflight-check.sh"
    "01-database-sudden-death.sh"
    "02-network-partition.sh"
    "03-memory-oom-kill.sh"
    "04-cpu-throttling.sh"
    "05-disk-io-starvation.sh"
    "06-connection-pool-exhaustion.sh"
    "07-redis-cache-failure.sh"
    "08-nginx-failure.sh"
    "09-network-latency.sh"
    "10-multi-service-failure.sh"
    "11-flamegraph-profiling.sh"
)

PASSED=0
FAILED=0
SKIPPED=0

declare -A RESULTS

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}" | tee -a "$SUMMARY_LOG"
echo -e "${BLUE}Experiments to run: ${#EXPERIMENTS[@]}${NC}" | tee -a "$SUMMARY_LOG"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}" | tee -a "$SUMMARY_LOG"
echo "" | tee -a "$SUMMARY_LOG"

for experiment in "${EXPERIMENTS[@]}"; do
    SCRIPT_PATH="$SCRIPT_DIR/$experiment"

    if [[ ! -f "$SCRIPT_PATH" ]]; then
        echo -e "${YELLOW}⚠  Skipping $experiment (not found)${NC}" | tee -a "$SUMMARY_LOG"
        RESULTS["$experiment"]="SKIPPED (not found)"
        ((SKIPPED++))
        continue
    fi

    if [[ ! -x "$SCRIPT_PATH" ]]; then
        echo -e "${YELLOW}⚠  Skipping $experiment (not executable)${NC}" | tee -a "$SUMMARY_LOG"
        RESULTS["$experiment"]="SKIPPED (not executable)"
        ((SKIPPED++))
        continue
    fi

    echo "" | tee -a "$SUMMARY_LOG"
    echo -e "${MAGENTA}╔═══════════════════════════════════════════════════════════════════════════╗${NC}" | tee -a "$SUMMARY_LOG"
    echo -e "${MAGENTA}║  Running: $experiment${NC}" | tee -a "$SUMMARY_LOG"
    echo -e "${MAGENTA}╚═══════════════════════════════════════════════════════════════════════════╝${NC}" | tee -a "$SUMMARY_LOG"
    echo "" | tee -a "$SUMMARY_LOG"

    START_TIME=$(date +%s)

    if "$SCRIPT_PATH" 2>&1 | tee -a "$SUMMARY_LOG"; then
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))
        echo -e "${GREEN}✓ $experiment completed successfully (${DURATION}s)${NC}" | tee -a "$SUMMARY_LOG"
        RESULTS["$experiment"]="PASSED (${DURATION}s)"
        ((PASSED++))
    else
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))
        echo -e "${RED}✗ $experiment failed (${DURATION}s)${NC}" | tee -a "$SUMMARY_LOG"
        RESULTS["$experiment"]="FAILED (${DURATION}s)"
        ((FAILED++))

        echo "" | tee -a "$SUMMARY_LOG"
        read -p "Continue with remaining experiments? (Y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Nn]$ ]]; then
            echo "Stopping experiment run." | tee -a "$SUMMARY_LOG"
            break
        fi
    fi

    # Cleanup and recovery time
    if [[ "$experiment" != "00-preflight-check.sh" && "$experiment" != "11-flamegraph-profiling.sh" ]]; then
        echo "" | tee -a "$SUMMARY_LOG"
        echo -e "${CYAN}⏱  Waiting 30 seconds for system recovery...${NC}" | tee -a "$SUMMARY_LOG"
        sleep 30
    fi
done

# Final Summary
echo "" | tee -a "$SUMMARY_LOG"
echo "╔═══════════════════════════════════════════════════════════════════════════╗" | tee -a "$SUMMARY_LOG"
echo "║                        FINAL SUMMARY                                      ║" | tee -a "$SUMMARY_LOG"
echo "╚═══════════════════════════════════════════════════════════════════════════╝" | tee -a "$SUMMARY_LOG"
echo "" | tee -a "$SUMMARY_LOG"

echo -e "Completed at: ${CYAN}$(date)${NC}" | tee -a "$SUMMARY_LOG"
echo "" | tee -a "$SUMMARY_LOG"

echo -e "${GREEN}Passed:${NC}  $PASSED" | tee -a "$SUMMARY_LOG"
echo -e "${RED}Failed:${NC}  $FAILED" | tee -a "$SUMMARY_LOG"
echo -e "${YELLOW}Skipped:${NC} $SKIPPED" | tee -a "$SUMMARY_LOG"
echo "" | tee -a "$SUMMARY_LOG"

echo "Detailed Results:" | tee -a "$SUMMARY_LOG"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$SUMMARY_LOG"

for experiment in "${EXPERIMENTS[@]}"; do
    if [[ -n "${RESULTS[$experiment]:-}" ]]; then
        STATUS="${RESULTS[$experiment]}"
        if [[ "$STATUS" =~ ^PASSED ]]; then
            echo -e "  ${GREEN}✓${NC} $experiment: $STATUS" | tee -a "$SUMMARY_LOG"
        elif [[ "$STATUS" =~ ^FAILED ]]; then
            echo -e "  ${RED}✗${NC} $experiment: $STATUS" | tee -a "$SUMMARY_LOG"
        else
            echo -e "  ${YELLOW}⚠${NC} $experiment: $STATUS" | tee -a "$SUMMARY_LOG"
        fi
    fi
done

echo "" | tee -a "$SUMMARY_LOG"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$SUMMARY_LOG"
echo "" | tee -a "$SUMMARY_LOG"

echo -e "${CYAN}Full summary saved to: $SUMMARY_LOG${NC}" | tee -a "$SUMMARY_LOG"
echo "" | tee -a "$SUMMARY_LOG"

# Individual experiment logs
echo "Individual experiment logs:" | tee -a "$SUMMARY_LOG"
ls -lht /tmp/chaos-exp-*.log 2>/dev/null | head -10 | tee -a "$SUMMARY_LOG" || echo "No logs found" | tee -a "$SUMMARY_LOG"

echo "" | tee -a "$SUMMARY_LOG"

if [[ $FAILED -eq 0 ]]; then
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════════════════════╗${NC}" | tee -a "$SUMMARY_LOG"
    echo -e "${GREEN}║  ✅ ALL EXPERIMENTS COMPLETED SUCCESSFULLY!                              ║${NC}" | tee -a "$SUMMARY_LOG"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════════════════╝${NC}" | tee -a "$SUMMARY_LOG"
    exit 0
else
    echo -e "${YELLOW}╔═══════════════════════════════════════════════════════════════════════════╗${NC}" | tee -a "$SUMMARY_LOG"
    echo -e "${YELLOW}║  ⚠️  SOME EXPERIMENTS FAILED                                             ║${NC}" | tee -a "$SUMMARY_LOG"
    echo -e "${YELLOW}╚═══════════════════════════════════════════════════════════════════════════╝${NC}" | tee -a "$SUMMARY_LOG"
    exit 1
fi
