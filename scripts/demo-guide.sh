#!/bin/bash
################################################################################
# Interactive Performance Review Demo Guide
#
# Step-by-step interactive guide for the 5-minute performance review demo
################################################################################

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

clear

echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║           INTERACTIVE PERFORMANCE REVIEW DEMO GUIDE                       ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}This interactive guide will walk you through the 5-minute demo${NC}"
echo -e "${CYAN}with step-by-step prompts and talking points.${NC}"
echo ""
echo -e "${YELLOW}Press ENTER at each step to continue...${NC}"
echo ""
read -p "Press ENTER to begin..."

clear

# ===== STEP 0: THE HOOK =====
echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║  STEP 0: THE HOOK (30 seconds)                                            ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${MAGENTA}${BOLD}TALKING POINTS:${NC}"
echo ""
echo "  'Who here has been paged at 3am for a production incident?'"
echo "  'How long did it take to find the root cause?'"
echo ""
echo "  'I built a platform that does it in 90 seconds.'"
echo "  'Let me show you how.'"
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
read -p "Press ENTER when ready for Step 1..."

clear

# ===== STEP 1: PROVE IT WORKS =====
echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║  STEP 1: Prove It Works (15 seconds)                                      ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${MAGENTA}${BOLD}SAY:${NC}"
echo "  'First, let me show you the system is actually running.'"
echo ""
echo -e "${GREEN}${BOLD}RUN THIS COMMAND:${NC}"
echo ""
echo -e "${CYAN}  curl -s http://localhost:8080/health | jq '.status'${NC}"
echo ""
echo -e "${MAGENTA}${BOLD}NARRATE:${NC}"
echo "  'This is a real trading system, processing orders right now.'"
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
read -p "Press ENTER to run the command..."

curl -s http://localhost:8080/health | jq '.status' 2>/dev/null || echo -e "${YELLOW}(Service might not be running - start with: docker compose up -d)${NC}"

echo ""
read -p "Press ENTER when ready for Step 2..."

clear

# ===== STEP 2: THE MAGIC COMMAND =====
echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║  STEP 2: The Magic Command (2 minutes)                                    ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${MAGENTA}${BOLD}SAY:${NC}"
echo "  'Watch this. ONE command demonstrates all four technologies:'"
echo "  '  1. Live trading system'"
echo "  '  2. Chaos engineering'"
echo "  '  3. eBPF kernel tracing'"
echo "  '  4. Flamegraph analysis with AI'"
echo ""
echo -e "${GREEN}${BOLD}RUN THIS COMMAND:${NC}"
echo ""
echo -e "${CYAN}  make chaos-flamegraph${NC}"
echo ""
echo -e "${MAGENTA}${BOLD}NARRATE WHILE IT RUNS:${NC}"
echo "  As baseline metrics capture:"
echo "    'Capturing baseline performance...'"
echo ""
echo "  When load starts:"
echo "    'Starting eBPF kernel-level tracing...'"
echo ""
echo "  When chaos is injected:"
echo "    'Now killing the database - injecting real chaos...'"
echo ""
echo "  During profiling:"
echo "    'Profiling CPU usage under stress with pprof...'"
echo ""
echo "  When generating flamegraphs:"
echo "    'Generating visual flamegraphs...'"
echo ""
echo "  At the end:"
echo "    'AI analyzing performance bottlenecks...'"
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${YELLOW}${BOLD}TIP:${NC} This takes about 90 seconds. Keep talking during execution!"
echo ""
read -p "Press ENTER to run the chaos experiment..."

cd "$PROJECT_ROOT"
make chaos-flamegraph || echo -e "${YELLOW}(Experiment may have failed - check if services are running)${NC}"

echo ""
read -p "Press ENTER when ready for Step 3..."

clear

# ===== STEP 3: SHOW eBPF TRACING =====
echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║  STEP 3: Show eBPF Kernel Tracing (30 seconds)                            ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${MAGENTA}${BOLD}SAY:${NC}"
echo "  'Let's look at what eBPF caught at the kernel level.'"
echo "  'This is showing TCP failures, retries, connection timing...'"
echo "  'All captured with ZERO overhead, no application changes needed.'"
echo ""
echo -e "${GREEN}${BOLD}RUN THIS COMMAND:${NC}"
echo ""
echo -e "${CYAN}  tail -30 chaos-results/flamegraph-profiling/experiment.log.bpf${NC}"
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
read -p "Press ENTER to show eBPF trace..."

if [[ -f "$PROJECT_ROOT/chaos-results/flamegraph-profiling/experiment.log.bpf" ]]; then
    tail -30 "$PROJECT_ROOT/chaos-results/flamegraph-profiling/experiment.log.bpf"
else
    echo -e "${YELLOW}eBPF trace not found (may not have been captured)${NC}"
fi

echo ""
read -p "Press ENTER when ready for Step 4..."

clear

# ===== STEP 4: SHOW VISUAL FLAMEGRAPH =====
echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║  STEP 4: Show Visual Flamegraph (45 seconds)                              ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${MAGENTA}${BOLD}SAY:${NC}"
echo "  'Now let's see the visual flamegraph.'"
echo "  'Width equals CPU time - wider bars are your hotspots.'"
echo ""
echo -e "${GREEN}${BOLD}RUN THIS COMMAND:${NC}"
echo ""

# Find the most recent CPU load flamegraph
FLAMEGRAPH=$(find "$PROJECT_ROOT/chaos-results/flamegraph-profiling/flamegraphs" -name "cpu_load_*.svg" 2>/dev/null | sort -r | head -1)

if [[ -z "$FLAMEGRAPH" ]]; then
    FLAMEGRAPH=$(find "$PROJECT_ROOT/flamegraphs" -name "cpu_*.svg" 2>/dev/null | sort -r | head -1)
fi

if [[ -n "$FLAMEGRAPH" ]]; then
    echo -e "${CYAN}  firefox '$FLAMEGRAPH' &${NC}"
else
    echo -e "${CYAN}  firefox chaos-results/flamegraph-profiling/flamegraphs/cpu_load.svg &${NC}"
fi

echo ""
echo -e "${MAGENTA}${BOLD}DEMONSTRATE:${NC}"
echo "  • Point to wide bars: 'These are your hotspots'"
echo "  • Click a box: 'Zoom in to see details'"
echo "  • Hover: 'See exact CPU time and percentage'"
echo "  • Ctrl+F: 'Search for specific functions'"
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
read -p "Press ENTER to open flamegraph in browser..."

if [[ -n "$FLAMEGRAPH" ]]; then
    if command -v firefox &> /dev/null; then
        firefox "$FLAMEGRAPH" &
        echo -e "${GREEN}✓ Opened flamegraph in Firefox${NC}"
    elif command -v chromium &> /dev/null; then
        chromium "$FLAMEGRAPH" &
        echo -e "${GREEN}✓ Opened flamegraph in Chromium${NC}"
    else
        echo -e "${YELLOW}No browser found. Open manually: $FLAMEGRAPH${NC}"
    fi
else
    echo -e "${YELLOW}No flamegraph found. Generate with: make chaos-flamegraph${NC}"
fi

echo ""
read -p "Press ENTER when ready for Step 5..."

clear

# ===== STEP 5: SHOW AI ANALYSIS =====
echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║  STEP 5: Show AI-Powered Analysis (30 seconds)                            ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${MAGENTA}${BOLD}SAY:${NC}"
echo "  'But wait - there's more. AI automatically analyzed this.'"
echo "  'Here's what it found:'"
echo ""
echo -e "${GREEN}${BOLD}RUN THIS COMMAND:${NC}"
echo ""

# Find the most recent analysis
ANALYSIS=$(find "$PROJECT_ROOT/flamegraphs" -name "*_analysis.md" 2>/dev/null | sort -r | head -1)

if [[ -z "$ANALYSIS" ]]; then
    ANALYSIS=$(find "$PROJECT_ROOT/chaos-results" -name "*_analysis.md" 2>/dev/null | sort -r | head -1)
fi

if [[ -n "$ANALYSIS" ]]; then
    echo -e "${CYAN}  head -40 '$ANALYSIS'${NC}"
else
    echo -e "${CYAN}  head -40 flamegraphs/cpu_load_analysis.md${NC}"
fi

echo ""
echo -e "${MAGENTA}${BOLD}HIGHLIGHT:${NC}"
echo "  • 'AI identified the top hotspot automatically'"
echo "  • 'Categorized by priority: HIGH, MEDIUM, LOW'"
echo "  • 'Specific code locations called out'"
echo "  • 'Actionable recommendations for each issue'"
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
read -p "Press ENTER to show AI analysis..."

if [[ -n "$ANALYSIS" ]]; then
    head -40 "$ANALYSIS"
else
    echo -e "${YELLOW}No analysis found. AI analyzer needs CPU profile to analyze.${NC}"
fi

echo ""
read -p "Press ENTER for the finale..."

clear

# ===== FINALE: THE IMPACT =====
echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║  FINALE: The Impact (15 seconds)                                          ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${MAGENTA}${BOLD}SAY:${NC}"
echo "  'So what does this mean?'"
echo ""
echo -e "${RED}${BOLD}TRADITIONAL ROOT CAUSE: 60-90 minutes${NC}"
echo "  • SSH to servers.................. 2 min"
echo "  • Attach profiler................. 5 min"
echo "  • Capture data.................... 5 min"
echo "  • Download & analyze.............. 15 min"
echo "  • Research optimization........... 30 min"
echo "  • Document findings............... 10 min"
echo ""
echo -e "${GREEN}${BOLD}THIS PLATFORM: 90 seconds (fully automated)${NC}"
echo ""
echo -e "${CYAN}${BOLD}ROI: 40-60x faster incident response${NC}"
echo ""
echo -e "${MAGENTA}${BOLD}SAY:${NC}"
echo "  'From hours to 90 seconds.'"
echo "  'That's the difference between being up at 3am...'"
echo "  '...and sleeping through the night.'"
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
read -p "Press ENTER to finish..."

clear

# ===== SUMMARY =====
echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║                         DEMO COMPLETE!                                    ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}✅ You just completed the 5-minute performance review demo!${NC}"
echo ""
echo -e "${CYAN}What we demonstrated:${NC}"
echo "  ✓ Live trading system (41,821 lines of Go)"
echo "  ✓ Chaos engineering (production-like failures)"
echo "  ✓ eBPF profiling (kernel-level tracing)"
echo "  ✓ Flamegraphs + AI analysis"
echo ""
echo -e "${CYAN}Key talking points covered:${NC}"
echo "  ✓ Real-world incident response scenario"
echo "  ✓ One-command comprehensive diagnostics"
echo "  ✓ Zero-overhead kernel tracing with eBPF"
echo "  ✓ Visual flamegraph analysis"
echo "  ✓ AI-powered bottleneck identification"
echo "  ✓ 40-60x ROI compared to traditional methods"
echo ""
echo -e "${YELLOW}Tips for next time:${NC}"
echo "  • Practice the narration timing"
echo "  • Have backup flamegraphs ready (make prep-demo)"
echo "  • Increase terminal font size before presenting"
echo "  • Keep the cheat sheet open in another window"
echo ""
echo -e "${CYAN}Useful commands:${NC}"
echo "  make prep-demo          # Full preparation with fresh chaos run"
echo "  make prep-demo-quick    # Quick prep with existing artifacts"
echo "  make run-demo           # Show all steps (no prompts)"
echo "  ./scripts/demo-guide.sh # This interactive guide"
echo ""
echo -e "${GREEN}Good luck with your presentation! 🚀${NC}"
echo ""
