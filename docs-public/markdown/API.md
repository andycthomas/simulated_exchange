# 📚 API Documentation

Comprehensive API reference for the Simulated Exchange Platform. This document provides detailed information about all available endpoints, request/response formats, and example usage.

## 🔗 Base URL

```
http://localhost:8080
```

## 🔐 Authentication

Currently, the API does not require authentication for demo purposes. In production, you would implement:
- JWT tokens for API access
- Rate limiting per API key
- Role-based access control

## 📋 API Overview

### Core Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | System health check |
| `/api/orders` | POST | Place new order |
| `/api/orders/{id}` | GET | Get order details |
| `/api/orders/{id}` | DELETE | Cancel order |
| `/api/metrics` | GET | Real-time system metrics |
| `/api/orderbook/{symbol}` | GET | Order book for symbol |

### Demo System Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/demo/load-test` | POST | Start load test scenario |
| `/demo/load-test` | DELETE | Stop load test |
| `/demo/chaos-test` | POST | Start chaos test scenario |
| `/demo/chaos-test` | DELETE | Stop chaos test |
| `/demo/status` | GET | Demo system status |
| `/demo/reset` | POST | Reset demo system |

### WebSocket Endpoints

| Endpoint | Protocol | Description |
|----------|----------|-------------|
| `/ws/demo` | WebSocket | Real-time demo updates |
| `/ws/metrics` | WebSocket | Live metrics stream |
| `/ws/orderbook` | WebSocket | Order book updates |

## 🏥 Health Check

### GET /health

Check system health and availability.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": {
    "metrics": "healthy",
    "simulation": "healthy",
    "demo": "healthy"
  },
  "version": "1.0.0"
}
```

**Status Codes:**
- `200 OK`: System is healthy
- `503 Service Unavailable`: System has issues

**Example:**
```bash
curl http://localhost:8080/health
```

## 📈 Trading API

### POST /api/orders

Place a new trading order.

**Request Body:**
```json
{
  "symbol": "BTCUSD",
  "side": "buy",
  "type": "market",
  "quantity": 1.0,
  "price": 50000.0
}
```

**Fields:**
- `symbol` (string, required): Trading pair (e.g., "BTCUSD", "ETHUSD")
- `side` (string, required): Order side ("buy" or "sell")
- `type` (string, required): Order type ("market" or "limit")
- `quantity` (number, required): Order quantity (> 0)
- `price` (number, required): Order price (> 0)

**Response:**
```json
{
  "success": true,
  "data": {
    "order_id": "order_1705312200123_BTCUSD",
    "status": "placed",
    "message": "Order placed successfully"
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_ORDER",
    "message": "Invalid order parameters",
    "details": "Quantity must be greater than 0"
  }
}
```

**Status Codes:**
- `201 Created`: Order placed successfully
- `400 Bad Request`: Invalid order parameters
- `422 Unprocessable Entity`: Business rule violation
- `500 Internal Server Error`: System error

**Examples:**

```bash
# Market Buy Order
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSD",
    "side": "buy",
    "type": "market",
    "quantity": 1.0,
    "price": 50000.0
  }'

# Limit Sell Order
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "ETHUSD",
    "side": "sell",
    "type": "limit",
    "quantity": 10.0,
    "price": 3000.0
  }'
```

### GET /api/orders/{id}

Retrieve order details by ID.

**Path Parameters:**
- `id` (string): Order ID

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "order_1705312200123_BTCUSD",
    "symbol": "BTCUSD",
    "side": "buy",
    "type": "market",
    "quantity": 1.0,
    "price": 50000.0,
    "status": "filled"
  }
}
```

**Status Codes:**
- `200 OK`: Order found
- `404 Not Found`: Order not found

**Example:**
```bash
curl http://localhost:8080/api/orders/order_1705312200123_BTCUSD
```

### DELETE /api/orders/{id}

Cancel an existing order.

**Path Parameters:**
- `id` (string): Order ID

**Response:**
```json
{
  "success": true,
  "data": {
    "order_id": "order_1705312200123_BTCUSD",
    "status": "cancelled",
    "message": "Order cancelled successfully"
  }
}
```

**Status Codes:**
- `200 OK`: Order cancelled
- `404 Not Found`: Order not found
- `409 Conflict`: Order cannot be cancelled (already filled)

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/orders/order_1705312200123_BTCUSD
```

### GET /api/orderbook/{symbol}

Get current order book for a trading symbol.

**Path Parameters:**
- `symbol` (string): Trading pair (e.g., "BTCUSD")

**Response:**
```json
{
  "success": true,
  "data": {
    "symbol": "BTCUSD",
    "bids": [
      {"price": 49990.0, "quantity": 2.5},
      {"price": 49980.0, "quantity": 1.0}
    ],
    "asks": [
      {"price": 50010.0, "quantity": 1.5},
      {"price": 50020.0, "quantity": 3.0}
    ]
  }
}
```

**Example:**
```bash
curl http://localhost:8080/api/orderbook/BTCUSD
```

## 📊 Metrics API

### GET /api/metrics

Get real-time system metrics.

**Response:**
```json
{
  "order_count": 1234,
  "trade_count": 567,
  "total_volume": 12345678.90,
  "avg_latency": "45ms",
  "orders_per_sec": 123.45,
  "trades_per_sec": 67.89,
  "symbol_metrics": {
    "BTCUSD": {
      "order_count": 890,
      "trade_count": 345,
      "volume": 8901234.56,
      "avg_price": 50123.45
    },
    "ETHUSD": {
      "order_count": 344,
      "trade_count": 222,
      "volume": 3444444.34,
      "avg_price": 3123.45
    }
  },
  "analysis": {
    "timestamp": "2024-01-15T10:30:00Z",
    "trend_direction": "stable",
    "bottlenecks": [],
    "recommendations": [
      "System performance is within normal parameters",
      "Consider scaling if traffic increases significantly"
    ]
  }
}
```

**Example:**
```bash
curl http://localhost:8080/api/metrics
```

## 🎭 Demo System API

### POST /demo/load-test

Start a load testing scenario.

**Request Body:**
```json
{
  "scenario": "light",
  "duration": "60s",
  "orders_per_second": 10,
  "concurrent_users": 5,
  "symbols": ["BTCUSD", "ETHUSD"]
}
```

**Predefined Scenarios:**
- `light`: 10 ops/sec, 5 users
- `medium`: 50 ops/sec, 25 users
- `heavy`: 200 ops/sec, 100 users
- `stress`: 1000+ ops/sec, 500+ users

**Response:**
```json
{
  "success": true,
  "data": {
    "scenario_id": "load_test_1705312200",
    "status": "started",
    "estimated_duration": "60s",
    "websocket_url": "ws://localhost:8081/ws/demo"
  }
}
```

**Example:**
```bash
# Start light load test
curl -X POST http://localhost:8080/demo/load-test \
  -H "Content-Type: application/json" \
  -d '{
    "scenario": "light",
    "duration": "60s",
    "orders_per_second": 10
  }'

# Start custom load test
curl -X POST http://localhost:8080/demo/load-test \
  -H "Content-Type: application/json" \
  -d '{
    "scenario": "custom",
    "duration": "120s",
    "orders_per_second": 25,
    "concurrent_users": 15,
    "symbols": ["BTCUSD", "ETHUSD", "ADAUSD"]
  }'
```

### DELETE /demo/load-test

Stop the current load test.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "stopped",
    "final_metrics": {
      "total_orders": 1234,
      "success_rate": 99.8,
      "avg_latency": "42ms",
      "duration": "45s"
    }
  }
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8080/demo/load-test
```

### POST /demo/chaos-test

Start a chaos engineering test.

**Request Body:**
```json
{
  "type": "latency_injection",
  "duration": "30s",
  "severity": "medium",
  "parameters": {
    "latency_ms": 100,
    "failure_rate": 0.1
  }
}
```

**Chaos Test Types:**
- `latency_injection`: Add artificial latency
- `error_simulation`: Inject random errors
- `resource_exhaustion`: Simulate resource constraints
- `network_partition`: Simulate network issues

**Severity Levels:**
- `low`: Minimal impact
- `medium`: Moderate impact
- `high`: Significant impact

**Response:**
```json
{
  "success": true,
  "data": {
    "chaos_id": "chaos_test_1705312200",
    "type": "latency_injection",
    "status": "active",
    "safety_limits": {
      "max_error_rate": 0.5,
      "max_latency_ms": 5000,
      "auto_recovery": true
    }
  }
}
```

**Examples:**

```bash
# Latency injection test
curl -X POST http://localhost:8080/demo/chaos-test \
  -H "Content-Type: application/json" \
  -d '{
    "type": "latency_injection",
    "duration": "30s",
    "severity": "medium"
  }'

# Error simulation test
curl -X POST http://localhost:8080/demo/chaos-test \
  -H "Content-Type: application/json" \
  -d '{
    "type": "error_simulation",
    "duration": "45s",
    "severity": "low",
    "parameters": {
      "error_rate": 0.05,
      "error_types": ["timeout", "connection_reset"]
    }
  }'
```

### DELETE /demo/chaos-test

Stop the current chaos test.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "stopped",
    "recovery_metrics": {
      "recovery_time": "2.3s",
      "final_error_rate": 0.02,
      "system_stability": "restored"
    }
  }
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8080/demo/chaos-test
```

### GET /demo/status

Get current demo system status.

**Response:**
```json
{
  "overall": "active",
  "active_scenarios": ["load_test_light"],
  "load_test": {
    "is_running": true,
    "scenario": "light",
    "progress": 45.2,
    "phase": "sustained",
    "metrics": {
      "orders_processed": 567,
      "success_rate": 99.1,
      "avg_latency": "38ms"
    }
  },
  "chaos_test": {
    "is_running": false,
    "last_completed": "2024-01-15T10:25:00Z"
  },
  "websocket_stats": {
    "active_connections": 3,
    "messages_sent": 1234
  }
}
```

**Example:**
```bash
curl http://localhost:8080/demo/status
```

### POST /demo/reset

Reset the demo system to clean state.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "reset_complete",
    "cleared_items": [
      "active_orders",
      "test_scenarios",
      "metrics_history"
    ]
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/demo/reset
```

## 🔌 WebSocket API

### /ws/demo

Real-time demo updates via WebSocket.

**Connection:**
```javascript
const ws = new WebSocket('ws://localhost:8081/ws/demo');
```

**Message Types:**

1. **Load Test Updates:**
```json
{
  "type": "load_test_status",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "progress": 67.5,
    "phase": "sustained",
    "orders_per_second": 23.4,
    "success_rate": 99.2,
    "avg_latency": "41ms"
  }
}
```

2. **Chaos Test Updates:**
```json
{
  "type": "chaos_test_status",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "type": "latency_injection",
    "phase": "injection",
    "impact_metrics": {
      "latency_increase": "150%",
      "error_rate": 0.12
    }
  }
}
```

3. **System Metrics:**
```json
{
  "type": "system_metrics",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "cpu_usage": 45.2,
    "memory_usage": 62.8,
    "active_connections": 125,
    "throughput": 234.5
  }
}
```

**Example Usage:**
```javascript
const ws = new WebSocket('ws://localhost:8081/ws/demo');

ws.onmessage = function(event) {
  const update = JSON.parse(event.data);
  console.log('Demo update:', update);

  switch(update.type) {
    case 'load_test_status':
      updateLoadTestDashboard(update.data);
      break;
    case 'chaos_test_status':
      updateChaosTestDashboard(update.data);
      break;
    case 'system_metrics':
      updateMetricsDashboard(update.data);
      break;
  }
};
```

## 📝 Error Handling

### Standard Error Response

All errors follow a consistent format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Additional error details",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Common Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `INVALID_REQUEST` | Malformed request | 400 |
| `INVALID_ORDER` | Invalid order parameters | 400 |
| `ORDER_NOT_FOUND` | Order does not exist | 404 |
| `SYSTEM_OVERLOAD` | System at capacity | 503 |
| `DEMO_CONFLICT` | Demo already running | 409 |
| `CHAOS_LIMIT_EXCEEDED` | Safety limits exceeded | 422 |

## 🔄 Rate Limiting

Current rate limits (per IP):
- **Trading API**: 100 requests/minute
- **Metrics API**: 60 requests/minute
- **Demo API**: 10 requests/minute
- **WebSocket**: 5 connections/IP

Rate limit headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1705312260
```

## 📊 Response Times

Typical response times:
- **Order Placement**: < 50ms (P99)
- **Order Query**: < 10ms (P99)
- **Metrics API**: < 100ms (P99)
- **Demo Control**: < 200ms (P99)

## 🔧 SDK Examples

### Python Example

```python
import requests
import json

class SimulatedExchangeClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url

    def place_order(self, symbol, side, order_type, quantity, price):
        url = f"{self.base_url}/api/orders"
        data = {
            "symbol": symbol,
            "side": side,
            "type": order_type,
            "quantity": quantity,
            "price": price
        }
        response = requests.post(url, json=data)
        return response.json()

    def get_metrics(self):
        url = f"{self.base_url}/api/metrics"
        response = requests.get(url)
        return response.json()

    def start_load_test(self, scenario="light", duration="60s"):
        url = f"{self.base_url}/demo/load-test"
        data = {
            "scenario": scenario,
            "duration": duration
        }
        response = requests.post(url, json=data)
        return response.json()

# Usage
client = SimulatedExchangeClient()
result = client.place_order("BTCUSD", "buy", "market", 1.0, 50000.0)
print(result)
```

### JavaScript Example

```javascript
class SimulatedExchangeClient {
  constructor(baseUrl = 'http://localhost:8080') {
    this.baseUrl = baseUrl;
  }

  async placeOrder(symbol, side, type, quantity, price) {
    const response = await fetch(`${this.baseUrl}/api/orders`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        symbol, side, type, quantity, price
      })
    });
    return response.json();
  }

  async getMetrics() {
    const response = await fetch(`${this.baseUrl}/api/metrics`);
    return response.json();
  }

  async startLoadTest(scenario = 'light', duration = '60s') {
    const response = await fetch(`${this.baseUrl}/demo/load-test`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ scenario, duration })
    });
    return response.json();
  }
}

// Usage
const client = new SimulatedExchangeClient();
client.placeOrder('BTCUSD', 'buy', 'market', 1.0, 50000.0)
  .then(result => console.log(result));
```

## 🧪 Testing the API

### Health Check Test
```bash
curl -s http://localhost:8080/health | jq '.status'
# Expected: "healthy"
```

### Complete Order Flow Test
```bash
#!/bin/bash

# Place order
ORDER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTCUSD","side":"buy","type":"market","quantity":1.0,"price":50000.0}')

ORDER_ID=$(echo $ORDER_RESPONSE | jq -r '.data.order_id')
echo "Order placed: $ORDER_ID"

# Check order status
curl -s http://localhost:8080/api/orders/$ORDER_ID | jq '.data.status'

# Get metrics
curl -s http://localhost:8080/api/metrics | jq '.order_count'
```

### Demo System Test
```bash
#!/bin/bash

# Start load test
curl -X POST http://localhost:8080/demo/load-test \
  -H "Content-Type: application/json" \
  -d '{"scenario":"light","duration":"30s"}'

# Monitor progress
for i in {1..5}; do
  sleep 5
  curl -s http://localhost:8080/demo/status | jq '.load_test.progress'
done

# Stop test
curl -X DELETE http://localhost:8080/demo/load-test
```

---

## 📞 Support

For API questions or issues:
- **Documentation**: [Full API Docs](API.md)
- **Examples**: [GitHub Repository](https://github.com/your-org/simulated_exchange)
- **Issues**: [Report API Issues](https://github.com/your-org/simulated_exchange/issues)

**Happy Trading!** 🚀