#!/bin/bash
# Performance Tier Runner Script
# Easy way to run the Simulated Exchange at different performance tiers

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Function to display usage
usage() {
    echo -e "${BLUE}Performance Tier Runner${NC}"
    echo ""
    echo "Usage: $0 [TIER]"
    echo ""
    echo "Available Tiers:"
    echo -e "  ${GREEN}light${NC}    - Tier 1: Light/Minimum (1k ops/sec, 4 cores, <20ms latency)"
    echo -e "  ${YELLOW}average${NC}  - Tier 2: Standard/Average (5k ops/sec, 8 cores, <50ms latency)"
    echo -e "  ${RED}maximum${NC}  - Tier 3: Peak/Maximum (12k ops/sec, 16 cores, <150ms latency)"
    echo ""
    echo "Commands:"
    echo "  $0 light      - Run with light tier"
    echo "  $0 average    - Run with average tier (default)"
    echo "  $0 maximum    - Run with maximum tier"
    echo "  $0 info       - Show tier information"
    echo "  $0 compare    - Compare all tiers"
    echo ""
    echo "Examples:"
    echo "  $0 average"
    echo "  $0 maximum"
    echo "  make run-light    # Alternative using Makefile"
    exit 1
}

# Function to show tier information
show_info() {
    local tier=$1
    local env_file="${PROJECT_DIR}/.env.${tier}"

    if [ ! -f "$env_file" ]; then
        echo -e "${RED}❌ Configuration file not found: $env_file${NC}"
        return 1
    fi

    echo -e "${BLUE}Configuration for '${tier}' tier:${NC}"
    echo ""
    grep -E '^[A-Z_]+=' "$env_file" | while IFS='=' read -r key value; do
        printf "  %-30s %s\n" "$key" "$value"
    done
    echo ""
}

# Function to compare all tiers
compare_tiers() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║            Simulated Exchange - Performance Tier Comparison            ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    printf "${BLUE}%-20s %-20s %-20s %-20s${NC}\n" "Metric" "Light (Tier 1)" "Average (Tier 2)" "Maximum (Tier 3)"
    echo "────────────────────────────────────────────────────────────────────────────"
    printf "%-20s %-20s %-20s %-20s\n" "Orders/Second" "1,000" "5,000" "12,000"
    printf "%-20s %-20s %-20s %-20s\n" "Trades/Second" "250-750" "3,750" "9,000"
    printf "%-20s %-20s %-20s %-20s\n" "Concurrent Users" "100-250" "1,000" "2,500"
    printf "%-20s %-20s %-20s %-20s\n" "P99 Latency" "<20ms" "<50ms" "<150ms"
    printf "%-20s %-20s %-20s %-20s\n" "Daily Orders" "86M" "432M" "1.04B"
    printf "%-20s %-20s %-20s %-20s\n" "CPU Cores" "4 (25%)" "8 (50%)" "16 (85%)"
    printf "%-20s %-20s %-20s %-20s\n" "Memory" "2-4 GB" "8-12 GB" "20-24 GB"
    printf "%-20s %-20s %-20s %-20s\n" "Daily Volume" "\$324M" "\$2.16B" "\$5.2B"
    printf "%-20s %-20s %-20s %-20s\n" "Workers" "4" "8" "16"
    printf "%-20s %-20s %-20s %-20s\n" "DB Connections" "25" "50" "100"
    echo ""

    echo -e "${BLUE}Use Cases:${NC}"
    echo -e "  ${GREEN}Light${NC}    - Initial deployment, dev/test, cost-sensitive"
    echo -e "  ${YELLOW}Average${NC}  - Standard production, typical workload, balanced"
    echo -e "  ${RED}Maximum${NC}  - Peak events, flash trading, stress testing"
    echo ""
}

# Function to run with specific tier
run_tier() {
    local tier=$1
    local env_file="${PROJECT_DIR}/.env.${tier}"

    if [ ! -f "$env_file" ]; then
        echo -e "${RED}❌ Configuration file not found: $env_file${NC}"
        echo ""
        echo "Available tiers: light, average, maximum"
        exit 1
    fi

    # Display tier info
    case $tier in
        light)
            echo -e "${GREEN}🟢 Starting with LIGHT performance tier...${NC}"
            echo "   • Target: 1,000 orders/sec"
            echo "   • Latency: <20ms P99"
            echo "   • Resources: 4 cores, 2-4GB RAM"
            export GOMAXPROCS=4
            ;;
        average)
            echo -e "${YELLOW}🟡 Starting with AVERAGE performance tier...${NC}"
            echo "   • Target: 5,000 orders/sec"
            echo "   • Latency: <50ms P99"
            echo "   • Resources: 8 cores, 8-12GB RAM"
            echo "   • Daily Volume: 432M orders, \$2.16B trading"
            export GOMAXPROCS=8
            ;;
        maximum)
            echo -e "${RED}🔴 Starting with MAXIMUM performance tier...${NC}"
            echo "   • Target: 12,000 orders/sec"
            echo "   • Latency: <150ms P99"
            echo "   • Resources: 16 cores, 20-24GB RAM"
            echo "   • Daily Volume: 1.04B orders, \$5.2B trading"
            export GOMAXPROCS=16
            export GOGC=100
            ;;
    esac

    echo ""
    echo -e "${BLUE}Loading configuration from: .env.${tier}${NC}"
    echo ""

    # Export environment variables
    set -a
    source "$env_file"
    set +a

    # Run the application
    cd "$PROJECT_DIR"
    exec go run cmd/api/main.go
}

# Main script logic
main() {
    local command=${1:-average}

    case $command in
        light|average|maximum)
            run_tier "$command"
            ;;
        info)
            if [ -n "$2" ]; then
                show_info "$2"
            else
                echo "Usage: $0 info [light|average|maximum]"
                echo ""
                echo "Showing all tiers:"
                echo ""
                for tier in light average maximum; do
                    show_info "$tier"
                done
            fi
            ;;
        compare)
            compare_tiers
            ;;
        help|-h|--help)
            usage
            ;;
        *)
            echo -e "${RED}❌ Unknown tier or command: $command${NC}"
            echo ""
            usage
            ;;
    esac
}

# Run main function
main "$@"
