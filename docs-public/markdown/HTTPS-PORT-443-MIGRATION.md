# HTTPS Port 443 Migration Guide

## Overview

This document describes the migration from SSH tunnel-based port forwarding to direct HTTPS on standard port 443.

## What Changed

### Before (Old Setup)
- SSH used port 443
- HTTPS traffic went through port 8443
- Required SSH tunnel for external access
- Complex port forwarding setup

### After (New Setup)
- HTTPS uses standard port 443 directly
- HTTP uses port 80 (auto-redirects to HTTPS)
- No SSH tunnel needed
- Simpler, more standard configuration

## Configuration Changes

### 1. Caddy Configuration (`docker/Caddyfile`)

**Removed:**
- `auto_https disable_redirects`
- `https_port 443` (now defaults to 443)
- All `http://` prefixes on domain configurations

**Added:**
- Automatic HTTPS with Let's Encrypt
- HSTS (HTTP Strict Transport Security) headers
- All domains now use HTTPS by default

### 2. Docker Compose (`docker/docker-compose.caddy.yml`)

**Changed:**
```yaml
# OLD
ports:
  - "80:80"
  - "8443:443"  # Non-standard port

# NEW
ports:
  - "80:80"     # HTTP - redirects to HTTPS
  - "443:443"   # HTTPS - standard secure port
```

### 3. Removed Files
- `scripts/setup-ssh-tunnel-service.sh` - No longer needed
- `docs/SSH_TUNNEL_SETUP.md` - No longer needed

## Migration Steps

### 1. Stop Existing Services

```bash
# Stop Caddy if running
docker-compose -f docker/docker-compose.caddy.yml down

# If you have SSH tunnel service (should not exist)
sudo systemctl stop simulated-exchange-ssh-tunnel 2>/dev/null || true
sudo systemctl disable simulated-exchange-ssh-tunnel 2>/dev/null || true
```

### 2. Ensure Firewall Allows Port 443

```bash
# Check firewall status
sudo ufw status

# Allow HTTPS on port 443
sudo ufw allow 443/tcp

# Allow HTTP on port 80 (for Let's Encrypt validation)
sudo ufw allow 80/tcp

# Verify
sudo ufw status numbered
```

### 3. Ensure Port 443 is Available

```bash
# Check if anything is using port 443
sudo lsof -i :443
sudo ss -tulnp | grep :443

# If SSH is using port 443, you'll need to move it to another port
# Edit /etc/ssh/sshd_config and change Port to something else (e.g., 22 or 2222)
```

### 4. Update DNS (if needed)

Ensure your DNS records point to the correct IP:

```bash
# Check current public IP
curl ifconfig.me

# Verify DNS resolution
dig andythomas-sre.com +short
dig api.andythomas-sre.com +short
dig grafana.andythomas-sre.com +short
```

If your IP changed, update DNS A records at your registrar.

### 5. Start Caddy with New Configuration

```bash
# Navigate to project directory
cd /home/andy/simulated_exchange

# Start Caddy with updated configuration
docker-compose -f docker/docker-compose.caddy.yml up -d

# Check logs
docker logs simulated-exchange-caddy -f
```

### 6. Verify HTTPS Access

```bash
# Test from command line
curl -I https://andythomas-sre.com
curl -I https://api.andythomas-sre.com
curl -I https://grafana.andythomas-sre.com

# Check certificate
curl -vI https://andythomas-sre.com 2>&1 | grep "subject:\|issuer:"
```

### 7. Test from Browser

Open browser and visit:
- https://andythomas-sre.com
- https://grafana.andythomas-sre.com
- https://prometheus.andythomas-sre.com

Verify:
- ✅ Green padlock in browser
- ✅ Valid Let's Encrypt certificate
- ✅ No certificate warnings
- ✅ Services load correctly

## Troubleshooting

### Certificate Not Loading

**Problem:** Browser shows SSL error or certificate warnings

**Solutions:**
1. Wait 1-2 minutes for Let's Encrypt to issue certificate
2. Check Caddy logs: `docker logs simulated-exchange-caddy`
3. Ensure ports 80 and 443 are accessible from internet
4. Verify DNS points to correct IP

### Port 443 Already in Use

**Problem:** Can't bind to port 443

**Solutions:**
```bash
# Find what's using the port
sudo lsof -i :443
sudo ss -tulnp | grep :443

# If it's SSH, move SSH to different port
sudo vi /etc/ssh/sshd_config
# Change: Port 22  (or 2222)
sudo systemctl restart sshd

# Then start Caddy again
```

### DNS Not Resolving

**Problem:** Domains don't resolve to your server

**Solutions:**
1. Check DNS propagation: https://dnschecker.org
2. Wait 30-60 minutes for DNS to propagate globally
3. Verify A records are correct at registrar
4. Use `dig` to check: `dig andythomas-sre.com +short`

### Let's Encrypt Rate Limits

**Problem:** Hit Let's Encrypt rate limits during testing

**Solutions:**
1. Wait 1 week for rate limits to reset
2. Use staging environment for testing (not recommended for production)
3. See: https://letsencrypt.org/docs/rate-limits/

## Benefits of Direct Port 443

### Simplicity
- Standard HTTPS port (443) - no need to specify port numbers
- No SSH tunnel management
- Fewer moving parts to maintain

### Security
- HSTS headers enforce HTTPS
- Let's Encrypt automatic certificate renewal
- Standard TLS best practices

### Performance
- Direct connection (no SSH overhead)
- Better for production deployments
- Easier to debug connection issues

### Compatibility
- Works with all clients without configuration
- No special firewall rules for non-standard ports
- Better mobile/app compatibility

## Rollback (If Needed)

If you need to rollback to the old configuration:

1. Restore old Caddyfile from git:
   ```bash
   git show HEAD~1:docker/Caddyfile > docker/Caddyfile
   ```

2. Restore old docker-compose.caddy.yml:
   ```bash
   git show HEAD~1:docker/docker-compose.caddy.yml > docker/docker-compose.caddy.yml
   ```

3. Restart Caddy:
   ```bash
   docker-compose -f docker/docker-compose.caddy.yml restart
   ```

## Related Documentation

- [QUICK-ACCESS-GUIDE.md](QUICK-ACCESS-GUIDE.md) - Quick setup instructions
- [EXTERNAL-ACCESS.md](EXTERNAL-ACCESS.md) - Detailed external access guide
- [CADDY-README.md](CADDY-README.md) - Caddy configuration reference
- [DNS-SETUP.md](DNS-SETUP.md) - DNS configuration guide

## Support

If you encounter issues:

1. Check Caddy logs: `docker logs simulated-exchange-caddy`
2. Verify DNS: `dig andythomas-sre.com +short`
3. Check firewall: `sudo ufw status`
4. Test connectivity: `curl -vI https://andythomas-sre.com`
