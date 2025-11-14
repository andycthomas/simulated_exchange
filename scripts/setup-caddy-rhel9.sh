#!/bin/bash
#
# Caddy Setup Script for RHEL 9
# Sets up Caddy as reverse proxy for SSH-tunneled trading API
#
# Usage:
#   ./setup-caddy-rhel9.sh [hostname]
#
# Example:
#   ./setup-caddy-rhel9.sh pcoactlin1.etc.cboe.net
#

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TUNNEL_PORT=7000
CADDY_HTTP_PORT=80
CADDY_HTTPS_PORT=443
HOSTNAME="${1:-localhost}"
USE_HTTPS="${USE_HTTPS:-false}"

# Functions
log_info() {
    printf "${BLUE}ℹ${NC} %s\n" "$1"
}

log_success() {
    printf "${GREEN}✓${NC} %s\n" "$1"
}

log_warning() {
    printf "${YELLOW}⚠${NC} %s\n" "$1"
}

log_error() {
    printf "${RED}✗${NC} %s\n" "$1"
}

check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

check_rhel() {
    if [ ! -f /etc/redhat-release ]; then
        log_error "This script is designed for RHEL/Rocky/AlmaLinux"
        exit 1
    fi

    VERSION=$(cat /etc/redhat-release)
    log_info "Detected: $VERSION"
}

install_caddy() {
    log_info "Installing Caddy..."

    # Check if already installed
    if command -v caddy &> /dev/null; then
        log_warning "Caddy is already installed ($(caddy version))"
        read -p "Reinstall? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            return 0
        fi
    fi

    # Install dependencies
    log_info "Installing dependencies..."
    dnf install -y yum-utils curl

    # Import Caddy GPG key
    log_info "Adding Caddy repository..."
    rpm --import https://dl.cloudsmith.io/public/caddy/stable/gpg.key

    # Add Caddy repository
    cat > /etc/yum.repos.d/caddy.repo << 'EOF'
[caddy]
name=Caddy
baseurl=https://dl.cloudsmith.io/public/caddy/stable/rpm/el/9/$basearch
gpgcheck=1
enabled=1
gpgkey=https://dl.cloudsmith.io/public/caddy/stable/gpg.key
EOF

    # Install Caddy
    log_info "Installing Caddy package..."
    dnf install -y caddy

    log_success "Caddy installed: $(caddy version)"
}

configure_firewall() {
    log_info "Configuring firewall..."

    # Check if firewalld is running
    if ! systemctl is-active --quiet firewalld; then
        log_warning "firewalld is not running, starting it..."
        systemctl start firewalld
        systemctl enable firewalld
    fi

    # Add HTTP/HTTPS services
    firewall-cmd --permanent --add-service=http
    firewall-cmd --permanent --add-service=https

    # Reload firewall
    firewall-cmd --reload

    log_success "Firewall configured (HTTP and HTTPS allowed)"

    # Show current rules
    log_info "Current firewall rules:"
    firewall-cmd --list-all | grep -E "(services|ports):"
}

configure_selinux() {
    log_info "Configuring SELinux..."

    # Check SELinux status
    SELINUX_STATUS=$(getenforce)
    log_info "SELinux status: $SELINUX_STATUS"

    if [ "$SELINUX_STATUS" != "Disabled" ]; then
        # Allow Caddy to make network connections
        log_info "Allowing HTTP network connections..."
        setsebool -P httpd_can_network_connect 1

        # Allow binding to tunnel port if needed
        log_info "Configuring port permissions..."
        semanage port -a -t http_port_t -p tcp $TUNNEL_PORT 2>/dev/null || \
            semanage port -m -t http_port_t -p tcp $TUNNEL_PORT 2>/dev/null || \
            log_warning "Port $TUNNEL_PORT already configured or couldn't be modified"

        log_success "SELinux configured for Caddy"
    else
        log_warning "SELinux is disabled"
    fi
}

create_caddyfile() {
    log_info "Creating Caddyfile configuration..."

    # Backup existing Caddyfile if present
    if [ -f /etc/caddy/Caddyfile ]; then
        BACKUP="/etc/caddy/Caddyfile.backup.$(date +%Y%m%d-%H%M%S)"
        cp /etc/caddy/Caddyfile "$BACKUP"
        log_info "Backed up existing Caddyfile to $BACKUP"
    fi

    # Create log directory
    mkdir -p /var/log/caddy
    chown caddy:caddy /var/log/caddy

    # Determine configuration based on hostname
    if [ "$HOSTNAME" = "localhost" ] || [ "$USE_HTTPS" = "false" ]; then
        # Simple HTTP configuration
        cat > /etc/caddy/Caddyfile << EOF
{
    # Global options
    admin off
    log {
        output file /var/log/caddy/caddy.log
        level INFO
    }
}

# HTTP proxy to tunneled trading API
:${CADDY_HTTP_PORT} {
    # Reverse proxy to SSH tunnel
    reverse_proxy localhost:${TUNNEL_PORT} {
        # Health checks
        health_uri /health
        health_interval 10s
        health_timeout 5s
        health_status 2xx

        # Failover settings
        fail_duration 30s
        max_fails 3
        unhealthy_status 5xx

        # Forward client information
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }

    # CORS support for web clients
    @cors_preflight {
        method OPTIONS
    }
    handle @cors_preflight {
        header {
            Access-Control-Allow-Origin "*"
            Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS"
            Access-Control-Allow-Headers "Content-Type, Authorization"
            Access-Control-Max-Age "3600"
        }
        respond 204
    }

    # Add CORS headers to all responses
    header {
        Access-Control-Allow-Origin "*"
    }

    # Compression
    encode gzip zstd

    # Access logging
    log {
        output file /var/log/caddy/trading-access.log {
            roll_size 100mb
            roll_keep 5
            roll_keep_days 7
        }
        format json
    }
}
EOF
        log_success "Created HTTP-only Caddyfile for :${CADDY_HTTP_PORT}"
    else
        # HTTPS configuration with hostname
        cat > /etc/caddy/Caddyfile << EOF
{
    # Global options
    admin off
    log {
        output file /var/log/caddy/caddy.log
        level INFO
    }
}

# Main trading API endpoint
${HOSTNAME} {
    # Reverse proxy to SSH tunnel
    reverse_proxy localhost:${TUNNEL_PORT} {
        # Health checks
        health_uri /health
        health_interval 10s
        health_timeout 5s
        health_status 2xx

        # Failover settings
        fail_duration 30s
        max_fails 3
        unhealthy_status 5xx

        # Forward client information
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }

    # CORS support
    @cors_preflight {
        method OPTIONS
    }
    handle @cors_preflight {
        header {
            Access-Control-Allow-Origin "*"
            Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS"
            Access-Control-Allow-Headers "Content-Type, Authorization"
            Access-Control-Max-Age "3600"
        }
        respond 204
    }

    header {
        Access-Control-Allow-Origin "*"
    }

    # Rate limiting
    rate_limit {
        zone trading {
            key {remote_host}
            events 500
            window 1m
        }
    }

    # Compression
    encode gzip zstd

    # Access logging
    log {
        output file /var/log/caddy/trading-access.log {
            roll_size 100mb
            roll_keep 5
            roll_keep_days 7
        }
        format json
    }

    # TLS configuration
    tls internal
}
EOF
        log_success "Created HTTPS Caddyfile for ${HOSTNAME}"
    fi

    # Set proper permissions
    chown root:caddy /etc/caddy/Caddyfile
    chmod 640 /etc/caddy/Caddyfile
}

validate_and_start_caddy() {
    log_info "Validating Caddy configuration..."

    if caddy validate --config /etc/caddy/Caddyfile; then
        log_success "Caddyfile is valid"
    else
        log_error "Caddyfile validation failed!"
        exit 1
    fi

    log_info "Starting Caddy service..."
    systemctl enable caddy
    systemctl restart caddy

    # Wait a moment for service to start
    sleep 2

    if systemctl is-active --quiet caddy; then
        log_success "Caddy is running"
    else
        log_error "Caddy failed to start"
        log_info "Check logs with: journalctl -u caddy -n 50"
        exit 1
    fi
}

create_verification_script() {
    log_info "Creating verification script..."

    cat > /usr/local/bin/check-trading-api << 'EOFSCRIPT'
#!/bin/bash
# Quick verification script for trading API setup

echo "==================================="
echo "Trading API Status Check"
echo "==================================="
echo ""

# Check Caddy
echo "1. Caddy Service:"
systemctl status caddy --no-pager -l | head -3
echo ""

# Check if tunnel port is listening
echo "2. SSH Tunnel (port 7000):"
if ss -tulnp | grep -q ":7000"; then
    echo "✓ Port 7000 is listening"
    ss -tulnp | grep ":7000"
else
    echo "✗ Port 7000 is NOT listening (tunnel may be down)"
fi
echo ""

# Check Caddy port
echo "3. Caddy HTTP Port:"
if ss -tulnp | grep -q ":80"; then
    echo "✓ Port 80 is listening"
else
    echo "✗ Port 80 is NOT listening"
fi
echo ""

# Test local health endpoint through tunnel
echo "4. Direct Tunnel Test:"
if timeout 3 curl -s http://localhost:7000/health > /dev/null 2>&1; then
    echo "✓ Tunnel endpoint responds"
    curl -s http://localhost:7000/health | head -3
else
    echo "✗ Tunnel endpoint not responding"
fi
echo ""

# Test through Caddy
echo "5. Caddy Proxy Test:"
if timeout 3 curl -s http://localhost/health > /dev/null 2>&1; then
    echo "✓ Caddy proxy responds"
    curl -s http://localhost/health | head -3
else
    echo "✗ Caddy proxy not responding"
fi
echo ""

# Recent Caddy logs
echo "6. Recent Caddy Logs:"
journalctl -u caddy -n 5 --no-pager
echo ""

echo "==================================="
echo "For detailed logs: journalctl -u caddy -f"
echo "For access logs: tail -f /var/log/caddy/trading-access.log | jq"
echo "==================================="
EOFSCRIPT

    chmod +x /usr/local/bin/check-trading-api
    log_success "Created verification script: /usr/local/bin/check-trading-api"
}

show_summary() {
    echo ""
    echo "========================================================================"
    printf "${GREEN}Caddy Setup Complete!${NC}\n"
    echo "========================================================================"
    echo ""
    echo "Configuration:"
    echo "  - Caddy is listening on port: $CADDY_HTTP_PORT"
    echo "  - Proxying to tunnel port: $TUNNEL_PORT"
    if [ "$HOSTNAME" != "localhost" ]; then
        echo "  - Hostname: $HOSTNAME"
    fi
    echo "  - Logs: /var/log/caddy/"
    echo ""
    echo "Next Steps:"
    echo ""
    echo "1. On the source host (157.180.81.233), create the SSH tunnel:"
    echo "   ${YELLOW}ssh -p 443 -N -R ${TUNNEL_PORT}:localhost:8080 andy@${HOSTNAME}${NC}"
    echo ""
    echo "2. Verify the setup on this host:"
    echo "   ${YELLOW}check-trading-api${NC}"
    echo ""
    echo "3. Test the API:"
    echo "   ${YELLOW}curl http://localhost/health${NC}"
    if [ "$HOSTNAME" != "localhost" ]; then
        echo "   ${YELLOW}curl http://${HOSTNAME}/health${NC}"
    fi
    echo ""
    echo "Useful Commands:"
    echo "  - Check Caddy status: ${YELLOW}systemctl status caddy${NC}"
    echo "  - View Caddy logs: ${YELLOW}journalctl -u caddy -f${NC}"
    echo "  - View access logs: ${YELLOW}tail -f /var/log/caddy/trading-access.log | jq${NC}"
    echo "  - Reload config: ${YELLOW}systemctl reload caddy${NC}"
    echo "  - Validate config: ${YELLOW}caddy validate --config /etc/caddy/Caddyfile${NC}"
    echo "  - Check tunnel: ${YELLOW}ss -tulnp | grep ${TUNNEL_PORT}${NC}"
    echo ""
    echo "========================================================================"
}

# Main execution
main() {
    echo "========================================================================"
    echo "Caddy Setup for RHEL 9 - Trading API Reverse Proxy"
    echo "========================================================================"
    echo ""

    check_root
    check_rhel

    log_info "Configuration:"
    log_info "  Tunnel port: $TUNNEL_PORT"
    log_info "  Caddy HTTP port: $CADDY_HTTP_PORT"
    log_info "  Hostname: $HOSTNAME"
    echo ""

    read -p "Continue with installation? [Y/n] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        log_info "Installation cancelled"
        exit 0
    fi

    install_caddy
    configure_firewall
    configure_selinux
    create_caddyfile
    validate_and_start_caddy
    create_verification_script
    show_summary
}

# Run main function
main "$@"
