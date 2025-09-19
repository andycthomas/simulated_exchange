-- Initialize database for development

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create schemas
CREATE SCHEMA IF NOT EXISTS trading;
CREATE SCHEMA IF NOT EXISTS analytics;

-- Set default search path
ALTER DATABASE trading_db_dev SET search_path = trading, public;

-- Create simple tables for development
CREATE TABLE IF NOT EXISTS trading.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS trading.orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES trading.users(id),
    symbol VARCHAR(10) NOT NULL,
    side VARCHAR(4) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert test data
INSERT INTO trading.users (username, email) VALUES
('dev_user', 'dev@localhost'),
('test_trader', 'test@localhost')
ON CONFLICT (username) DO NOTHING;

INSERT INTO trading.orders (user_id, symbol, side, quantity, price) VALUES
((SELECT id FROM trading.users WHERE username = 'dev_user'), 'BTCUSD', 'BUY', 1.0, 50000.00),
((SELECT id FROM trading.users WHERE username = 'test_trader'), 'ETHUSD', 'SELL', 2.0, 3000.00)
ON CONFLICT DO NOTHING;