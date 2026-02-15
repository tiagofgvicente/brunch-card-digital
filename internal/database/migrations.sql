-- 1. Limpeza TOTAL (Reset)
DROP TABLE IF EXISTS loyalty_cards;
DROP TABLE IF EXISTS stores;

-- 2. Tabela de LOJAS (Atualizada com Username e Email para Login Global)
CREATE TABLE stores (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    
    -- CREDENCIAIS DE ACESSO (Login Global)
    admin_username VARCHAR(50) UNIQUE NOT NULL,
    admin_email VARCHAR(150) UNIQUE NOT NULL,
    admin_password TEXT NOT NULL, 

    -- CONFIGURAÇÃO VISUAL
    logo_url TEXT,
    primary_color VARCHAR(7) DEFAULT '#00a896',
    stamp_icon VARCHAR(50) DEFAULT '🍳',
    card_skin VARCHAR(50) DEFAULT 'default', 
    theme_mode VARCHAR(10) DEFAULT 'dark',

    -- REGRAS DE NEGÓCIO
    bronze_threshold INTEGER DEFAULT 15,
    silver_threshold INTEGER DEFAULT 40,
    gold_threshold INTEGER DEFAULT 100,
    
    created_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

-- 3. Tabela de Cartões (Clientes)
CREATE TABLE loyalty_cards (
    id UUID PRIMARY KEY,
    store_id UUID REFERENCES stores(id) ON DELETE CASCADE,
    
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
    consent_date TIMESTAMP DEFAULT NOW(),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(store_id, email)
);

-- 4. Índices para performance
CREATE INDEX idx_store_slug ON stores(slug);
CREATE INDEX idx_store_login ON stores(admin_username, admin_email);
CREATE INDEX idx_card_lookup ON loyalty_cards(store_id, email, phone);

-- 5. SEED (Dados Iniciais Obrigatórios)
INSERT INTO stores (
    id, name, slug, 
    admin_username, admin_email, admin_password, 
    primary_color, stamp_icon, card_skin, theme_mode
)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Brunch VIP Project',
    'brunch',
    'admin_brunch',      -- Username
    'admin@brunch.com',  -- Email
    'brunch2026vip',     -- Password
    '#00a896',         
    '🍳',
    'default',
    'dark'
);