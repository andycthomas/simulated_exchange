#!/bin/bash
################################################################################
# Chaos Experiment 11: Comprehensive Flamegraph Profiling
#
# This experiment generates comprehensive CPU flamegraphs under various
# load conditions, including during chaos events. It demonstrates the
# integration of profiling with chaos engineering.
#
# Duration: ~90 seconds
# Root Required: No
# Impact: Performance profiling overhead (~5%)
################################################################################

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
EXPERIMENT_NAME="flamegraph-profiling"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
RESULTS_DIR="$PROJECT_ROOT/chaos-results/$EXPERIMENT_NAME"
LOG_FILE="$RESULTS_DIR/experiment.log"
FLAMEGRAPH_DIR="$RESULTS_DIR/flamegraphs"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
TRADING_API_PORT=8080
MARKET_SIM_PORT=8081
PROFILE_DURATION=30

log() {
    echo -e "${CYAN}[$(date +'%H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[$(date +'%H:%M:%S')] ✓${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[$(date +'%H:%M:%S')] ✗${NC} $1" | tee -a "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}[$(date +'%H:%M:%S')] ⚠${NC} $1" | tee -a "$LOG_FILE"
}

log_info() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')] ℹ${NC} $1" | tee -a "$LOG_FILE"
}

# Setup
mkdir -p "$RESULTS_DIR" "$FLAMEGRAPH_DIR"

echo "╔═══════════════════════════════════════════════════════════════════════════╗" | tee "$LOG_FILE"
echo "║           CHAOS EXPERIMENT 11: FLAMEGRAPH PROFILING                      ║" | tee -a "$LOG_FILE"
echo "╚═══════════════════════════════════════════════════════════════════════════╝" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

log "Experiment started at: $(date)"
log "Results will be saved to: $RESULTS_DIR"

# Check if services are running
log_info "Checking service health..."
if ! curl -s -f http://localhost:$TRADING_API_PORT/health > /dev/null 2>&1; then
    log_error "Trading API not responding on port $TRADING_API_PORT"
    log_error "Start services with: docker compose up -d"
    exit 1
fi
log_success "Trading API is healthy"

if ! curl -s -f http://localhost:$MARKET_SIM_PORT/health > /dev/null 2>&1; then
    log_warn "Market Simulator not responding (optional)"
else
    log_success "Market Simulator is healthy"
fi

echo "" | tee -a "$LOG_FILE"
log_info "This experiment will:"
log_info "  1. Capture baseline CPU profile (30s)"
log_info "  2. Generate load and capture under-stress profile (30s)"
log_info "  3. Inject chaos (DB kill) and profile recovery (30s)"
log_info "  4. Generate flamegraph visualizations"
log_info "  5. Analyze profiles with AI (if available)"
echo "" | tee -a "$LOG_FILE"

# Phase 1: Baseline Profile
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_FILE"
log "${MAGENTA}Phase 1: Baseline CPU Profile${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_FILE"

log "Capturing baseline CPU profile for ${PROFILE_DURATION}s..."
log "URL: http://localhost:$TRADING_API_PORT/debug/pprof/profile?seconds=$PROFILE_DURATION"

BASELINE_PROFILE="$FLAMEGRAPH_DIR/cpu_baseline_${TIMESTAMP}.pb.gz"
if curl -s "http://localhost:$TRADING_API_PORT/debug/pprof/profile?seconds=$PROFILE_DURATION" -o "$BASELINE_PROFILE" 2>&1 | tee -a "$LOG_FILE"; then
    log_success "Baseline profile captured: $(du -h "$BASELINE_PROFILE" | cut -f1)"
else
    log_error "Failed to capture baseline profile"
fi

# Phase 2: Load Profile
echo "" | tee -a "$LOG_FILE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_FILE"
log "${MAGENTA}Phase 2: Under-Load CPU Profile${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_FILE"

log "Starting load generation in background..."

# Start load generation
(
    for i in {1..100}; do
        curl -s -X POST http://localhost:$TRADING_API_PORT/api/orders \
            -H "Content-Type: application/json" \
            -d '{"symbol":"AAPL","side":"buy","quantity":100,"price":150.00}' > /dev/null 2>&1 &
        sleep 0.3
    done
) &
LOAD_PID=$!

sleep 2
log_success "Load generation started (PID: $LOAD_PID)"

log "Capturing under-load CPU profile for ${PROFILE_DURATION}s..."
LOAD_PROFILE="$FLAMEGRAPH_DIR/cpu_load_${TIMESTAMP}.pb.gz"
if curl -s "http://localhost:$TRADING_API_PORT/debug/pprof/profile?seconds=$PROFILE_DURATION" -o "$LOAD_PROFILE" 2>&1 | tee -a "$LOG_FILE"; then
    log_success "Load profile captured: $(du -h "$LOAD_PROFILE" | cut -f1)"
else
    log_error "Failed to capture load profile"
fi

# Stop load generation
kill $LOAD_PID 2>/dev/null || true
wait $LOAD_PID 2>/dev/null || true
log_success "Load generation stopped"

# Phase 3: Chaos + Recovery Profile
echo "" | tee -a "$LOG_FILE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_FILE"
log "${MAGENTA}Phase 3: Chaos Injection + Recovery Profile${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_FILE"

log "Starting chaos profile capture..."
CHAOS_PROFILE="$FLAMEGRAPH_DIR/cpu_chaos_${TIMESTAMP}.pb.gz"

# Start profile capture in background
(
    sleep 5
    curl -s "http://localhost:$TRADING_API_PORT/debug/pprof/profile?seconds=$PROFILE_DURATION" -o "$CHAOS_PROFILE"
) &
PROFILE_PID=$!

sleep 2
log "Injecting chaos: Killing database..."

# Kill database
cd "$PROJECT_ROOT"
if docker compose ps postgres 2>/dev/null | grep -q "Up\|running"; then
    docker compose kill postgres 2>&1 | tee -a "$LOG_FILE"
    log_success "Database killed"
else
    log_warn "Database container not found, skipping kill"
fi

sleep 5
log "Restarting database..."
docker compose up -d postgres 2>&1 | tee -a "$LOG_FILE"
sleep 5

log "Waiting for recovery and profile completion..."
wait $PROFILE_PID 2>/dev/null || true

if [[ -f "$CHAOS_PROFILE" ]]; then
    log_success "Chaos profile captured: $(du -h "$CHAOS_PROFILE" | cut -f1)"
else
    log_error "Failed to capture chaos profile"
fi

# Phase 4: Generate Flamegraphs
echo "" | tee -a "$LOG_FILE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_FILE"
log "${MAGENTA}Phase 4: Generating Flamegraph Visualizations${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_FILE"

# Check if pprof is available
if ! command -v go &> /dev/null; then
    log_error "Go not found, cannot generate flamegraphs"
else
    log "Generating flamegraphs with go tool pprof..."

    for profile in "$BASELINE_PROFILE" "$LOAD_PROFILE" "$CHAOS_PROFILE"; do
        if [[ -f "$profile" ]]; then
            PROFILE_NAME=$(basename "$profile" .pb.gz)
            SVG_FILE="$FLAMEGRAPH_DIR/${PROFILE_NAME}.svg"

            log "  Generating $PROFILE_NAME.svg..."
            if go tool pprof -svg "$profile" > "$SVG_FILE" 2>&1 | tee -a "$LOG_FILE"; then
                log_success "  Created $SVG_FILE ($(du -h "$SVG_FILE" | cut -f1))"
            else
                log_warn "  Failed to generate SVG for $PROFILE_NAME"
            fi
        fi
    done
fi

# Capture additional profile types (quick snapshots)
log "Capturing additional profile types..."

PROFILE_TYPES=("heap" "goroutine" "allocs" "block" "mutex")
for ptype in "${PROFILE_TYPES[@]}"; do
    PROFILE_FILE="$FLAMEGRAPH_DIR/${ptype}_${TIMESTAMP}.pb.gz"

    if [[ "$ptype" == "allocs" ]]; then
        # Allocs needs a duration
        curl -s "http://localhost:$TRADING_API_PORT/debug/pprof/allocs?seconds=5" -o "$PROFILE_FILE" 2>&1 | tee -a "$LOG_FILE" || true
    else
        curl -s "http://localhost:$TRADING_API_PORT/debug/pprof/$ptype" -o "$PROFILE_FILE" 2>&1 | tee -a "$LOG_FILE" || true
    fi

    if [[ -f "$PROFILE_FILE" && -s "$PROFILE_FILE" ]]; then
        SVG_FILE="$FLAMEGRAPH_DIR/${ptype}_${TIMESTAMP}.svg"
        go tool pprof -svg "$PROFILE_FILE" > "$SVG_FILE" 2>/dev/null || true
        if [[ -f "$SVG_FILE" ]]; then
            log_success "  $ptype profile captured and visualized"
        fi
    fi
done

# Phase 5: Analysis
echo "" | tee -a "$LOG_FILE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_FILE"
log "${MAGENTA}Phase 5: Profile Analysis${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_FILE"

# Generate text analysis
log "Generating text analysis of profiles..."

for profile in "$BASELINE_PROFILE" "$LOAD_PROFILE" "$CHAOS_PROFILE"; do
    if [[ -f "$profile" ]]; then
        PROFILE_NAME=$(basename "$profile" .pb.gz)
        TXT_FILE="$FLAMEGRAPH_DIR/${PROFILE_NAME}_top.txt"

        log "  Analyzing $PROFILE_NAME..."
        echo "Top 20 functions by CPU time for $PROFILE_NAME:" > "$TXT_FILE"
        echo "=====================================================" >> "$TXT_FILE"
        go tool pprof -top -nodecount=20 "$profile" >> "$TXT_FILE" 2>&1 || true

        if [[ -f "$TXT_FILE" ]]; then
            log_success "  Analysis saved to ${PROFILE_NAME}_top.txt"
        fi
    fi
done

# Try AI analysis if available
if [[ -f "$PROJECT_ROOT/scripts/ai-flamegraph-analyzer.py" ]]; then
    log "Running AI-powered analysis..."
    if command -v python3 &> /dev/null; then
        python3 "$PROJECT_ROOT/scripts/ai-flamegraph-analyzer.py" "$LOAD_PROFILE" 2>&1 | tee -a "$LOG_FILE" || log_warn "AI analysis failed"
    else
        log_warn "Python3 not found, skipping AI analysis"
    fi
else
    log_info "AI analyzer not found (optional feature)"
fi

# Summary
echo "" | tee -a "$LOG_FILE"
echo "╔═══════════════════════════════════════════════════════════════════════════╗" | tee -a "$LOG_FILE"
echo "║                           EXPERIMENT SUMMARY                              ║" | tee -a "$LOG_FILE"
echo "╚═══════════════════════════════════════════════════════════════════════════╝" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

log_success "Experiment completed successfully!"
echo "" | tee -a "$LOG_FILE"

log_info "Generated Artifacts:"
ls -lh "$FLAMEGRAPH_DIR"/*.svg 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}' | tee -a "$LOG_FILE" || log_warn "No SVG files generated"

echo "" | tee -a "$LOG_FILE"
log_info "View Results:"
log_info "  All files:  ls -lh $FLAMEGRAPH_DIR/"
log_info "  In browser: firefox $FLAMEGRAPH_DIR/*.svg"
log_info "  Via Make:   make profile-view"
echo "" | tee -a "$LOG_FILE"

log_info "Profile Analysis:"
log_info "  Interactive: go tool pprof $LOAD_PROFILE"
log_info "  Top funcs:   go tool pprof -top $LOAD_PROFILE | head -20"
log_info "  Text files:  cat $FLAMEGRAPH_DIR/*_top.txt"
echo "" | tee -a "$LOG_FILE"

log "Full experiment log: $LOG_FILE"

# Copy flamegraphs to main flamegraphs directory for easy access
log_info "Copying flamegraphs to main directory..."
mkdir -p "$PROJECT_ROOT/flamegraphs"
cp "$FLAMEGRAPH_DIR"/*.svg "$PROJECT_ROOT/flamegraphs/" 2>/dev/null || true
cp "$FLAMEGRAPH_DIR"/*.pb.gz "$PROJECT_ROOT/flamegraphs/" 2>/dev/null || true
log_success "Flamegraphs available in: $PROJECT_ROOT/flamegraphs/"

echo "" | tee -a "$LOG_FILE"
log_success "✅ Flamegraph profiling chaos experiment complete!"
exit 0
