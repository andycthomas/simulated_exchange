#!/bin/bash

# Load Generator for Docker Microservices Architecture
# Directly submits orders to the trading-api to generate Grafana metrics

set -e

# Configuration
TRADING_API_URL="http://localhost:8080"
DURATION=${1:-60}  # Default 60 seconds
RATE=${2:-10}      # Default 10 requests/sec
SYMBOLS=("BTCUSD" "ETHUSD" "ADAUSD" "SOLUSD" "DOTUSD" "MATICUSD")
SIDES=("BUY" "SELL")
TYPES=("LIMIT" "MARKET")

echo "=========================================="
echo "Load Generator for Trading API"
echo "=========================================="
echo "Duration: ${DURATION} seconds"
echo "Target Rate: ${RATE} requests/sec"
echo "Endpoint: ${TRADING_API_URL}"
echo "=========================================="
echo ""

# Function to generate a random order
generate_order() {
    local symbol=${SYMBOLS[$RANDOM % ${#SYMBOLS[@]}]}
    local side=${SIDES[$RANDOM % ${#SIDES[@]}]}
    local type=${TYPES[$RANDOM % ${#TYPES[@]}]}
    local quantity=$(awk -v seed=$RANDOM 'BEGIN { srand(seed); printf "%.2f", rand() * 10 + 0.1 }')
    local price=$(awk -v seed=$RANDOM 'BEGIN { srand(seed); printf "%.2f", rand() * 10000 + 1000 }')

    if [ "$type" == "MARKET" ]; then
        # Market orders don't need price
        echo "{\"symbol\":\"$symbol\",\"type\":\"$type\",\"side\":\"$side\",\"quantity\":$quantity}"
    else
        echo "{\"symbol\":\"$symbol\",\"type\":\"$type\",\"side\":\"$side\",\"quantity\":$quantity,\"price\":$price}"
    fi
}

# Function to submit order
submit_order() {
    local order=$(generate_order)
    curl -s -X POST "${TRADING_API_URL}/orders" \
        -H "Content-Type: application/json" \
        -d "$order" > /dev/null 2>&1
}

# Calculate sleep time between requests
sleep_time=$(awk -v rate=$RATE 'BEGIN { printf "%.4f", 1.0 / rate }')

echo "Starting load generation..."
echo "Sleep time between requests: ${sleep_time}s"
echo ""

# Track metrics
total_requests=0
start_time=$(date +%s)
end_time=$((start_time + DURATION))

# Main loop
while [ $(date +%s) -lt $end_time ]; do
    submit_order &
    total_requests=$((total_requests + 1))

    # Print progress every 10 requests
    if [ $((total_requests % 10)) -eq 0 ]; then
        elapsed=$(($(date +%s) - start_time))
        if [ $elapsed -gt 0 ]; then
            current_rate=$(awk -v total=$total_requests -v elapsed=$elapsed 'BEGIN { printf "%.2f", total / elapsed }')
            echo "[${elapsed}s] Sent: $total_requests orders | Rate: ${current_rate} req/s"
        fi
    fi

    sleep $sleep_time
done

# Wait for all background jobs to finish
wait

# Final statistics
elapsed=$(($(date +%s) - start_time))
if [ $elapsed -eq 0 ]; then elapsed=1; fi
avg_rate=$(awk -v total=$total_requests -v elapsed=$elapsed 'BEGIN { printf "%.2f", total / elapsed }')

echo ""
echo "=========================================="
echo "Load Generation Complete"
echo "=========================================="
echo "Total requests: $total_requests"
echo "Duration: ${elapsed}s"
echo "Average rate: ${avg_rate} req/s"
echo "=========================================="
echo ""
echo "Check Grafana dashboards at http://localhost:3000"
