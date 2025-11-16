# External Access Guide for andythomas-sre.com

Complete guide to accessing your SRE lab from outside your local network.

## 🌍 Overview

To access your SRE lab domains from anywhere on the internet, you need:

1. ✅ **Public IP Address** - Your server must be reachable from the internet
2. ✅ **DNS Configuration** - Domain pointing to your public IP
3. ✅ **Firewall Rules** - Allow HTTP/HTTPS traffic
4. ✅ **Port Forwarding** - If behind NAT/router (home network)
5. ✅ **Caddy Running** - Reverse proxy handling HTTPS

## 📋 Step-by-Step Setup

### Step 1: Find Your Public IP Address

```bash
# From your server
curl ifconfig.me

# Or
curl icanhazip.com

# Or
dig +short myip.opendns.com @resolver1.opendns.com
```

**Note your public IP address** - Example: `203.0.113.45`

### Step 2: Check Your Network Type

#### Scenario A: VPS/Cloud Server (AWS, DigitalOcean, Linode, etc.)

✅ **Already has public IP** - Skip to Step 3

Common providers:
- AWS EC2
- DigitalOcean Droplets
- Linode
- Vultr
- Google Cloud Compute
- Azure VMs

#### Scenario B: Home/Office Network (Behind Router)

⚠️ **Need port forwarding** - See "Port Forwarding Setup" below

#### Scenario C: Behind Corporate/University Firewall

⚠️ **May need VPN or tunneling** - See "Tunneling Solutions" below

### Step 3: Configure DNS Records

Log into your domain registrar for `andythomas-sre.com` and add:

```dns
Type    Name        Value                   TTL
────────────────────────────────────────────────
A       @           YOUR_PUBLIC_IP          3600
A       *           YOUR_PUBLIC_IP          3600
```

**Example with IP 203.0.113.45:**

```dns
Type    Name        Value           TTL
A       @           203.0.113.45    3600
A       *           203.0.113.45    3600
```

#### Popular Registrars

**Cloudflare:**
1. Dashboard → Select Domain → DNS
2. Add Record → Type: A, Name: @, IPv4: YOUR_IP
3. Add Record → Type: A, Name: *, IPv4: YOUR_IP
4. **Important:** Set Proxy Status to "DNS only" (gray cloud)

**Namecheap:**
1. Domain List → Manage → Advanced DNS
2. Add New Record → A Record, Host: @, Value: YOUR_IP
3. Add New Record → A Record, Host: *, Value: YOUR_IP

**GoDaddy:**
1. My Products → DNS → Manage DNS
2. Add → A, Name: @, Value: YOUR_IP
3. Add → A, Name: *, Value: YOUR_IP

**Google Domains:**
1. My Domains → Manage → DNS
2. Custom Resource Records → A, @, YOUR_IP
3. Custom Resource Records → A, *, YOUR_IP

### Step 4: Configure Firewall (Linux Server)

#### Check Current Firewall Status

```bash
# For UFW (Ubuntu/Debian)
sudo ufw status

# For firewalld (CentOS/RHEL)
sudo firewall-cmd --list-all

# For iptables
sudo iptables -L -n
```

#### Allow HTTP and HTTPS Traffic

**Using UFW (Ubuntu/Debian):**

```bash
# Allow HTTP (port 80)
sudo ufw allow 80/tcp comment 'HTTP for Let's Encrypt'

# Allow HTTPS (port 443)
sudo ufw allow 443/tcp comment 'HTTPS for web traffic'

# Check status
sudo ufw status numbered

# If UFW is inactive, enable it
sudo ufw enable
```

**Using firewalld (CentOS/RHEL):**

```bash
# Allow HTTP
sudo firewall-cmd --permanent --add-service=http

# Allow HTTPS
sudo firewall-cmd --permanent --add-service=https

# Reload firewall
sudo firewall-cmd --reload

# Verify
sudo firewall-cmd --list-services
```

**Using iptables:**

```bash
# Allow HTTP
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT

# Allow HTTPS
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# Save rules (Ubuntu/Debian)
sudo netfilter-persistent save

# Save rules (CentOS/RHEL)
sudo service iptables save
```

### Step 5: Port Forwarding (Home/Office Network Only)

If your server is behind a home router, configure port forwarding:

#### Access Router Admin Panel

Common router IPs:
- `192.168.1.1`
- `192.168.0.1`
- `10.0.0.1`
- Check router label or manual

#### Forward Ports 80 and 443

**Generic Steps:**

1. Log into router admin panel
2. Navigate to "Port Forwarding" or "NAT" section
3. Add new port forwarding rules:

```
Service Name: HTTP
External Port: 80
Internal IP: YOUR_SERVER_LOCAL_IP (e.g., 192.168.1.100)
Internal Port: 80
Protocol: TCP

Service Name: HTTPS
External Port: 443
Internal IP: YOUR_SERVER_LOCAL_IP
Internal Port: 443
Protocol: TCP
```

#### Find Your Server's Local IP

```bash
# On the server
ip addr show | grep inet

# Or
hostname -I

# Or
ifconfig
```

#### Popular Router Configurations

**Netgear:**
- Advanced → Advanced Setup → Port Forwarding

**TP-Link:**
- Advanced → NAT Forwarding → Virtual Servers

**Linksys:**
- Security → Apps and Gaming → Single Port Forwarding

**ASUS:**
- Advanced Settings → WAN → Port Forwarding

**D-Link:**
- Advanced → Port Forwarding

### Step 6: Verify DNS Propagation

```bash
# Check from multiple locations
dig @8.8.8.8 andythomas-sre.com +short
dig @1.1.1.1 andythomas-sre.com +short

# Check globally
# Visit: https://dnschecker.org
# Enter: andythomas-sre.com
```

**Wait Time:**
- Minimum: 5-10 minutes
- Typical: 30-60 minutes
- Maximum: 24-48 hours

### Step 7: Start Caddy

```bash
cd /home/andy/simulated_exchange
make caddy-start
```

Caddy will automatically:
- Detect DNS is configured
- Request Let's Encrypt certificate
- Enable HTTPS for all domains

### Step 8: Test External Access

```bash
# From your local machine (not the server)
curl -I https://andythomas-sre.com

# Expected response:
HTTP/2 200
server: Caddy
...
```

**Test all subdomains:**

```bash
curl -I https://api.andythomas-sre.com
curl -I https://grafana.andythomas-sre.com
curl -I https://prometheus.andythomas-sre.com
```

## 🔍 Troubleshooting

### Issue: Cannot access from external network

**Diagnosis Steps:**

1. **Test if ports are open from internet:**
   ```bash
   # Visit: https://www.yougetsignal.com/tools/open-ports/
   # Enter your public IP and test ports 80, 443
   ```

2. **Check server firewall:**
   ```bash
   sudo ufw status
   # Should show 80/tcp and 443/tcp as ALLOW
   ```

3. **Verify Caddy is running:**
   ```bash
   make caddy
   docker compose -f docker-compose.yml -f docker/docker-compose.caddy.yml ps
   ```

4. **Test from server locally:**
   ```bash
   curl -I http://localhost
   curl -I http://localhost:2019/config/
   ```

5. **Check DNS resolution:**
   ```bash
   dig andythomas-sre.com +short
   # Should return your public IP
   ```

### Issue: SSL/TLS certificate errors

**Common Causes:**

1. **DNS not propagated yet**
   - Wait 30-60 minutes after DNS changes
   - Check: https://dnschecker.org

2. **Ports 80/443 not accessible from internet**
   - Let's Encrypt needs port 80 for verification
   - Check firewall and port forwarding

3. **Caddy logs show errors:**
   ```bash
   make caddy-logs
   # Look for certificate acquisition errors
   ```

**Solutions:**

```bash
# Restart Caddy after DNS propagates
make caddy-restart

# Check certificate status
make caddy-certs

# View detailed logs
docker logs simulated-exchange-caddy --tail 100
```

### Issue: Works locally but not externally

**Likely Causes:**

1. **Router port forwarding not configured** (home network)
2. **ISP blocking ports 80/443**
3. **Public IP changed** (dynamic IP)

**Solutions:**

1. Double-check port forwarding rules
2. Contact ISP about port blocking
3. Use Dynamic DNS service (see below)

### Issue: Only HTTP works, HTTPS fails

**Check:**

```bash
# Verify port 443 is open
sudo ufw status | grep 443

# Verify Caddy is listening on 443
sudo netstat -tlnp | grep :443

# Check Caddy logs
make caddy-logs
```

## 🏠 Dynamic IP Solutions (Home Networks)

If your ISP provides a **dynamic IP** that changes periodically:

### Option 1: Dynamic DNS Service

**Free Services:**
- No-IP (https://www.noip.com)
- DuckDNS (https://www.duckdns.org)
- FreeDNS (https://freedns.afraid.org)

**Setup:**

1. Register for Dynamic DNS service
2. Get a hostname (e.g., `yourlab.noip.com`)
3. Install DDNS client on server:

```bash
# Install No-IP client
cd /usr/local/src
wget https://www.noip.com/client/linux/noip-duc-linux.tar.gz
tar xzf noip-duc-linux.tar.gz
cd noip-2.1.9-1
make && sudo make install

# Configure
sudo /usr/local/bin/noip2 -C
```

4. Create CNAME record:
   ```dns
   Type      Name    Value                   TTL
   CNAME     @       yourlab.noip.com        3600
   CNAME     *       yourlab.noip.com        3600
   ```

### Option 2: Cloudflare API Update

Use Cloudflare API to update IP automatically:

```bash
#!/bin/bash
# /usr/local/bin/update-cloudflare-ip.sh

ZONE_ID="your_zone_id"
API_TOKEN="your_api_token"
DOMAIN="andythomas-sre.com"

CURRENT_IP=$(curl -s ifconfig.me)
RECORD_ID=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records?name=$DOMAIN" \
  -H "Authorization: Bearer $API_TOKEN" | jq -r '.result[0].id')

curl -X PUT "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records/$RECORD_ID" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  --data "{\"type\":\"A\",\"name\":\"$DOMAIN\",\"content\":\"$CURRENT_IP\"}"
```

Add to crontab:
```bash
# Update IP every 5 minutes
*/5 * * * * /usr/local/bin/update-cloudflare-ip.sh
```

## 🔒 Security Recommendations for External Access

### 1. Enable Fail2Ban

Protect against brute force attacks:

```bash
# Install
sudo apt install fail2ban

# Configure
sudo cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local

# Edit settings
sudo nano /etc/fail2ban/jail.local

# Start service
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

### 2. Configure SSH Security

```bash
# Edit SSH config
sudo nano /etc/ssh/sshd_config

# Recommended settings:
Port 2222                           # Change from default 22
PermitRootLogin no
PasswordAuthentication no           # Use keys only
MaxAuthTries 3

# Restart SSH
sudo systemctl restart sshd
```

### 3. Enable Automatic Security Updates

```bash
# Ubuntu/Debian
sudo apt install unattended-upgrades
sudo dpkg-reconfigure -plow unattended-upgrades
```

### 4. Set Up Monitoring

```bash
# Monitor failed login attempts
sudo tail -f /var/log/auth.log

# Monitor Caddy access
make caddy-logs

# Monitor all traffic
sudo tcpdump -i any port 80 or port 443
```

### 5. Add Basic Auth to Sensitive Services

Edit `docker/Caddyfile`:

```caddy
prometheus.andythomas-sre.com {
    basicauth {
        admin $2a$14$YOUR_HASH_HERE
    }
    reverse_proxy prometheus:9090
}
```

Generate password hash:
```bash
docker exec simulated-exchange-caddy caddy hash-password --plaintext yourpassword
```

## 🚇 Tunneling Solutions (Alternative to Port Forwarding)

If you **cannot** configure port forwarding or firewall:

### Option 1: Cloudflare Tunnel (Recommended)

**Advantages:**
- No port forwarding needed
- No firewall changes
- DDoS protection
- Free tier available

**Setup:**

```bash
# Install cloudflared
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared-linux-amd64.deb

# Authenticate
cloudflared tunnel login

# Create tunnel
cloudflared tunnel create sre-lab

# Configure tunnel
cat > ~/.cloudflared/config.yml << EOF
tunnel: YOUR_TUNNEL_ID
credentials-file: /home/andy/.cloudflared/YOUR_TUNNEL_ID.json

ingress:
  - hostname: andythomas-sre.com
    service: http://localhost:80
  - hostname: api.andythomas-sre.com
    service: http://localhost:8080
  - hostname: grafana.andythomas-sre.com
    service: http://localhost:3000
  - service: http_status:404
EOF

# Run tunnel
cloudflared tunnel run sre-lab

# Or install as service
sudo cloudflared service install
sudo systemctl start cloudflared
```

### Option 2: Ngrok

**Quick setup for testing:**

```bash
# Install ngrok
wget https://bin.equinox.io/c/bNyj1mQVY4c/ngrok-v3-stable-linux-amd64.tgz
tar xvzf ngrok-v3-stable-linux-amd64.tgz
sudo mv ngrok /usr/local/bin/

# Start tunnel
ngrok http 80

# Use provided URL (e.g., https://abc123.ngrok.io)
```

### Option 3: Tailscale (Private Access)

**For secure team access:**

```bash
# Install Tailscale
curl -fsSL https://tailscale.com/install.sh | sh

# Start Tailscale
sudo tailscale up

# Share with team members
# They can access via Tailscale IP
```

## ✅ Verification Checklist

Before considering setup complete:

- [ ] DNS resolves to correct public IP (`dig andythomas-sre.com`)
- [ ] Ports 80/443 open from internet (test at yougetsignal.com)
- [ ] Firewall allows HTTP/HTTPS traffic
- [ ] Port forwarding configured (if home network)
- [ ] Caddy running (`make caddy`)
- [ ] SSL certificates obtained (`make caddy-certs`)
- [ ] Can access from external device (phone on cellular)
- [ ] All subdomains work (api, grafana, prometheus, etc.)
- [ ] HTTPS works without errors
- [ ] Monitoring in place (fail2ban, logs)

## 📱 Testing from External Network

```bash
# From your phone (on cellular, not WiFi)
# Visit in browser:
https://andythomas-sre.com
https://api.andythomas-sre.com/health
https://grafana.andythomas-sre.com

# Or use online tools:
# - https://www.isitdownrightnow.com
# - https://downforeveryoneorjustme.com
# - https://httpstatus.io
```

## 🎯 Quick Start Summary

1. **Find public IP:** `curl ifconfig.me`
2. **Configure DNS:** A record `@` and `*` → your IP
3. **Open firewall:** `sudo ufw allow 80,443/tcp`
4. **Port forward** (home only): Router → 80, 443 → server
5. **Wait for DNS:** 30-60 minutes
6. **Start Caddy:** `make caddy-start`
7. **Test access:** `curl -I https://andythomas-sre.com`
8. **Enable security:** fail2ban, SSH hardening

Your SRE lab will now be accessible from anywhere in the world! 🌍
