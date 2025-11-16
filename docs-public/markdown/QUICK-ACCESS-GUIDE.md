# Quick Access Setup - TL;DR

## 🚀 5-Minute Setup

### 1. Get Your Public IP
```bash
curl ifconfig.me
```
**Note this IP** (e.g., `203.0.113.45`)

### 2. Configure DNS
Go to your domain registrar and add:

```
Type    Name    Value           TTL
A       @       203.0.113.45    3600
A       *       203.0.113.45    3600
```

### 3. Open Firewall
```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw status
```

### 4. Start Caddy
```bash
cd /home/andy/simulated_exchange
make caddy-start
```

### 5. Wait & Test
```bash
# Wait 10-60 minutes for DNS propagation
# Then test from external device:
curl -I https://andythomas-sre.com
```

## ✅ That's it!

Your lab is now accessible at:
- https://andythomas-sre.com
- https://api.andythomas-sre.com
- https://grafana.andythomas-sre.com

## 🏠 Home Network? Add Port Forwarding

Router Settings → Port Forwarding:
```
External Port 80  → Internal IP: YOUR_SERVER_IP → Port 80
External Port 443 → Internal IP: YOUR_SERVER_IP → Port 443
```

## 🔍 Troubleshooting

| Problem | Solution |
|---------|----------|
| DNS not working | Wait 30-60 min, check https://dnschecker.org |
| Can't get SSL | Make sure ports 80/443 open from internet |
| Works locally only | Check port forwarding on router |

## 📖 Full Guide
See `EXTERNAL-ACCESS.md` for complete instructions.
