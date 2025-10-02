package domain

import (
	"context"
	"log/slog"
	"math"
	"math/rand"
	"sync"
	"time"

	"simulated_exchange/pkg/shared"
)

// UserBehavior represents different types of user trading behavior
type UserBehavior struct {
	Type                string             `json:"type"`
	OrderFrequency      float64            `json:"order_frequency"`      // Orders per minute
	PriceReactivity     float64            `json:"price_reactivity"`     // How much price changes affect behavior
	TrendFollowing      float64            `json:"trend_following"`      // Tendency to follow price trends
	RiskTolerance       float64            `json:"risk_tolerance"`       // Risk tolerance level
	PreferredOrderTypes []shared.OrderType `json:"preferred_order_types"`
}

// UserSession represents an active user trading session
type UserSession struct {
	UserID      string       `json:"user_id"`
	Behavior    UserBehavior `json:"behavior"`
	LastActive  time.Time    `json:"last_active"`
	OrderCount  int          `json:"order_count"`
	IsActive    bool         `json:"is_active"`
	StartTime   time.Time    `json:"start_time"`
	SessionType string       `json:"session_type"` // "brief", "normal", "extended"
}

// MarketState tracks current market conditions that affect user behavior
type MarketState struct {
	Symbol           string    `json:"symbol"`
	CurrentPrice     float64   `json:"current_price"`
	PriceChange      float64   `json:"price_change"`      // % change from previous price
	Volume           float64   `json:"volume"`            // Recent trading volume
	Volatility       float64   `json:"volatility"`        // Current volatility measure
	LastUpdate       time.Time `json:"last_update"`
	TrendDirection   string    `json:"trend_direction"`   // "up", "down", "sideways"
	TrendStrength    float64   `json:"trend_strength"`    // 0.0 to 1.0
}

// UserSimulator simulates realistic user trading behavior
type UserSimulator struct {
	orderGenerator *OrderGenerator
	logger         *slog.Logger
	random         *rand.Rand

	// User behavior definitions
	userBehaviors map[string]UserBehavior

	// Active user sessions
	activeSessions map[string]*UserSession
	sessionsMutex  sync.RWMutex

	// Market state tracking
	marketStates map[string]*MarketState
	marketMutex  sync.RWMutex

	// Configuration
	maxConcurrentUsers int
	sessionDuration    map[string]time.Duration
}

// NewUserSimulator creates a new user simulator
func NewUserSimulator(orderGenerator *OrderGenerator, logger *slog.Logger) *UserSimulator {
	us := &UserSimulator{
		orderGenerator:     orderGenerator,
		logger:             logger,
		random:             rand.New(rand.NewSource(time.Now().UnixNano())),
		userBehaviors:      make(map[string]UserBehavior),
		activeSessions:     make(map[string]*UserSession),
		marketStates:       make(map[string]*MarketState),
		maxConcurrentUsers: 100,
		sessionDuration: map[string]time.Duration{
			"brief":    5 * time.Minute,
			"normal":   30 * time.Minute,
			"extended": 2 * time.Hour,
		},
	}

	us.initializeUserBehaviors()
	return us
}

// StartSimulation begins user behavior simulation
func (us *UserSimulator) StartSimulation(ctx context.Context) error {
	us.logger.Info("Starting user behavior simulation")

	// Start session management goroutine
	go us.manageUserSessions(ctx)

	// Start behavior simulation goroutine
	go us.simulateUserBehavior(ctx)

	return nil
}

// OnPriceUpdate handles price update events from the market
func (us *UserSimulator) OnPriceUpdate(symbol string, newPrice float64) {
	us.marketMutex.Lock()
	defer us.marketMutex.Unlock()

	state, exists := us.marketStates[symbol]
	if !exists {
		state = &MarketState{
			Symbol:       symbol,
			CurrentPrice: newPrice,
			LastUpdate:   time.Now(),
		}
		us.marketStates[symbol] = state
		return
	}

	// Calculate price change
	priceChange := (newPrice - state.CurrentPrice) / state.CurrentPrice * 100
	state.PriceChange = priceChange
	state.CurrentPrice = newPrice
	state.LastUpdate = time.Now()

	// Update trend information
	us.updateTrendInfo(state)

	us.logger.Debug("Price update processed",
		"symbol", symbol,
		"new_price", newPrice,
		"price_change", priceChange,
		"trend", state.TrendDirection,
	)

	// Trigger reactive user behavior
	go us.reactToMarketChange(symbol, priceChange)
}

// OnTradeExecuted handles trade execution events
func (us *UserSimulator) OnTradeExecuted(symbol string, price float64, quantity float64) {
	us.marketMutex.Lock()
	state, exists := us.marketStates[symbol]
	if exists {
		// Update volume tracking
		state.Volume += quantity
		// Simple volatility calculation based on trade activity
		state.Volatility = math.Min(1.0, state.Volume/1000.0)
	}
	us.marketMutex.Unlock()

	us.logger.Debug("Trade execution processed",
		"symbol", symbol,
		"price", price,
		"quantity", quantity,
	)

	// Some users react to trade executions
	go us.reactToTradeExecution(symbol, price, quantity)
}

// GetActiveUserCount returns the number of currently active users
func (us *UserSimulator) GetActiveUserCount() int {
	us.sessionsMutex.RLock()
	defer us.sessionsMutex.RUnlock()
	return len(us.activeSessions)
}

// GetMarketState returns the current market state for a symbol
func (us *UserSimulator) GetMarketState(symbol string) (*MarketState, bool) {
	us.marketMutex.RLock()
	defer us.marketMutex.RUnlock()
	state, exists := us.marketStates[symbol]
	if !exists {
		return nil, false
	}
	// Return a copy to avoid race conditions
	stateCopy := *state
	return &stateCopy, true
}

// Private methods

func (us *UserSimulator) initializeUserBehaviors() {
	us.userBehaviors["conservative"] = UserBehavior{
		Type:                "conservative",
		OrderFrequency:      2.0, // 2 orders per minute
		PriceReactivity:     0.3,
		TrendFollowing:      0.2,
		RiskTolerance:       0.3,
		PreferredOrderTypes: []shared.OrderType{shared.OrderTypeLimit},
	}

	us.userBehaviors["aggressive"] = UserBehavior{
		Type:                "aggressive",
		OrderFrequency:      8.0, // 8 orders per minute
		PriceReactivity:     0.8,
		TrendFollowing:      0.7,
		RiskTolerance:       0.8,
		PreferredOrderTypes: []shared.OrderType{shared.OrderTypeMarket, shared.OrderTypeLimit},
	}

	us.userBehaviors["momentum"] = UserBehavior{
		Type:                "momentum",
		OrderFrequency:      5.0, // 5 orders per minute
		PriceReactivity:     0.9,
		TrendFollowing:      0.9,
		RiskTolerance:       0.6,
		PreferredOrderTypes: []shared.OrderType{shared.OrderTypeMarket, shared.OrderTypeStopLoss},
	}
}

func (us *UserSimulator) manageUserSessions(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			us.cleanupExpiredSessions()
			us.createNewSessions()
		}
	}
}

func (us *UserSimulator) cleanupExpiredSessions() {
	us.sessionsMutex.Lock()
	defer us.sessionsMutex.Unlock()

	now := time.Now()
	expired := make([]string, 0)

	for userID, session := range us.activeSessions {
		sessionDuration := us.sessionDuration[session.SessionType]
		if now.Sub(session.StartTime) > sessionDuration {
			expired = append(expired, userID)
		}
	}

	for _, userID := range expired {
		delete(us.activeSessions, userID)
		us.logger.Debug("User session expired", "user_id", userID)
	}
}

func (us *UserSimulator) createNewSessions() {
	us.sessionsMutex.Lock()
	currentUsers := len(us.activeSessions)
	us.sessionsMutex.Unlock()

	if currentUsers >= us.maxConcurrentUsers {
		return
	}

	// Probability of new user joining (higher when fewer users are active)
	joinProbability := float64(us.maxConcurrentUsers-currentUsers) / float64(us.maxConcurrentUsers)

	if us.random.Float64() < joinProbability*0.3 { // 30% base probability
		us.createNewUserSession()
	}
}

func (us *UserSimulator) createNewUserSession() {
	userID := "SIM-" + us.generateRandomString(8)

	// Randomly select user behavior type
	behaviorTypes := []string{"conservative", "aggressive", "momentum"}
	behaviorType := behaviorTypes[us.random.Intn(len(behaviorTypes))]
	behavior := us.userBehaviors[behaviorType]

	// Randomly select session duration
	sessionTypes := []string{"brief", "normal", "extended"}
	sessionType := sessionTypes[us.random.Intn(len(sessionTypes))]

	session := &UserSession{
		UserID:      userID,
		Behavior:    behavior,
		LastActive:  time.Now(),
		OrderCount:  0,
		IsActive:    true,
		StartTime:   time.Now(),
		SessionType: sessionType,
	}

	us.sessionsMutex.Lock()
	us.activeSessions[userID] = session
	us.sessionsMutex.Unlock()

	us.logger.Debug("New user session created",
		"user_id", userID,
		"behavior_type", behaviorType,
		"session_type", sessionType,
	)
}

func (us *UserSimulator) simulateUserBehavior(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			us.processUserActions()
		}
	}
}

func (us *UserSimulator) processUserActions() {
	us.sessionsMutex.RLock()
	sessions := make([]*UserSession, 0, len(us.activeSessions))
	for _, session := range us.activeSessions {
		sessions = append(sessions, session)
	}
	us.sessionsMutex.RUnlock()

	for _, session := range sessions {
		if us.shouldUserAct(session) {
			go us.executeUserAction(session)
		}
	}
}

func (us *UserSimulator) shouldUserAct(session *UserSession) bool {
	// Time since last action
	timeSinceLastAction := time.Since(session.LastActive)

	// Calculate probability based on user behavior and time
	baseInterval := 60.0 / session.Behavior.OrderFrequency // seconds between orders
	probability := float64(timeSinceLastAction.Seconds()) / baseInterval

	return us.random.Float64() < probability*0.1 // 10% chance per check
}

func (us *UserSimulator) executeUserAction(session *UserSession) {
	// Select a random symbol
	symbols := us.orderGenerator.GetSupportedSymbols()
	symbol := symbols[us.random.Intn(len(symbols))]

	// Get current market state
	marketState, exists := us.GetMarketState(symbol)
	if !exists {
		// Use a default price if no market data available
		marketState = &MarketState{
			Symbol:       symbol,
			CurrentPrice: us.getDefaultPrice(symbol),
		}
	}

	// Generate an order based on user behavior
	order, err := us.orderGenerator.GenerateOrder(
		context.Background(),
		session.Behavior.Type,
		symbol,
		marketState.CurrentPrice,
	)
	if err != nil {
		us.logger.Error("Failed to generate order", "error", err, "user_id", session.UserID)
		return
	}

	// Update session
	us.sessionsMutex.Lock()
	session.LastActive = time.Now()
	session.OrderCount++
	us.sessionsMutex.Unlock()

	us.logger.Info("User action executed",
		"user_id", session.UserID,
		"order_id", order.ID,
		"symbol", symbol,
		"behavior_type", session.Behavior.Type,
	)
}

func (us *UserSimulator) reactToMarketChange(symbol string, priceChange float64) {
	if math.Abs(priceChange) < 1.0 { // Only react to significant price changes
		return
	}

	us.sessionsMutex.RLock()
	sessions := make([]*UserSession, 0)
	for _, session := range us.activeSessions {
		// Users with high reactivity are more likely to react
		if us.random.Float64() < session.Behavior.PriceReactivity*math.Abs(priceChange)/10.0 {
			sessions = append(sessions, session)
		}
	}
	us.sessionsMutex.RUnlock()

	for _, session := range sessions {
		go us.executeUserAction(session)
	}
}

func (us *UserSimulator) reactToTradeExecution(symbol string, price float64, quantity float64) {
	// Only large trades trigger reactions
	if quantity < 1.0 {
		return
	}

	us.sessionsMutex.RLock()
	sessions := make([]*UserSession, 0)
	for _, session := range us.activeSessions {
		// Momentum traders are more likely to react to large trades
		if session.Behavior.Type == "momentum" && us.random.Float64() < 0.3 {
			sessions = append(sessions, session)
		}
	}
	us.sessionsMutex.RUnlock()

	for _, session := range sessions {
		go us.executeUserAction(session)
	}
}

func (us *UserSimulator) updateTrendInfo(state *MarketState) {
	absChange := math.Abs(state.PriceChange)

	if absChange < 0.5 {
		state.TrendDirection = "sideways"
		state.TrendStrength = 0.1
	} else if state.PriceChange > 0 {
		state.TrendDirection = "up"
		state.TrendStrength = math.Min(1.0, absChange/5.0)
	} else {
		state.TrendDirection = "down"
		state.TrendStrength = math.Min(1.0, absChange/5.0)
	}
}

func (us *UserSimulator) getDefaultPrice(symbol string) float64 {
	prices := map[string]float64{
		"BTC":   45000.0,
		"ETH":   3000.0,
		"ADA":   1.20,
		"DOT":   25.0,
		"SOL":   150.0,
		"MATIC": 2.0,
	}

	if price, exists := prices[symbol]; exists {
		return price
	}
	return 100.0 // Default fallback price
}

func (us *UserSimulator) generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[us.random.Intn(len(charset))]
	}
	return string(result)
}