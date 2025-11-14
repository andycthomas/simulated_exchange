# Simulated Exchange Platform - Unified Makefile
# ============================================================
# Complete orchestration: development, testing, demos, chaos experiments, and monitoring
#
# Quick Start:
#   make help              - Show all available commands
#   make start-average     - Start app with average performance tier (background)
#   make watch-metrics     - Monitor live metrics
#   make test-all          - Run all tests
#   make demo-full         - Run complete demo
#
# Sections:
#   1. Configuration & Variables
#   2. Help & Documentation
#   3. Application Management (start/stop/status)
#   4. Performance Tier Run Targets
#   5. Demo Orchestration
#   6. Chaos Experiments
#   7. Monitoring & Dashboards
#   8. Testing & Validation
#   9. Docker & CI/CD
#  10. Utilities & Maintenance

# Force bash shell for all commands (required for read -p, [[ ]], etc.)
SHELL := /bin/bash

.PHONY: help setup
.DEFAULT_GOAL := help

# ============================================================
# Section 1: Configuration & Variables
# ============================================================

# Go configuration
GO := go
GO_VERSION := $(shell go version | cut -d' ' -f3)
GOPATH := $(HOME)/go
GO_BIN := $(HOME)/.local/go/bin
LOCAL_BIN := $(HOME)/.local/bin
export PATH := $(LOCAL_BIN):$(GO_BIN):$(PATH)

# Project configuration
PROJECT_NAME := simulated_exchange
COVERAGE_FILE := coverage.out
TEST_RESULTS_DIR := test-output

# Application configuration
APP_PORT := 8080
APP_HOST := localhost
APP_BINARY := cmd/api/main.go
APP_PID_FILE := /tmp/simulated-exchange.pid
APP_LOG_FILE := /tmp/simulated-exchange.log

# Demo configuration
LIGHT_LOAD_DURATION := 60s
LIGHT_LOAD_OPS := 25
MEDIUM_LOAD_DURATION := 90s
MEDIUM_LOAD_OPS := 100
HEAVY_LOAD_DURATION := 120s
HEAVY_LOAD_OPS := 200

# Chaos experiment configuration
CHAOS_DIR := chaos-experiments
CHAOS_LOG_DIR := /tmp/chaos-logs

# Docker configuration
DOCKER_COMPOSE_FILE := docker-compose.test.yml

# URLs
HEALTH_URL := http://$(APP_HOST):$(APP_PORT)/health
METRICS_URL := http://$(APP_HOST):$(APP_PORT)/api/metrics
DASHBOARD_URL := http://$(APP_HOST):$(APP_PORT)/dashboard
ORDERS_URL := http://$(APP_HOST):$(APP_PORT)/api/orders
LOADTEST_URL := http://$(APP_HOST):$(APP_PORT)/demo/load-test
CHAOSTEST_URL := http://$(APP_HOST):$(APP_PORT)/demo/chaos-test

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
MAGENTA := \033[0;35m
CYAN := \033[0;36m
NC := \033[0m
ECHO := printf '%b\n'

# ============================================================
# Section 2: Help & Documentation
# ============================================================

help: ## Show comprehensive help message
	@$(ECHO) ""
	@$(ECHO) "$(CYAN)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@$(ECHO) "$(CYAN)║     Simulated Exchange Platform - Unified Makefile            ║$(NC)"
	@$(ECHO) "$(CYAN)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)INITIAL SETUP:$(NC)"
	@$(ECHO) "  $(GREEN)make setup$(NC)                  - Run full system setup with conflict resolution"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)APPLICATION MANAGEMENT:$(NC)"
	@$(ECHO) "  $(GREEN)make start-light$(NC)            - Start with light tier (1k ops/sec, background)"
	@$(ECHO) "  $(GREEN)make start-average$(NC)          - Start with average tier (5k ops/sec, background)"
	@$(ECHO) "  $(GREEN)make start-maximum$(NC)          - Start with maximum tier (12k ops/sec, background)"
	@$(ECHO) "  $(GREEN)make stop$(NC)                   - Stop Go application instances"
	@$(ECHO) "  $(GREEN)make stop-all$(NC)               - Stop BOTH Docker containers AND Go apps"
	@$(ECHO) "  $(GREEN)make restart-average$(NC)        - Restart with average tier"
	@$(ECHO) "  $(GREEN)make status$(NC)                 - Check application status"
	@$(ECHO) "  $(GREEN)make health$(NC)                 - Check system health"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)MONITORING:$(NC)"
	@$(ECHO) "  $(GREEN)make watch-metrics$(NC)          - Live metrics updates (every 5s)"
	@$(ECHO) "  $(GREEN)make watch-health$(NC)           - Live health check (every 2s)"
	@$(ECHO) "  $(GREEN)make watch-logs$(NC)             - Tail application logs"
	@$(ECHO) "  $(GREEN)make dashboard$(NC)              - Open web dashboard in browser"
	@$(ECHO) "  $(GREEN)make grafana$(NC)                - Open Grafana dashboard (port 3000)"
	@$(ECHO) "  $(GREEN)make grafana-start$(NC)          - Start Grafana container"
	@$(ECHO) "  $(GREEN)make metrics-snapshot$(NC)       - Save current metrics to file"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)CADDY REVERSE PROXY:$(NC)"
	@$(ECHO) "  $(GREEN)make caddy-start$(NC)            - Start Caddy with SSL for andythomas-sre.com"
	@$(ECHO) "  $(GREEN)make caddy$(NC)                  - Show Caddy status and URLs"
	@$(ECHO) "  $(GREEN)make caddy-reload$(NC)           - Reload config (zero-downtime)"
	@$(ECHO) "  $(GREEN)make caddy-certs$(NC)            - Show SSL certificates"
	@$(ECHO) "  $(GREEN)make caddy-test$(NC)             - Test all endpoints"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)DEMOS:$(NC)"
	@$(ECHO) "  $(GREEN)make demo-full$(NC)              - Complete automated demo (30 min)"
	@$(ECHO) "  $(GREEN)make demo-interactive$(NC)       - Interactive guided demo"
	@$(ECHO) "  $(GREEN)make demo-presentation$(NC)      - Presentation-ready demo (45 min)"
	@$(ECHO) "  $(GREEN)make demo-trading$(NC)           - Basic trading operations"
	@$(ECHO) "  $(GREEN)make demo-loadtest-light$(NC)    - Light load test (25 ops/sec)"
	@$(ECHO) "  $(GREEN)make demo-loadtest-medium$(NC)   - Medium load test (100 ops/sec)"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)CHAOS EXPERIMENTS:$(NC)"
	@$(ECHO) "  $(GREEN)make chaos-all$(NC)              - Run all 10 chaos experiments"
	@$(ECHO) "  $(GREEN)make chaos-db-death$(NC)         - Database sudden death"
	@$(ECHO) "  $(GREEN)make chaos-network-partition$(NC) - Network partition"
	@$(ECHO) "  $(GREEN)make chaos-latency$(NC)          - Network latency injection"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)PROFILING & FLAMEGRAPHS:$(NC)"
	@$(ECHO) "  $(GREEN)make profile-cpu$(NC)            - Generate CPU flamegraph (30s)"
	@$(ECHO) "  $(GREEN)make profile-all$(NC)            - Generate all flamegraph types"
	@$(ECHO) "  $(GREEN)make chaos-flamegraph$(NC)       - Full profiling chaos experiment"
	@$(ECHO) "  $(GREEN)make profile-view$(NC)           - Open flamegraphs in browser"
	@$(ECHO) "  $(GREEN)make profile-help$(NC)           - Show detailed profiling help"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)TESTING & DEVELOPMENT:$(NC)"
	@$(ECHO) "  $(GREEN)make test-all$(NC)               - Run all tests (lint, unit, e2e)"
	@$(ECHO) "  $(GREEN)make test-unit$(NC)              - Run unit tests with coverage"
	@$(ECHO) "  $(GREEN)make test-e2e$(NC)               - Run E2E tests"
	@$(ECHO) "  $(GREEN)make benchmark$(NC)              - Run performance benchmarks"
	@$(ECHO) "  $(GREEN)make lint$(NC)                   - Run code linter"
	@$(ECHO) "  $(GREEN)make coverage$(NC)               - Generate coverage report"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)UTILITIES:$(NC)"
	@$(ECHO) "  $(GREEN)make clean$(NC)                  - Clean temporary files and logs"
	@$(ECHO) "  $(GREEN)make clean-logs$(NC)             - Clean all log files (app, tests, chaos)"
	@$(ECHO) "  $(GREEN)make deps$(NC)                   - Download and verify dependencies"
	@$(ECHO) "  $(GREEN)make install-deps$(NC)           - Install system dependencies"
	@$(ECHO) "  $(GREEN)make check-prereqs$(NC)          - Check all prerequisites"
	@$(ECHO) "  $(RED)make nuke$(NC)                   - ⚠️  DANGER: Stop ALL Docker containers system-wide"
	@$(ECHO) ""
	@$(ECHO) "$(CYAN)Quick Examples:$(NC)"
	@$(ECHO) "  $(BLUE)# Start and monitor:$(NC)"
	@$(ECHO) "  make start-average && make watch-metrics"
	@$(ECHO) ""
	@$(ECHO) "  $(BLUE)# Run tests and start:$(NC)"
	@$(ECHO) "  make test-all && make start-light"
	@$(ECHO) ""
	@$(ECHO) "  $(BLUE)# Full demo:$(NC)"
	@$(ECHO) "  make demo-full"
	@$(ECHO) ""

# ============================================================
# Section 2.5: Initial Setup
# ============================================================

setup: ## Run the full setup script with automatic conflict resolution
	@echo "🧹 Cleaning up existing containers and processes..."
	@docker compose down 2>/dev/null || true
	@sleep 2
	@echo "🔍 Checking for port 443 conflict with SSH..."
	@if sudo ss -tlnp | grep ":443 " | grep -q "sshd" 2>/dev/null; then \
		echo "⚠️  Detected: SSH is running on port 443 (non-standard)"; \
		echo "   Disabling HTTPS (port 443) binding in docker-compose..."; \
		if [ ! -f docker-compose.yml.backup ]; then \
			cp docker-compose.yml docker-compose.yml.backup; \
		fi && \
		sed -i.tmp 's/- "443:443"/# - "443:443"  # Disabled: SSH conflict/' docker-compose.yml && \
		rm -f docker-compose.yml.tmp; \
		echo "   ✅ Port 443 disabled - application will use HTTP only"; \
	else \
		echo "✅ Port 443 is available - HTTPS will be enabled"; \
	fi
	@echo "🔍 Checking for other conflicting processes on ports 80, 5432, 6379, 8080, 8081, 8082..."
	@for port in 80 8080 8081 8082 5432 6379; do \
		PIDS=$$(sudo lsof -ti:$$port 2>/dev/null | grep -v "^1$$" || true); \
		if [ -n "$$PIDS" ]; then \
			echo "🛑 Found processes on port $$port: $$PIDS"; \
			echo "   Killing processes..."; \
			echo "$$PIDS" | xargs -r sudo kill -9 2>/dev/null || true; \
		fi; \
	done
	@sleep 1
	@echo "✅ Port cleanup complete"
	@echo "🚀 Running Simulated Trading Exchange Setup..."
	@./setup.sh
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Setup Complete! Docker containers are now running."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "💡 Next steps:"
	@echo "  • Run demos with Docker:  make demo-interactive"
	@echo "  • View dashboard:         http://localhost/dashboard"
	@echo "  • Check API health:       http://localhost:8080/health"
	@echo "  • Monitor containers:     docker compose ps"
	@echo ""
	@echo "📝 Note: Demo targets will automatically use the running Docker containers."
	@echo "    No need to run 'make start-*' targets separately."
	@echo ""

# ============================================================
# Section 3: Application Management
# ============================================================

.PHONY: start-light start-average start-maximum stop stop-all restart-light restart-average restart-maximum status health

start-light: check-go ## Start with light performance tier (1k ops/sec, 4 cores, background)
	@$(ECHO) "$(GREEN)🟢 Starting with LIGHT performance tier (background)...$(NC)"
	@$(ECHO) "   • Target: 1,000 orders/sec"
	@$(ECHO) "   • Latency: <20ms P99"
	@$(ECHO) "   • Resources: 4 cores, 2-4GB RAM"
	@$(ECHO) ""
	@if [ ! -f .env.light ]; then \
		$(ECHO) "$(RED)✗ .env.light not found$(NC)"; \
		exit 1; \
	fi
	@export $$(cat .env.light | grep -v '^#' | xargs) && \
	export GOMAXPROCS=4 && \
	nohup $(GO) run $(APP_BINARY) > $(APP_LOG_FILE) 2>&1 & echo $$! > $(APP_PID_FILE)
	@sleep 3
	@if $(MAKE) -s health >/dev/null 2>&1; then \
		$(ECHO) "$(GREEN)✓ Application started successfully$(NC)"; \
		$(ECHO) "$(BLUE)PID: $$(cat $(APP_PID_FILE))$(NC)"; \
		$(ECHO) "$(BLUE)Logs: $(APP_LOG_FILE)$(NC)"; \
		$(ECHO) "$(YELLOW)Use 'make watch-logs' to monitor$(NC)"; \
	else \
		$(ECHO) "$(RED)✗ Application failed to start$(NC)"; \
		$(ECHO) "$(YELLOW)Check logs: tail -f $(APP_LOG_FILE)$(NC)"; \
		exit 1; \
	fi

start-average: check-go ## Start with average performance tier (5k ops/sec, 8 cores, background)
	@$(ECHO) "$(GREEN)🟡 Starting with AVERAGE performance tier (background)...$(NC)"
	@$(ECHO) "   • Target: 5,000 orders/sec"
	@$(ECHO) "   • Latency: <50ms P99"
	@$(ECHO) "   • Resources: 8 cores, 8-12GB RAM"
	@$(ECHO) "   • Daily Volume: 432M orders, \$$2.16B trading"
	@$(ECHO) ""
	@if [ ! -f .env.average ]; then \
		$(ECHO) "$(RED)✗ .env.average not found$(NC)"; \
		exit 1; \
	fi
	@export $$(cat .env.average | grep -v '^#' | xargs) && \
	export GOMAXPROCS=8 && \
	nohup $(GO) run $(APP_BINARY) > $(APP_LOG_FILE) 2>&1 & echo $$! > $(APP_PID_FILE)
	@sleep 3
	@if $(MAKE) -s health >/dev/null 2>&1; then \
		$(ECHO) "$(GREEN)✓ Application started successfully$(NC)"; \
		$(ECHO) "$(BLUE)PID: $$(cat $(APP_PID_FILE))$(NC)"; \
		$(ECHO) "$(BLUE)Logs: $(APP_LOG_FILE)$(NC)"; \
		$(ECHO) "$(YELLOW)Use 'make watch-logs' to monitor$(NC)"; \
	else \
		$(ECHO) "$(RED)✗ Application failed to start$(NC)"; \
		$(ECHO) "$(YELLOW)Check logs: tail -f $(APP_LOG_FILE)$(NC)"; \
		exit 1; \
	fi

start-maximum: check-go ## Start with maximum performance tier (12k ops/sec, 16 cores, background)
	@$(ECHO) "$(GREEN)🔴 Starting with MAXIMUM performance tier (background)...$(NC)"
	@$(ECHO) "   • Target: 12,000 orders/sec"
	@$(ECHO) "   • Latency: <150ms P99"
	@$(ECHO) "   • Resources: 16 cores, 20-24GB RAM"
	@$(ECHO) "   • Daily Volume: 1.04B orders, \$$5.2B trading"
	@$(ECHO) ""
	@if [ ! -f .env.maximum ]; then \
		$(ECHO) "$(RED)✗ .env.maximum not found$(NC)"; \
		exit 1; \
	fi
	@export $$(cat .env.maximum | grep -v '^#' | xargs) && \
	export GOMAXPROCS=16 && \
	export GOGC=100 && \
	nohup $(GO) run $(APP_BINARY) > $(APP_LOG_FILE) 2>&1 & echo $$! > $(APP_PID_FILE)
	@sleep 3
	@if $(MAKE) -s health >/dev/null 2>&1; then \
		$(ECHO) "$(GREEN)✓ Application started successfully$(NC)"; \
		$(ECHO) "$(BLUE)PID: $$(cat $(APP_PID_FILE))$(NC)"; \
		$(ECHO) "$(BLUE)Logs: $(APP_LOG_FILE)$(NC)"; \
		$(ECHO) "$(YELLOW)Use 'make watch-logs' to monitor$(NC)"; \
	else \
		$(ECHO) "$(RED)✗ Application failed to start$(NC)"; \
		$(ECHO) "$(YELLOW)Check logs: tail -f $(APP_LOG_FILE)$(NC)"; \
		exit 1; \
	fi

# Aliases for backward compatibility
start: start-average ## Alias for start-average
start-background: start-average ## Alias for start-average (backward compatibility with Makefile.demo)

stop: ## Stop the application gracefully
	@$(ECHO) "$(BLUE)Stopping Simulated Exchange Platform...$(NC)"
	@STOPPED=0; \
	if [ -f $(APP_PID_FILE) ]; then \
		PID=$$(cat $(APP_PID_FILE)); \
		if ps -p $$PID > /dev/null 2>&1; then \
			kill -SIGTERM $$PID 2>/dev/null || true; \
			$(ECHO) "$(GREEN)✓ Sent shutdown signal to PID $$PID$(NC)"; \
			sleep 2; \
			if ps -p $$PID > /dev/null 2>&1; then \
				$(ECHO) "$(YELLOW)Process still running, forcing...$(NC)"; \
				kill -9 $$PID 2>/dev/null || true; \
			fi; \
			STOPPED=1; \
		fi; \
		rm -f $(APP_PID_FILE); \
	fi; \
	PIDS=$$(ps aux | grep -E "[g]o run.*cmd/api/main.go|[m]ain" | grep -v grep | awk '{print $$2}' || true); \
	if [ -n "$$PIDS" ]; then \
		$(ECHO) "$(YELLOW)Found additional running instances: $$PIDS$(NC)"; \
		for pid in $$PIDS; do \
			kill -SIGTERM $$pid 2>/dev/null || true; \
			sleep 1; \
			if ps -p $$pid > /dev/null 2>&1; then \
				kill -9 $$pid 2>/dev/null || true; \
			fi; \
		done; \
		STOPPED=1; \
	fi; \
	PORT_PIDS=$$(sudo lsof -ti:8080 2>/dev/null | grep -v docker || true); \
	if [ -n "$$PORT_PIDS" ]; then \
		$(ECHO) "$(YELLOW)Found processes on port 8080: $$PORT_PIDS$(NC)"; \
		for pid in $$PORT_PIDS; do \
			if ps -p $$pid | grep -qE "go|main"; then \
				$(ECHO) "$(YELLOW)Stopping Go process $$pid on port 8080...$(NC)"; \
				kill -9 $$pid 2>/dev/null || true; \
				STOPPED=1; \
			fi; \
		done; \
	fi; \
	if [ $$STOPPED -eq 1 ]; then \
		$(ECHO) "$(GREEN)✓ Application stopped$(NC)"; \
	else \
		$(ECHO) "$(YELLOW)No running instances found$(NC)"; \
	fi

stop-all: ## Stop both Docker containers AND Go applications
	@$(ECHO) "$(BLUE)Stopping ALL instances (Docker + Go apps)...$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)1. Stopping Docker containers...$(NC)"
	@docker compose down 2>/dev/null || true
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)2. Stopping Go applications...$(NC)"
	@$(MAKE) -s stop
	@$(ECHO) ""
	@$(ECHO) "$(GREEN)✓ All instances stopped$(NC)"

restart-light: stop start-light ## Restart with light performance tier
restart-average: stop start-average ## Restart with average performance tier
restart-maximum: stop start-maximum ## Restart with maximum performance tier
restart: restart-average ## Restart with average tier (default)

status: ## Check if application is running
	@if [ -f $(APP_PID_FILE) ]; then \
		PID=$$(cat $(APP_PID_FILE)); \
		if ps -p $$PID > /dev/null 2>&1; then \
			$(ECHO) "$(GREEN)✓ Application is running (PID: $$PID)$(NC)"; \
			$(ECHO) "$(BLUE)Uptime: $$(ps -p $$PID -o etime=)$(NC)"; \
		else \
			$(ECHO) "$(RED)✗ Application is not running (stale PID file)$(NC)"; \
			exit 1; \
		fi; \
	else \
		$(ECHO) "$(RED)✗ Application is not running$(NC)"; \
		exit 1; \
	fi

health: ## Check system health
	@curl -s $(HEALTH_URL) | jq '.' 2>/dev/null || ($(ECHO) "$(RED)✗ Health check failed$(NC)" && exit 1)

# ============================================================
# Section 4: Performance Tier Run Targets (Foreground)
# ============================================================
# Note: These run in foreground (for development)
# Use start-* targets for background mode

.PHONY: run-light-fg run-average-fg run-maximum-fg run

run-light-fg: check-go ## Run with light tier (foreground, for development)
	@$(ECHO) "$(GREEN)🟢 Starting with LIGHT performance tier (foreground)...$(NC)"
	@$(ECHO) "$(YELLOW)Press Ctrl+C to stop$(NC)"
	@if [ -f .env.light ]; then \
		export $$(cat .env.light | grep -v '^#' | xargs) && \
		export GOMAXPROCS=4 && \
		$(GO) run $(APP_BINARY); \
	else \
		$(ECHO) "$(RED)✗ .env.light not found$(NC)"; \
		exit 1; \
	fi

run-average-fg: check-go ## Run with average tier (foreground, for development)
	@$(ECHO) "$(GREEN)🟡 Starting with AVERAGE performance tier (foreground)...$(NC)"
	@$(ECHO) "$(YELLOW)Press Ctrl+C to stop$(NC)"
	@if [ -f .env.average ]; then \
		export $$(cat .env.average | grep -v '^#' | xargs) && \
		export GOMAXPROCS=8 && \
		$(GO) run $(APP_BINARY); \
	else \
		$(ECHO) "$(RED)✗ .env.average not found$(NC)"; \
		exit 1; \
	fi

run-maximum-fg: check-go ## Run with maximum tier (foreground, for development)
	@$(ECHO) "$(GREEN)🔴 Starting with MAXIMUM performance tier (foreground)...$(NC)"
	@$(ECHO) "$(YELLOW)Press Ctrl+C to stop$(NC)"
	@if [ -f .env.maximum ]; then \
		export $$(cat .env.maximum | grep -v '^#' | xargs) && \
		export GOMAXPROCS=16 && \
		export GOGC=100 && \
		$(GO) run $(APP_BINARY); \
	else \
		$(ECHO) "$(RED)✗ .env.maximum not found$(NC)"; \
		exit 1; \
	fi

run: run-average-fg ## Run with average tier (foreground, default)

# ============================================================
# Section 5: Demo Orchestration
# ============================================================

.PHONY: demo-full demo-interactive demo-presentation demo-trading demo-metrics demo-loadtest-light demo-loadtest-medium demo-loadtest-heavy demo-chaos-latency demo-chaos-errors

demo-full: ## Complete automated demo (30 minutes)
	@$(ECHO) "$(CYAN)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@$(ECHO) "$(CYAN)║        SIMULATED EXCHANGE - COMPLETE AUTOMATED DEMO            ║$(NC)"
	@$(ECHO) "$(CYAN)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(BLUE)This will run:$(NC)"
	@$(ECHO) "  1. System startup and health verification"
	@$(ECHO) "  2. Basic trading operations"
	@$(ECHO) "  3. Metrics monitoring"
	@$(ECHO) "  4. Light load test"
	@$(ECHO) "  5. Medium load test"
	@$(ECHO) "  6. Chaos: Latency injection"
	@$(ECHO) "  7. Chaos: Error simulation"
	@$(ECHO) "  8. Final metrics and cleanup"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Estimated duration: 30 minutes$(NC)"
	@$(ECHO) ""
	@$(MAKE) -s _demo-full-internal

_demo-full-internal:
	@$(ECHO) "$(MAGENTA)[1/8] Starting system...$(NC)"
	@if docker compose ps | grep -q "simulated-exchange-trading-api.*Up"; then \
		$(ECHO) "$(GREEN)✓ Trading API already running in Docker$(NC)"; \
	else \
		$(MAKE) -s start-average; \
	fi
	@sleep 5
	@$(ECHO) ""
	@$(ECHO) "$(MAGENTA)[2/8] Running trading demo...$(NC)"
	@$(MAKE) -s demo-trading
	@sleep 5
	@$(ECHO) ""
	@$(ECHO) "$(MAGENTA)[3/8] Displaying metrics...$(NC)"
	@$(MAKE) -s demo-metrics
	@sleep 5
	@$(ECHO) ""
	@$(ECHO) "$(MAGENTA)[4/8] Running light load test (60s)...$(NC)"
	@$(MAKE) -s demo-loadtest-light
	@sleep 10
	@$(ECHO) ""
	@$(ECHO) "$(MAGENTA)[5/8] Running medium load test (90s)...$(NC)"
	@$(MAKE) -s demo-loadtest-medium
	@sleep 10
	@$(ECHO) ""
	@$(ECHO) "$(MAGENTA)[6/8] Chaos: Latency injection (60s)...$(NC)"
	@$(MAKE) -s demo-chaos-latency
	@sleep 10
	@$(ECHO) ""
	@$(ECHO) "$(MAGENTA)[7/8] Chaos: Error simulation (45s)...$(NC)"
	@$(MAKE) -s demo-chaos-errors
	@sleep 10
	@$(ECHO) ""
	@$(ECHO) "$(MAGENTA)[8/8] Final metrics snapshot...$(NC)"
	@$(MAKE) -s metrics-snapshot
	@$(ECHO) ""
	@$(ECHO) "$(GREEN)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@$(ECHO) "$(GREEN)║              DEMO COMPLETED SUCCESSFULLY!                      ║$(NC)"
	@$(ECHO) "$(GREEN)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(BLUE)Results saved to:$(NC)"
	@$(ECHO) "  - Metrics: /tmp/metrics-*.json"
	@$(ECHO) "  - Logs: $(APP_LOG_FILE)"
	@$(ECHO) ""

demo-interactive: ## Interactive guided demo with prompts
	@$(ECHO) "$(CYAN)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@$(ECHO) "$(CYAN)║          SIMULATED EXCHANGE - INTERACTIVE DEMO                 ║$(NC)"
	@$(ECHO) "$(CYAN)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(BLUE)This is an interactive demo with pauses between steps.$(NC)"
	@$(ECHO) "$(YELLOW)Press ENTER to continue to each next step.$(NC)"
	@$(ECHO) ""
	@read -p "Press ENTER to start..."
	@$(MAKE) -s _demo-interactive-internal

_demo-interactive-internal:
	@$(ECHO) ""
	@$(ECHO) "$(MAGENTA)══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) "$(MAGENTA)STEP 1: System Startup$(NC)"
	@$(ECHO) "$(MAGENTA)══════════════════════════════════════════════════════════════$(NC)"
	@if docker compose ps | grep -q "simulated-exchange-trading-api.*Up"; then \
		$(ECHO) "$(GREEN)✓ Trading API already running in Docker$(NC)"; \
	else \
		$(MAKE) -s start-average; \
	fi
	@$(ECHO) ""
	@printf '%b' "$(YELLOW)Press ENTER to continue to trading demo...$(NC)"; read -r
	@$(ECHO) ""
	@$(ECHO) "$(MAGENTA)══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) "$(MAGENTA)STEP 2: Basic Trading Operations$(NC)"
	@$(ECHO) "$(MAGENTA)══════════════════════════════════════════════════════════════$(NC)"
	@$(MAKE) -s demo-trading
	@$(ECHO) ""
	@printf '%b' "$(YELLOW)Press ENTER to view metrics...$(NC)"; read -r
	@$(ECHO) ""
	@$(ECHO) "$(MAGENTA)══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) "$(MAGENTA)STEP 3: Real-Time Metrics$(NC)"
	@$(ECHO) "$(MAGENTA)══════════════════════════════════════════════════════════════$(NC)"
	@$(MAKE) -s demo-metrics
	@$(ECHO) ""
	@printf '%b' "$(YELLOW)Press ENTER to start load testing...$(NC)"; read -r
	@$(ECHO) ""
	@$(ECHO) "$(MAGENTA)══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) "$(MAGENTA)STEP 4: Load Testing$(NC)"
	@$(ECHO) "$(MAGENTA)══════════════════════════════════════════════════════════════$(NC)"
	@$(MAKE) -s demo-loadtest-light
	@$(ECHO) ""
	@$(ECHO) "$(GREEN)Interactive demo complete!$(NC)"

demo-presentation: ## Presentation-ready demo with timing (45 min)
	@$(ECHO) "$(CYAN)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@$(ECHO) "$(CYAN)║      SIMULATED EXCHANGE - PRESENTATION DEMO (45 MIN)           ║$(NC)"
	@$(ECHO) "$(CYAN)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@$(ECHO) ""
	@printf '%b' "$(YELLOW)Ready to start presentation? Press ENTER...$(NC)"; read -r
	@$(MAKE) -s _demo-presentation-internal

_demo-presentation-internal:
	@$(ECHO) ""
	@$(ECHO) "$(CYAN)[00:00] Starting presentation...$(NC)"
	@if docker compose ps | grep -q "simulated-exchange-trading-api.*Up"; then \
		$(ECHO) "$(GREEN)✓ Trading API already running in Docker$(NC)"; \
	else \
		$(MAKE) -s start-average; \
	fi
	@sleep 5
	@$(ECHO) ""
	@$(ECHO) "$(CYAN)[02:00] Demo 1: Trading Operations$(NC)"
	@$(MAKE) -s demo-trading
	@sleep 3
	@$(ECHO) ""
	@$(ECHO) "$(CYAN)[07:00] Demo 2: Real-Time Metrics$(NC)"
	@$(MAKE) -s demo-metrics
	@sleep 3
	@$(ECHO) ""
	@$(ECHO) "$(CYAN)[12:00] Demo 3: Load Testing$(NC)"
	@$(MAKE) -s demo-loadtest-medium
	@sleep 5
	@$(ECHO) ""
	@$(ECHO) "$(CYAN)[20:00] Demo 4: Chaos Engineering$(NC)"
	@$(MAKE) -s demo-chaos-latency
	@sleep 5
	@$(ECHO) ""
	@$(ECHO) "$(CYAN)[32:00] Presentation demos complete!$(NC)"
	@$(MAKE) -s metrics-snapshot

demo-trading: ## Basic trading operations demo
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) "$(BLUE)DEMO: Basic Trading Operations$(NC)"
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Placing buy order for 1 BTC at \$$50,000...$(NC)"
	@curl -s -X POST $(ORDERS_URL) \
		-H "Content-Type: application/json" \
		-d '{"user_id":"ce77abc4-2496-40b2-a7ad-72c57368879a","symbol":"BTCUSD","side":"BUY","type":"LIMIT","quantity":1.0,"price":50000.0}' | jq '.'
	@sleep 2
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Placing matching sell order...$(NC)"
	@curl -s -X POST $(ORDERS_URL) \
		-H "Content-Type: application/json" \
		-d '{"user_id":"a1a6ac0e-b1bc-48e2-8f68-a77bdb236680","symbol":"BTCUSD","side":"SELL","type":"LIMIT","quantity":1.0,"price":50000.0}' | jq '.'
	@sleep 2
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Placing market order for 5 ETH...$(NC)"
	@curl -s -X POST $(ORDERS_URL) \
		-H "Content-Type: application/json" \
		-d '{"user_id":"b4450f6c-f3f3-491d-b126-2a0d14addf5a","symbol":"ETHUSD","side":"BUY","type":"MARKET","quantity":5.0,"price":0}' | jq '.'
	@$(ECHO) ""
	@$(ECHO) "$(GREEN)✓ Trading demo complete$(NC)"

demo-metrics: ## Real-time metrics demo
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) "$(BLUE)DEMO: Real-Time Metrics$(NC)"
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Current System Metrics:$(NC)"
	@curl -s $(METRICS_URL) | jq '.data | {order_count, trade_count, avg_latency, orders_per_sec, trades_per_sec, total_volume}'
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Symbol-Specific Metrics:$(NC)"
	@curl -s $(METRICS_URL) | jq '.data.symbol_metrics'
	@$(ECHO) ""
	@$(ECHO) "$(GREEN)✓ Metrics demo complete$(NC)"

demo-loadtest-light: ## Light load test demo (25 ops/sec, 60s)
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) "$(BLUE)DEMO: Light Load Test$(NC)"
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Starting light load test:$(NC)"
	@$(ECHO) "  - Orders/second: $(LIGHT_LOAD_OPS)"
	@$(ECHO) "  - Duration: $(LIGHT_LOAD_DURATION)"
	@$(ECHO) "  - Concurrent users: 10"
	@$(ECHO) ""
	@curl -s -X POST $(LOADTEST_URL) \
		-H "Content-Type: application/json" \
		-d '{"intensity":"light","duration":60,"orders_per_second":$(LIGHT_LOAD_OPS),"concurrent_users":10,"symbols":["BTCUSD","ETHUSD","ADAUSD"]}' | jq '.'
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Monitoring progress (will check every 10s)...$(NC)"
	@for i in 1 2 3 4 5 6; do \
		sleep 10; \
		printf '%b\n' "$(CYAN)[$$i0s]$(NC) Load test status:"; \
		curl -s $(LOADTEST_URL)/status | jq '.data | {is_running, progress, phase, current_metrics: {orders_per_second: .current_metrics.orders_per_second, average_latency: .current_metrics.average_latency, error_rate: .current_metrics.error_rate}}'; \
		echo ""; \
	done
	@$(ECHO) "$(GREEN)✓ Light load test complete$(NC)"

demo-loadtest-medium: ## Medium load test demo (100 ops/sec, 90s)
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) "$(BLUE)DEMO: Medium Load Test$(NC)"
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Starting medium load test:$(NC)"
	@$(ECHO) "  - Orders/second: $(MEDIUM_LOAD_OPS)"
	@$(ECHO) "  - Duration: $(MEDIUM_LOAD_DURATION)"
	@$(ECHO) "  - Concurrent users: 25"
	@$(ECHO) ""
	@curl -s -X POST $(LOADTEST_URL) \
		-H "Content-Type: application/json" \
		-d '{"intensity":"medium","duration":90,"orders_per_second":$(MEDIUM_LOAD_OPS),"concurrent_users":25,"symbols":["BTCUSD","ETHUSD","ADAUSD"]}' | jq '.'
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Monitoring progress...$(NC)"
	@for i in 1 2 3 4 5 6 7 8 9; do \
		sleep 10; \
		printf '%b\n' "$(CYAN)[$$i0s]$(NC) Status:"; \
		curl -s $(LOADTEST_URL)/status | jq '.data | {progress, phase, current_metrics: {orders_per_second: .current_metrics.orders_per_second, average_latency: .current_metrics.average_latency}}'; \
		echo ""; \
	done
	@$(ECHO) "$(GREEN)✓ Medium load test complete$(NC)"

demo-loadtest-heavy: ## Heavy load test demo (200 ops/sec, 120s)
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) "$(BLUE)DEMO: Heavy Load Test$(NC)"
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(RED)⚠ WARNING: This test generates significant load$(NC)"
	@$(ECHO) ""
	@curl -s -X POST $(LOADTEST_URL) \
		-H "Content-Type: application/json" \
		-d '{"scenario":"heavy","duration":"$(HEAVY_LOAD_DURATION)","orders_per_second":$(HEAVY_LOAD_OPS),"concurrent_users":50,"symbols":["BTCUSD","ETHUSD","ADAUSD"]}' | jq '.'
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Monitoring...$(NC)"
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12; do \
		sleep 10; \
		curl -s $(LOADTEST_URL)/status | jq '.data | {progress, phase, current_metrics: {orders_per_second: .current_metrics.orders_per_second, average_latency: .current_metrics.average_latency, error_rate: .current_metrics.error_rate}}'; \
		echo ""; \
	done

demo-chaos-latency: ## Chaos: Latency injection demo
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) "$(BLUE)DEMO: Chaos Engineering - Latency Injection$(NC)"
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Capturing baseline metrics...$(NC)"
	@curl -s $(METRICS_URL) | jq '.data | {avg_latency, orders_per_sec}' > /tmp/baseline-chaos.json
	@cat /tmp/baseline-chaos.json
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Injecting 100ms latency into 50% of requests...$(NC)"
	@curl -s -X POST $(CHAOSTEST_URL) \
		-H "Content-Type: application/json" \
		-d '{"type":"latency_injection","duration":60,"severity":"medium","target":{"component":"trading_engine","percentage":50},"parameters":{"latency_ms":100},"recovery":{"auto_recover":true,"graceful_recover":true}}' | jq '.'
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Monitoring chaos impact (every 15s)...$(NC)"
	@for i in 1 2 3 4; do \
		sleep 15; \
		printf '%b\n' "$(CYAN)[$$i5s]$(NC) Chaos status:"; \
		curl -s $(CHAOSTEST_URL)/status | jq '.data | {phase, metrics: {service_degradation: .metrics.service_degradation, resilience_score: .metrics.resilience_score, errors_generated: .metrics.errors_generated}}'; \
		echo ""; \
	done
	@$(ECHO) "$(GREEN)✓ Chaos test complete$(NC)"

demo-chaos-errors: ## Chaos: Error simulation demo
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) "$(BLUE)DEMO: Chaos Engineering - Error Simulation$(NC)"
	@$(ECHO) "$(BLUE)═══════════════════════════════════════════════════════════════$(NC)"
	@$(ECHO) ""
	@curl -s -X POST $(CHAOSTEST_URL) \
		-H "Content-Type: application/json" \
		-d '{"type":"error_simulation","duration":45,"severity":"low","target":{"component":"order_service","percentage":25},"parameters":{"error_rate":0.15},"recovery":{"auto_recover":true,"graceful_recover":true}}' | jq '.'
	@$(ECHO) ""
	@for i in 1 2 3; do \
		sleep 15; \
		curl -s $(CHAOSTEST_URL)/status | jq '.data | {phase, metrics}'; \
		echo ""; \
	done

# ============================================================
# Section 6: Chaos Experiments (BPF/eBPF-based)
# ============================================================

.PHONY: chaos-all chaos-check chaos-db-death chaos-network-partition chaos-oom-kill chaos-cpu-throttle chaos-disk-io chaos-conn-pool chaos-redis chaos-nginx chaos-latency chaos-cascade chaos-logs

chaos-check:
	@if [ ! -d "$(CHAOS_DIR)" ]; then \
		$(ECHO) "$(RED)✗ Chaos experiments directory not found: $(CHAOS_DIR)$(NC)"; \
		exit 1; \
	fi
	@mkdir -p $(CHAOS_LOG_DIR)

chaos-all: chaos-check ## Run all 10 chaos experiments sequentially
	@$(ECHO) "$(CYAN)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@$(ECHO) "$(CYAN)║          RUNNING ALL CHAOS EXPERIMENTS                         ║$(NC)"
	@$(ECHO) "$(CYAN)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)This will run 10 comprehensive chaos experiments with BPF monitoring$(NC)"
	@$(ECHO) "$(YELLOW)Estimated duration: 90-120 minutes$(NC)"
	@$(ECHO) ""
	@read -p "Continue? [y/N] " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		for exp in $(CHAOS_DIR)/0*.sh $(CHAOS_DIR)/10-*.sh; do \
			if [ -f "$$exp" ]; then \
				echo ""; \
				printf '%b\n' "$(MAGENTA)═══════════════════════════════════════════════════════════════$(NC)"; \
				printf '%b\n' "$(MAGENTA)Running: $$(basename $$exp)$(NC)"; \
				printf '%b\n' "$(MAGENTA)═══════════════════════════════════════════════════════════════$(NC)"; \
				sudo "$$exp" 2>&1 | tee "$(CHAOS_LOG_DIR)/$$(basename $$exp .sh)-$$(date +%Y%m%d-%H%M%S).log"; \
				printf '%b\n' "$(YELLOW)Waiting 30s before next experiment...$(NC)"; \
				sleep 30; \
			fi; \
		done; \
		echo ""; \
		printf '%b\n' "$(GREEN)All chaos experiments complete!$(NC)"; \
		printf '%b\n' "$(BLUE)Logs saved to: $(CHAOS_LOG_DIR)$(NC)"; \
	fi

chaos-db-death: chaos-check ## Chaos Experiment 1: Database sudden death
	@$(ECHO) "$(MAGENTA)Chaos Experiment 1: Database Sudden Death$(NC)"
	@$(ECHO) "$(YELLOW)Tests: Connection retry, fallback mechanisms, recovery$(NC)"
	@sudo $(CHAOS_DIR)/01-database-sudden-death.sh 2>&1 | tee $(CHAOS_LOG_DIR)/01-db-death-$$(date +%Y%m%d-%H%M%S).log

chaos-network-partition: chaos-check ## Chaos Experiment 2: Network partition (iptables + BPF)
	@$(ECHO) "$(MAGENTA)Chaos Experiment 2: Network Partition$(NC)"
	@$(ECHO) "$(YELLOW)Tests: TCP timeout, retries, iptables filtering$(NC)"
	@sudo $(CHAOS_DIR)/02-network-partition.sh 2>&1 | tee $(CHAOS_LOG_DIR)/02-network-partition-$$(date +%Y%m%d-%H%M%S).log

chaos-oom-kill: chaos-check ## Chaos Experiment 3: Memory OOM kill (eBPF monitoring)
	@$(ECHO) "$(MAGENTA)Chaos Experiment 3: Memory Pressure & OOM Kill$(NC)"
	@$(ECHO) "$(YELLOW)Tests: Memory limits, OOM killer, auto-restart$(NC)"
	@sudo $(CHAOS_DIR)/03-memory-oom-kill.sh 2>&1 | tee $(CHAOS_LOG_DIR)/03-oom-kill-$$(date +%Y%m%d-%H%M%S).log

chaos-cpu-throttle: chaos-check ## Chaos Experiment 4: CPU throttling (BPF scheduler monitoring)
	@$(ECHO) "$(MAGENTA)Chaos Experiment 4: CPU Throttling$(NC)"
	@$(ECHO) "$(YELLOW)Tests: CPU limits, scheduler behavior, queue buildup$(NC)"
	@sudo $(CHAOS_DIR)/04-cpu-throttling.sh 2>&1 | tee $(CHAOS_LOG_DIR)/04-cpu-throttle-$$(date +%Y%m%d-%H%M%S).log

chaos-disk-io: chaos-check ## Chaos Experiment 5: Disk I/O starvation (BPF I/O tracing)
	@$(ECHO) "$(MAGENTA)Chaos Experiment 5: Disk I/O Starvation$(NC)"
	@$(ECHO) "$(YELLOW)Tests: I/O saturation, query performance, checkpoint delays$(NC)"
	@sudo $(CHAOS_DIR)/05-disk-io-starvation.sh 2>&1 | tee $(CHAOS_LOG_DIR)/05-disk-io-$$(date +%Y%m%d-%H%M%S).log

chaos-conn-pool: chaos-check ## Chaos Experiment 6: Connection pool exhaustion
	@$(ECHO) "$(MAGENTA)Chaos Experiment 6: Connection Pool Exhaustion$(NC)"
	@$(ECHO) "$(YELLOW)Tests: Connection limits, queue behavior, error handling$(NC)"
	@sudo $(CHAOS_DIR)/06-connection-pool-exhaustion.sh 2>&1 | tee $(CHAOS_LOG_DIR)/06-conn-pool-$$(date +%Y%m%d-%H%M%S).log

chaos-redis: chaos-check ## Chaos Experiment 7: Redis cache failure
	@$(ECHO) "$(MAGENTA)Chaos Experiment 7: Redis Cache Failure$(NC)"
	@$(ECHO) "$(YELLOW)Tests: Cache fallback, database load increase$(NC)"
	@$(CHAOS_DIR)/07-redis-cache-failure.sh 2>&1 | tee $(CHAOS_LOG_DIR)/07-redis-$$(date +%Y%m%d-%H%M%S).log

chaos-nginx: chaos-check ## Chaos Experiment 8: Nginx reverse proxy failure
	@$(ECHO) "$(MAGENTA)Chaos Experiment 8: Nginx Proxy Failure$(NC)"
	@$(ECHO) "$(YELLOW)Tests: Reverse proxy dependency, direct access paths$(NC)"
	@$(CHAOS_DIR)/08-nginx-failure.sh 2>&1 | tee $(CHAOS_LOG_DIR)/08-nginx-$$(date +%Y%m%d-%H%M%S).log

chaos-latency: chaos-check ## Chaos Experiment 9: Network latency injection (tc netem)
	@$(ECHO) "$(MAGENTA)Chaos Experiment 9: Network Latency Injection$(NC)"
	@$(ECHO) "$(YELLOW)Tests: Latency tolerance, timeout tuning, retry behavior$(NC)"
	@sudo $(CHAOS_DIR)/09-network-latency.sh 2>&1 | tee $(CHAOS_LOG_DIR)/09-latency-$$(date +%Y%m%d-%H%M%S).log || $(ECHO) "$(RED)Note: This experiment requires tc netem kernel module$(NC)"

chaos-cascade: chaos-check ## Chaos Experiment 10: Multi-service cascade failure
	@$(ECHO) "$(MAGENTA)Chaos Experiment 10: Multi-Service Cascade Failure$(NC)"
	@$(ECHO) "$(YELLOW)Tests: Dependency management, startup ordering, recovery$(NC)"
	@$(CHAOS_DIR)/10-multi-service-failure.sh 2>&1 | tee $(CHAOS_LOG_DIR)/10-cascade-$$(date +%Y%m%d-%H%M%S).log

chaos-logs: ## View chaos experiment logs
	@if [ -d "$(CHAOS_LOG_DIR)" ] && [ -n "$$(ls -A $(CHAOS_LOG_DIR) 2>/dev/null)" ]; then \
		$(ECHO) "$(BLUE)Available chaos experiment logs:$(NC)"; \
		ls -lht $(CHAOS_LOG_DIR) | head -20; \
		echo ""; \
		$(ECHO) "$(YELLOW)View specific log with: less $(CHAOS_LOG_DIR)/<filename>$(NC)"; \
	else \
		$(ECHO) "$(YELLOW)No chaos experiment logs found$(NC)"; \
		echo "Run chaos experiments first with: make chaos-all"; \
	fi

# ============================================================
# Section 6.5: Profiling & Flamegraphs
# ============================================================

.PHONY: profile-cpu profile-cpu-long profile-heap profile-goroutine profile-allocs profile-block profile-mutex profile-all chaos-flamegraph profile-view profile-clean profile-serve profile-analyze profile-help

profile-cpu: ## Generate CPU flamegraph (30 seconds)
	@echo "🔥 Generating CPU flamegraph..."
	@./scripts/generate-flamegraph.sh cpu 30 8080

profile-cpu-long: ## Generate CPU flamegraph (60 seconds)
	@echo "🔥 Generating CPU flamegraph (60s)..."
	@./scripts/generate-flamegraph.sh cpu 60 8080

profile-heap: ## Generate heap memory flamegraph
	@echo "🔥 Generating heap flamegraph..."
	@./scripts/generate-flamegraph.sh heap 0 8080

profile-goroutine: ## Generate goroutine flamegraph
	@echo "🔥 Generating goroutine flamegraph..."
	@./scripts/generate-flamegraph.sh goroutine 0 8080

profile-allocs: ## Generate memory allocations flamegraph (30 seconds)
	@echo "🔥 Generating allocations flamegraph..."
	@./scripts/generate-flamegraph.sh allocs 30 8080

profile-block: ## Generate blocking operations flamegraph
	@echo "🔥 Generating block flamegraph..."
	@./scripts/generate-flamegraph.sh block 0 8080

profile-mutex: ## Generate mutex contention flamegraph
	@echo "🔥 Generating mutex flamegraph..."
	@./scripts/generate-flamegraph.sh mutex 0 8080

profile-all: ## Generate all flamegraph types
	@echo "🔥 Generating all flamegraphs..."
	@echo "This will take approximately 2-3 minutes..."
	@./scripts/generate-flamegraph.sh cpu 30 8080
	@./scripts/generate-flamegraph.sh heap 0 8080
	@./scripts/generate-flamegraph.sh goroutine 0 8080
	@./scripts/generate-flamegraph.sh allocs 30 8080
	@./scripts/generate-flamegraph.sh block 0 8080
	@./scripts/generate-flamegraph.sh mutex 0 8080
	@echo "✅ All flamegraphs generated in ./flamegraphs/"

chaos-flamegraph: ## Run flamegraph profiling chaos experiment
	@echo "🔥 Running flamegraph profiling chaos experiment..."
	@./chaos-experiments/11-flamegraph-profiling.sh

profile-view: ## Open most recent flamegraphs in browser
	@echo "🌐 Opening flamegraphs in browser..."
	@if command -v firefox >/dev/null 2>&1; then \
		firefox flamegraphs/*.svg 2>/dev/null & \
	elif command -v chromium >/dev/null 2>&1; then \
		chromium flamegraphs/*.svg 2>/dev/null & \
	elif command -v google-chrome >/dev/null 2>&1; then \
		google-chrome flamegraphs/*.svg 2>/dev/null & \
	else \
		echo "No browser found. Please open files in ./flamegraphs/ manually."; \
	fi

profile-clean: ## Clean up flamegraph files
	@echo "🧹 Cleaning flamegraph files..."
	@rm -rf flamegraphs/*.svg flamegraphs/*.pb.gz flamegraphs/*.txt
	@echo "✅ Flamegraph files cleaned"

profile-serve: ## Start web server to view flamegraphs remotely
	@echo "🌐 Starting flamegraph web server..."
	@./scripts/serve-flamegraphs.sh 8888

profile-analyze: ## AI-powered analysis of most recent CPU profile
	@echo "🤖 Analyzing profile with AI..."
	@LATEST=$$(ls -t flamegraphs/cpu_profile_*.pb.gz 2>/dev/null | head -1); \
	if [ -n "$$LATEST" ]; then \
		python3 scripts/ai-flamegraph-analyzer.py "$$LATEST"; \
	else \
		echo "No CPU profiles found. Run 'make profile-cpu' first."; \
	fi

profile-help: ## Show profiling and flamegraph help
	@echo "🔥 Profiling & Flamegraph Targets"
	@echo "=================================="
	@echo ""
	@echo "Quick Start:"
	@echo "  make profile-cpu           - Generate CPU flamegraph (30s)"
	@echo "  make profile-all           - Generate all flamegraph types"
	@echo "  make chaos-flamegraph      - Full profiling chaos experiment"
	@echo ""
	@echo "Individual Profile Types:"
	@echo "  make profile-cpu           - CPU usage (where time is spent)"
	@echo "  make profile-cpu-long      - CPU usage (60s for rare code paths)"
	@echo "  make profile-heap          - Memory allocations"
	@echo "  make profile-goroutine     - Goroutine states and stacks"
	@echo "  make profile-allocs        - Memory allocation patterns"
	@echo "  make profile-block         - Blocking operations"
	@echo "  make profile-mutex         - Mutex contention"
	@echo ""
	@echo "Utilities:"
	@echo "  make profile-view          - Open flamegraphs in browser"
	@echo "  make profile-serve         - Start web server for remote access"
	@echo "  make profile-analyze       - AI analysis of latest CPU profile"
	@echo "  make profile-clean         - Clean up flamegraph files"
	@echo ""
	@echo "Manual Profiling:"
	@echo "  ./scripts/generate-flamegraph.sh [type] [duration] [port]"
	@echo ""
	@echo "Profile Types:"
	@echo "  cpu        - CPU profiling (requires duration)"
	@echo "  heap       - Heap memory snapshot (duration=0)"
	@echo "  goroutine  - Goroutine snapshot (duration=0)"
	@echo "  allocs     - Allocations over time (requires duration)"
	@echo "  block      - Blocking operations snapshot (duration=0)"
	@echo "  mutex      - Mutex contention snapshot (duration=0)"
	@echo ""
	@echo "Documentation:"
	@echo "  docs/FLAMEGRAPH_GUIDE.md   - Complete flamegraph guide"
	@echo ""
	@echo "Examples:"
	@echo "  # CPU profile for 60 seconds on port 8080"
	@echo "  ./scripts/generate-flamegraph.sh cpu 60 8080"
	@echo ""
	@echo "  # Heap snapshot on market simulator (port 8081)"
	@echo "  ./scripts/generate-flamegraph.sh heap 0 8081"
	@echo ""
	@echo "  # Interactive pprof analysis"
	@echo "  go tool pprof flamegraphs/cpu_profile_*.pb.gz"
	@echo ""
	@echo "💡 Tip: Run 'make chaos-flamegraph' for comprehensive profiling"
	@echo "    under various load conditions with detailed analysis!"

# ============================================================
# Section 7: Monitoring & Dashboards
# ============================================================

.PHONY: dashboard grafana grafana-start grafana-stop grafana-restart grafana-logs caddy caddy-start caddy-stop caddy-restart caddy-logs caddy-reload caddy-certs caddy-test watch-metrics watch-health watch-logs metrics-snapshot

dashboard: ## Open web dashboard in browser
	@$(ECHO) "$(BLUE)Opening dashboard in browser...$(NC)"
	@$(ECHO) "$(YELLOW)URL: $(DASHBOARD_URL)$(NC)"
	@if command -v xdg-open > /dev/null; then \
		xdg-open $(DASHBOARD_URL) & \
	elif command -v open > /dev/null; then \
		open $(DASHBOARD_URL) & \
	else \
		$(ECHO) "$(YELLOW)Please open manually: $(DASHBOARD_URL)$(NC)"; \
	fi

grafana: ## Open Grafana dashboard in browser
	@$(ECHO) "$(BLUE)Opening Grafana dashboard...$(NC)"
	@$(ECHO) ""
	@if docker compose ps | grep -q "simulated-exchange-grafana.*Up"; then \
		$(ECHO) "$(GREEN)✓ Grafana is running$(NC)"; \
		$(ECHO) "$(CYAN)URL: http://localhost:3000$(NC)"; \
		$(ECHO) "$(YELLOW)Username: admin$(NC)"; \
		$(ECHO) "$(YELLOW)Password: admin123$(NC)"; \
		$(ECHO) ""; \
		if command -v xdg-open > /dev/null; then \
			xdg-open http://localhost:3000 & \
		elif command -v open > /dev/null; then \
			open http://localhost:3000 & \
		else \
			$(ECHO) "$(YELLOW)Please open manually: http://localhost:3000$(NC)"; \
		fi; \
	else \
		$(ECHO) "$(RED)✗ Grafana is not running$(NC)"; \
		$(ECHO) "$(YELLOW)Start Grafana with: make grafana-start$(NC)"; \
		exit 1; \
	fi

grafana-start: ## Start Grafana container
	@$(ECHO) "$(BLUE)Starting Grafana...$(NC)"
	@docker compose up -d grafana
	@sleep 3
	@if docker compose ps | grep -q "simulated-exchange-grafana.*Up"; then \
		$(ECHO) "$(GREEN)✓ Grafana started successfully$(NC)"; \
		$(ECHO) "$(CYAN)URL: http://localhost:3000$(NC)"; \
		$(ECHO) "$(YELLOW)Username: admin$(NC)"; \
		$(ECHO) "$(YELLOW)Password: admin123$(NC)"; \
	else \
		$(ECHO) "$(RED)✗ Failed to start Grafana$(NC)"; \
		$(ECHO) "$(YELLOW)Check logs: make grafana-logs$(NC)"; \
		exit 1; \
	fi

grafana-stop: ## Stop Grafana container
	@$(ECHO) "$(BLUE)Stopping Grafana...$(NC)"
	@docker compose stop grafana
	@$(ECHO) "$(GREEN)✓ Grafana stopped$(NC)"

grafana-restart: ## Restart Grafana container
	@$(ECHO) "$(BLUE)Restarting Grafana...$(NC)"
	@docker compose restart grafana
	@sleep 3
	@$(ECHO) "$(GREEN)✓ Grafana restarted$(NC)"

grafana-logs: ## View Grafana logs
	@$(ECHO) "$(BLUE)Grafana logs (Ctrl+C to stop)...$(NC)"
	@docker compose logs -f grafana

caddy: ## Open Caddy admin interface
	@$(ECHO) "$(BLUE)Caddy Configuration...$(NC)"
	@$(ECHO) ""
	@if docker compose -f docker-compose.yml -f docker/docker-compose.caddy.yml ps | grep -q "simulated-exchange-caddy.*Up"; then \
		$(ECHO) "$(GREEN)✓ Caddy is running$(NC)"; \
		$(ECHO) "$(CYAN)Admin API: http://localhost:2019$(NC)"; \
		$(ECHO) "$(CYAN)Dev Interface: http://localhost:8888$(NC)"; \
		$(ECHO) ""; \
		$(ECHO) "$(YELLOW)Domain Services:$(NC)"; \
		$(ECHO) "  • Main: https://andythomas-sre.com"; \
		$(ECHO) "  • API: https://api.andythomas-sre.com"; \
		$(ECHO) "  • Grafana: https://grafana.andythomas-sre.com"; \
		$(ECHO) "  • Prometheus: https://prometheus.andythomas-sre.com"; \
	else \
		$(ECHO) "$(RED)✗ Caddy is not running$(NC)"; \
		$(ECHO) "$(YELLOW)Start Caddy with: make caddy-start$(NC)"; \
		exit 1; \
	fi

caddy-start: ## Start Caddy reverse proxy
	@$(ECHO) "$(BLUE)Starting Caddy reverse proxy...$(NC)"
	@mkdir -p logs/caddy
	@docker compose -f docker-compose.yml -f docker/docker-compose.caddy.yml up -d caddy
	@sleep 3
	@if docker compose -f docker-compose.yml -f docker/docker-compose.caddy.yml ps | grep -q "simulated-exchange-caddy.*Up"; then \
		$(ECHO) "$(GREEN)✓ Caddy started successfully$(NC)"; \
		$(ECHO) "$(CYAN)Admin API: http://localhost:2019$(NC)"; \
		$(ECHO) ""; \
		$(ECHO) "$(YELLOW)⚠️  Important: Configure DNS for andythomas-sre.com$(NC)"; \
		$(ECHO) "$(YELLOW)See: docker/DNS-SETUP.md$(NC)"; \
	else \
		$(ECHO) "$(RED)✗ Failed to start Caddy$(NC)"; \
		$(ECHO) "$(YELLOW)Check logs: make caddy-logs$(NC)"; \
		exit 1; \
	fi

caddy-stop: ## Stop Caddy reverse proxy
	@$(ECHO) "$(BLUE)Stopping Caddy...$(NC)"
	@docker compose -f docker-compose.yml -f docker/docker-compose.caddy.yml stop caddy
	@$(ECHO) "$(GREEN)✓ Caddy stopped$(NC)"

caddy-restart: ## Restart Caddy reverse proxy
	@$(ECHO) "$(BLUE)Restarting Caddy...$(NC)"
	@docker compose -f docker-compose.yml -f docker/docker-compose.caddy.yml restart caddy
	@sleep 3
	@$(ECHO) "$(GREEN)✓ Caddy restarted$(NC)"

caddy-reload: ## Reload Caddy configuration (zero-downtime)
	@$(ECHO) "$(BLUE)Reloading Caddy configuration...$(NC)"
	@docker exec simulated-exchange-caddy caddy reload --config /etc/caddy/Caddyfile
	@$(ECHO) "$(GREEN)✓ Configuration reloaded$(NC)"

caddy-logs: ## View Caddy logs
	@$(ECHO) "$(BLUE)Caddy logs (Ctrl+C to stop)...$(NC)"
	@docker compose -f docker-compose.yml -f docker/docker-compose.caddy.yml logs -f caddy

caddy-certs: ## Show SSL certificate information
	@$(ECHO) "$(BLUE)Caddy SSL Certificates...$(NC)"
	@$(ECHO) ""
	@docker exec simulated-exchange-caddy caddy list-certificates 2>/dev/null || \
		$(ECHO) "$(YELLOW)No certificates found. Make sure DNS is configured.$(NC)"

caddy-test: ## Test all Caddy endpoints
	@$(ECHO) "$(BLUE)Testing Caddy endpoints...$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Testing localhost endpoints:$(NC)"
	@curl -s -o /dev/null -w "  Admin API (2019):    %{http_code}\n" http://localhost:2019/config/ || echo "  Admin API: Failed"
	@curl -s -o /dev/null -w "  Dev Interface (8888): %{http_code}\n" http://localhost:8888/ || echo "  Dev Interface: Failed"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)Testing service proxies (if DNS configured):$(NC)"
	@$(ECHO) "  Main: curl -I https://andythomas-sre.com"
	@$(ECHO) "  API: curl -I https://api.andythomas-sre.com"
	@$(ECHO) "  Grafana: curl -I https://grafana.andythomas-sre.com"

caddy-external-check: ## Check if external access is configured correctly
	@$(ECHO) "$(BLUE)Checking External Access Configuration$(NC)"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)1. Finding Public IP Address:$(NC)"
	@PUBLIC_IP=$$(curl -s ifconfig.me 2>/dev/null || curl -s icanhazip.com 2>/dev/null || echo "Unable to detect"); \
	$(ECHO) "   Your public IP: $$PUBLIC_IP"
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)2. Checking DNS Resolution:$(NC)"
	@DNS_IP=$$(dig +short andythomas-sre.com @8.8.8.8 | head -1); \
	if [ -n "$$DNS_IP" ]; then \
		$(ECHO) "   andythomas-sre.com → $$DNS_IP"; \
		PUBLIC_IP=$$(curl -s ifconfig.me); \
		if [ "$$DNS_IP" = "$$PUBLIC_IP" ]; then \
			$(ECHO) "   $(GREEN)✓ DNS correctly points to your server$(NC)"; \
		else \
			$(ECHO) "   $(RED)✗ DNS points to $$DNS_IP but your public IP is $$PUBLIC_IP$(NC)"; \
		fi; \
	else \
		$(ECHO) "   $(RED)✗ DNS not configured yet$(NC)"; \
		$(ECHO) "   $(YELLOW)See: docker/DNS-SETUP.md$(NC)"; \
	fi
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)3. Checking Firewall:$(NC)"
	@if command -v ufw >/dev/null 2>&1; then \
		if sudo ufw status | grep -q "80/tcp.*ALLOW"; then \
			$(ECHO) "   $(GREEN)✓ Port 80 open$(NC)"; \
		else \
			$(ECHO) "   $(RED)✗ Port 80 blocked$(NC)"; \
			$(ECHO) "   Fix: sudo ufw allow 80/tcp"; \
		fi; \
		if sudo ufw status | grep -q "443/tcp.*ALLOW"; then \
			$(ECHO) "   $(GREEN)✓ Port 443 open$(NC)"; \
		else \
			$(ECHO) "   $(RED)✗ Port 443 blocked$(NC)"; \
			$(ECHO) "   Fix: sudo ufw allow 443/tcp"; \
		fi; \
	else \
		$(ECHO) "   $(YELLOW)⚠ UFW not detected, manual check needed$(NC)"; \
	fi
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)4. Checking Caddy Status:$(NC)"
	@if docker compose -f docker-compose.yml -f docker/docker-compose.caddy.yml ps | grep -q "simulated-exchange-caddy.*Up"; then \
		$(ECHO) "   $(GREEN)✓ Caddy is running$(NC)"; \
	else \
		$(ECHO) "   $(RED)✗ Caddy not running$(NC)"; \
		$(ECHO) "   Fix: make caddy-start"; \
	fi
	@$(ECHO) ""
	@$(ECHO) "$(YELLOW)5. Next Steps:$(NC)"
	@$(ECHO) "   • Quick guide: docker/QUICK-ACCESS-GUIDE.md"
	@$(ECHO) "   • Full guide: docker/EXTERNAL-ACCESS.md"
	@$(ECHO) "   • Test from external device: curl -I https://andythomas-sre.com"

watch-metrics: ## Live metrics updates (every 5s)
	@$(ECHO) "$(BLUE)Watching metrics (Ctrl+C to stop)...$(NC)"
	@$(ECHO) ""
	@watch -n 5 'curl -s $(METRICS_URL) | jq ".data | {order_count, trade_count, total_volume, avg_latency, orders_per_sec, trades_per_sec}"'

watch-health: ## Live health check (every 2s)
	@$(ECHO) "$(BLUE)Watching health (Ctrl+C to stop)...$(NC)"
	@$(ECHO) ""
	@watch -n 2 'curl -s $(HEALTH_URL) | jq "."'

watch-logs: ## Tail application logs
	@if [ -f $(APP_LOG_FILE) ]; then \
		$(ECHO) "$(BLUE)Watching application logs (Ctrl+C to stop)...$(NC)"; \
		tail -f $(APP_LOG_FILE); \
	else \
		$(ECHO) "$(RED)✗ Log file not found: $(APP_LOG_FILE)$(NC)"; \
		$(ECHO) ""; \
		if ps aux | grep -q "[g]o run.*api/main.go"; then \
			$(ECHO) "$(YELLOW)The application is running, but not in the expected mode.$(NC)"; \
			$(ECHO) "$(YELLOW)To enable log watching:$(NC)"; \
			$(ECHO) "  1. Stop current instances: make stop"; \
			$(ECHO) "  2. Start in background mode: make start-average"; \
		else \
			$(ECHO) "$(YELLOW)The application is not running.$(NC)"; \
			$(ECHO) "$(YELLOW)Start it with: make start-average$(NC)"; \
		fi; \
		exit 1; \
	fi

metrics-snapshot: ## Save current metrics to file
	@FILENAME="/tmp/metrics-snapshot-$$(date +%Y%m%d-%H%M%S).json"; \
	curl -s $(METRICS_URL) > $$FILENAME; \
	$(ECHO) "$(GREEN)✓ Metrics saved to: $$FILENAME$(NC)"; \
	cat $$FILENAME | jq '.'

# ============================================================
# Section 8: Testing & Validation
# ============================================================

.PHONY: test-all test-unit test-demo test-e2e test-perf benchmark deps lint fmt vet security coverage coverage-ci

# Dependency management
deps: ## Download and verify dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod verify
	go mod tidy

# Code quality
fmt: ## Format code
	@echo "🎨 Formatting code..."
	go fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	go vet ./...

lint: ## Run linter
	@echo "🧹 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

security: ## Run security checks
	@echo "🔒 Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Testing targets
test-unit: ## Run unit tests with coverage
	@echo "🧪 Running unit tests..."
	mkdir -p $(TEST_RESULTS_DIR)
	go test -v -race -coverprofile=$(TEST_RESULTS_DIR)/$(COVERAGE_FILE) -covermode=atomic ./internal/... | tee $(TEST_RESULTS_DIR)/unit-tests.log

test-demo: ## Run demo system tests
	@echo "🎭 Running demo system tests..."
	mkdir -p $(TEST_RESULTS_DIR)
	go test -v -race ./internal/demo/... -timeout=5m | tee $(TEST_RESULTS_DIR)/demo-tests.log

test-e2e: ## Run E2E tests
	@echo "🌐 Running E2E tests..."
	mkdir -p $(TEST_RESULTS_DIR)
	go test -v -race ./test/e2e/... -timeout=15m | tee $(TEST_RESULTS_DIR)/e2e-tests.log

test-perf: ## Run performance tests
	@echo "⚡ Running performance tests..."
	mkdir -p $(TEST_RESULTS_DIR)
	go test -v -race ./test/e2e/ -run=TestPerformanceTestSuite -timeout=10m | tee $(TEST_RESULTS_DIR)/perf-tests.log

benchmark: ## Run benchmarks
	@echo "📊 Running benchmarks..."
	mkdir -p $(TEST_RESULTS_DIR)
	go test -v -bench=. -benchmem -run=^$$ ./test/e2e/... | tee $(TEST_RESULTS_DIR)/benchmarks.log

test-all: clean lint test-unit test-demo test-e2e ## Run all tests (except Docker)
	@echo "✅ All tests completed!"

# Coverage targets
coverage: test-unit ## Generate coverage report
	@echo "📈 Generating coverage report..."
	go tool cover -html=$(TEST_RESULTS_DIR)/$(COVERAGE_FILE) -o $(TEST_RESULTS_DIR)/coverage.html
	go tool cover -func=$(TEST_RESULTS_DIR)/$(COVERAGE_FILE) -o $(TEST_RESULTS_DIR)/coverage.txt
	@echo "Coverage report generated: $(TEST_RESULTS_DIR)/coverage.html"
	@echo "Coverage summary:"
	@tail -1 $(TEST_RESULTS_DIR)/coverage.txt

coverage-ci: ## Generate coverage for CI (with specific format)
	@echo "📈 Generating CI coverage report..."
	mkdir -p $(TEST_RESULTS_DIR)
	go test -v -race -coverprofile=$(TEST_RESULTS_DIR)/$(COVERAGE_FILE) -covermode=atomic ./internal/...
	go tool cover -func=$(TEST_RESULTS_DIR)/$(COVERAGE_FILE) > $(TEST_RESULTS_DIR)/coverage-summary.txt
	@echo "Coverage for CI generated"

# ============================================================
# Section 9: Docker & CI/CD
# ============================================================

.PHONY: docker-build docker-test-unit docker-test-e2e docker-test-all ci-test ci-e2e ci-full

docker-build: ## Build Docker test image
	@echo "🐳 Building Docker test image..."
	docker build -f Dockerfile.test -t $(PROJECT_NAME):test .

docker-test-unit: ## Run unit tests in Docker
	@echo "🐳 Running unit tests in Docker..."
	docker run --rm -v $(PWD)/$(TEST_RESULTS_DIR):/app/$(TEST_RESULTS_DIR) $(PROJECT_NAME):test \
		sh -c "go test -v -race -coverprofile=/app/$(TEST_RESULTS_DIR)/$(COVERAGE_FILE) ./internal/..."

docker-test-e2e: ## Run E2E tests with Docker Compose
	@echo "🐳 Running E2E tests with Docker Compose..."
	docker compose -f $(DOCKER_COMPOSE_FILE) up --build --abort-on-container-exit
	docker compose -f $(DOCKER_COMPOSE_FILE) down --volumes

docker-test-all: docker-build ## Run comprehensive Docker test suite
	@echo "🐳 Running comprehensive Docker test suite..."
	./scripts/run-tests.sh

# CI/CD targets
ci-test: deps lint test-unit coverage-ci ## Run CI test suite
	@echo "🔄 CI test suite completed"

ci-e2e: test-e2e test-demo ## Run CI E2E tests
	@echo "🔄 CI E2E tests completed"

ci-full: ci-test ci-e2e benchmark ## Run full CI suite
	@echo "🔄 Full CI suite completed"

# ============================================================
# Section 10: Utilities & Maintenance
# ============================================================

.PHONY: clean clean-chaos clean-docker clean-logs nuke install-deps install-bpf check-prereqs check-go dev-setup version

clean: ## Clean temporary files and logs
	@$(ECHO) "$(BLUE)Cleaning temporary files...$(NC)"
	@rm -f $(APP_LOG_FILE)
	@rm -f $(APP_PID_FILE)
	@rm -f /tmp/metrics-*.json
	@rm -f /tmp/baseline-*.json
	@rm -f /tmp/chaos-exp-*.log
	@rm -f /tmp/chaos-exp-*.log.bpf
	@rm -rf $(TEST_RESULTS_DIR)
	@rm -f coverage.out coverage.html coverage.txt
	@$(ECHO) "$(GREEN)✓ Cleanup complete$(NC)"

clean-chaos: ## Clean chaos experiment artifacts
	@$(ECHO) "$(BLUE)Cleaning chaos experiment logs...$(NC)"
	@rm -rf $(CHAOS_LOG_DIR)
	@mkdir -p $(CHAOS_LOG_DIR)
	@rm -f /tmp/chaos-exp-*.log*
	@rm -f /tmp/*-bpf.bt
	@$(ECHO) "$(GREEN)✓ Chaos artifacts cleaned$(NC)"

clean-docker: ## Clean Docker images and containers
	@echo "🐳 Cleaning Docker images and containers..."
	docker compose -f $(DOCKER_COMPOSE_FILE) down --volumes --remove-orphans --rmi all 2>/dev/null || true
	docker image prune -f --filter "label=project=$(PROJECT_NAME)" 2>/dev/null || true

clean-logs: ## Clean all log files (application, tests, chaos experiments)
	@$(ECHO) "$(BLUE)Cleaning all log files...$(NC)"
	@echo ""
	@echo "📋 Removing log files:"
	@# Application logs
	@if [ -f $(APP_LOG_FILE) ]; then \
		echo "  • Application log: $(APP_LOG_FILE)"; \
		rm -f $(APP_LOG_FILE); \
	fi
	@# Test output logs
	@if [ -d "$(TEST_RESULTS_DIR)" ]; then \
		echo "  • Test result logs: $(TEST_RESULTS_DIR)/*.log"; \
		rm -f $(TEST_RESULTS_DIR)/*.log; \
	fi
	@# Chaos experiment logs in /tmp
	@if ls /tmp/chaos-exp-*.log* >/dev/null 2>&1; then \
		echo "  • Chaos experiment logs: /tmp/chaos-exp-*.log*"; \
		rm -f /tmp/chaos-exp-*.log*; \
	fi
	@# Chaos logs directory
	@if [ -d "$(CHAOS_LOG_DIR)" ] && [ -n "$$(ls -A $(CHAOS_LOG_DIR) 2>/dev/null)" ]; then \
		echo "  • Chaos log directory: $(CHAOS_LOG_DIR)/*"; \
		rm -f $(CHAOS_LOG_DIR)/*; \
	fi
	@# BPF trace logs
	@if ls /tmp/*-bpf.bt >/dev/null 2>&1; then \
		echo "  • BPF trace files: /tmp/*-bpf.bt"; \
		rm -f /tmp/*-bpf.bt; \
	fi
	@# Temporary metric files
	@if ls /tmp/metrics-*.json >/dev/null 2>&1; then \
		echo "  • Metric snapshots: /tmp/metrics-*.json"; \
		rm -f /tmp/metrics-*.json; \
	fi
	@if ls /tmp/baseline-*.json >/dev/null 2>&1; then \
		echo "  • Baseline metrics: /tmp/baseline-*.json"; \
		rm -f /tmp/baseline-*.json; \
	fi
	@# Docker build logs
	@if [ -f /tmp/docker-build.log ]; then \
		echo "  • Docker build log: /tmp/docker-build.log"; \
		rm -f /tmp/docker-build.log; \
	fi
	@# Preflight check logs
	@if [ -f /tmp/preflight-check.log ]; then \
		echo "  • Preflight check log: /tmp/preflight-check.log"; \
		rm -f /tmp/preflight-check.log; \
	fi
	@# BPF validation logs
	@if [ -f /tmp/bpf-validation.log ]; then \
		echo "  • BPF validation log: /tmp/bpf-validation.log"; \
		rm -f /tmp/bpf-validation.log; \
	fi
	@# Prep demo logs
	@if [ -f /tmp/prep-demo-chaos.log ]; then \
		echo "  • Demo prep log: /tmp/prep-demo-chaos.log"; \
		rm -f /tmp/prep-demo-chaos.log; \
	fi
	@echo ""
	@$(ECHO) "$(GREEN)✓ All log files cleaned$(NC)"

nuke: ## ⚠️  DANGER: Stop and remove ALL Docker containers system-wide
	@echo "☢️  NUCLEAR OPTION - STOPPING ALL DOCKER CONTAINERS ☢️"
	@echo "=================================================="
	@echo "⚠️  WARNING: This will stop and remove ALL Docker containers"
	@echo "⚠️  on this system, regardless of which project started them!"
	@echo ""
	@CONTAINER_COUNT=$$(docker ps -q | wc -l | tr -d ' '); \
	if [ "$$CONTAINER_COUNT" -eq 0 ]; then \
		echo "✅ No running containers found. System is clean."; \
	else \
		echo "📋 Currently running containers ($$CONTAINER_COUNT):"; \
		docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"; \
		echo ""; \
		echo "🛑 Stopping all containers..."; \
		docker stop $$(docker ps -q) 2>/dev/null || true; \
		echo "🗑️  Removing all containers..."; \
		docker rm $$(docker ps -aq) 2>/dev/null || true; \
		echo ""; \
		echo "✅ All containers stopped and removed"; \
	fi
	@echo ""
	@echo "🧹 Cleaning up dangling resources..."
	@docker network prune -f 2>/dev/null || true
	@docker volume prune -f 2>/dev/null || true
	@echo "✅ Cleanup complete!"
	@echo ""
	@echo "📊 Final status:"
	@echo "  Containers: $$(docker ps -a | wc -l | tr -d ' ') total (should be 1 for header only)"
	@echo "  Networks:   $$(docker network ls | wc -l | tr -d ' ')"
	@echo "  Volumes:    $$(docker volume ls | wc -l | tr -d ' ')"

install-deps: ## Install system dependencies
	@$(ECHO) "$(BLUE)Installing system dependencies...$(NC)"
	@if command -v apt-get > /dev/null; then \
		sudo apt-get update && sudo apt-get install -y curl jq watch; \
	elif command -v yum > /dev/null; then \
		sudo yum install -y curl jq procps-ng; \
	elif command -v dnf > /dev/null; then \
		sudo dnf install -y curl jq procps-ng; \
	else \
		$(ECHO) "$(YELLOW)Please install manually: curl, jq, watch$(NC)"; \
	fi

install-bpf: ## Install BPF/eBPF tools (for chaos experiments)
	@$(ECHO) "$(BLUE)Installing BPF/eBPF tools...$(NC)"
	@if command -v apt-get > /dev/null; then \
		$(ECHO) "$(YELLOW)Ubuntu/Debian:$(NC)"; \
		sudo apt-get install -y bpfcc-tools linux-headers-$$(uname -r) bpftrace; \
	elif command -v dnf > /dev/null; then \
		$(ECHO) "$(YELLOW)RHEL/Fedora:$(NC)"; \
		sudo dnf install -y bpftrace kernel-devel; \
	else \
		$(ECHO) "$(RED)Unsupported package manager$(NC)"; \
		echo "Please install bpftrace manually"; \
		exit 1; \
	fi
	@$(ECHO) "$(GREEN)✓ BPF tools installed$(NC)"

check-prereqs: ## Check all prerequisites
	@$(ECHO) "$(BLUE)Checking prerequisites...$(NC)"
	@$(ECHO) ""
	@printf "%-20s" "Go:"; \
	if command -v go > /dev/null; then \
		printf '%b\n' "$(GREEN)✓$(NC) $$(go version)"; \
	else \
		printf '%b\n' "$(RED)✗ Not found$(NC)"; \
	fi
	@printf "%-20s" "curl:"; \
	if command -v curl > /dev/null; then \
		printf '%b\n' "$(GREEN)✓$(NC) $$(curl --version | head -1)"; \
	else \
		printf '%b\n' "$(RED)✗ Not found$(NC)"; \
	fi
	@printf "%-20s" "jq:"; \
	if command -v jq > /dev/null; then \
		printf '%b\n' "$(GREEN)✓$(NC) $$(jq --version)"; \
	else \
		printf '%b\n' "$(RED)✗ Not found$(NC)"; \
	fi
	@printf "%-20s" "docker:"; \
	if command -v docker > /dev/null; then \
		printf '%b\n' "$(GREEN)✓$(NC) $$(docker --version)"; \
	else \
		printf '%b\n' "$(YELLOW)⚠ Not found (optional)$(NC)"; \
	fi
	@printf "%-20s" "bpftrace:"; \
	if command -v bpftrace > /dev/null; then \
		printf '%b\n' "$(GREEN)✓$(NC) $$(bpftrace --version 2>&1 | head -1)"; \
	else \
		printf '%b\n' "$(YELLOW)⚠ Not found (for chaos experiments)$(NC)"; \
	fi
	@$(ECHO) ""
	@$(ECHO) "$(BLUE)Run 'make install-deps' to install missing dependencies$(NC)"
	@$(ECHO) "$(BLUE)Run 'make install-bpf' to install BPF tools for chaos experiments$(NC)"

check-go:
	@if ! command -v go > /dev/null; then \
		$(ECHO) "$(RED)✗ Go not found in PATH$(NC)"; \
		$(ECHO) "$(YELLOW)Add Go to PATH: export PATH=\$$HOME/.local/go/bin:\$$PATH$(NC)"; \
		exit 1; \
	fi

dev-setup: deps ## Set up development environment
	@echo "⚙️ Setting up development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	@echo "✅ Development environment ready!"

version: ## Show version information
	@echo "Go Version: $(GO_VERSION)"
	@echo "Project: $(PROJECT_NAME)"
	@git describe --tags --always --dirty 2>/dev/null || echo "No git tags found"

# ============================================================
# End of Unified Makefile
# ============================================================
