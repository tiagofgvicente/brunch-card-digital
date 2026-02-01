-- Initial migration to create the loyalty card table
-- This runs automatically when the Go app starts

CREATE TABLE IF NOT EXISTS brunch_cards (
    id UUID PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    stamps_count INTEGER DEFAULT 0,
    is_reward_ready BOOLEAN DEFAULT FALSE,
    design VARCHAR(50) DEFAULT 'minimalist',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index to speed up lookups by customer ID
CREATE INDEX IF NOT EXISTS idx_customer_id ON brunch_cards(customer_id);

CREATE TABLE IF NOT EXISTS brunch_cards (
    id UUID PRIMARY KEY,
    customer_id TEXT NOT NULL, -- First Name
    last_name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    nif TEXT,
    stamps_count INTEGER DEFAULT 0,
    total_stamps INTEGER DEFAULT 0,
    is_reward_ready BOOLEAN DEFAULT FALSE,
    design TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for high-speed searches
CREATE INDEX IF NOT EXISTS idx_brunch_search ON brunch_cards (customer_id, last_name, phone, email, nif);