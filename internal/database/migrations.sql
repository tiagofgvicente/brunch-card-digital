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

DROP TABLE IF EXISTS brunch_cards;

CREATE TABLE IF NOT EXISTS brunch_cards (
    id UUID PRIMARY KEY,
    member_number SERIAL, -- Gerado automaticamente (1, 2, 3...)
    customer_id TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT CHECK (email IS NULL OR email LIKE '%@%'),
    phone TEXT CHECK (phone IS NULL OR (phone ~ '^[0-9]+$')),
    nif TEXT CHECK (nif IS NULL OR (nif ~ '^[0-9]{9}$')),
    stamps_count INTEGER DEFAULT 0, -- Ciclo 0-10
    total_stamps INTEGER DEFAULT 0, -- Histórico Vitalício (Sempre cresce)
    total_redeemed_bonuses INTEGER DEFAULT 0, -- Contador de prémios ganhos
    is_reward_ready BOOLEAN DEFAULT FALSE,
    design TEXT DEFAULT 'minimalist',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customer_id ON brunch_cards(customer_id);

-- Index for high-speed searches
CREATE INDEX IF NOT EXISTS idx_brunch_search ON brunch_cards (customer_id, last_name, phone, email, nif);