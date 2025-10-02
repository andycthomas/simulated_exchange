-- Critical Database Indexes for Trading Performance
-- These indexes will dramatically improve query performance

-- Orders table indexes (most critical for trading)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_symbol_status ON orders(symbol, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_user_created ON orders(user_id, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_status_created ON orders(status, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_symbol_side_status ON orders(symbol, side, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_symbol_price_pending ON orders(symbol, price) WHERE status IN ('PENDING', 'PARTIAL');

-- Trades table indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_symbol_timestamp ON trades(symbol, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_timestamp ON trades(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_order_ids ON trades(buy_order_id, sell_order_id);

-- Users table indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_created ON users(created_at DESC);

-- Composite indexes for order book queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_orderbook_buy ON orders(symbol, price DESC, created_at) WHERE side = 'BUY' AND status IN ('PENDING', 'PARTIAL');
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_orderbook_sell ON orders(symbol, price ASC, created_at) WHERE side = 'SELL' AND status IN ('PENDING', 'PARTIAL');

-- Update table statistics
ANALYZE orders;
ANALYZE trades;
ANALYZE users;

-- Show index creation results
\d+ orders
\d+ trades