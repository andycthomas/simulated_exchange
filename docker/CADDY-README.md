# Caddy Reverse Proxy Setup for SRE Lab

Complete guide to setting up Caddy as a reverse proxy with automatic HTTPS for your SRE lab at `andythomas-sre.com`.

## 🎯 Overview

This setup provides:

- ✅ **Automatic HTTPS** with Let's Encrypt certificates
- ✅ **Zero-downtime reloads** when configuration changes
- ✅ **Subdomain routing** for all services
- ✅ **Production-ready** SSL/TLS configuration
- ✅ **Comprehensive logging** for all services
- ✅ **Health monitoring** and status endpoints

## 📋 Prerequisites

### 1. Domain Ownership
- You must own `andythomas-sre.com`
- DNS access to configure A records

### 2. Server Requirements
- Public-facing server with ports 80 and 443 accessible
- Docker and Docker Compose installed
- Firewall configured to allow HTTP/HTTPS traffic

### 3. DNS Configuration
See [`DNS-SETUP.md`](./DNS-SETUP.md) for detailed instructions.

**Quick DNS setup:**
```dns
Type    Name        Value               TTL
A       @           YOUR_SERVER_IP      3600
A       *           YOUR_SERVER_IP      3600
```

## 🚀 Quick Start

### Step 1: Configure DNS

1. Point `andythomas-sre.com` and `*.andythomas-sre.com` to your server's IP
2. Wait for DNS propagation (5-60 minutes)
3. Verify: `dig andythomas-sre.com +short`

### Step 2: Start Caddy

```bash
make caddy-start
```

This will:
- Start Caddy container
- Automatically obtain Let's Encrypt SSL certificates
- Configure reverse proxy for all services
- Enable HTTPS for all subdomains

### Step 3: Verify Setup

```bash
# Check Caddy status
make caddy

# Test endpoints
make caddy-test

# View certificates
make caddy-certs
```

### Step 4: Access Services

Your services are now available at:

| Service | URL | Description |
|---------|-----|-------------|
| Dashboard | https://andythomas-sre.com | Main trading dashboard |
| Trading API | https://api.andythomas-sre.com | REST API endpoints |
| Market Simulator | https://market.andythomas-sre.com | Market data simulator |
| Order Flow | https://orders.andythomas-sre.com | Order flow generator |
| Grafana | https://grafana.andythomas-sre.com | Metrics dashboards |
| Prometheus | https://prometheus.andythomas-sre.com | Metrics collection |
| Status Page | https://status.andythomas-sre.com | Health status |
| Documentation | https://docs.andythomas-sre.com | API docs |

## 📖 Available Commands

### Container Management

```bash
# Start Caddy
make caddy-start

# Stop Caddy
make caddy-stop

# Restart Caddy
make caddy-restart

# View status
make caddy
```

### Configuration Management

```bash
# Reload configuration (zero-downtime)
make caddy-reload

# Edit Caddyfile
vim docker/Caddyfile

# After editing, reload:
make caddy-reload
```

### Monitoring & Debugging

```bash
# View logs
make caddy-logs

# Test all endpoints
make caddy-test

# Show SSL certificates
make caddy-certs

# Check admin API
curl http://localhost:2019/config/
```

## 🔧 Configuration Files

### Main Configuration: `docker/Caddyfile`

```caddy
andythomas-sre.com {
    reverse_proxy nginx:80

    log {
        output file /var/log/caddy/main.log
        format json
    }
}
```

### Docker Compose: `docker/docker-compose.caddy.yml`

Defines the Caddy service container with:
- Port mappings (80, 443, 2019, 8888)
- Volume mounts for config and data
- Health checks
- Network configuration

## 🌐 Service Routing

### Main Dashboard (`andythomas-sre.com`)
- **Routes to**: nginx:80
- **Purpose**: Trading exchange dashboard
- **Features**: Static files, API proxy

### API (`api.andythomas-sre.com`)
- **Routes to**: trading-api:8080
- **Purpose**: Direct API access
- **Features**: CORS enabled, JSON responses

### Grafana (`grafana.andythomas-sre.com`)
- **Routes to**: grafana:3000
- **Purpose**: Metrics visualization
- **Features**: WebSocket support for live updates

### Prometheus (`prometheus.andythomas-sre.com`)
- **Routes to**: prometheus:9090
- **Purpose**: Metrics collection
- **Features**: Optional basic auth (commented out)

### Market Simulator (`market.andythomas-sre.com`)
- **Routes to**: market-simulator:8081
- **Purpose**: Market data generation

### Order Flow (`orders.andythomas-sre.com`)
- **Routes to**: order-flow-simulator:8082
- **Purpose**: Order flow simulation

## 🔐 Security Features

### Automatic HTTPS
- Let's Encrypt certificates auto-renewed
- TLS 1.2+ enforced
- Modern cipher suites
- OCSP stapling enabled

### Security Headers
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

### Access Control
- Admin API on localhost only (port 2019)
- Optional basic auth for sensitive services
- CORS configured per service

## 📊 Logging

### Log Files Location

```bash
# View all logs
ls -lah logs/caddy/

# Individual service logs
logs/caddy/main.log          # Main dashboard
logs/caddy/api.log           # Trading API
logs/caddy/grafana.log       # Grafana
logs/caddy/prometheus.log    # Prometheus
logs/caddy/market.log        # Market simulator
logs/caddy/orders.log        # Order flow
```

### Log Format

All logs use JSON format for easy parsing:

```json
{
  "level": "info",
  "ts": 1699564800,
  "logger": "http.log.access",
  "msg": "handled request",
  "request": {
    "remote_ip": "1.2.3.4",
    "proto": "HTTP/2.0",
    "method": "GET",
    "host": "api.andythomas-sre.com",
    "uri": "/health"
  },
  "status": 200,
  "duration": 0.001234
}
```

## 🔄 Certificate Management

### Automatic Renewal

Caddy automatically:
- Obtains certificates from Let's Encrypt
- Renews certificates before expiration
- Handles ACME challenges (HTTP-01)

### Manual Operations

```bash
# List certificates
make caddy-certs

# Force certificate renewal (rarely needed)
docker exec simulated-exchange-caddy caddy reload --force

# Certificate storage location
docker volume inspect simulated_exchange_caddy_data
```

## 🐛 Troubleshooting

### Issue: Certificates not obtained

**Symptoms**: HTTPS not working, certificate errors

**Solutions**:
1. Verify DNS is configured and propagated
   ```bash
   dig andythomas-sre.com +short
   ```

2. Check ports 80/443 are accessible from internet
   ```bash
   # From external machine
   curl -I http://YOUR_SERVER_IP
   ```

3. Check Caddy logs
   ```bash
   make caddy-logs
   ```

4. Verify firewall allows traffic
   ```bash
   sudo ufw status
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   ```

### Issue: Service not accessible

**Symptoms**: 502 Bad Gateway or connection refused

**Solutions**:
1. Verify backend service is running
   ```bash
   docker compose ps
   ```

2. Check service health
   ```bash
   curl http://localhost:8080/health  # Trading API
   curl http://localhost:3000/api/health  # Grafana
   ```

3. Verify docker network
   ```bash
   docker network inspect simulated_exchange_network
   ```

### Issue: Configuration changes not applied

**Solutions**:
1. Reload configuration
   ```bash
   make caddy-reload
   ```

2. If reload fails, restart
   ```bash
   make caddy-restart
   ```

3. Check Caddyfile syntax
   ```bash
   docker exec simulated-exchange-caddy caddy validate --config /etc/caddy/Caddyfile
   ```

### Issue: Let's Encrypt rate limits

**Symptoms**: "too many certificates already issued" error

**Solutions**:
1. Use staging environment for testing
   ```caddy
   {
       acme_ca https://acme-staging-v02.api.letsencrypt.org/directory
   }
   ```

2. Wait for rate limit reset (weekly)
3. Reuse existing certificates (check data volume)

## 📈 Performance Tips

### Enable HTTP/2
Already enabled by default in Caddy

### Enable Compression
Add to Caddyfile:
```caddy
andythomas-sre.com {
    encode gzip zstd
    reverse_proxy nginx:80
}
```

### Enable Caching
Add to Caddyfile:
```caddy
andythomas-sre.com {
    @static {
        path *.css *.js *.png *.jpg *.gif *.svg
    }
    header @static Cache-Control "public, max-age=31536000"

    reverse_proxy nginx:80
}
```

## 🔒 Additional Security

### Add Basic Auth (Prometheus example)

1. Generate password hash:
```bash
docker exec simulated-exchange-caddy caddy hash-password --plaintext mypassword
```

2. Update Caddyfile:
```caddy
prometheus.andythomas-sre.com {
    basicauth {
        admin $2a$14$HASH_HERE
    }
    reverse_proxy prometheus:9090
}
```

3. Reload:
```bash
make caddy-reload
```

### IP Whitelisting

```caddy
api.andythomas-sre.com {
    @allowed {
        remote_ip 1.2.3.4 5.6.7.8/24
    }
    handle @allowed {
        reverse_proxy trading-api:8080
    }
    handle {
        respond "Access denied" 403
    }
}
```

## 📚 Additional Resources

- [Caddy Documentation](https://caddyserver.com/docs/)
- [Caddyfile Syntax](https://caddyserver.com/docs/caddyfile)
- [Let's Encrypt Limits](https://letsencrypt.org/docs/rate-limits/)
- [DNS Setup Guide](./DNS-SETUP.md)

## 🆘 Support

If you encounter issues:

1. Check logs: `make caddy-logs`
2. Verify DNS: `dig andythomas-sre.com`
3. Test backend: `docker compose ps`
4. Check firewall: `sudo ufw status`
5. Review Caddy docs: https://caddyserver.com/docs/

## 📝 Next Steps

After Caddy is running:

1. ✅ Configure DNS (see DNS-SETUP.md)
2. ✅ Start Caddy (`make caddy-start`)
3. ✅ Verify certificates (`make caddy-certs`)
4. ✅ Test all services (`make caddy-test`)
5. ✅ Set up monitoring dashboards in Grafana
6. ✅ Configure alerts for certificate expiration
7. ✅ Set up log aggregation (optional)
8. ✅ Add custom dashboards to docs.andythomas-sre.com
