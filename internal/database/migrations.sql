-- 1. Limpeza TOTAL
DROP TABLE IF EXISTS loyalty_cards CASCADE;
DROP TABLE IF EXISTS stores CASCADE;
DROP TABLE IF EXISTS skins CASCADE;
DROP TABLE IF EXISTS global_users CASCADE;

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
    
    -- Design & Personalização do Cartão
    card_skin VARCHAR(50) DEFAULT 'default', 
    theme_mode VARCHAR(20) DEFAULT 'dark',
    primary_color VARCHAR(20) DEFAULT '#00a896',
    text_color VARCHAR(20) DEFAULT '#ffffff',     
    border_color VARCHAR(20) DEFAULT '#ffffff',  
    card_image_url TEXT,                        
    card_image_zoom INTEGER DEFAULT 100,
    card_image_pos_x INTEGER DEFAULT 0,
    card_image_pos_y INTEGER DEFAULT 0,
    card_scope VARCHAR(25) DEFAULT 'Geral',
    social_instagram VARCHAR(255) DEFAULT '',
    social_facebook VARCHAR(255) DEFAULT '',
    social_twitter VARCHAR(255) DEFAULT '',
    social_whatsapp VARCHAR(255) DEFAULT '',
    social_tiktok VARCHAR(255) DEFAULT '',
    social_youtube VARCHAR(255) DEFAULT '',
    social_website VARCHAR(255) DEFAULT '',
    menu_url TEXT DEFAULT '',
    location_url TEXT DEFAULT '',
    
    -- SaaS & MONETIZAÇÃO
    tier VARCHAR(20) DEFAULT 'free_trial',
    tier_expiration TIMESTAMP NOT NULL,
    billing_cycle VARCHAR(20) DEFAULT 'monthly',
    max_users INTEGER DEFAULT 1,
    status VARCHAR(20) DEFAULT 'active',
    is_active BOOLEAN DEFAULT TRUE,

    -- Thresholds
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
    member_number SERIAL,
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

CREATE TABLE global_users (
    id VARCHAR(36) PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    password VARCHAR(255) NOT NULL,
    rgpd_accepted BOOLEAN DEFAULT FALSE,
    marketing_accepted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE store_scopes (
    id VARCHAR(36) PRIMARY KEY,
    store_id TEXT REFERENCES stores(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    stamp_icon VARCHAR(50) DEFAULT '🍳',
    is_main BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, name) -- Garante que a loja não cria dois âmbitos com o mesmo nome
);

ALTER TABLE loyalty_cards ADD COLUMN scope_id VARCHAR(36) REFERENCES store_scopes(id) ON DELETE CASCADE;

ALTER TABLE loyalty_cards DROP CONSTRAINT IF EXISTS loyalty_cards_store_id_email_key;
ALTER TABLE loyalty_cards ADD CONSTRAINT loyalty_cards_store_scope_email_key UNIQUE(store_id, scope_id, email);

CREATE TABLE wallet_notifications (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    store_id VARCHAR(36) REFERENCES stores(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) DEFAULT 'info', -- Pode ser: 'info', 'warning', 'error', 'success'
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);



-- ==========================================
-- 6. SEED DATA (DADOS DE TESTE)
-- ==========================================

-- LOJA 1: Brunch VIP 
INSERT INTO stores (id, name, slug, admin_username, admin_email, admin_password, tier, tier_expiration, billing_cycle, max_users, is_active, account_activated, stamp_icon)
VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Brunch VIP', 'brunch', 'admin@brunch.com', 'admin@brunch.com', '$2a$10$n4KCN/mYuBvOSluV4LT1iuQ7MwIUSDryOCjw4b/9wBrei0WzYeBye', 'pro', '2030-01-01', 'annual', 10, TRUE, TRUE, '🍳');

-- LOJA 2: Sushi Zen 
INSERT INTO stores (id, name, slug, admin_username, admin_email, admin_password, tier, tier_expiration, billing_cycle, max_users, is_active, account_activated, stamp_icon)
VALUES ('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Sushi Zen', 'sushi', 'manager@sushizen.com', 'manager@sushizen.com', '$2a$10$n4KCN/mYuBvOSluV4LT1iuQ7MwIUSDryOCjw4b/9wBrei0WzYeBye', 'free_trial', CURRENT_TIMESTAMP + INTERVAL '3 days', 'monthly', 1, TRUE, FALSE, '🍣');

-- LOJA 3: Burger Kingz 
INSERT INTO stores (id, name, slug, admin_username, admin_email, admin_password, tier, tier_expiration, billing_cycle, max_users, is_active, account_activated, stamp_icon)
VALUES ('c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'Burger Kingz', 'burger', 'boss@burgerkingz.com', 'boss@burgerkingz.com', '$2a$10$n4KCN/mYuBvOSluV4LT1iuQ7MwIUSDryOCjw4b/9wBrei0WzYeBye', 'basic', CURRENT_TIMESTAMP + INTERVAL '25 days', 'monthly', 2, TRUE, TRUE, '🍔');

-- LOJA 4: Coffee Lovers 
INSERT INTO stores (id, name, slug, admin_username, admin_email, admin_password, tier, tier_expiration, billing_cycle, max_users, is_active, account_activated, stamp_icon)
VALUES ('d3eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'Coffee Lovers', 'coffee', 'hello@coffeelovers.pt', 'hello@coffeelovers.pt', '$2a$10$n4KCN/mYuBvOSluV4LT1iuQ7MwIUSDryOCjw4b/9wBrei0WzYeBye', 'lite', CURRENT_TIMESTAMP - INTERVAL '1 day', 'biannual', 3, FALSE, TRUE, '☕');

-- LOJA 5: Pastelaria Antiga
INSERT INTO stores (id, name, slug, admin_username, admin_email, admin_password, tier, tier_expiration, billing_cycle, max_users, is_active, account_activated, stamp_icon)
VALUES ('e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'Pastelaria Antiga', 'antiga', 'gerencia@antiga.pt', 'gerencia@antiga.pt', '$2a$10$n4KCN/mYuBvOSluV4LT1iuQ7MwIUSDryOCjw4b/9wBrei0WzYeBye', 'free_trial', CURRENT_TIMESTAMP - INTERVAL '2 days', 'monthly', 1, TRUE, TRUE, '🥐');

-- ==========================================
-- INSERIR OS ÂMBITOS PRINCIPAIS AUTOMÁTICOS
-- (Isto é essencial para o Scanner ter um "Cartão" por defeito)
-- ==========================================
INSERT INTO store_scopes (id, store_id, name, stamp_icon, is_main, is_active) VALUES 
('scope-brunch-1', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Geral', '🍳', TRUE, TRUE),
('scope-sushi-1', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Geral', '🍣', TRUE, TRUE),
('scope-burger-1', 'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'Geral', '🍔', TRUE, TRUE),
('scope-coffee-1', 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'Geral', '☕', TRUE, TRUE),
('scope-antiga-1', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'Geral', '🥐', TRUE, TRUE);

-- SKINS INICIAIS
INSERT INTO skins (id, name, type, color_bg, color_text, color_border, is_global) VALUES 
('default', 'Default Teal', 'standard', '#00a896', '#ffffff', '#ffd166', TRUE),
('black', 'Black Edition', 'standard', '#1a1a1a', '#ffffff', '#ffd166', TRUE),
('valentine', 'Valentines', 'seasonal', '#ff4d6d', '#ffffff', '#ffffff', TRUE);