package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"simulated_exchange/pkg/shared"
)

// RedisEventBus implements shared.EventBus using Redis Pub/Sub
type RedisEventBus struct {
	client      *redis.Client
	subscribers map[shared.EventType][]shared.EventHandler
	pubsub      *redis.PubSub
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewRedisEventBus creates a new Redis-based event bus
func NewRedisEventBus(client *redis.Client) *RedisEventBus {
	ctx, cancel := context.WithCancel(context.Background())

	return &RedisEventBus{
		client:      client,
		subscribers: make(map[shared.EventType][]shared.EventHandler),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Publish publishes an event to Redis
func (r *RedisEventBus) Publish(ctx context.Context, event *shared.Event) error {
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	channel := string(event.Type)
	err = r.client.Publish(ctx, channel, data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish event to channel %s: %w", channel, err)
	}

	log.Printf("Published event %s to channel %s", event.ID, channel)
	return nil
}

// Subscribe subscribes to an event type with a handler
func (r *RedisEventBus) Subscribe(ctx context.Context, eventType shared.EventType, handler shared.EventHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Add handler to local subscribers
	r.subscribers[eventType] = append(r.subscribers[eventType], handler)

	// If this is the first subscriber for this event type, start listening
	if len(r.subscribers[eventType]) == 1 {
		channel := string(eventType)

		// Create or update pubsub subscription
		if r.pubsub == nil {
			r.pubsub = r.client.Subscribe(ctx, channel)
		} else {
			err := r.pubsub.Subscribe(ctx, channel)
			if err != nil {
				return fmt.Errorf("failed to subscribe to channel %s: %w", channel, err)
			}
		}

		// Start listening goroutine if not already started
		if len(r.subscribers) == 1 {
			r.wg.Add(1)
			go r.listenForMessages()
		}

		log.Printf("Subscribed to event type %s", eventType)
	}

	return nil
}

// Unsubscribe unsubscribes from an event type
func (r *RedisEventBus) Unsubscribe(ctx context.Context, eventType shared.EventType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.subscribers, eventType)

	if r.pubsub != nil {
		channel := string(eventType)
		err := r.pubsub.Unsubscribe(ctx, channel)
		if err != nil {
			return fmt.Errorf("failed to unsubscribe from channel %s: %w", channel, err)
		}
		log.Printf("Unsubscribed from event type %s", eventType)
	}

	return nil
}

// Close closes the event bus and all connections
func (r *RedisEventBus) Close() error {
	r.cancel()

	if r.pubsub != nil {
		err := r.pubsub.Close()
		if err != nil {
			log.Printf("Error closing pubsub: %v", err)
		}
	}

	r.wg.Wait()
	log.Println("Event bus closed")
	return nil
}

// listenForMessages listens for incoming messages from Redis
func (r *RedisEventBus) listenForMessages() {
	defer r.wg.Done()

	ch := r.pubsub.Channel()

	for {
		select {
		case <-r.ctx.Done():
			log.Println("Stopping event bus listener")
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}

			// Parse the event
			var event shared.Event
			err := json.Unmarshal([]byte(msg.Payload), &event)
			if err != nil {
				log.Printf("Failed to unmarshal event from channel %s: %v", msg.Channel, err)
				continue
			}

			// Get event type from channel name
			eventType := shared.EventType(msg.Channel)

			// Handle the event
			r.handleEvent(eventType, &event)
		}
	}
}

// handleEvent processes an incoming event by calling all registered handlers
func (r *RedisEventBus) handleEvent(eventType shared.EventType, event *shared.Event) {
	r.mu.RLock()
	handlers := r.subscribers[eventType]
	r.mu.RUnlock()

	if len(handlers) == 0 {
		log.Printf("No handlers registered for event type %s", eventType)
		return
	}

	log.Printf("Handling event %s of type %s with %d handlers", event.ID, eventType, len(handlers))

	// Call each handler in a separate goroutine
	for _, handler := range handlers {
		go func(h shared.EventHandler) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err := h(ctx, event)
			if err != nil {
				log.Printf("Error handling event %s: %v", event.ID, err)
			}
		}(handler)
	}
}

// PublishOrderPlaced publishes an order placed event
func (r *RedisEventBus) PublishOrderPlaced(ctx context.Context, order *shared.Order) error {
	event := &shared.Event{
		Type:   shared.EventTypeOrderPlaced,
		Source: "trading-api",
		Data: map[string]interface{}{
			"order_id": order.ID,
			"user_id":  order.UserID,
			"symbol":   order.Symbol,
			"side":     order.Side,
			"type":     order.Type,
			"price":    order.Price,
			"quantity": order.Quantity,
		},
	}
	return r.Publish(ctx, event)
}

// PublishOrderCancelled publishes an order cancelled event
func (r *RedisEventBus) PublishOrderCancelled(ctx context.Context, orderID string, userID string) error {
	event := &shared.Event{
		Type:   shared.EventTypeOrderCancelled,
		Source: "trading-api",
		Data: map[string]interface{}{
			"order_id": orderID,
			"user_id":  userID,
		},
	}
	return r.Publish(ctx, event)
}

// PublishTradeExecuted publishes a trade executed event
func (r *RedisEventBus) PublishTradeExecuted(ctx context.Context, trade *shared.Trade) error {
	event := &shared.Event{
		Type:   shared.EventTypeTradeExecuted,
		Source: "trading-api",
		Data: map[string]interface{}{
			"trade_id":      trade.ID,
			"buy_order_id":  trade.BuyOrderID,
			"sell_order_id": trade.SellOrderID,
			"symbol":        trade.Symbol,
			"price":         trade.Price,
			"quantity":      trade.Quantity,
		},
	}
	return r.Publish(ctx, event)
}

// PublishPriceUpdate publishes a price update event
func (r *RedisEventBus) PublishPriceUpdate(ctx context.Context, update *shared.PriceUpdate) error {
	event := &shared.Event{
		Type:   shared.EventTypePriceUpdate,
		Source: "market-simulator",
		Data: map[string]interface{}{
			"symbol":    update.Symbol,
			"price":     update.Price,
			"volume":    update.Volume,
			"timestamp": update.Timestamp,
		},
	}
	return r.Publish(ctx, event)
}

// PublishMarketData publishes market data event
func (r *RedisEventBus) PublishMarketData(ctx context.Context, data *shared.MarketData) error {
	event := &shared.Event{
		Type:   shared.EventTypeMarketData,
		Source: "market-simulator",
		Data: map[string]interface{}{
			"symbol":              data.Symbol,
			"current_price":       data.CurrentPrice,
			"previous_price":      data.PreviousPrice,
			"daily_high":          data.DailyHigh,
			"daily_low":           data.DailyLow,
			"daily_volume":        data.DailyVolume,
			"price_change":        data.PriceChange,
			"price_change_percent": data.PriceChangePerc,
			"timestamp":           data.Timestamp,
		},
	}
	return r.Publish(ctx, event)
}

// Helper function to generate unique event IDs
func generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}

// EventBusHealthChecker implements the shared.HealthChecker interface for Redis EventBus
type EventBusHealthChecker struct {
	eventBus *RedisEventBus
}

// NewEventBusHealthChecker creates a new health checker for the event bus
func NewEventBusHealthChecker(eventBus *RedisEventBus) *EventBusHealthChecker {
	return &EventBusHealthChecker{eventBus: eventBus}
}

// Check performs a health check on the event bus
func (h *EventBusHealthChecker) Check(ctx context.Context) error {
	// Test by pinging Redis
	return h.eventBus.client.Ping(ctx).Err()
}

// Name returns the name of the health checker
func (h *EventBusHealthChecker) Name() string {
	return "event-bus"
}