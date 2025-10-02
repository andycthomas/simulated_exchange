package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"simulated_exchange/services/trading-api/internal/app"
)

func main() {
	// Create application instance
	application, err := app.NewApplication()
	if err != nil {
		log.Fatalf("Failed to create trading API application: %v", err)
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start application
	if err := application.Start(ctx); err != nil {
		log.Fatalf("Failed to start trading API: %v", err)
	}

	// Wait for shutdown signal
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
	case <-ctx.Done():
		log.Println("Context cancelled, initiating graceful shutdown...")
	}

	// Stop application gracefully
	if err := application.Stop(); err != nil {
		log.Printf("Error during trading API shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("Trading API shutdown completed successfully")
}