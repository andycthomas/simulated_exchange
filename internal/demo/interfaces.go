package demo

import (
	"context"
	"time"

	"simulated_exchange/internal/api/dto"
)

// DemoController interface defines the contract for demo control operations
type DemoController interface {
	// Load testing operations
	StartLoadTest(ctx context.Context, scenario LoadTestScenario) error
	StopLoadTest(ctx context.Context) error
	GetLoadTestStatus(ctx context.Context) (*LoadTestStatus, error)

	// Chaos testing operations
	TriggerChaosTest(ctx context.Context, scenario ChaosTestScenario) error
	StopChaosTest(ctx context.Context) error
	GetChaosTestStatus(ctx context.Context) (*ChaosTestStatus, error)

	// System management
	ResetSystem(ctx context.Context) error
	GetSystemStatus(ctx context.Context) (*DemoSystemStatus, error)

	// Real-time updates
	Subscribe(subscriber DemoSubscriber) error
	Unsubscribe(subscriberID string) error
	BroadcastUpdate(update DemoUpdate) error
}

// ScenarioManager interface for managing demo scenarios
type ScenarioManager interface {
	// Load test scenarios
	ExecuteLoadScenario(ctx context.Context, scenario LoadTestScenario) error
	StopLoadScenario(ctx context.Context) error

	// Chaos test scenarios
	ExecuteChaosScenario(ctx context.Context, scenario ChaosTestScenario) error
	StopChaosScenario(ctx context.Context) error

	// Scenario configuration
	GetAvailableLoadScenarios() []LoadTestScenario
	GetAvailableChaosScenarios() []ChaosTestScenario
}

// DemoSubscriber interface for real-time updates
type DemoSubscriber interface {
	GetID() string
	SendUpdate(update DemoUpdate) error
	IsActive() bool
	Close() error
}

// LoadTestScenario defines load testing parameters
type LoadTestScenario struct {
	Name                string             `json:"name"`
	Description         string             `json:"description"`
	Intensity           LoadIntensity      `json:"intensity"`
	Duration            time.Duration      `json:"duration"`
	OrdersPerSecond     int                `json:"orders_per_second"`
	ConcurrentUsers     int                `json:"concurrent_users"`
	Symbols             []string           `json:"symbols"`
	OrderTypes          []string           `json:"order_types"`
	PriceVariation      float64            `json:"price_variation"`
	VolumeRange         VolumeRange        `json:"volume_range"`
	UserBehaviorPattern UserBehaviorPattern `json:"user_behavior_pattern"`
	RampUp              RampUpConfig       `json:"ramp_up"`
	Metrics             MetricsConfig      `json:"metrics"`
}

// ChaosTestScenario defines chaos engineering parameters
type ChaosTestScenario struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        ChaosType     `json:"type"`
	Duration    time.Duration `json:"duration"`
	Severity    ChaosSeverity `json:"severity"`
	Target      ChaosTarget   `json:"target"`
	Parameters  ChaosParams   `json:"parameters"`
	Recovery    RecoveryConfig `json:"recovery"`
}

// Load testing enums and types
type LoadIntensity string

const (
	LoadLight  LoadIntensity = "light"
	LoadMedium LoadIntensity = "medium"
	LoadHeavy  LoadIntensity = "heavy"
	LoadStress LoadIntensity = "stress"
)

type VolumeRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type UserBehaviorPattern struct {
	BuyRatio         float64 `json:"buy_ratio"`
	SellRatio        float64 `json:"sell_ratio"`
	MarketOrderRatio float64 `json:"market_order_ratio"`
	LimitOrderRatio  float64 `json:"limit_order_ratio"`
	CancelRatio      float64 `json:"cancel_ratio"`
}

type RampUpConfig struct {
	Enabled      bool          `json:"enabled"`
	Duration     time.Duration `json:"duration"`
	StartPercent float64       `json:"start_percent"`
	EndPercent   float64       `json:"end_percent"`
}

type MetricsConfig struct {
	CollectLatency     bool `json:"collect_latency"`
	CollectThroughput  bool `json:"collect_throughput"`
	CollectErrorRate   bool `json:"collect_error_rate"`
	CollectResourceUse bool `json:"collect_resource_use"`
	SampleRate         int  `json:"sample_rate"`
}

// Chaos testing enums and types
type ChaosType string

const (
	ChaosLatencyInjection   ChaosType = "latency_injection"
	ChaosErrorSimulation    ChaosType = "error_simulation"
	ChaosResourceExhaustion ChaosType = "resource_exhaustion"
	ChaosNetworkPartition   ChaosType = "network_partition"
	ChaosServiceFailure     ChaosType = "service_failure"
)

type ChaosSeverity string

const (
	ChaosLow      ChaosSeverity = "low"
	ChaosMedium   ChaosSeverity = "medium"
	ChaosHigh     ChaosSeverity = "high"
	ChaosCritical ChaosSeverity = "critical"
)

type ChaosTarget struct {
	Component string   `json:"component"`
	Services  []string `json:"services"`
	Endpoints []string `json:"endpoints"`
	Percentage float64 `json:"percentage"`
}

type ChaosParams struct {
	LatencyMs        int     `json:"latency_ms,omitempty"`
	ErrorRate        float64 `json:"error_rate,omitempty"`
	MemoryLimitMB    int     `json:"memory_limit_mb,omitempty"`
	CPULimitPercent  float64 `json:"cpu_limit_percent,omitempty"`
	NetworkDelayMs   int     `json:"network_delay_ms,omitempty"`
	PacketLossPercent float64 `json:"packet_loss_percent,omitempty"`
}

type RecoveryConfig struct {
	AutoRecover     bool          `json:"auto_recover"`
	RecoveryTime    time.Duration `json:"recovery_time"`
	GracefulRecover bool          `json:"graceful_recover"`
}

// Status tracking structures
type LoadTestStatus struct {
	IsRunning           bool                   `json:"is_running"`
	Scenario            *LoadTestScenario      `json:"scenario,omitempty"`
	StartTime           time.Time              `json:"start_time,omitempty"`
	ElapsedTime         time.Duration          `json:"elapsed_time"`
	RemainingTime       time.Duration          `json:"remaining_time"`
	Progress            float64                `json:"progress"`
	CurrentMetrics      *LoadTestMetrics       `json:"current_metrics,omitempty"`
	HistoricalMetrics   []LoadTestMetrics      `json:"historical_metrics"`
	ActiveOrders        int                    `json:"active_orders"`
	CompletedOrders     int                    `json:"completed_orders"`
	FailedOrders        int                    `json:"failed_orders"`
	CurrentUsers        int                    `json:"current_users"`
	Phase               LoadTestPhase          `json:"phase"`
	Errors              []LoadTestError        `json:"errors"`
}

type ChaosTestStatus struct {
	IsRunning       bool               `json:"is_running"`
	Scenario        *ChaosTestScenario `json:"scenario,omitempty"`
	StartTime       time.Time          `json:"start_time,omitempty"`
	ElapsedTime     time.Duration      `json:"elapsed_time"`
	RemainingTime   time.Duration      `json:"remaining_time"`
	Progress        float64            `json:"progress"`
	AffectedTargets []string           `json:"affected_targets"`
	Metrics         *ChaosTestMetrics  `json:"metrics,omitempty"`
	Phase           ChaosTestPhase     `json:"phase"`
	Errors          []ChaosTestError   `json:"errors"`
}

type DemoSystemStatus struct {
	Timestamp       time.Time            `json:"timestamp"`
	Overall         SystemHealth         `json:"overall"`
	TradingEngine   ComponentHealth      `json:"trading_engine"`
	OrderService    ComponentHealth      `json:"order_service"`
	MetricsService  ComponentHealth      `json:"metrics_service"`
	Database        ComponentHealth      `json:"database"`
	ActiveScenarios []string             `json:"active_scenarios"`
	SystemMetrics   SystemMetrics        `json:"system_metrics"`
	Alerts          []SystemAlert        `json:"alerts"`
}

// Metrics structures
type LoadTestMetrics struct {
	Timestamp          time.Time `json:"timestamp"`
	OrdersPerSecond    float64   `json:"orders_per_second"`
	AverageLatency     float64   `json:"average_latency"`
	P95Latency         float64   `json:"p95_latency"`
	P99Latency         float64   `json:"p99_latency"`
	ErrorRate          float64   `json:"error_rate"`
	Throughput         float64   `json:"throughput"`
	ActiveConnections  int       `json:"active_connections"`
	MemoryUsageMB      float64   `json:"memory_usage_mb"`
	CPUUsagePercent    float64   `json:"cpu_usage_percent"`
	DatabaseConnections int      `json:"database_connections"`
	QueueDepth         int       `json:"queue_depth"`
}

type ChaosTestMetrics struct {
	Timestamp           time.Time `json:"timestamp"`
	TargetsAffected     int       `json:"targets_affected"`
	SuccessfulInjections int      `json:"successful_injections"`
	FailedInjections    int       `json:"failed_injections"`
	SystemRecoveryTime  float64   `json:"system_recovery_time"`
	ErrorsGenerated     int       `json:"errors_generated"`
	ServiceDegradation  float64   `json:"service_degradation"`
	ResilienceScore     float64   `json:"resilience_score"`
}

type SystemMetrics struct {
	CPU           float64 `json:"cpu"`
	Memory        float64 `json:"memory"`
	DiskIO        float64 `json:"disk_io"`
	NetworkIO     float64 `json:"network_io"`
	Goroutines    int     `json:"goroutines"`
	HeapSize      int64   `json:"heap_size"`
	GCPauses      float64 `json:"gc_pauses"`
}

// Enums for status tracking
type LoadTestPhase string

const (
	LoadPhaseRampUp    LoadTestPhase = "ramp_up"
	LoadPhaseSustained LoadTestPhase = "sustained"
	LoadPhaseRampDown  LoadTestPhase = "ramp_down"
	LoadPhaseCompleted LoadTestPhase = "completed"
	LoadPhaseError     LoadTestPhase = "error"
)

type ChaosTestPhase string

const (
	ChaosPhaseInjection ChaosTestPhase = "injection"
	ChaosPhaseSustained ChaosTestPhase = "sustained"
	ChaosPhaseRecovery  ChaosTestPhase = "recovery"
	ChaosPhaseCompleted ChaosTestPhase = "completed"
	ChaosPhaseError     ChaosTestPhase = "error"
)

type SystemHealth string

const (
	HealthHealthy   SystemHealth = "healthy"
	HealthDegraded  SystemHealth = "degraded"
	HealthUnhealthy SystemHealth = "unhealthy"
	HealthCritical  SystemHealth = "critical"
)

type ComponentHealth struct {
	Status      SystemHealth `json:"status"`
	LastCheck   time.Time    `json:"last_check"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorRate   float64      `json:"error_rate"`
	Message     string       `json:"message,omitempty"`
}

// Error structures
type LoadTestError struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	OrderID   string    `json:"order_id,omitempty"`
	Symbol    string    `json:"symbol,omitempty"`
	Severity  string    `json:"severity"`
}

type ChaosTestError struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Target    string    `json:"target,omitempty"`
	Severity  string    `json:"severity"`
}

type SystemAlert struct {
	ID        string       `json:"id"`
	Timestamp time.Time    `json:"timestamp"`
	Level     AlertLevel   `json:"level"`
	Component string       `json:"component"`
	Message   string       `json:"message"`
	Resolved  bool         `json:"resolved"`
}

type AlertLevel string

const (
	AlertInfo     AlertLevel = "info"
	AlertWarning  AlertLevel = "warning"
	AlertError    AlertLevel = "error"
	AlertCritical AlertLevel = "critical"
)

// Real-time update structures
type DemoUpdate struct {
	Type      UpdateType  `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	Source    string      `json:"source"`
}

type UpdateType string

const (
	UpdateLoadTestStatus    UpdateType = "load_test_status"
	UpdateChaosTestStatus   UpdateType = "chaos_test_status"
	UpdateSystemStatus      UpdateType = "system_status"
	UpdateMetrics          UpdateType = "metrics"
	UpdateError            UpdateType = "error"
	UpdateAlert            UpdateType = "alert"
	UpdateScenarioComplete UpdateType = "scenario_complete"
)

// WebSocket message structures
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type WebSocketSubscription struct {
	SubscriberID string   `json:"subscriber_id"`
	UpdateTypes  []string `json:"update_types"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
}

// Trading engine integration structures
type TradingEngineIntegration interface {
	PlaceOrder(order dto.PlaceOrderRequest) (dto.OrderResponse, error)
	CancelOrder(orderID string) error
	GetOrderStatus(orderID string) (dto.OrderResponse, error)
	GetMetrics() (interface{}, error)
	Reset() error
}

// Configuration structures
type DemoConfig struct {
	LoadTest LoadTestConfig `json:"load_test"`
	Chaos    ChaosConfig    `json:"chaos"`
	WebSocket WebSocketConfig `json:"websocket"`
	Metrics  DemoMetricsConfig `json:"metrics"`
}

type LoadTestConfig struct {
	MaxConcurrentTests int           `json:"max_concurrent_tests"`
	DefaultTimeout     time.Duration `json:"default_timeout"`
	MetricsInterval    time.Duration `json:"metrics_interval"`
	MaxDuration        time.Duration `json:"max_duration"`
	EnabledScenarios   []string      `json:"enabled_scenarios"`
}

type ChaosConfig struct {
	MaxConcurrentTests int           `json:"max_concurrent_tests"`
	DefaultTimeout     time.Duration `json:"default_timeout"`
	SafetyLimits       SafetyLimits  `json:"safety_limits"`
	EnabledTypes       []string      `json:"enabled_types"`
}

type WebSocketConfig struct {
	MaxConnections    int           `json:"max_connections"`
	PingInterval      time.Duration `json:"ping_interval"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	MaxMessageSize    int64         `json:"max_message_size"`
}

type DemoMetricsConfig struct {
	CollectionInterval time.Duration `json:"collection_interval"`
	RetentionPeriod    time.Duration `json:"retention_period"`
	EnabledMetrics     []string      `json:"enabled_metrics"`
	ExportFormats      []string      `json:"export_formats"`
}

type SafetyLimits struct {
	MaxLatencyMs      int     `json:"max_latency_ms"`
	MaxErrorRate      float64 `json:"max_error_rate"`
	MaxCPUUsage       float64 `json:"max_cpu_usage"`
	MaxMemoryUsage    float64 `json:"max_memory_usage"`
	RequireConfirmation bool  `json:"require_confirmation"`
}