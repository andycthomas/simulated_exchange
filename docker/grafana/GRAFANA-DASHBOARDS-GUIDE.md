# Grafana Dashboards Guide

Complete guide to the Grafana dashboards for the Simulated Exchange system.

## 📊 Dashboard Overview

Six comprehensive dashboards have been created to monitor all aspects of your simulated trading exchange:

1. **Exchange Overview** - High-level system metrics and trading activity
2. **Order Book Performance** - Detailed order processing and execution metrics
3. **System Health & Resources** - Service health, connections, and infrastructure
4. **Trading Activity & Analytics** - In-depth trading patterns and analytics
5. **Event Bus & Integration** - Event processing and inter-service communication
6. **Performance Metrics** - Latency, throughput, and performance analysis

## 🔗 Accessing Dashboards

### Via HTTPS (Recommended)
```
https://grafana.andythomas-sre.com
```

**Default Credentials:**
- Username: `admin`
- Password: `admin123`

**⚠️ IMPORTANT: Change the default password immediately!**

## 📈 Dashboard Details

### 1. Exchange Overview
**UID:** `exchange-overview`

**Key Metrics:**
- Orders Per Minute
- Trades Per Minute
- Order Processing P95 Latency
- System Health
- Order Fill Rate by Side
- Active Orders by Symbol
- Current Market Prices
- Market Volatility
- Order Status Distribution

### 2. Order Book Performance
**UID:** `order-book-performance`

**Key Metrics:**
- Order Processing Latency Distribution (P50, P95, P99)
- Average Order Processing Time by Side
- Order Types Distribution
- Order Book Depth
- Order Throughput by Status
- Order Fill Rate %
- Trade Execution Rate by Symbol

### 3. System Health & Resources
**UID:** `system-health`

**Key Metrics:**
- Service Health Status
- System Uptime
- HTTP Request Rate
- HTTP Response Times
- Redis Connections
- PostgreSQL Connections
- HTTP Status Codes
- Simulation Errors

### 4. Trading Activity & Analytics
**UID:** `trading-activity`

**Key Metrics:**
- Order Flow Analysis
- Trading Volume by Symbol
- Order Status & Type Distribution
- Active Orders Breakdown
- Order Book Depth Heatmap
- Processing Latency by Type

### 5. Event Bus & Integration
**UID:** `event-bus`

**Key Metrics:**
- Events Published/Processed
- Event Success Rate
- Event Processing Latency
- Event Type Distribution
- Event Flow by Service

### 6. Performance Metrics
**UID:** `performance-metrics`

**Key Metrics:**
- Total Request Rate
- Success Rate
- P50, P95, P99 Response Times
- Response Time by Service
- Endpoint Performance Table
- Request Count by Status

## 🎯 Quick Access URLs

Once logged in:
- https://grafana.andythomas-sre.com/d/exchange-overview
- https://grafana.andythomas-sre.com/d/order-book-performance
- https://grafana.andythomas-sre.com/d/system-health
- https://grafana.andythomas-sre.com/d/trading-activity
- https://grafana.andythomas-sre.com/d/event-bus
- https://grafana.andythomas-sre.com/d/performance-metrics

## 🔐 Security

### Change Default Password
```bash
docker exec -it simulated-exchange-grafana grafana-cli admin reset-admin-password YOUR_NEW_PASSWORD
```

## 🐛 Troubleshooting

### No Data Showing
1. Check Prometheus is running: `docker ps | grep prometheus`
2. Verify targets: https://prometheus.andythomas-sre.com/targets
3. Test datasource in Grafana UI

### Dashboards Not Loading
```bash
docker restart simulated-exchange-grafana
docker logs simulated-exchange-grafana | grep provisioning
```

## 📊 Dashboard Files

All dashboards are located in:
```
docker/grafana/dashboards/
├── 01-exchange-overview.json
├── 02-order-book-performance.json
├── 03-system-health.json
├── 04-trading-activity.json
├── 05-event-bus.json
└── 06-performance-metrics.json
```

---

**Last Updated:** 2025-11-14
