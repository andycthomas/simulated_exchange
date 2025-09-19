package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"simulated_exchange/internal/api"
	"simulated_exchange/internal/api/dto"
	"simulated_exchange/internal/app"
	"simulated_exchange/internal/config"
	"simulated_exchange/internal/demo"
	"simulated_exchange/test/fixtures"
)

// TestServer provides a complete test server setup
type TestServer struct {
	Server     *httptest.Server
	Container  *app.Container
	Config     *config.Config
	DemoSystem *demo.DemoSystem
	Logger     *slog.Logger
	t          *testing.T
}

// NewTestServer creates a new test server instance
func NewTestServer(t *testing.T) *TestServer {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         8080,
			Environment:  "test",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Metrics: config.MetricsConfig{
			Enabled:        true,
			CollectionTime: time.Second,
			Endpoint:       "/metrics",
		},
		Simulation: config.SimulationConfig{
			Enabled: true,
		},
		Health: config.HealthConfig{
			Enabled:  true,
			Endpoint: "/health",
		},
	}

	// Create container with dependencies
	container, err := app.NewContainer(cfg, logger)
	require.NoError(t, err)

	// Create demo system
	demoSystem, err := demo.CreateDemoSystemWithDefaults(
		container.GetMarketSimulator().(*app.SimulationTradingEngine),
		container.GetOrderService().(*app.MockOrderService),
		container.GetMetricsService().(*app.MockMetricsService),
		logger,
	)
	require.NoError(t, err)

	// Create HTTP server
	server := httptest.NewServer(container.GetServer().GetRouter())

	return &TestServer{
		Server:     server,
		Container:  container,
		Config:     cfg,
		DemoSystem: demoSystem,
		Logger:     logger,
		t:          t,
	}
}

// Close shuts down the test server
func (ts *TestServer) Close() {
	if ts.DemoSystem != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ts.DemoSystem.Stop(ctx)
	}
	if ts.Container != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ts.Container.Shutdown(ctx)
	}
	if ts.Server != nil {
		ts.Server.Close()
	}
}

// GetURL returns the server URL with path
func (ts *TestServer) GetURL(path string) string {
	return ts.Server.URL + path
}

// HTTPClient provides HTTP client utilities for testing
type HTTPClient struct {
	client *http.Client
	t      *testing.T
}

// NewHTTPClient creates a new HTTP client for testing
func NewHTTPClient(t *testing.T) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		t: t,
	}
}

// PlaceOrder places an order via HTTP API
func (hc *HTTPClient) PlaceOrder(baseURL string, order dto.PlaceOrderRequest) (*dto.APIResponse, error) {
	data, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	resp, err := hc.client.Post(baseURL+"/api/orders", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp dto.APIResponse
	err = json.Unmarshal(body, &apiResp)
	return &apiResp, err
}

// GetOrder retrieves an order via HTTP API
func (hc *HTTPClient) GetOrder(baseURL, orderID string) (*dto.APIResponse, error) {
	resp, err := hc.client.Get(baseURL + "/api/orders/" + orderID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp dto.APIResponse
	err = json.Unmarshal(body, &apiResp)
	return &apiResp, err
}

// CancelOrder cancels an order via HTTP API
func (hc *HTTPClient) CancelOrder(baseURL, orderID string) (*dto.APIResponse, error) {
	req, err := http.NewRequest("DELETE", baseURL+"/api/orders/"+orderID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp dto.APIResponse
	err = json.Unmarshal(body, &apiResp)
	return &apiResp, err
}

// GetMetrics retrieves metrics via HTTP API
func (hc *HTTPClient) GetMetrics(baseURL string) (*dto.MetricsResponse, error) {
	resp, err := hc.client.Get(baseURL + "/api/metrics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var metricsResp dto.MetricsResponse
	err = json.Unmarshal(body, &metricsResp)
	return &metricsResp, err
}

// GetHealth checks health via HTTP API
func (hc *HTTPClient) GetHealth(baseURL string) (*dto.HealthResponse, error) {
	resp, err := hc.client.Get(baseURL + "/health")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var healthResp dto.HealthResponse
	err = json.Unmarshal(body, &healthResp)
	return &healthResp, err
}

// WebSocketClient provides WebSocket client utilities for testing
type WebSocketClient struct {
	conn     *websocket.Conn
	messages chan []byte
	errors   chan error
	done     chan bool
	mu       sync.RWMutex
	t        *testing.T
}

// NewWebSocketClient creates a new WebSocket client for testing
func NewWebSocketClient(t *testing.T, url string) (*WebSocketClient, error) {
	// Convert HTTP URL to WebSocket URL
	wsURL := strings.Replace(url, "http://", "ws://", 1) + "/ws/demo"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	client := &WebSocketClient{
		conn:     conn,
		messages: make(chan []byte, 100),
		errors:   make(chan error, 10),
		done:     make(chan bool),
		t:        t,
	}

	// Start message reader
	go client.readMessages()

	return client, nil
}

// readMessages reads messages from WebSocket connection
func (wsc *WebSocketClient) readMessages() {
	defer close(wsc.messages)
	defer close(wsc.errors)

	for {
		select {
		case <-wsc.done:
			return
		default:
			_, message, err := wsc.conn.ReadMessage()
			if err != nil {
				wsc.errors <- err
				return
			}
			wsc.messages <- message
		}
	}
}

// WaitForMessage waits for a WebSocket message with timeout
func (wsc *WebSocketClient) WaitForMessage(timeout time.Duration) ([]byte, error) {
	select {
	case message := <-wsc.messages:
		return message, nil
	case err := <-wsc.errors:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for message")
	}
}

// SendMessage sends a message via WebSocket
func (wsc *WebSocketClient) SendMessage(message []byte) error {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()
	return wsc.conn.WriteMessage(websocket.TextMessage, message)
}

// Close closes the WebSocket connection
func (wsc *WebSocketClient) Close() error {
	close(wsc.done)
	return wsc.conn.Close()
}

// LoadTestRunner provides utilities for running load tests
type LoadTestRunner struct {
	httpClient *HTTPClient
	baseURL    string
	t          *testing.T
}

// NewLoadTestRunner creates a new load test runner
func NewLoadTestRunner(t *testing.T, baseURL string) *LoadTestRunner {
	return &LoadTestRunner{
		httpClient: NewHTTPClient(t),
		baseURL:    baseURL,
		t:          t,
	}
}

// RunConcurrentOrders runs multiple orders concurrently
func (ltr *LoadTestRunner) RunConcurrentOrders(orders []dto.PlaceOrderRequest, concurrency int) (*LoadTestResults, error) {
	results := &LoadTestResults{
		TotalOrders:    len(orders),
		SuccessCount:   0,
		ErrorCount:     0,
		AverageLatency: 0,
		Latencies:      make([]time.Duration, 0, len(orders)),
		Errors:         make([]error, 0),
	}

	// Create semaphore for concurrency control
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	startTime := time.Now()

	for i, order := range orders {
		wg.Add(1)
		go func(orderIndex int, orderReq dto.PlaceOrderRequest) {
			defer wg.Done()
			sem <- struct{}{} // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			orderStart := time.Now()
			resp, err := ltr.httpClient.PlaceOrder(ltr.baseURL, orderReq)
			latency := time.Since(orderStart)

			mu.Lock()
			defer mu.Unlock()

			results.Latencies = append(results.Latencies, latency)

			if err != nil {
				results.ErrorCount++
				results.Errors = append(results.Errors, err)
			} else if resp.Success {
				results.SuccessCount++
			} else {
				results.ErrorCount++
				results.Errors = append(results.Errors, fmt.Errorf("order failed: %v", resp.Error))
			}
		}(i, order)
	}

	wg.Wait()
	results.TotalDuration = time.Since(startTime)

	// Calculate average latency
	if len(results.Latencies) > 0 {
		var totalLatency time.Duration
		for _, latency := range results.Latencies {
			totalLatency += latency
		}
		results.AverageLatency = totalLatency / time.Duration(len(results.Latencies))
	}

	return results, nil
}

// LoadTestResults contains the results of a load test
type LoadTestResults struct {
	TotalOrders     int
	SuccessCount    int
	ErrorCount      int
	TotalDuration   time.Duration
	AverageLatency  time.Duration
	Latencies       []time.Duration
	Errors          []error
}

// GetSuccessRate returns the success rate as a percentage
func (ltr *LoadTestResults) GetSuccessRate() float64 {
	if ltr.TotalOrders == 0 {
		return 0
	}
	return float64(ltr.SuccessCount) / float64(ltr.TotalOrders)
}

// GetThroughput returns orders per second
func (ltr *LoadTestResults) GetThroughput() float64 {
	if ltr.TotalDuration == 0 {
		return 0
	}
	return float64(ltr.TotalOrders) / ltr.TotalDuration.Seconds()
}

// GetLatencyPercentile returns the latency at the given percentile
func (ltr *LoadTestResults) GetLatencyPercentile(percentile float64) time.Duration {
	if len(ltr.Latencies) == 0 {
		return 0
	}

	// Sort latencies
	latencies := make([]time.Duration, len(ltr.Latencies))
	copy(latencies, ltr.Latencies)

	// Simple bubble sort for small arrays
	for i := 0; i < len(latencies)-1; i++ {
		for j := 0; j < len(latencies)-i-1; j++ {
			if latencies[j] > latencies[j+1] {
				latencies[j], latencies[j+1] = latencies[j+1], latencies[j]
			}
		}
	}

	index := int(float64(len(latencies)-1) * percentile / 100.0)
	return latencies[index]
}

// DemoTestRunner provides utilities for testing demo functionality
type DemoTestRunner struct {
	demoSystem *demo.DemoSystem
	wsClient   *WebSocketClient
	t          *testing.T
}

// NewDemoTestRunner creates a new demo test runner
func NewDemoTestRunner(t *testing.T, demoSystem *demo.DemoSystem, wsURL string) (*DemoTestRunner, error) {
	wsClient, err := NewWebSocketClient(t, wsURL)
	if err != nil {
		return nil, err
	}

	return &DemoTestRunner{
		demoSystem: demoSystem,
		wsClient:   wsClient,
		t:          t,
	}, nil
}

// RunLoadTestScenario runs a load test scenario and monitors via WebSocket
func (dtr *DemoTestRunner) RunLoadTestScenario(ctx context.Context, scenario demo.LoadTestScenario) (*DemoTestResults, error) {
	results := &DemoTestResults{
		Scenario:    scenario.Name,
		StartTime:   time.Now(),
		Updates:     make([]demo.DemoUpdate, 0),
		Completed:   false,
	}

	// Start monitoring WebSocket updates
	go dtr.monitorUpdates(results)

	// Start the load test
	err := dtr.demoSystem.Controller.StartLoadTest(ctx, scenario)
	if err != nil {
		return results, err
	}

	// Wait for completion or timeout
	timeout := scenario.Duration + 30*time.Second
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case <-time.After(timeout):
			return results, fmt.Errorf("test timed out after %v", timeout)
		case <-ticker.C:
			status, err := dtr.demoSystem.Controller.GetLoadTestStatus(ctx)
			if err == nil && !status.IsRunning {
				results.EndTime = time.Now()
				results.Completed = true
				results.FinalStatus = status
				return results, nil
			}
		}
	}
}

// monitorUpdates monitors WebSocket updates during demo test
func (dtr *DemoTestRunner) monitorUpdates(results *DemoTestResults) {
	for {
		message, err := dtr.wsClient.WaitForMessage(5 * time.Second)
		if err != nil {
			// Timeout is expected when no updates are available
			continue
		}

		var update demo.DemoUpdate
		if err := json.Unmarshal(message, &update); err == nil {
			results.Updates = append(results.Updates, update)
		}
	}
}

// Close closes the demo test runner
func (dtr *DemoTestRunner) Close() error {
	return dtr.wsClient.Close()
}

// DemoTestResults contains the results of a demo test
type DemoTestResults struct {
	Scenario    string
	StartTime   time.Time
	EndTime     time.Time
	Updates     []demo.DemoUpdate
	Completed   bool
	FinalStatus *demo.LoadTestStatus
}

// GetDuration returns the test duration
func (dtr *DemoTestResults) GetDuration() time.Duration {
	if dtr.EndTime.IsZero() {
		return time.Since(dtr.StartTime)
	}
	return dtr.EndTime.Sub(dtr.StartTime)
}

// GetUpdateCount returns the number of updates received
func (dtr *DemoTestResults) GetUpdateCount() int {
	return len(dtr.Updates)
}

// AssertionHelpers provides common test assertions
type AssertionHelpers struct {
	t *testing.T
}

// NewAssertionHelpers creates new assertion helpers
func NewAssertionHelpers(t *testing.T) *AssertionHelpers {
	return &AssertionHelpers{t: t}
}

// AssertOrderPlacementSuccess asserts successful order placement
func (ah *AssertionHelpers) AssertOrderPlacementSuccess(resp *dto.APIResponse, expectedSymbol string) {
	assert.True(ah.t, resp.Success, "Order placement should succeed")
	assert.Nil(ah.t, resp.Error, "Should not have error")
	assert.NotNil(ah.t, resp.Data, "Should have response data")

	// Check if data contains order information
	if orderData, ok := resp.Data.(map[string]interface{}); ok {
		if symbol, exists := orderData["symbol"]; exists {
			assert.Equal(ah.t, expectedSymbol, symbol, "Order symbol should match")
		}
	}
}

// AssertLoadTestCompletion asserts load test completion
func (ah *AssertionHelpers) AssertLoadTestCompletion(results *DemoTestResults, expectedDuration time.Duration) {
	assert.True(ah.t, results.Completed, "Load test should complete")
	assert.NotNil(ah.t, results.FinalStatus, "Should have final status")
	assert.False(ah.t, results.FinalStatus.IsRunning, "Load test should not be running")

	// Check duration is within reasonable bounds
	actualDuration := results.GetDuration()
	tolerance := 30 * time.Second
	assert.True(ah.t,
		actualDuration >= expectedDuration-tolerance && actualDuration <= expectedDuration+tolerance,
		"Duration should be within tolerance: expected %v, got %v", expectedDuration, actualDuration)
}

// AssertPerformanceTargets asserts performance targets are met
func (ah *AssertionHelpers) AssertPerformanceTargets(results *LoadTestResults, targets fixtures.PerformanceTestConfig) {
	// Check latency
	p99Latency := results.GetLatencyPercentile(99)
	assert.True(ah.t,
		p99Latency <= time.Duration(targets.MaxLatencyMs)*time.Millisecond,
		"P99 latency should be under %dms, got %v", targets.MaxLatencyMs, p99Latency)

	// Check throughput
	throughput := results.GetThroughput()
	assert.True(ah.t,
		throughput >= targets.MinThroughput,
		"Throughput should be at least %.2f ops/s, got %.2f", targets.MinThroughput, throughput)

	// Check error rate
	errorRate := 1.0 - results.GetSuccessRate()
	assert.True(ah.t,
		errorRate <= targets.MaxErrorRate,
		"Error rate should be under %.2f%%, got %.2f%%", targets.MaxErrorRate*100, errorRate*100)
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			if condition() {
				return
			}
		case <-timeoutChan:
			require.Fail(t, "Timeout waiting for condition: "+message)
			return
		}
	}
}

// RandomOrder generates a random order for testing
func RandomOrder(symbols []string) dto.PlaceOrderRequest {
	symbol := symbols[time.Now().UnixNano()%int64(len(symbols))]
	side := "buy"
	if time.Now().UnixNano()%2 == 0 {
		side = "sell"
	}
	orderType := "limit"
	if time.Now().UnixNano()%3 == 0 {
		orderType = "market"
	}

	return dto.PlaceOrderRequest{
		Symbol:   symbol,
		Side:     side,
		Type:     orderType,
		Quantity: float64(time.Now().UnixNano()%1000 + 1) / 10.0,
		Price:    float64(time.Now().UnixNano()%50000 + 10000) / 100.0,
	}
}