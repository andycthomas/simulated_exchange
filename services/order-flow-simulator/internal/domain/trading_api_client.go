package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"simulated_exchange/pkg/shared"
)

// TradingAPIClient handles communication with the Trading API service
type TradingAPIClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

// APIResponse represents the standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents error information in API responses
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// OrderSubmissionRequest represents the request for submitting an order
type OrderSubmissionRequest struct {
	UserID   string             `json:"user_id"`
	Symbol   string             `json:"symbol"`
	Type     shared.OrderType   `json:"type"`
	Side     shared.OrderSide   `json:"side"`
	Quantity float64            `json:"quantity"`
	Price    float64            `json:"price,omitempty"`
}

// OrderResponse represents the response when submitting an order
type OrderResponse struct {
	OrderID   string               `json:"order_id"`
	Status    shared.OrderStatus   `json:"status"`
	Message   string               `json:"message,omitempty"`
}

// NewTradingAPIClient creates a new trading API client
func NewTradingAPIClient(baseURL string, logger *slog.Logger) *TradingAPIClient {
	return &TradingAPIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
			},
		},
		logger: logger,
	}
}

// SubmitOrder submits an order to the trading API
func (c *TradingAPIClient) SubmitOrder(ctx context.Context, order *shared.Order) error {
	// Convert order to API request format
	request := OrderSubmissionRequest{
		UserID:   order.UserID,
		Symbol:   order.Symbol,
		Type:     order.Type,
		Side:     order.Side,
		Quantity: order.Quantity,
		Price:    order.Price,
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal order request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/orders", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "order-flow-simulator/1.0")

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var apiResponse APIResponse
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	// Check for API errors
	if !apiResponse.Success {
		if apiResponse.Error != nil {
			return fmt.Errorf("API error: %s - %s", apiResponse.Error.Code, apiResponse.Error.Message)
		}
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Log successful submission
	c.logger.Debug("Order submitted successfully",
		"order_id", order.ID,
		"symbol", order.Symbol,
		"type", order.Type,
		"side", order.Side,
		"quantity", order.Quantity,
		"price", order.Price,
		"status_code", resp.StatusCode,
	)

	return nil
}

// GetOrderStatus retrieves the status of an order
func (c *TradingAPIClient) GetOrderStatus(ctx context.Context, orderID string) (*shared.Order, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("User-Agent", "order-flow-simulator/1.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse APIResponse
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if !apiResponse.Success {
		if apiResponse.Error != nil {
			return nil, fmt.Errorf("API error: %s - %s", apiResponse.Error.Code, apiResponse.Error.Message)
		}
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Parse order data
	orderData, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order data: %w", err)
	}

	var order shared.Order
	if err := json.Unmarshal(orderData, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}

// CancelOrder cancels an existing order
func (c *TradingAPIClient) CancelOrder(ctx context.Context, orderID string) error {
	url := fmt.Sprintf("%s/api/orders/%s/cancel", c.baseURL, orderID)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("User-Agent", "order-flow-simulator/1.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse APIResponse
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	if !apiResponse.Success {
		if apiResponse.Error != nil {
			return fmt.Errorf("API error: %s - %s", apiResponse.Error.Code, apiResponse.Error.Message)
		}
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	c.logger.Debug("Order cancelled successfully", "order_id", orderID)
	return nil
}

// GetMarketData retrieves current market data for a symbol
func (c *TradingAPIClient) GetMarketData(ctx context.Context, symbol string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/market/%s", c.baseURL, symbol)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("User-Agent", "order-flow-simulator/1.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse APIResponse
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if !apiResponse.Success {
		if apiResponse.Error != nil {
			return nil, fmt.Errorf("API error: %s - %s", apiResponse.Error.Code, apiResponse.Error.Message)
		}
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Convert data to map
	data, ok := apiResponse.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid market data format")
	}

	return data, nil
}

// HealthCheck performs a health check against the trading API
func (c *TradingAPIClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("User-Agent", "order-flow-simulator/1.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// Close cleans up resources used by the client
func (c *TradingAPIClient) Close() {
	// Close idle connections
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}