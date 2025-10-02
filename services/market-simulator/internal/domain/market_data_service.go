package domain

import (
	"context"
	"log/slog"
	"time"

	"simulated_exchange/pkg/cache"
	"simulated_exchange/pkg/messaging"
	"simulated_exchange/pkg/shared"
)

// MarketDataService manages market data and publishes updates
type MarketDataService struct {
	cache    *cache.RedisClient
	eventBus *messaging.RedisEventBus
	logger   *slog.Logger
}

// NewMarketDataService creates a new market data service
func NewMarketDataService(cache *cache.RedisClient, eventBus *messaging.RedisEventBus, logger *slog.Logger) *MarketDataService {
	return &MarketDataService{
		cache:    cache,
		eventBus: eventBus,
		logger:   logger,
	}
}

// UpdateMarketData updates and publishes market data for a symbol
func (mds *MarketDataService) UpdateMarketData(ctx context.Context, priceUpdate *shared.PriceUpdate) error {
	// Get previous market data from cache
	previousData, err := mds.cache.GetMarketData(ctx, priceUpdate.Symbol)
	if err != nil {
		// If no previous data, create initial data
		previousData = &shared.MarketData{
			Symbol:           priceUpdate.Symbol,
			CurrentPrice:     priceUpdate.Price,
			PreviousPrice:    priceUpdate.Price,
			DailyHigh:        priceUpdate.Price,
			DailyLow:         priceUpdate.Price,
			DailyVolume:      0,
			PriceChange:      0,
			PriceChangePerc:  0,
			Timestamp:        priceUpdate.Timestamp,
		}
	}

	// Create new market data
	newData := &shared.MarketData{
		Symbol:           priceUpdate.Symbol,
		CurrentPrice:     priceUpdate.Price,
		PreviousPrice:    previousData.CurrentPrice,
		DailyHigh:        max(previousData.DailyHigh, priceUpdate.Price),
		DailyLow:         min(previousData.DailyLow, priceUpdate.Price),
		DailyVolume:      previousData.DailyVolume + priceUpdate.Volume,
		Timestamp:        priceUpdate.Timestamp,
	}

	// Calculate price change
	newData.PriceChange = newData.CurrentPrice - newData.PreviousPrice
	if newData.PreviousPrice > 0 {
		newData.PriceChangePerc = (newData.PriceChange / newData.PreviousPrice) * 100
	}

	// Check if it's a new trading day
	if mds.isNewTradingDay(previousData.Timestamp, priceUpdate.Timestamp) {
		// Reset daily values
		newData.DailyHigh = priceUpdate.Price
		newData.DailyLow = priceUpdate.Price
		newData.DailyVolume = priceUpdate.Volume
		newData.PriceChange = 0
		newData.PriceChangePerc = 0
	}

	// Cache the market data
	if err := mds.cache.SetMarketData(ctx, priceUpdate.Symbol, newData); err != nil {
		mds.logger.Warn("Failed to cache market data", "error", err, "symbol", priceUpdate.Symbol)
	}

	// Publish price update event
	if err := mds.eventBus.PublishPriceUpdate(ctx, priceUpdate); err != nil {
		mds.logger.Warn("Failed to publish price update event", "error", err)
	}

	// Publish market data event
	if err := mds.eventBus.PublishMarketData(ctx, newData); err != nil {
		mds.logger.Warn("Failed to publish market data event", "error", err)
	}

	mds.logger.Debug("Market data updated",
		"symbol", priceUpdate.Symbol,
		"price", priceUpdate.Price,
		"change", newData.PriceChange,
		"change_pct", newData.PriceChangePerc,
	)

	return nil
}

// GetMarketData retrieves current market data for a symbol
func (mds *MarketDataService) GetMarketData(ctx context.Context, symbol string) (*shared.MarketData, error) {
	return mds.cache.GetMarketData(ctx, symbol)
}

// GetCurrentPrice retrieves the current price for a symbol
func (mds *MarketDataService) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	marketData, err := mds.cache.GetMarketData(ctx, symbol)
	if err != nil {
		return 0, err
	}
	return marketData.CurrentPrice, nil
}

// GetAllMarketData retrieves market data for all available symbols
func (mds *MarketDataService) GetAllMarketData(ctx context.Context, symbols []string) (map[string]*shared.MarketData, error) {
	result := make(map[string]*shared.MarketData)

	for _, symbol := range symbols {
		marketData, err := mds.cache.GetMarketData(ctx, symbol)
		if err != nil {
			mds.logger.Warn("Failed to get market data for symbol", "symbol", symbol, "error", err)
			continue
		}
		result[symbol] = marketData
	}

	return result, nil
}

// PublishSystemStatus publishes overall market system status
func (mds *MarketDataService) PublishSystemStatus(ctx context.Context, status map[string]interface{}) error {
	event := &shared.Event{
		Type:   "system.status",
		Source: "market-simulator",
		Data:   status,
	}

	return mds.eventBus.Publish(ctx, event)
}

// Helper functions

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func (mds *MarketDataService) isNewTradingDay(previous, current time.Time) bool {
	// Simple check: different day
	return previous.Day() != current.Day()
}