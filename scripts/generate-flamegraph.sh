#!/bin/bash
################################################################################
# Flamegraph Generation Script
#
# Generates CPU, heap, goroutine, and other flamegraphs from Go pprof data
#
# Usage: ./generate-flamegraph.sh [type] [duration] [port]
#   type:     cpu, heap, goroutine, allocs, block, mutex
#   duration: seconds (0 for snapshot types)
#   port:     service port (default: 8080)
################################################################################

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Default values
PROFILE_TYPE="${1:-cpu}"
DURATION="${2:-30}"
PORT="${3:-8080}"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
OUTPUT_DIR="$PROJECT_ROOT/flamegraphs"

log() {
    echo -e "${CYAN}[$(date +'%H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[$(date +'%H:%M:%S')] ✓${NC} $1"
}

log_error() {
    echo -e "${RED}[$(date +'%H:%M:%S')] ✗${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[$(date +'%H:%M:%S')] ⚠${NC} $1"
}

log_info() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')] ℹ${NC} $1"
}

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo "╔═══════════════════════════════════════════════════════════════════════════╗"
echo "║                    FLAMEGRAPH GENERATOR                                   ║"
echo "╚═══════════════════════════════════════════════════════════════════════════╝"
echo ""

log_info "Configuration:"
log_info "  Profile Type: $PROFILE_TYPE"
log_info "  Duration:     ${DURATION}s"
log_info "  Port:         $PORT"
log_info "  Output Dir:   $OUTPUT_DIR"
echo ""

# Validate profile type
VALID_TYPES=("cpu" "heap" "goroutine" "allocs" "block" "mutex")
if [[ ! " ${VALID_TYPES[@]} " =~ " ${PROFILE_TYPE} " ]]; then
    log_error "Invalid profile type: $PROFILE_TYPE"
    log_info "Valid types: ${VALID_TYPES[*]}"
    exit 1
fi

# Check if service is running
log "Checking service health on port $PORT..."
if ! curl -s -f "http://localhost:$PORT/health" > /dev/null 2>&1; then
    log_error "Service not responding on port $PORT"
    log_info "Make sure the service is running: docker compose up -d"
    exit 1
fi
log_success "Service is healthy"

# Determine profile URL and parameters
case "$PROFILE_TYPE" in
    cpu)
        if [[ "$DURATION" -le 0 ]]; then
            log_error "$PROFILE_TYPE profiling requires duration > 0"
            exit 1
        fi
        PROFILE_URL="http://localhost:$PORT/debug/pprof/profile?seconds=$DURATION"
        ;;
    allocs)
        if [[ "$DURATION" -le 0 ]]; then
            log_error "$PROFILE_TYPE profiling requires duration > 0"
            exit 1
        fi
        PROFILE_URL="http://localhost:$PORT/debug/pprof/${PROFILE_TYPE}?seconds=$DURATION"
        ;;
    heap|goroutine|block|mutex)
        PROFILE_URL="http://localhost:$PORT/debug/pprof/${PROFILE_TYPE}"
        ;;
esac

# File paths
PROFILE_FILE="$OUTPUT_DIR/${PROFILE_TYPE}_profile_${TIMESTAMP}.pb.gz"
SVG_FILE="$OUTPUT_DIR/${PROFILE_TYPE}_${TIMESTAMP}.svg"
TXT_FILE="$OUTPUT_DIR/${PROFILE_TYPE}_${TIMESTAMP}_top.txt"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log "Phase 1: Capturing Profile Data"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ "$PROFILE_TYPE" == "cpu" || "$PROFILE_TYPE" == "allocs" ]]; then
    log "Capturing $PROFILE_TYPE profile for ${DURATION}s..."
    log_warn "This will take ${DURATION} seconds, please wait..."
else
    log "Capturing $PROFILE_TYPE snapshot..."
fi

if curl -s -f "$PROFILE_URL" -o "$PROFILE_FILE"; then
    if [[ -s "$PROFILE_FILE" ]]; then
        SIZE=$(du -h "$PROFILE_FILE" | cut -f1)
        log_success "Profile captured: $PROFILE_FILE ($SIZE)"
    else
        log_error "Profile file is empty"
        rm -f "$PROFILE_FILE"
        exit 1
    fi
else
    log_error "Failed to capture profile"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log "Phase 2: Generating Flamegraph Visualization"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if pprof is available
if ! command -v go &> /dev/null; then
    log_error "Go not found, cannot generate flamegraph"
    exit 1
fi

# Check for FlameGraph tools
FLAMEGRAPH_DIR="${FLAMEGRAPH_DIR:-$HOME/FlameGraph}"
if [[ ! -f "$FLAMEGRAPH_DIR/flamegraph.pl" ]]; then
    log_warn "FlameGraph tools not found at $FLAMEGRAPH_DIR"
    log_warn "Falling back to go tool pprof -svg (less interactive)"

    log "Generating SVG flamegraph..."
    if go tool pprof -svg "$PROFILE_FILE" > "$SVG_FILE" 2>&1; then
        SIZE=$(du -h "$SVG_FILE" | cut -f1)
        log_success "Flamegraph generated: $SVG_FILE ($SIZE)"
    else
        log_error "Failed to generate flamegraph"
        exit 1
    fi
else
    log "Generating interactive flamegraph using Brendan Gregg's tools..."

    # Generate folded stacks from pprof data
    FOLDED_FILE="${PROFILE_FILE%.pb.gz}.folded"

    # Convert pprof to text format, then fold stacks
    if go tool pprof -raw "$PROFILE_FILE" 2>/dev/null | \
       perl "$FLAMEGRAPH_DIR/stackcollapse-go.pl" > "$FOLDED_FILE" 2>&1; then

        # Generate interactive SVG flamegraph
        if perl "$FLAMEGRAPH_DIR/flamegraph.pl" \
            --title="$PROFILE_TYPE Profile - $(date)" \
            --width=1600 \
            "$FOLDED_FILE" > "$SVG_FILE" 2>&1; then

            SIZE=$(du -h "$SVG_FILE" | cut -f1)
            log_success "Interactive flamegraph generated: $SVG_FILE ($SIZE)"

            # Clean up intermediate file
            rm -f "$FOLDED_FILE"
        else
            log_error "Failed to generate flamegraph from folded stacks"
            rm -f "$FOLDED_FILE"
            exit 1
        fi
    else
        log_error "Failed to generate folded stacks"
        exit 1
    fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log "Phase 3: Generating Text Analysis"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

log "Extracting top functions..."
{
    echo "Top 30 Functions by $PROFILE_TYPE"
    echo "Generated: $(date)"
    echo "Profile: $PROFILE_FILE"
    echo "=========================================="
    echo ""
    go tool pprof -top -nodecount=30 "$PROFILE_FILE" 2>&1
} > "$TXT_FILE"

log_success "Text analysis saved: $TXT_FILE"

# Show preview
echo ""
log_info "Preview of top functions:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
head -20 "$TXT_FILE" | tail -15
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Try AI analysis if available
echo ""
if [[ -f "$PROJECT_ROOT/scripts/ai-flamegraph-analyzer.py" && "$PROFILE_TYPE" == "cpu" ]]; then
    log "Attempting AI-powered analysis..."
    if command -v python3 &> /dev/null; then
        python3 "$PROJECT_ROOT/scripts/ai-flamegraph-analyzer.py" "$PROFILE_FILE" 2>&1 || log_warn "AI analysis failed (optional)"
    else
        log_warn "Python3 not found, skipping AI analysis"
    fi
fi

# Summary
echo ""
echo "╔═══════════════════════════════════════════════════════════════════════════╗"
echo "║                              SUMMARY                                      ║"
echo "╚═══════════════════════════════════════════════════════════════════════════╝"
echo ""

log_success "Flamegraph generation complete!"
echo ""

log_info "Generated Files:"
echo "  📊 Flamegraph:  $SVG_FILE"
echo "  📦 Profile:     $PROFILE_FILE"
echo "  📄 Analysis:    $TXT_FILE"
echo ""

log_info "View Results:"
echo "  Browser:     firefox '$SVG_FILE' &"
echo "  Interactive: go tool pprof '$PROFILE_FILE'"
echo "  Analysis:    cat '$TXT_FILE'"
echo ""

log_info "Next Steps:"
case "$PROFILE_TYPE" in
    cpu)
        echo "  • Look for wide bars in the flamegraph (hotspots)"
        echo "  • Focus on functions with high 'flat' percentage"
        echo "  • Consider generating heap profile: ./$(basename $0) heap 0 $PORT"
        ;;
    heap)
        echo "  • Check for memory leaks (growing allocations)"
        echo "  • Look for large allocations"
        echo "  • Consider goroutine profile: ./$(basename $0) goroutine 0 $PORT"
        ;;
    goroutine)
        echo "  • Check for goroutine leaks (too many active)"
        echo "  • Look for blocked goroutines"
        echo "  • Consider block profile: ./$(basename $0) block 0 $PORT"
        ;;
    *)
        echo "  • Analyze the flamegraph for bottlenecks"
        echo "  • Run 'make profile-all' for comprehensive analysis"
        ;;
esac

echo ""
log_info "All flamegraphs: ls -lh '$OUTPUT_DIR/'"
echo ""

exit 0
