package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("HEALTH_CHECK_PORT")
	if port == "" {
		port = "8080"
	}

	url := fmt.Sprintf("http://localhost:%s/health", port)

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Health check returned status: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	os.Exit(0)
}
