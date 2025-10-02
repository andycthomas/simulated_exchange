-- Critical Database Optimization Script
-- Run this to fix performance issues

-- Add critical indexes for trading queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_symbol_status ON orders(symbol, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_user_created ON orders(user_id, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_status_created ON orders(status, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_symbol_timestamp ON trades(symbol, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_timestamp ON trades(created_at DESC);

-- Add index for order matching queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_symbol_side_status ON orders(symbol, side, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_symbol_price ON orders(symbol, price) WHERE status IN ('PENDING', 'PARTIAL');

-- Analyze tables to update statistics
ANALYZE orders;
ANALYZE trades;
ANALYZE users;

-- Show index usage
SELECT schemaname, tablename, indexname, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_tup_read DESC;