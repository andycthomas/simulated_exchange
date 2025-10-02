package domain

import (
	"context"
	"log/slog"
	"time"

	"simulated_exchange/pkg/shared"
)

// PriceService implements shared.PriceService interface
type PriceService struct {
	priceGenerator    *PriceGenerator
	marketDataService *MarketDataService
	logger            *slog.Logger
}

// NewPriceService creates a new price service
func NewPriceService(
	priceGenerator *PriceGenerator,
	marketDataService *MarketDataService,
	logger *slog.Logger,
) *PriceService {
	return &PriceService{
		priceGenerator:    priceGenerator,
		marketDataService: marketDataService,
		logger:            logger,
	}
}

// GetCurrentPrice returns the current price for a symbol
func (ps *PriceService) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	return ps.priceGenerator.GetCurrentPrice(symbol)
}

// GetMarketData returns market data for a symbol
func (ps *PriceService) GetMarketData(ctx context.Context, symbol string) (*shared.MarketData, error) {
	return ps.marketDataService.GetMarketData(ctx, symbol)
}

// GetPriceHistory returns price history for a symbol
func (ps *PriceService) GetPriceHistory(ctx context.Context, symbol string, from, to time.Time) ([]shared.PriceUpdate, error) {
	// For now, return recent history from price generator
	// In a production system, this would query historical data from database
	history, err := ps.priceGenerator.GetPriceHistory(symbol, 100)
	if err != nil {
		return nil, err
	}

	// Filter by time range
	var filtered []shared.PriceUpdate
	for _, update := range history {
		if update.Timestamp.After(from) && update.Timestamp.Before(to) {
			filtered = append(filtered, update)
		}
	}

	return filtered, nil
}

// UpdatePrice generates and publishes a new price update
func (ps *PriceService) UpdatePrice(ctx context.Context, update *shared.PriceUpdate) error {
	// Update market data and publish events
	return ps.marketDataService.UpdateMarketData(ctx, update)
}

// GenerateAndPublishPrice generates a new price and publishes it
func (ps *PriceService) GenerateAndPublishPrice(ctx context.Context, symbol string) error {
	// Generate new price
	priceUpdate, err := ps.priceGenerator.GeneratePrice(symbol)
	if err != nil {
		return err
	}

	// Update market data and publish
	return ps.marketDataService.UpdateMarketData(ctx, priceUpdate)
}

// SetInitialPrice sets the initial price for a symbol
func (ps *PriceService) SetInitialPrice(symbol string, price float64) error {
	return ps.priceGenerator.SetBasePrice(symbol, price)
}

// SimulateVolatility applies volatility patterns to price generation
func (ps *PriceService) SimulateVolatility(symbol, pattern string, intensity float64) error {
	return ps.priceGenerator.SimulateVolatility(symbol, pattern, intensity)
}

// GetAllSymbols returns all symbols that have price data
func (ps *PriceService) GetAllSymbols() []string {
	// This would typically come from configuration or database
	return []string{"BTCUSD", "ETHUSD", "ADAUSD"}
}