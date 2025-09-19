package app

import (
	"context"
	"os"
	"testing"
	"time"

	"simulated_exchange/internal/simulation"
)

func TestApplication_NewApplication(t *testing.T) {
	// Test application creation
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	if app == nil {
		t.Fatal("Expected application to be created, got nil")
	}

	if app.config == nil {
		t.Fatal("Expected configuration to be loaded")
	}

	if app.logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	if app.container == nil {
		t.Fatal("Expected container to be initialized")
	}
}

func TestApplication_StartStop(t *testing.T) {
	// Set environment variables for test
	originalEnv := map[string]string{
		"SERVER_PORT":        os.Getenv("SERVER_PORT"),
		"SIMULATION_ENABLED": os.Getenv("SIMULATION_ENABLED"),
		"HEALTH_ENABLED":     os.Getenv("HEALTH_ENABLED"),
	}
	defer func() {
		// Restore original environment
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Set test environment
	os.Setenv("SERVER_PORT", "18080") // Use different port for testing
	os.Setenv("SIMULATION_ENABLED", "false") // Disable simulation for faster tests
	os.Setenv("HEALTH_ENABLED", "true")

	app, err := NewApplication()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Test that application is not running initially
	if app.IsRunning() {
		t.Fatal("Expected application to not be running initially")
	}

	// Start application
	if err := app.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}

	// Test that application is running
	if !app.IsRunning() {
		t.Fatal("Expected application to be running after start")
	}

	// Give some time for components to start
	time.Sleep(100 * time.Millisecond)

	// Test uptime
	uptime := app.GetUptime()
	if uptime <= 0 {
		t.Fatal("Expected positive uptime")
	}

	// Test health status
	health := app.GetHealthStatus()
	if health.Status == "" {
		t.Fatal("Expected health status to be available")
	}

	// Stop application
	if err := app.Stop(); err != nil {
		t.Fatalf("Failed to stop application: %v", err)
	}

	// Test that application is not running after stop
	if app.IsRunning() {
		t.Fatal("Expected application to not be running after stop")
	}
}

func TestApplication_StartStopWithSimulation(t *testing.T) {
	// Set environment variables for test with simulation enabled
	originalEnv := map[string]string{
		"SERVER_PORT":        os.Getenv("SERVER_PORT"),
		"SIMULATION_ENABLED": os.Getenv("SIMULATION_ENABLED"),
		"HEALTH_ENABLED":     os.Getenv("HEALTH_ENABLED"),
	}
	defer func() {
		// Restore original environment
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Set test environment with simulation
	os.Setenv("SERVER_PORT", "18081") // Use different port for testing
	os.Setenv("SIMULATION_ENABLED", "true")
	os.Setenv("HEALTH_ENABLED", "true")

	app, err := NewApplication()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Start application
	if err := app.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}

	// Give some time for simulation to start
	time.Sleep(500 * time.Millisecond)

	// Test simulation status
	status, err := app.GetSimulationStatus()
	if err != nil {
		t.Fatalf("Failed to get simulation status: %v", err)
	}

	if !status.IsRunning {
		t.Fatalf("Expected simulation to be running, got IsRunning: %v", status.IsRunning)
	}

	// Test volatility injection
	err = app.InjectVolatility(simulation.VolatilityPattern("SPIKE"), 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to inject volatility: %v", err)
	}

	// Stop application
	if err := app.Stop(); err != nil {
		t.Fatalf("Failed to stop application: %v", err)
	}
}

func TestApplication_Configuration(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	config := app.GetConfig()
	if config == nil {
		t.Fatal("Expected configuration to be available")
	}

	// Test configuration defaults
	if config.Server.Port == "" {
		t.Fatal("Expected server port to be configured")
	}

	if config.Server.Environment == "" {
		t.Fatal("Expected server environment to be configured")
	}
}

func TestApplication_Container(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	container := app.GetContainer()
	if container == nil {
		t.Fatal("Expected container to be available")
	}

	// Test that services are available
	orderService := container.GetOrderService()
	if orderService == nil {
		t.Fatal("Expected order service to be available")
	}

	metricsService := container.GetMetricsService()
	if metricsService == nil {
		t.Fatal("Expected metrics service to be available")
	}

	healthService := container.GetHealthService()
	if healthService == nil {
		t.Fatal("Expected health service to be available")
	}

	server := container.GetServer()
	if server == nil {
		t.Fatal("Expected server to be available")
	}
}

func TestApplication_HealthCheck(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	healthService := app.GetContainer().GetHealthService()
	if healthService == nil {
		t.Fatal("Expected health service to be available")
	}

	// Test health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := healthService.Check(ctx)
	if status.Status == "" {
		t.Fatal("Expected health status to be returned")
	}

	if status.Timestamp.IsZero() {
		t.Fatal("Expected health status timestamp to be set")
	}
}

func TestApplication_ErrorHandling(t *testing.T) {
	// Test error when trying to start already running application
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Set test port
	originalPort := os.Getenv("SERVER_PORT")
	os.Setenv("SERVER_PORT", "18082")
	defer func() {
		if originalPort == "" {
			os.Unsetenv("SERVER_PORT")
		} else {
			os.Setenv("SERVER_PORT", originalPort)
		}
	}()

	if err := app.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}

	// Try to start again
	err = app.Start()
	if err == nil {
		t.Fatal("Expected error when starting already running application")
	}

	// Stop application
	if err := app.Stop(); err != nil {
		t.Fatalf("Failed to stop application: %v", err)
	}

	// Try to stop again
	err = app.Stop()
	if err == nil {
		t.Fatal("Expected error when stopping already stopped application")
	}
}

func TestApplication_SimulationDisabled(t *testing.T) {
	// Test application behavior when simulation is disabled
	originalEnv := os.Getenv("SIMULATION_ENABLED")
	os.Setenv("SIMULATION_ENABLED", "false")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("SIMULATION_ENABLED")
		} else {
			os.Setenv("SIMULATION_ENABLED", originalEnv)
		}
	}()

	app, err := NewApplication()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Test that simulation methods return errors when disabled
	err = app.InjectVolatility(simulation.VolatilityPattern("SPIKE"), 1*time.Second)
	if err == nil {
		t.Fatal("Expected error when injecting volatility with simulation disabled")
	}

	_, err = app.GetSimulationStatus()
	if err == nil {
		t.Fatal("Expected error when getting simulation status with simulation disabled")
	}
}

// Benchmark tests
func BenchmarkApplication_StartStop(b *testing.B) {
	// Set test environment
	os.Setenv("SERVER_PORT", "18083")
	os.Setenv("SIMULATION_ENABLED", "false")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SIMULATION_ENABLED")
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app, err := NewApplication()
		if err != nil {
			b.Fatalf("Failed to create application: %v", err)
		}

		if err := app.Start(); err != nil {
			b.Fatalf("Failed to start application: %v", err)
		}

		if err := app.Stop(); err != nil {
			b.Fatalf("Failed to stop application: %v", err)
		}
	}
}

func BenchmarkApplication_Creation(b *testing.B) {
	// Set test environment
	os.Setenv("SERVER_PORT", "18084")
	os.Setenv("SIMULATION_ENABLED", "false")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SIMULATION_ENABLED")
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app, err := NewApplication()
		if err != nil {
			b.Fatalf("Failed to create application: %v", err)
		}
		_ = app // Use the app to avoid optimization
	}
}