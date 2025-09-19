# Makefile for Simulated Exchange Testing
.PHONY: help test test-unit test-e2e test-demo test-perf test-docker test-all clean coverage benchmark deps lint fmt vet security docker-build docker-test

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Variables
GO_VERSION := $(shell go version | cut -d' ' -f3)
PROJECT_NAME := simulated_exchange
COVERAGE_FILE := coverage.out
TEST_RESULTS_DIR := test-output
DOCKER_COMPOSE_FILE := docker-compose.test.yml

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

test-docker: ## Run tests in Docker environment
	@echo "🐳 Running tests in Docker..."
	./scripts/run-tests.sh

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

# Docker targets
docker-build: ## Build Docker test image
	@echo "🐳 Building Docker test image..."
	docker build -f Dockerfile.test -t $(PROJECT_NAME):test .

docker-test-unit: ## Run unit tests in Docker
	@echo "🐳 Running unit tests in Docker..."
	docker run --rm -v $(PWD)/$(TEST_RESULTS_DIR):/app/$(TEST_RESULTS_DIR) $(PROJECT_NAME):test \
		sh -c "go test -v -race -coverprofile=/app/$(TEST_RESULTS_DIR)/$(COVERAGE_FILE) ./internal/..."

docker-test-e2e: ## Run E2E tests with Docker Compose
	@echo "🐳 Running E2E tests with Docker Compose..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) up --build --abort-on-container-exit
	docker-compose -f $(DOCKER_COMPOSE_FILE) down --volumes

docker-test-all: docker-build ## Run comprehensive Docker test suite
	@echo "🐳 Running comprehensive Docker test suite..."
	./scripts/run-tests.sh

# Load testing targets
load-test-light: ## Run light load test
	@echo "🚀 Running light load test..."
	mkdir -p $(TEST_RESULTS_DIR)
	go test -v ./test/e2e/ -run=TestPerformanceTestSuite/TestOrderPlacementPerformance -timeout=5m | tee $(TEST_RESULTS_DIR)/load-light.log

load-test-medium: ## Run medium load test
	@echo "🚀 Running medium load test..."
	mkdir -p $(TEST_RESULTS_DIR)
	go test -v ./test/e2e/ -run=TestPerformanceTestSuite/TestScalabilityCharacteristics -timeout=8m | tee $(TEST_RESULTS_DIR)/load-medium.log

load-test-heavy: ## Run heavy load test (requires Docker)
	@echo "🚀 Running heavy load test..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) up load-tester --abort-on-container-exit
	docker-compose -f $(DOCKER_COMPOSE_FILE) down --volumes

# Demo testing targets
demo-test-scenarios: ## Test demo scenarios
	@echo "🎭 Testing demo scenarios..."
	mkdir -p $(TEST_RESULTS_DIR)
	go test -v ./test/e2e/ -run=TestDemoFlowTestSuite/TestLoadTestScenarios -timeout=5m | tee $(TEST_RESULTS_DIR)/demo-scenarios.log

demo-test-chaos: ## Test chaos scenarios
	@echo "🌪️ Testing chaos scenarios..."
	mkdir -p $(TEST_RESULTS_DIR)
	go test -v ./test/e2e/ -run=TestDemoFlowTestSuite/TestChaosTestScenarios -timeout=5m | tee $(TEST_RESULTS_DIR)/demo-chaos.log

demo-test-full: ## Run full demo system test
	@echo "🎭 Running full demo system test..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) up demo-tester --abort-on-container-exit
	docker-compose -f $(DOCKER_COMPOSE_FILE) down --volumes

# Cleanup targets
clean: ## Clean test results and artifacts
	@echo "🧹 Cleaning test results and artifacts..."
	rm -rf $(TEST_RESULTS_DIR)
	rm -f coverage.out coverage.html coverage.txt
	docker-compose -f $(DOCKER_COMPOSE_FILE) down --volumes --remove-orphans 2>/dev/null || true
	docker system prune -f --filter "label=project=$(PROJECT_NAME)" 2>/dev/null || true

clean-docker: ## Clean Docker images and containers
	@echo "🐳 Cleaning Docker images and containers..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) down --volumes --remove-orphans --rmi all 2>/dev/null || true
	docker image prune -f --filter "label=project=$(PROJECT_NAME)" 2>/dev/null || true

# Development targets
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

dev-test: ## Quick development test (unit tests only)
	@echo "🚀 Running quick development tests..."
	go test -v -race ./internal/...

dev-watch: ## Watch for changes and run tests
	@echo "👀 Watching for changes..."
	@if command -v fswatch >/dev/null 2>&1; then \
		fswatch -o . | xargs -n1 -I{} make dev-test; \
	else \
		echo "fswatch not installed. Install with: brew install fswatch (macOS) or apt-get install fswatch (Ubuntu)"; \
	fi

# CI/CD targets
ci-test: deps lint test-unit coverage-ci ## Run CI test suite
	@echo "🔄 CI test suite completed"

ci-e2e: test-e2e test-demo ## Run CI E2E tests
	@echo "🔄 CI E2E tests completed"

ci-full: ci-test ci-e2e benchmark ## Run full CI suite
	@echo "🔄 Full CI suite completed"

# Reporting targets
report: ## Generate comprehensive test report
	@echo "📋 Generating test report..."
	mkdir -p $(TEST_RESULTS_DIR)
	@echo "# Test Report" > $(TEST_RESULTS_DIR)/report.md
	@echo "" >> $(TEST_RESULTS_DIR)/report.md
	@echo "Generated on: $$(date)" >> $(TEST_RESULTS_DIR)/report.md
	@echo "" >> $(TEST_RESULTS_DIR)/report.md
	@echo "## Go Version" >> $(TEST_RESULTS_DIR)/report.md
	@echo "\`\`\`" >> $(TEST_RESULTS_DIR)/report.md
	@echo "$(GO_VERSION)" >> $(TEST_RESULTS_DIR)/report.md
	@echo "\`\`\`" >> $(TEST_RESULTS_DIR)/report.md
	@echo "" >> $(TEST_RESULTS_DIR)/report.md
	@if [ -f "$(TEST_RESULTS_DIR)/coverage.txt" ]; then \
		echo "## Coverage Summary" >> $(TEST_RESULTS_DIR)/report.md; \
		echo "\`\`\`" >> $(TEST_RESULTS_DIR)/report.md; \
		cat $(TEST_RESULTS_DIR)/coverage.txt >> $(TEST_RESULTS_DIR)/report.md; \
		echo "\`\`\`" >> $(TEST_RESULTS_DIR)/report.md; \
	fi
	@echo "📋 Test report generated: $(TEST_RESULTS_DIR)/report.md"

# Status targets
status: ## Show project status
	@echo "📊 Project Status"
	@echo "================"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Project: $(PROJECT_NAME)"
	@echo ""
	@echo "Available test targets:"
	@echo "  - make test-unit     (Unit tests)"
	@echo "  - make test-demo     (Demo system tests)"
	@echo "  - make test-e2e      (End-to-end tests)"
	@echo "  - make test-perf     (Performance tests)"
	@echo "  - make test-docker   (Docker-based tests)"
	@echo "  - make test-all      (All tests except Docker)"
	@echo ""
	@echo "Quick commands:"
	@echo "  - make dev-test      (Quick development test)"
	@echo "  - make coverage      (Generate coverage report)"
	@echo "  - make benchmark     (Run benchmarks)"
	@echo ""
	@if [ -d "$(TEST_RESULTS_DIR)" ]; then \
		echo "Test results available in: $(TEST_RESULTS_DIR)"; \
		ls -la $(TEST_RESULTS_DIR)/; \
	else \
		echo "No test results found. Run 'make test-unit' to get started."; \
	fi

# Version target
version: ## Show version information
	@echo "Go Version: $(GO_VERSION)"
	@echo "Project: $(PROJECT_NAME)"
	@git describe --tags --always --dirty 2>/dev/null || echo "No git tags found"

# Default target when no target is specified
.DEFAULT_GOAL := help