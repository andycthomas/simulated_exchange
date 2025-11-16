# GoDaddy DNS Setup for andythomas-sre.com

Step-by-step guide to configure your GoDaddy domain for the SRE lab.

## 🎯 What You Need

1. **Your Server's Public IP Address**
2. **GoDaddy Account Access** for andythomas-sre.com

## 📋 Step 1: Find Your Public IP

On your server, run:

```bash
curl ifconfig.me
```

**Example Output:** `203.0.113.45`

**Write this down** - you'll need it for GoDaddy.

## 🌐 Step 2: Log Into GoDaddy

1. Go to https://www.godaddy.com
2. Click **Sign In**
3. Enter your credentials
4. Navigate to **My Products**

## ⚙️ Step 3: Access DNS Management

1. Find **andythomas-sre.com** in your domain list
2. Click the **DNS** button next to the domain
   - Or click **Manage DNS** if you see that option
3. You'll see the DNS Management page

## 🔧 Step 4: Configure DNS Records

### Option A: Simple Setup (Wildcard)

This is the **easiest** method - uses a wildcard to handle all subdomains.

#### Delete Existing Records (if any)

1. Look for existing **A records** with these names:
   - `@` (root domain)
   - `www`
   - Any other A records pointing to different IPs

2. Click the **pencil icon** (Edit) or **trash icon** (Delete) for each

#### Add New A Records

Click **Add** and create these two records:

**Record 1 - Root Domain:**
```
Type:    A
Name:    @
Value:   YOUR_SERVER_IP (e.g., 203.0.113.45)
TTL:     1 Hour (or 600 seconds)
```

**Record 2 - Wildcard for Subdomains:**
```
Type:    A
Name:    *
Value:   YOUR_SERVER_IP (e.g., 203.0.113.45)
TTL:     1 Hour (or 600 seconds)
```

Click **Save** after adding each record.

### Option B: Explicit Subdomains

If you prefer to explicitly list each subdomain instead of using wildcard:

Click **Add** and create these records:

**1. Root Domain:**
```
Type:    A
Name:    @
Value:   YOUR_SERVER_IP
TTL:     1 Hour
```

**2. WWW:**
```
Type:    A
Name:    www
Value:   YOUR_SERVER_IP
TTL:     1 Hour
```

**3. API:**
```
Type:    A
Name:    api
Value:   YOUR_SERVER_IP
TTL:     1 Hour
```

**4. Grafana:**
```
Type:    A
Name:    grafana
Value:   YOUR_SERVER_IP
TTL:     1 Hour
```

**5. Prometheus:**
```
Type:    A
Name:    prometheus
Value:   YOUR_SERVER_IP
TTL:     1 Hour
```

**6. Market:**
```
Type:    A
Name:    market
Value:   YOUR_SERVER_IP
TTL:     1 Hour
```

**7. Orders:**
```
Type:    A
Name:    orders
Value:   YOUR_SERVER_IP
TTL:     1 Hour
```

**8. Status:**
```
Type:    A
Name:    status
Value:   YOUR_SERVER_IP
TTL:     1 Hour
```

**9. Docs:**
```
Type:    A
Name:    docs
Value:   YOUR_SERVER_IP
TTL:     1 Hour
```

## ✅ Step 5: Verify Your Configuration

After saving, your DNS records should look like this:

```
Type    Name           Value           TTL
────────────────────────────────────────────────
A       @              203.0.113.45    1 Hour
A       *              203.0.113.45    1 Hour
```

Or if using explicit subdomains:

```
Type    Name           Value           TTL
────────────────────────────────────────────────
A       @              203.0.113.45    1 Hour
A       www            203.0.113.45    1 Hour
A       api            203.0.113.45    1 Hour
A       grafana        203.0.113.45    1 Hour
A       prometheus     203.0.113.45    1 Hour
A       market         203.0.113.45    1 Hour
A       orders         203.0.113.45    1 Hour
A       status         203.0.113.45    1 Hour
A       docs           203.0.113.45    1 Hour
```

## ⏰ Step 6: Wait for DNS Propagation

**Time Required:** 10 minutes to 48 hours (usually 30-60 minutes)

GoDaddy typically propagates changes within:
- **Minimum:** 10-30 minutes
- **Typical:** 1-2 hours
- **Maximum:** 48 hours

## 🔍 Step 7: Check DNS Propagation

### From Your Server

```bash
# Wait 10-15 minutes, then test
dig andythomas-sre.com +short

# Should return your server IP
```

### From Online Tools

Visit these websites and enter `andythomas-sre.com`:

1. **https://dnschecker.org**
   - Shows DNS status globally
   - Wait until most locations show your IP

2. **https://www.whatsmydns.net**
   - Alternative DNS checker
   - Check multiple locations

3. **https://mxtoolbox.com/DNSLookup.aspx**
   - Professional DNS lookup tool

### Expected Results

```bash
$ dig andythomas-sre.com +short
203.0.113.45

$ dig api.andythomas-sre.com +short
203.0.113.45

$ dig grafana.andythomas-sre.com +short
203.0.113.45
```

## 🚀 Step 8: Start Caddy

Once DNS is propagated (shows your correct IP):

```bash
cd /home/andy/simulated_exchange
make caddy-start
```

Caddy will automatically:
- Detect DNS is configured
- Request SSL certificate from Let's Encrypt
- Enable HTTPS for all domains
- Start serving traffic

## ✅ Step 9: Test Access

### From Your Phone (Cellular Network, NOT WiFi)

Open browser and visit:
- https://andythomas-sre.com
- https://api.andythomas-sre.com
- https://grafana.andythomas-sre.com

### From Command Line

```bash
curl -I https://andythomas-sre.com

# Expected response:
HTTP/2 200
server: Caddy
...
```

### Test All Subdomains

```bash
for sub in "" api grafana prometheus market orders status docs; do
  if [ -z "$sub" ]; then
    domain="andythomas-sre.com"
  else
    domain="$sub.andythomas-sre.com"
  fi
  echo -n "$domain: "
  curl -s -o /dev/null -w "%{http_code}\n" https://$domain || echo "Failed"
done
```

## 🔒 Optional: Add CAA Records (Recommended)

CAA records specify which Certificate Authorities can issue certificates for your domain.

### Add CAA Record

In GoDaddy DNS Management:

Click **Add** and create:

```
Type:    CAA
Name:    @
Tag:     issue
Value:   letsencrypt.org
TTL:     1 Hour
```

This tells the world that only Let's Encrypt can issue certificates for your domain.

## 🐛 Troubleshooting

### DNS Not Updating

**Problem:** DNS still shows old IP or no IP

**Solutions:**

1. **Wait longer** - Can take up to 48 hours
2. **Clear your local DNS cache:**
   ```bash
   # Linux
   sudo systemd-resolve --flush-caches

   # macOS
   sudo dscacheutil -flushcache

   # Windows
   ipconfig /flushdns
   ```

3. **Check GoDaddy DNS Management**
   - Verify records are saved
   - Check for typos in IP address

4. **Verify using GoDaddy's servers:**
   ```bash
   dig @ns1.domaincontrol.com andythomas-sre.com
   ```

### GoDaddy-Specific Issues

**Problem:** Changes not saving

**Solutions:**
- Log out and log back in
- Try a different browser
- Clear browser cache
- Contact GoDaddy support

**Problem:** Can't find DNS Management

**Solutions:**
- From **My Products**, look for your domain
- Click **DNS** button (may be under a "..." menu)
- Or look for **Manage DNS** link

**Problem:** Records keep reverting

**Solutions:**
- Check if domain is using **Forwarding** instead of DNS
  - Go to Domain Settings
  - Disable Forwarding if enabled
- Ensure you're not using a different nameserver

### SSL Certificate Not Obtained

**Problem:** Caddy can't get Let's Encrypt certificate

**Common Causes:**

1. **DNS not propagated yet**
   - Wait 30-60 minutes
   - Verify with `dig andythomas-sre.com`

2. **Port 80 not accessible**
   - Let's Encrypt needs port 80 for verification
   - Check firewall: `sudo ufw allow 80/tcp`
   - Check router port forwarding (if home network)

3. **Wrong IP in DNS**
   - Verify: `curl ifconfig.me`
   - Compare with DNS: `dig andythomas-sre.com +short`
   - Update GoDaddy if different

**Check Caddy Logs:**
```bash
make caddy-logs
```

Look for certificate acquisition errors.

## 📋 Verification Checklist

Before considering setup complete:

- [ ] GoDaddy DNS records added and saved
- [ ] DNS propagation complete (`dig andythomas-sre.com` shows correct IP)
- [ ] Server firewall allows ports 80/443
- [ ] Port forwarding configured (if home network)
- [ ] Caddy started (`make caddy-start`)
- [ ] SSL certificates obtained (`make caddy-certs`)
- [ ] Can access from external device (phone on cellular)
- [ ] All subdomains work (api, grafana, etc.)
- [ ] HTTPS works without errors

## 🆘 Need Help?

### Check External Access Setup

```bash
make caddy-external-check
```

This will verify:
- Your public IP
- DNS configuration
- Firewall status
- Caddy status

### View Logs

```bash
# Caddy logs
make caddy-logs

# System logs
sudo tail -f /var/log/syslog
```

### GoDaddy Support

If you're stuck with GoDaddy-specific issues:

- **Phone:** 480-505-8877
- **Chat:** Available in GoDaddy account
- **Help Center:** https://www.godaddy.com/help

## 📝 Summary

**What you changed in GoDaddy:**

1. ✅ Added A record: `@` → Your Server IP
2. ✅ Added A record: `*` → Your Server IP
3. ✅ (Optional) Added CAA record for Let's Encrypt

**What happens next:**

1. DNS propagates (30-60 minutes)
2. You start Caddy (`make caddy-start`)
3. Caddy gets SSL certificates automatically
4. Your lab is accessible worldwide via HTTPS

**Your services will be at:**

- https://andythomas-sre.com
- https://api.andythomas-sre.com
- https://grafana.andythomas-sre.com
- https://prometheus.andythomas-sre.com
- https://market.andythomas-sre.com
- https://orders.andythomas-sre.com
- https://status.andythomas-sre.com
- https://docs.andythomas-sre.com

All with automatic HTTPS and valid SSL certificates! 🎉
