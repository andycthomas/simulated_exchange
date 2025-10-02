#!/bin/bash

# Database Load Simulation Script
# Gradually increases database connections until exhaustion
# Uses the trading-api container to create connections

set -e

# Configuration
CONTAINER_NAME="simulated-exchange-trading-api"
DB_HOST="postgres"
DB_PORT="5432"
DB_NAME="trading_db"
DB_USER="trading_user"
DB_PASSWORD="trading_password"
MAX_CONNECTIONS=300
STEP_SIZE=10
SLEEP_INTERVAL=5
LOG_FILE="/tmp/db_load_test.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Database Load Simulation Started ===${NC}"
echo "Target: Gradually increase connections until exhaustion (max: $MAX_CONNECTIONS)"
echo "Container: $CONTAINER_NAME"
echo "Step size: $STEP_SIZE connections every $SLEEP_INTERVAL seconds"
echo "Log file: $LOG_FILE"
echo ""

# Initialize log file
echo "$(date): Database load simulation started" > $LOG_FILE

# Function to check current connection count
check_connections() {
    local count=$(docker exec $CONTAINER_NAME psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname='$DB_NAME';" 2>/dev/null | tr -d ' ')
    echo $count
}

# Function to create a persistent connection
create_connection() {
    local conn_id=$1
    local script="
import psycopg2
import time
import sys

try:
    conn = psycopg2.connect(
        host='$DB_HOST',
        port='$DB_PORT',
        database='$DB_NAME',
        user='$DB_USER',
        password='$DB_PASSWORD'
    )
    print(f'Connection {conn_id} established')

    # Keep connection alive with periodic queries
    cursor = conn.cursor()
    while True:
        try:
            cursor.execute('SELECT 1')
            time.sleep(30)  # Query every 30 seconds to keep alive
        except Exception as e:
            print(f'Connection {conn_id} error: {e}')
            break
except Exception as e:
    print(f'Connection {conn_id} failed: {e}')
    sys.exit(1)
finally:
    if 'conn' in locals():
        conn.close()
"

    # Run the connection script in the background
    docker exec -d $CONTAINER_NAME python3 -c "$script" &
    echo $! # Return process ID
}

# Function to kill all connection processes
cleanup() {
    echo -e "\n${YELLOW}Cleaning up connections...${NC}"
    # Kill all python processes in the container
    docker exec $CONTAINER_NAME pkill -f "python3 -c" 2>/dev/null || true
    sleep 2

    local final_count=$(check_connections)
    echo -e "${GREEN}Cleanup complete. Current connections: $final_count${NC}"
    echo "$(date): Simulation ended. Final connection count: $final_count" >> $LOG_FILE
}

# Set up cleanup on script exit
trap cleanup EXIT INT TERM

# Check if container is running
if ! docker ps | grep -q $CONTAINER_NAME; then
    echo -e "${RED}Error: Container $CONTAINER_NAME is not running${NC}"
    exit 1
fi

# Check initial connection count
initial_count=$(check_connections)
echo -e "${GREEN}Initial connection count: $initial_count${NC}"
echo "$(date): Initial connection count: $initial_count" >> $LOG_FILE

# Array to store process IDs
declare -a pids=()
current_connections=$initial_count
step=1

echo -e "\n${BLUE}Starting connection creation...${NC}"

while [ $current_connections -lt $MAX_CONNECTIONS ]; do
    echo -e "\n--- Step $step: Adding $STEP_SIZE connections ---"

    # Create STEP_SIZE new connections
    for i in $(seq 1 $STEP_SIZE); do
        conn_id="${step}_${i}"
        echo "Creating connection $conn_id..."

        # Try to create connection
        if pid=$(create_connection $conn_id 2>/dev/null); then
            pids+=($pid)
        else
            echo -e "${RED}Failed to create connection $conn_id${NC}"
        fi

        # Small delay between individual connections
        sleep 0.5
    done

    # Wait for connections to establish
    sleep $SLEEP_INTERVAL

    # Check current connection count
    new_count=$(check_connections)

    if [ "$new_count" -eq "$current_connections" ]; then
        echo -e "${RED}Warning: Connection count didn't increase. Possible connection limit reached.${NC}"
        echo "$(date): Step $step - Connection limit may have been reached. Count: $new_count" >> $LOG_FILE
        break
    fi

    current_connections=$new_count
    echo -e "${GREEN}Current connections: $current_connections${NC}"
    echo "$(date): Step $step completed. Current connections: $current_connections" >> $LOG_FILE

    # Check if we're approaching the limit
    if [ $current_connections -ge $(($MAX_CONNECTIONS - $STEP_SIZE)) ]; then
        echo -e "${YELLOW}Approaching connection limit. Slowing down...${NC}"
        STEP_SIZE=5
        SLEEP_INTERVAL=10
    fi

    step=$((step + 1))
done

echo -e "\n${BLUE}=== Load Test Complete ===${NC}"
echo -e "${GREEN}Final connection count: $current_connections${NC}"
echo -e "${YELLOW}Maximum theoretical connections: $MAX_CONNECTIONS${NC}"

if [ $current_connections -ge $(($MAX_CONNECTIONS - 10)) ]; then
    echo -e "${RED}🔥 Connection exhaustion achieved!${NC}"
else
    echo -e "${YELLOW}⚠️  Connection limit not fully reached${NC}"
fi

echo ""
echo "Connection creation summary:"
echo "- Initial connections: $initial_count"
echo "- Final connections: $current_connections"
echo "- New connections created: $((current_connections - initial_count))"
echo "- Total steps: $((step - 1))"

echo ""
echo -e "${BLUE}Log file available at: $LOG_FILE${NC}"
echo "Press Ctrl+C to exit and cleanup connections..."

# Keep script running to maintain connections
while true; do
    sleep 10
    count=$(check_connections)
    echo "$(date): Active connections: $count"
done