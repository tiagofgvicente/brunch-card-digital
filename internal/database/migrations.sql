-- 1. Limpeza TOTAL
DROP TABLE IF EXISTS loyalty_cards CASCADE;
DROP TABLE IF EXISTS stores CASCADE;
DROP TABLE IF EXISTS skins CASCADE;

-- 2. Tabela de LOJAS
CREATE TABLE stores (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    admin_username VARCHAR(100) UNIQUE NOT NULL,
    admin_email VARCHAR(255) UNIQUE NOT NULL,
    admin_password TEXT NOT NULL, 
    account_activated BOOLEAN DEFAULT FALSE,
    logo_url TEXT,
    stamp_icon VARCHAR(50) DEFAULT '🍳',
    card_skin VARCHAR(50) DEFAULT 'default', 
    theme_mode VARCHAR(20) DEFAULT 'dark',
    primary_color VARCHAR(20) DEFAULT '#00a896',
    
    -- SaaS & MONETIZAÇÃO
    tier VARCHAR(20) DEFAULT 'free_trial',
    tier_expiration TIMESTAMP NOT NULL,
    billing_cycle VARCHAR(20) DEFAULT 'monthly',
    max_users INTEGER DEFAULT 1,
    status VARCHAR(20) DEFAULT 'active',
    is_active BOOLEAN DEFAULT TRUE,

    -- Thresholds (Necessários para o teu Scan do Go)
    bronze_threshold INTEGER DEFAULT 15,
    silver_threshold INTEGER DEFAULT 40,
    gold_threshold INTEGER DEFAULT 100,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Tabela de Clientes
CREATE TABLE loyalty_cards (
    id TEXT PRIMARY KEY,
    store_id TEXT REFERENCES stores(id) ON DELETE CASCADE,
    member_number SERIAL, -- SERIAL para auto-incremento no Postgres
    customer_id TEXT NOT NULL,
    last_name TEXT,
    email TEXT,
    phone TEXT,
    nif TEXT,
    stamps_count INTEGER DEFAULT 0,
    total_stamps INTEGER DEFAULT 0,
    total_redeemed_bonuses INTEGER DEFAULT 0,
    is_reward_ready BOOLEAN DEFAULT FALSE,
    design VARCHAR(50) DEFAULT NULL, 
    rgpd_accepted BOOLEAN DEFAULT FALSE,
    marketing_accepted BOOLEAN DEFAULT FALSE,
    consent_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, email)
);

-- 4. Tabela de Skins
CREATE TABLE skins (
    id TEXT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, 
    image_data TEXT,           
    color_bg VARCHAR(20) DEFAULT '#cbd5e1',
    color_text VARCHAR(20) DEFAULT '#ffffff',
    color_border VARCHAR(20) DEFAULT '#ffd166',
    is_global BOOLEAN DEFAULT FALSE,
    store_id TEXT REFERENCES stores(id) ON DELETE SET NULL,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Índices
CREATE INDEX idx_store_slug ON stores(slug);
CREATE INDEX idx_store_status ON stores(status);

-- 6. SEED (Password: brunch2026vip)
INSERT INTO stores (id, name, slug, admin_username, admin_email, admin_password, tier, tier_expiration, max_users, is_active)
VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Brunch VIP', 'brunch', 'admin', 'admin@brunch.com', '$2a$10$C82H9C1f1F5J3Z9oR7u2m.I8R8v2p4z6H7x9N9v1w2L7q3t0r1s2e', 'pro', '2030-01-01', 3, TRUE);

INSERT INTO skins (id, name, type, color_bg, color_text, color_border, is_global) VALUES 
('default', 'Default Teal', 'standard', '#00a896', '#ffffff', '#ffd166', TRUE),
('black', 'Black Edition', 'standard', '#1a1a1a', '#ffffff', '#ffd166', TRUE),
('valentine', 'Valentines', 'seasonal', '#ff4d6d', '#ffffff', '#ffffff', TRUE);