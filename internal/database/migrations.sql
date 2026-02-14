-- Initial migration to create the loyalty card table
-- This runs automatically when the Go app starts

CREATE TABLE IF NOT EXISTS brunch_cards (
    id UUID PRIMARY KEY,
    member_number SERIAL,
    customer_id TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT CHECK (email IS NULL OR email LIKE '%@%'),
    phone TEXT CHECK (phone IS NULL OR (phone ~ '^[0-9]+$')),
    nif TEXT CHECK (nif IS NULL OR (nif ~ '^[0-9]{9}$')),
    
    -- Game Logic
    stamps_count INTEGER DEFAULT 0,
    total_stamps INTEGER DEFAULT 0,
    total_redeemed_bonuses INTEGER DEFAULT 0,
    is_reward_ready BOOLEAN DEFAULT FALSE,
    design TEXT DEFAULT 'minimalist',
    
    -- RGPD & Consent (NOVOS CAMPOS)
    rgpd_accepted BOOLEAN DEFAULT FALSE,
    marketing_accepted BOOLEAN DEFAULT FALSE,
    consent_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS system_settings (
    id SERIAL PRIMARY KEY,
    store_name TEXT DEFAULT 'Brunch.co',
    store_logo TEXT DEFAULT '',
    theme_mode TEXT DEFAULT 'dark',
    primary_color TEXT DEFAULT '#00a896',
    
    -- Loyalty Thresholds (NOVOS CAMPOS)
    bronze_threshold INTEGER DEFAULT 15,
    silver_threshold INTEGER DEFAULT 40,
    gold_threshold INTEGER DEFAULT 100,

    -- Security (NOVO CAMPO)
    admin_password TEXT DEFAULT 'brunch2026vip',

    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert the initial default row with default thresholds and password
INSERT INTO system_settings (id, store_name, bronze_threshold, silver_threshold, gold_threshold, admin_password) 
VALUES (1, 'Brunch.co', 15, 40, 100, 'brunch2026vip') 
ON CONFLICT (id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_customer_id ON brunch_cards(customer_id);
CREATE INDEX IF NOT EXISTS idx_brunch_search ON brunch_cards (customer_id, last_name, phone, email, nif);