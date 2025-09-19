# Docker Configuration for Simulated Exchange

This directory contains Docker configuration files for the Simulated Exchange trading platform.

## Quick Start

### Production Deployment

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f trading-service

# Stop services
docker-compose down
```

### Development Environment

```bash
# Start development environment with hot reload
docker-compose -f docker/docker-compose.dev.yml up -d

# View development logs
docker-compose -f docker/docker-compose.dev.yml logs -f trading-service-dev

# Stop development environment
docker-compose -f docker/docker-compose.dev.yml down
```

## Architecture

### Production Services

- **trading-service**: Main Go application
- **redis**: Caching and session storage
- **postgres**: Primary database
- **prometheus**: Metrics collection (optional)
- **grafana**: Metrics visualization (optional)
- **nginx**: Reverse proxy and load balancer (optional)

### Development Services

- **trading-service-dev**: Go application with hot reload
- **redis-dev**: Redis for development
- **postgres-dev**: PostgreSQL for development
- **adminer**: Database management UI
- **mailhog**: Email testing
- **jaeger**: Distributed tracing

## Configuration

### Environment Variables

#### Trading Service
- `GIN_MODE`: Application mode (release/debug)
- `PORT`: Application port (default: 8080)
- `LOG_LEVEL`: Logging level (debug/info/warn/error)
- `METRICS_ENABLED`: Enable metrics collection
- `AI_ANALYSIS_ENABLED`: Enable AI performance analysis

#### Database
- `POSTGRES_DB`: Database name
- `POSTGRES_USER`: Database user
- `POSTGRES_PASSWORD`: Database password

### Profiles

Use Docker Compose profiles to control which services start:

```bash
# Start with monitoring stack
docker-compose --profile monitoring up -d

# Start with production proxy
docker-compose --profile production up -d

# Start core services only (default)
docker-compose up -d
```

## File Structure

```
docker/
├── README.md                 # This file
├── docker-compose.dev.yml    # Development configuration
├── prometheus.yml            # Prometheus configuration
├── nginx.conf               # NGINX configuration
├── init-db.sql              # Production database schema
└── init-db-dev.sql          # Development database schema

# Root directory files
├── Dockerfile               # Production multi-stage build
├── Dockerfile.dev          # Development with hot reload
├── docker-compose.yml      # Production services
├── .dockerignore           # Files to exclude from build
└── .air.toml              # Hot reload configuration
```

## Development Workflow

### Hot Reload Development

1. Start the development environment:
   ```bash
   docker-compose -f docker/docker-compose.dev.yml up -d
   ```

2. The application will automatically reload when you make changes to Go files.

3. Access services:
   - Trading API: http://localhost:8080
   - Database Admin: http://localhost:8081
   - Email Testing: http://localhost:8025
   - Tracing UI: http://localhost:16686

### Debugging

The development container includes Delve debugger on port 2345:

```bash
# Connect with your IDE or dlv command
dlv connect localhost:2345
```

### Database Management

Access the database via Adminer:
1. Open http://localhost:8081
2. Server: `postgres-dev`
3. Username: `dev_user`
4. Password: `dev_pass`
5. Database: `trading_db_dev`

## Production Deployment

### Building Images

```bash
# Build production image
docker build -t simulated-exchange:latest .

# Build development image
docker build -f Dockerfile.dev -t simulated-exchange:dev .
```

### Health Checks

All services include health checks:

```bash
# Check service health
docker-compose ps

# View health check logs
docker inspect <container_name> | jq '.[0].State.Health'
```

### Scaling

Scale services horizontally:

```bash
# Scale trading service to 3 replicas
docker-compose up -d --scale trading-service=3
```

### Monitoring

#### Prometheus Metrics

Access Prometheus at http://localhost:9090 when using the monitoring profile.

Available metrics endpoints:
- Trading service: http://localhost:8080/metrics
- Redis: Requires redis_exporter (not included)
- PostgreSQL: Requires postgres_exporter (not included)

#### Grafana Dashboards

Access Grafana at http://localhost:3000:
- Username: admin
- Password: admin123

## Security

### Production Security Considerations

1. **Non-root containers**: All services run as non-root users
2. **Network isolation**: Services communicate via dedicated Docker network
3. **Secret management**: Use Docker secrets for sensitive data
4. **SSL/TLS**: Configure NGINX with proper certificates
5. **Firewall**: Restrict external access to necessary ports only

### Development Security

Development containers include additional debugging tools and may have relaxed security for easier development.

## Volumes and Data Persistence

### Production Volumes
- `trading_data`: Application data
- `trading_logs`: Application logs
- `postgres_data`: Database data
- `redis_data`: Redis data

### Development Volumes
- Source code is mounted for hot reload
- Separate development database volume

### Backup and Recovery

```bash
# Backup database
docker-compose exec postgres pg_dump -U trading_user trading_db > backup.sql

# Restore database
docker-compose exec -T postgres psql -U trading_user trading_db < backup.sql
```

## Troubleshooting

### Common Issues

1. **Port conflicts**: Change port mappings in docker-compose.yml
2. **Permission denied**: Ensure proper file permissions
3. **Build failures**: Clear Docker cache: `docker system prune -a`
4. **Database connection**: Check network connectivity and credentials

### Logs

```bash
# View all logs
docker-compose logs

# Follow specific service logs
docker-compose logs -f trading-service

# View last 100 lines
docker-compose logs --tail=100 trading-service
```

### Performance Monitoring

Monitor container resource usage:

```bash
# Real-time stats
docker stats

# Container processes
docker-compose top
```

## Custom Configuration

### Adding New Services

1. Add service definition to `docker-compose.yml`
2. Create necessary configuration files in `docker/`
3. Update this documentation
4. Test with development environment first

### Environment-Specific Overrides

Create environment-specific compose files:

```bash
# docker-compose.staging.yml
# docker-compose.production.yml
```

Use with:
```bash
docker-compose -f docker-compose.yml -f docker-compose.staging.yml up -d
```

## Support

For issues and questions:
1. Check container logs first
2. Verify network connectivity
3. Review this documentation
4. Check Docker and Docker Compose versions