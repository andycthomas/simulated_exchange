# Simulated Exchange Platform - Architecture Diagrams

This document contains visual architecture diagrams for the Simulated Exchange Platform using Mermaid syntax. These diagrams can be viewed in GitHub, GitLab, or any markdown viewer that supports Mermaid.

## Table of Contents
1. [System Context Diagram](#system-context-diagram)
2. [High-Level Architecture](#high-level-architecture)
3. [Component Diagram](#component-diagram)
4. [Deployment Diagram](#deployment-diagram)
5. [Order Processing Flow](#order-processing-flow)
6. [Market Simulation Flow](#market-simulation-flow)
7. [Demo Load Test Flow](#demo-load-test-flow)
8. [Chaos Engineering Flow](#chaos-engineering-flow)
9. [Dependency Graph](#dependency-graph)
10. [Data Model Diagram](#data-model-diagram)

---

## System Context Diagram

Shows the system and its interactions with external actors.

```mermaid
graph TB
    subgraph External Actors
        User[User/Trader]
        Admin[System Administrator]
        Monitor[Monitoring System]
        LoadTester[Load Tester]
    end

    subgraph Simulated Exchange Platform
        API[REST API]
        WS[WebSocket Server]
        Dashboard[Web Dashboard]
    end

    User -->|Place Orders| API
    User -->|View Dashboard| Dashboard
    User -->|Real-time Updates| WS

    Admin -->|Configure| API
    Admin -->|Monitor| Dashboard

    LoadTester -->|Load Tests| API
    LoadTester -->|Chaos Tests| API

    Monitor -->|Scrape Metrics| API

    API -->|Metrics| Monitor
    WS -->|Live Updates| User
    Dashboard -->|Data| API
```

---

## High-Level Architecture

Shows the layered architecture of the system.

```mermaid
graph TB
    subgraph Presentation Layer
        REST[REST API<br/>Gin Framework]
        WS[WebSocket<br/>gorilla/websocket]
        UI[Web Dashboard<br/>HTML/CSS/JS]
    end

    subgraph Application Layer
        OH[Order Handler]
        MH[Metrics Handler]
        DC[Demo Controller]
    end

    subgraph Business Layer
        TE[Trading Engine]
        MS[Market Simulator]
        MT[Metrics Service]
        HS[Health Service]
    end

    subgraph Data Layer
        OR[Order Repository]
        TR[Trade Repository]
        MC[Metrics Collector]
    end

    REST --> OH
    REST --> MH
    REST --> DC
    WS --> DC
    UI --> REST

    OH --> TE
    MH --> MT
    DC --> MS
    DC --> TE

    TE --> OR
    TE --> TR
    TE --> MC
    MS --> TE
    MT --> MC

    OR -.->|In-Memory| DB[(Storage)]
    TR -.->|In-Memory| DB
```

---

## Component Diagram

Detailed view of all system components and their relationships.

```mermaid
graph TB
    subgraph Application Container
        APP[Application<br/>Orchestrator]
        CONT[Dependency<br/>Container]
    end

    subgraph API Components
        SRV[HTTP Server]
        OrderH[Order Handler]
        MetricsH[Metrics Handler]
        MW[Middleware Stack]
    end

    subgraph Core Services
        OS[Order Service]
        MS[Metrics Service]
        HS[Health Service]
    end

    subgraph Trading Components
        TE[Trading Engine]
        OM[Order Matcher]
        TX[Trade Executor]
    end

    subgraph Simulation Components
        Sim[Market Simulator]
        PG[Price Generator]
        OG[Order Generator]
        EG[Event Generator]
    end

    subgraph Demo Components
        DemoCtrl[Demo Controller]
        Scenarios[Scenario Manager]
        WSMgr[WebSocket Manager]
    end

    subgraph Metrics Components
        Coll[Metrics Collector]
        Analyzer[AI Analyzer]
    end

    APP --> CONT
    CONT --> SRV
    CONT --> OS
    CONT --> MS
    CONT --> HS
    CONT --> TE
    CONT --> Sim

    SRV --> MW
    SRV --> OrderH
    SRV --> MetricsH

    OrderH --> OS
    MetricsH --> MS

    OS --> TE
    TE --> OM
    TE --> TX

    Sim --> PG
    Sim --> OG
    Sim --> EG
    Sim --> OS

    DemoCtrl --> Scenarios
    DemoCtrl --> WSMgr
    DemoCtrl --> Sim

    MS --> Coll
    MS --> Analyzer

    TE -.->|Records| Coll
```

---

## Deployment Diagram

Shows deployment architecture for different environments.

```mermaid
graph TB
    subgraph Production Kubernetes Cluster
        LB[Load Balancer<br/>Ingress]

        subgraph Application Pods
            POD1[App Pod 1<br/>Port: 8080]
            POD2[App Pod 2<br/>Port: 8080]
            POD3[App Pod 3<br/>Port: 8080]
        end

        subgraph Data Layer
            PG[(PostgreSQL<br/>StatefulSet)]
            RD[(Redis<br/>StatefulSet)]
        end

        subgraph Monitoring
            PROM[Prometheus<br/>Deployment]
            GRAF[Grafana<br/>Deployment]
        end
    end

    Internet((Internet)) --> LB
    LB --> POD1
    LB --> POD2
    LB --> POD3

    POD1 --> PG
    POD2 --> PG
    POD3 --> PG

    POD1 --> RD
    POD2 --> RD
    POD3 --> RD

    POD1 -.->|Metrics| PROM
    POD2 -.->|Metrics| PROM
    POD3 -.->|Metrics| PROM

    GRAF --> PROM
```

---

## Order Processing Flow

Sequence diagram showing order placement and execution.

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Validator
    participant OrderHandler
    participant TradingEngine
    participant OrderMatcher
    participant TradeExecutor
    participant Repository
    participant MetricsCollector

    Client->>API: POST /api/orders
    API->>Validator: Validate Request
    Validator-->>API: Valid
    API->>OrderHandler: PlaceOrder(request)
    OrderHandler->>TradingEngine: PlaceOrder(order)

    TradingEngine->>TradingEngine: Validate Order
    TradingEngine->>TradingEngine: Generate Order ID
    TradingEngine->>Repository: Save(order)
    Repository-->>TradingEngine: Success

    TradingEngine->>Repository: GetBySymbol(symbol)
    Repository-->>TradingEngine: existing_orders[]

    TradingEngine->>OrderMatcher: FindMatches(order, existing_orders)
    OrderMatcher-->>TradingEngine: matches[]

    loop For each match
        TradingEngine->>TradeExecutor: ExecuteTrade(buy, sell, qty, price)
        TradeExecutor->>Repository: SaveTrade(trade)
        TradeExecutor-->>TradingEngine: trade
        TradingEngine->>Repository: UpdateOrderQuantities()
        TradingEngine->>MetricsCollector: RecordTrade(trade)
    end

    TradingEngine->>MetricsCollector: RecordOrder(order)
    TradingEngine-->>OrderHandler: Success
    OrderHandler-->>API: OrderResponse
    API-->>Client: 200 OK
```

---

## Market Simulation Flow

Shows how market simulation generates and places orders.

```mermaid
sequenceDiagram
    participant App as Application
    participant Sim as Market Simulator
    participant PG as Price Generator
    participant OG as Order Generator
    participant EG as Event Generator
    participant TE as Trading Engine
    participant Metrics as Metrics Collector

    App->>Sim: StartSimulation(config)
    activate Sim

    loop Every Tick Interval (100ms)
        Sim->>PG: GeneratePrice(symbol, currentPrice)
        PG->>PG: Apply Volatility
        PG->>PG: Apply Trend
        PG->>PG: Apply Mean Reversion
        PG-->>Sim: newPrice

        Sim->>EG: GenerateMarketEvent()
        EG-->>Sim: event (optional)

        alt Event Occurred
            Sim->>PG: SimulateVolatility(pattern, intensity)
            Sim->>OG: UpdateMarketSentiment(sentiment)
        end

        Sim->>OG: GenerateRealisticOrders(symbol, price, condition)
        OG->>OG: Simulate User Behavior
        OG->>OG: Calculate Order Sizes
        OG->>OG: Determine Order Types
        OG-->>Sim: orders[]

        loop For each order
            Sim->>TE: PlaceOrder(order)
            TE-->>Sim: Success/Error
            Sim->>Metrics: RecordSimulationMetric()
        end

        Sim->>Sim: Update Simulation Status
    end

    deactivate Sim
```

---

## Demo Load Test Flow

Illustrates the load testing execution flow.

```mermaid
sequenceDiagram
    participant Client
    participant DemoCtrl as Demo Controller
    participant ScenMgr as Scenario Manager
    participant OrderGen as Order Generator
    participant API as Trading API
    participant WS as WebSocket
    participant Metrics as Metrics Collector

    Client->>DemoCtrl: StartLoadTest(scenario)
    DemoCtrl->>DemoCtrl: Validate Scenario
    DemoCtrl->>DemoCtrl: Check Safety Limits
    DemoCtrl->>ScenMgr: ExecuteLoadScenario(scenario)
    activate ScenMgr

    alt Ramp-Up Enabled
        ScenMgr->>ScenMgr: Ramp-Up Phase
        loop Gradual Increase
            ScenMgr->>OrderGen: Generate Orders (increasing rate)
            OrderGen-->>ScenMgr: orders[]
            ScenMgr->>API: Place Orders
            ScenMgr->>Metrics: Collect Metrics
            ScenMgr->>WS: Broadcast Update
            WS-->>Client: Load Test Update
        end
    end

    ScenMgr->>ScenMgr: Sustained Load Phase
    loop Duration or Until Stopped
        ScenMgr->>OrderGen: Generate Orders (target rate)
        OrderGen-->>ScenMgr: orders[]

        par Concurrent Users
            ScenMgr->>API: Place Orders (User 1)
            ScenMgr->>API: Place Orders (User 2)
            ScenMgr->>API: Place Orders (User N)
        end

        ScenMgr->>Metrics: Collect Metrics
        Metrics-->>ScenMgr: metrics_snapshot

        ScenMgr->>WS: Broadcast Update
        WS-->>Client: Real-time Metrics

        alt Error Threshold Exceeded
            ScenMgr->>ScenMgr: Stop Test
        end
    end

    ScenMgr->>ScenMgr: Ramp-Down Phase (if configured)
    ScenMgr->>Metrics: Generate Final Report
    ScenMgr-->>DemoCtrl: Test Complete
    deactivate ScenMgr

    DemoCtrl->>WS: Broadcast Completion
    WS-->>Client: Test Results
```

---

## Chaos Engineering Flow

Shows chaos testing execution with failure injection and recovery.

```mermaid
sequenceDiagram
    participant Client
    participant DemoCtrl as Demo Controller
    participant ChaosEng as Chaos Engine
    participant Target as Target Component
    participant Monitor as System Monitor
    participant Recovery as Recovery Manager

    Client->>DemoCtrl: TriggerChaosTest(scenario)
    DemoCtrl->>DemoCtrl: Validate Scenario
    DemoCtrl->>DemoCtrl: Check Safety Limits
    DemoCtrl->>ChaosEng: ExecuteChaosScenario(scenario)
    activate ChaosEng

    ChaosEng->>ChaosEng: Injection Phase

    alt Latency Injection
        ChaosEng->>Target: Inject Latency (delay_ms)
        Target->>Target: Add artificial delay
    else Error Simulation
        ChaosEng->>Target: Inject Errors (error_rate)
        Target->>Target: Return random errors
    else Resource Exhaustion
        ChaosEng->>Target: Limit Resources (cpu/memory)
        Target->>Target: Apply resource limits
    else Network Partition
        ChaosEng->>Target: Simulate Network Issues
        Target->>Target: Drop packets/connections
    end

    ChaosEng->>Monitor: Start Monitoring
    activate Monitor

    ChaosEng->>ChaosEng: Sustained Phase
    loop Duration
        Monitor->>Target: Check Health
        Target-->>Monitor: Status
        Monitor->>Monitor: Measure Degradation
        Monitor->>Monitor: Calculate Resilience Score
        Monitor->>Client: Broadcast Metrics

        alt Critical Threshold Exceeded
            ChaosEng->>ChaosEng: Trigger Early Recovery
        end
    end

    ChaosEng->>ChaosEng: Recovery Phase
    ChaosEng->>Recovery: InitiateRecovery(graceful?)

    alt Graceful Recovery
        Recovery->>Target: Gradually Remove Chaos
        loop Recovery Steps
            Recovery->>Target: Reduce chaos intensity
            Recovery->>Monitor: Verify stability
        end
    else Immediate Recovery
        Recovery->>Target: Remove All Chaos
    end

    Recovery->>Monitor: Measure Recovery Time
    Monitor-->>ChaosEng: recovery_metrics
    deactivate Monitor

    ChaosEng->>ChaosEng: Generate Report
    ChaosEng-->>DemoCtrl: Test Complete
    deactivate ChaosEng

    DemoCtrl->>Client: Final Results
```

---

## Dependency Graph

Shows the dependency relationships between major components.

```mermaid
graph TD
    subgraph cmd
        Main[cmd/api/main.go]
    end

    subgraph internal/app
        App[Application]
        Container[Container]
    end

    subgraph internal/api
        Server[Server]
        Handlers[Handlers]
        Middleware[Middleware]
    end

    subgraph internal/engine
        TradingEngine[Trading Engine]
    end

    subgraph internal/simulation
        MarketSim[Market Simulator]
        PriceGen[Price Generator]
        OrderGen[Order Generator]
    end

    subgraph internal/demo
        DemoCtrl[Demo Controller]
    end

    subgraph internal/metrics
        MetricsSvc[Metrics Service]
        Collector[Collector]
        Analyzer[Analyzer]
    end

    subgraph internal/config
        Config[Config]
    end

    subgraph internal/domain
        Domain[Domain Models]
    end

    Main --> App
    App --> Container
    App --> Config

    Container --> Server
    Container --> TradingEngine
    Container --> MarketSim
    Container --> MetricsSvc
    Container --> DemoCtrl

    Server --> Handlers
    Server --> Middleware
    Handlers --> TradingEngine
    Handlers --> MetricsSvc

    MarketSim --> PriceGen
    MarketSim --> OrderGen
    MarketSim --> TradingEngine

    DemoCtrl --> MarketSim
    DemoCtrl --> TradingEngine

    MetricsSvc --> Collector
    MetricsSvc --> Analyzer

    TradingEngine --> Domain
    MarketSim --> Domain
    Handlers --> Domain
```

---

## Data Model Diagram

Entity-relationship diagram showing core data models.

```mermaid
erDiagram
    Order ||--o{ Trade : "matched_in"
    Order {
        string ID PK
        string UserID
        string Symbol
        OrderSide Side
        OrderType Type
        float64 Price
        float64 Quantity
        OrderStatus Status
        time Timestamp
    }

    Trade {
        string ID PK
        string BuyOrderID FK
        string SellOrderID FK
        string Symbol
        float64 Quantity
        float64 Price
        time Timestamp
    }

    OrderBook ||--|{ Order : "contains"
    OrderBook {
        string Symbol PK
        Order[] Bids
        Order[] Asks
    }

    Match {
        Order BuyOrder
        Order SellOrder
        float64 Quantity
        float64 Price
    }

    MetricsSnapshot {
        int64 OrderCount
        int64 TradeCount
        float64 TotalVolume
        duration AvgLatency
        float64 OrdersPerSec
        float64 TradesPerSec
        map SymbolMetrics
    }

    SimulationStatus {
        bool IsRunning
        time StartTime
        duration RunningDuration
        int64 OrdersGenerated
        int64 PriceUpdates
        map CurrentPrices
        string[] ActivePatterns
        MarketCondition Condition
    }

    UserProfile {
        string Name
        UserBehaviorPattern Pattern
        float64 RiskTolerance
        OrderSizeRange SizeRange
        float64 TradingFrequency
        duration ReactionTime
        float64 Wealth
    }
```

---

## State Machine Diagrams

### Order Status State Machine

```mermaid
stateDiagram-v2
    [*] --> PENDING: Order Created
    PENDING --> PARTIAL: Partial Match
    PENDING --> FILLED: Full Match
    PENDING --> CANCELLED: User Cancels
    PENDING --> REJECTED: Validation Failed

    PARTIAL --> FILLED: Remaining Matched
    PARTIAL --> CANCELLED: User Cancels

    FILLED --> [*]
    CANCELLED --> [*]
    REJECTED --> [*]
```

### Simulation State Machine

```mermaid
stateDiagram-v2
    [*] --> IDLE: Initialize
    IDLE --> STARTING: StartSimulation()
    STARTING --> RUNNING: Initialization Complete

    RUNNING --> RUNNING: Generate Orders/Prices
    RUNNING --> PAUSED: PauseSimulation()
    RUNNING --> STOPPING: StopSimulation()

    PAUSED --> RUNNING: ResumeSimulation()
    PAUSED --> STOPPING: StopSimulation()

    STOPPING --> STOPPED: Cleanup Complete
    STOPPED --> [*]

    RUNNING --> ERROR: Fatal Error
    ERROR --> STOPPED: Forced Stop
```

### Load Test State Machine

```mermaid
stateDiagram-v2
    [*] --> IDLE: Initialize
    IDLE --> RAMP_UP: Start Load Test

    RAMP_UP --> SUSTAINED: Ramp-Up Complete
    RAMP_UP --> ERROR: Safety Limit Exceeded

    SUSTAINED --> RAMP_DOWN: Duration Complete
    SUSTAINED --> STOPPED: Manual Stop
    SUSTAINED --> ERROR: Safety Limit Exceeded

    RAMP_DOWN --> COMPLETED: Ramp-Down Complete

    ERROR --> [*]: Cleanup
    COMPLETED --> [*]: Cleanup
    STOPPED --> [*]: Cleanup
```

---

## Technology Stack Diagram

Shows the technology layers and their components.

```mermaid
graph TB
    subgraph Application
        GO[Go 1.23+<br/>Runtime]
    end

    subgraph Web Framework
        GIN[Gin<br/>HTTP Router]
        GORILLA[Gorilla<br/>WebSocket]
    end

    subgraph Data Access
        SQLX[sqlx<br/>SQL Toolkit]
        REDIS[go-redis<br/>Redis Client]
    end

    subgraph Utilities
        UUID[google/uuid<br/>ID Generation]
        VALIDATOR[validator<br/>Validation]
        SLOG[slog<br/>Logging]
    end

    subgraph Testing
        TESTIFY[testify<br/>Testing]
        MOCK[go-mock<br/>Mocking]
    end

    subgraph Monitoring
        PROM[prometheus<br/>Metrics]
    end

    subgraph Storage
        PG[(PostgreSQL)]
        REDISDB[(Redis)]
        MEM[(In-Memory)]
    end

    GO --> GIN
    GO --> GORILLA
    GO --> SQLX
    GO --> REDIS
    GO --> UUID
    GO --> VALIDATOR
    GO --> SLOG
    GO --> PROM

    SQLX --> PG
    REDIS --> REDISDB

    GIN -.->|Default| MEM
```

---

## Metrics Collection Architecture

Shows how metrics flow through the system.

```mermaid
graph LR
    subgraph Event Sources
        TE[Trading Engine]
        SIM[Market Simulator]
        API[API Layer]
        DEMO[Demo System]
    end

    subgraph Collection Layer
        COLL[Metrics Collector]
        BUF[Time-Windowed Buffer]
    end

    subgraph Processing Layer
        AGG[Aggregator]
        CALC[Calculator]
        AI[AI Analyzer]
    end

    subgraph Storage Layer
        MEM[(In-Memory Store)]
        CACHE[(Cache)]
    end

    subgraph Exposure Layer
        PROM[Prometheus Metrics]
        REST[REST API]
        WS[WebSocket Stream]
    end

    TE -->|Order Events| COLL
    TE -->|Trade Events| COLL
    SIM -->|Simulation Events| COLL
    API -->|Request Metrics| COLL
    DEMO -->|Test Metrics| COLL

    COLL --> BUF
    BUF --> AGG
    AGG --> CALC
    CALC --> AI

    CALC --> MEM
    CALC --> CACHE
    AI --> MEM

    MEM --> PROM
    MEM --> REST
    MEM --> WS
```

---

## Middleware Pipeline

Shows the HTTP request processing pipeline.

```mermaid
graph TB
    REQ[Incoming Request] --> RECOVERY[Recovery Middleware]
    RECOVERY --> ERROR[Error Handler]
    ERROR --> LOG[Logging Middleware]
    LOG --> SEC[Security Headers]
    SEC --> CORS[CORS Middleware]
    CORS --> CONTENT[Content Type Validation]
    CONTENT --> RATE[Rate Limiting]
    RATE --> VALID[Validation Middleware]
    VALID --> HANDLER[Route Handler]
    HANDLER --> RESP[Response]

    RECOVERY -.->|Panic| ERROR_RESP[500 Error Response]
    ERROR -.->|Error| ERROR_RESP
    RATE -.->|Too Many Requests| ERROR_RESP
    VALID -.->|Invalid| ERROR_RESP
```

---

**Document Version**: 1.0
**Last Updated**: 2025-10-24
**Format**: Mermaid Diagrams
**Compatible With**: GitHub, GitLab, VS Code (with Mermaid plugin)

## Viewing These Diagrams

### GitHub/GitLab
These diagrams will render automatically when viewing this file on GitHub or GitLab.

### VS Code
Install the "Markdown Preview Mermaid Support" extension.

### Standalone
Use the Mermaid Live Editor: https://mermaid.live/

### Export
You can export these diagrams to PNG/SVG using:
- Mermaid CLI
- Mermaid Live Editor
- VS Code extensions
