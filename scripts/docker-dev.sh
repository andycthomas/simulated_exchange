#!/bin/bash

# Docker development helper script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

case "${1:-help}" in
    "build")
        echo "Building Docker images..."
        docker build -t simulated-exchange:latest .
        docker build -f Dockerfile.dev -t simulated-exchange:dev .
        echo "✅ Images built successfully"
        ;;

    "dev")
        echo "Starting development environment..."
        docker-compose -f docker/docker-compose.dev.yml up -d
        echo "✅ Development environment started"
        echo "Services:"
        echo "  - Trading API: http://localhost:8080"
        echo "  - Database Admin: http://localhost:8081"
        echo "  - Email Testing: http://localhost:8025"
        echo "  - Tracing UI: http://localhost:16686"
        ;;

    "prod")
        echo "Starting production environment..."
        docker-compose up -d
        echo "✅ Production environment started"
        echo "Services:"
        echo "  - Trading API: http://localhost:8080"
        ;;

    "monitoring")
        echo "Starting with monitoring stack..."
        docker-compose --profile monitoring up -d
        echo "✅ Production environment with monitoring started"
        echo "Services:"
        echo "  - Trading API: http://localhost:8080"
        echo "  - Prometheus: http://localhost:9090"
        echo "  - Grafana: http://localhost:3000 (admin/admin123)"
        ;;

    "stop")
        echo "Stopping all environments..."
        docker-compose down 2>/dev/null || true
        docker-compose -f docker/docker-compose.dev.yml down 2>/dev/null || true
        echo "✅ All environments stopped"
        ;;

    "logs")
        service="${2:-trading-service}"
        if docker-compose ps | grep -q "$service"; then
            docker-compose logs -f "$service"
        elif docker-compose -f docker/docker-compose.dev.yml ps | grep -q "$service"; then
            docker-compose -f docker/docker-compose.dev.yml logs -f "$service"
        else
            echo "❌ Service '$service' not found"
            exit 1
        fi
        ;;

    "shell")
        service="${2:-trading-service}"
        if docker-compose ps | grep -q "$service" | grep -q "Up"; then
            docker-compose exec "$service" /bin/sh
        elif docker-compose -f docker/docker-compose.dev.yml ps | grep -q "$service" | grep -q "Up"; then
            docker-compose -f docker/docker-compose.dev.yml exec "$service" /bin/sh
        else
            echo "❌ Service '$service' is not running"
            exit 1
        fi
        ;;

    "test")
        echo "Running tests in container..."
        docker-compose -f docker/docker-compose.dev.yml exec trading-service-dev go test ./... -v
        ;;

    "clean")
        echo "Cleaning up Docker resources..."
        docker-compose down -v 2>/dev/null || true
        docker-compose -f docker/docker-compose.dev.yml down -v 2>/dev/null || true
        docker system prune -f
        echo "✅ Cleanup complete"
        ;;

    "status")
        echo "Docker environment status:"
        echo ""
        echo "Production services:"
        docker-compose ps 2>/dev/null || echo "  Not running"
        echo ""
        echo "Development services:"
        docker-compose -f docker/docker-compose.dev.yml ps 2>/dev/null || echo "  Not running"
        ;;

    "help"|*)
        echo "Docker Development Helper"
        echo ""
        echo "Usage: $0 <command>"
        echo ""
        echo "Commands:"
        echo "  build       Build Docker images"
        echo "  dev         Start development environment"
        echo "  prod        Start production environment"
        echo "  monitoring  Start production with monitoring"
        echo "  stop        Stop all environments"
        echo "  logs [svc]  Show logs for service (default: trading-service)"
        echo "  shell [svc] Open shell in service container"
        echo "  test        Run tests in development container"
        echo "  clean       Clean up all Docker resources"
        echo "  status      Show status of all environments"
        echo "  help        Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0 dev                    # Start development environment"
        echo "  $0 logs trading-service   # View trading service logs"
        echo "  $0 shell postgres-dev     # Open shell in postgres container"
        ;;
esac