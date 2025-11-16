# DNS Configuration for andythomas-sre.com

This document provides the DNS records needed to configure your SRE lab domain.

## Required DNS Records

Configure these DNS records in your domain registrar's DNS management panel:

### A Records (IPv4)

Point these to your server's public IPv4 address (replace `YOUR_SERVER_IP`):

```
Type    Name                        Value               TTL
────────────────────────────────────────────────────────────
A       @                           YOUR_SERVER_IP      3600
A       www                         YOUR_SERVER_IP      3600
A       api                         YOUR_SERVER_IP      3600
A       market                      YOUR_SERVER_IP      3600
A       orders                      YOUR_SERVER_IP      3600
A       grafana                     YOUR_SERVER_IP      3600
A       prometheus                  YOUR_SERVER_IP      3600
A       status                      YOUR_SERVER_IP      3600
A       docs                        YOUR_SERVER_IP      3600
```

### Alternative: Wildcard Record

Instead of individual A records, you can use a wildcard:

```
Type    Name                        Value               TTL
────────────────────────────────────────────────────────────
A       @                           YOUR_SERVER_IP      3600
A       *                           YOUR_SERVER_IP      3600
```

### AAAA Records (IPv6) - Optional

If your server has IPv6 (replace `YOUR_SERVER_IPv6`):

```
Type    Name                        Value               TTL
────────────────────────────────────────────────────────────
AAAA    @                           YOUR_SERVER_IPv6    3600
AAAA    *                           YOUR_SERVER_IPv6    3600
```

### CAA Records (Certificate Authority Authorization) - Recommended

Allow Let's Encrypt to issue certificates:

```
Type    Name                        Value                           TTL
──────────────────────────────────────────────────────────────────────
CAA     @                           0 issue "letsencrypt.org"       3600
CAA     @                           0 issuewild "letsencrypt.org"   3600
```

## Service URLs

Once DNS is configured and Caddy is running, your services will be available at:

| Service                  | URL                                  | Description                    |
|--------------------------|--------------------------------------|--------------------------------|
| Main Dashboard           | https://andythomas-sre.com           | Trading exchange dashboard     |
| Trading API              | https://api.andythomas-sre.com       | REST API endpoints             |
| Market Simulator         | https://market.andythomas-sre.com    | Market data simulator          |
| Order Flow Simulator     | https://orders.andythomas-sre.com    | Order flow generator           |
| Grafana Dashboards       | https://grafana.andythomas-sre.com   | Metrics visualization          |
| Prometheus               | https://prometheus.andythomas-sre.com| Metrics collection             |
| Status Page              | https://status.andythomas-sre.com    | Service health status          |
| Documentation            | https://docs.andythomas-sre.com      | API & system documentation     |

## DNS Propagation

After updating DNS records:

1. **Wait for propagation**: Can take 5 minutes to 48 hours (usually < 1 hour)
2. **Check DNS propagation**: Use https://dnschecker.org
3. **Test locally**: `dig andythomas-sre.com` or `nslookup andythomas-sre.com`

## Verification Commands

### Check DNS Resolution

```bash
# Check main domain
dig andythomas-sre.com +short

# Check subdomain
dig api.andythomas-sre.com +short

# Check all subdomains
for sub in www api market orders grafana prometheus status docs; do
  echo "$sub.andythomas-sre.com: $(dig +short $sub.andythomas-sre.com)"
done
```

### Check SSL Certificate

```bash
# After Caddy is running with DNS configured
curl -vI https://andythomas-sre.com 2>&1 | grep -i "SSL\|subject\|issuer"
```

## Example DNS Configuration (Cloudflare)

If using Cloudflare:

1. Log in to Cloudflare Dashboard
2. Select your domain: `andythomas-sre.com`
3. Go to **DNS** > **Records**
4. Click **Add record**
5. Add each A record:
   - Type: `A`
   - Name: `@` (or subdomain like `api`)
   - IPv4 address: Your server IP
   - Proxy status: **DNS only** (gray cloud) - Important for Let's Encrypt
   - TTL: Auto

### Cloudflare Settings

- **SSL/TLS Mode**: Full (not Full Strict, to avoid certificate issues during initial setup)
- **Always Use HTTPS**: Disable initially, enable after Caddy is working
- **Automatic HTTPS Rewrites**: Enable after initial setup

## Example DNS Configuration (Namecheap)

If using Namecheap:

1. Log in to Namecheap
2. Go to **Domain List** > Select `andythomas-sre.com`
3. Click **Manage** > **Advanced DNS**
4. Add records:

```
Type          Host                    Value               TTL
────────────────────────────────────────────────────────────────
A Record      @                       YOUR_SERVER_IP      Automatic
A Record      *                       YOUR_SERVER_IP      Automatic
CAA Record    @                       0 issue "letsencrypt.org"
```

## Troubleshooting

### DNS not resolving

```bash
# Check if DNS is configured
dig andythomas-sre.com

# Check from different DNS servers
dig @8.8.8.8 andythomas-sre.com
dig @1.1.1.1 andythomas-sre.com
```

### Certificate issues

```bash
# Check Caddy logs
docker logs simulated-exchange-caddy

# Check if port 80/443 are accessible from internet
# Use: https://www.yougetsignal.com/tools/open-ports/
```

### Common Issues

1. **Port 80/443 blocked by firewall**
   ```bash
   # Check firewall
   sudo ufw status

   # Allow ports if needed
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   ```

2. **DNS still pointing to old IP**
   - Clear local DNS cache: `sudo systemd-resolve --flush-caches`
   - Wait for TTL to expire
   - Use incognito/private browser

3. **Let's Encrypt rate limits**
   - Limit: 50 certificates per registered domain per week
   - Use staging environment for testing (see Caddyfile comments)

## Security Considerations

1. **DNS Security**: Enable DNSSEC if your registrar supports it
2. **CAA Records**: Restrict which CAs can issue certificates
3. **HTTPS Only**: After setup, enforce HTTPS everywhere
4. **Firewall**: Only expose ports 80, 443 to the internet
5. **DDoS Protection**: Consider using Cloudflare or similar CDN

## Next Steps

After DNS is configured:

1. Start Caddy: `make caddy-start`
2. Check certificates: `make caddy-certs`
3. Monitor logs: `make caddy-logs`
4. Verify all services: `make caddy-test`
