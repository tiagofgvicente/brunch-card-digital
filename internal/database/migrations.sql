-- 1. Limpeza TOTAL (Reset) com CASCADE para apagar dependências
DROP TABLE IF EXISTS loyalty_cards CASCADE;
DROP TABLE IF EXISTS stores CASCADE;
DROP TABLE IF EXISTS skins CASCADE;

-- 2. Tabela de LOJAS
CREATE TABLE stores (
    id TEXT PRIMARY KEY, -- Mantemos TEXT para compatibilidade com o UUID string do Go
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    
    -- CREDENCIAIS DE ACESSO
    admin_username VARCHAR(100) UNIQUE NOT NULL,
    admin_email VARCHAR(255) UNIQUE NOT NULL,
    admin_password TEXT NOT NULL, 

    -- CONFIGURAÇÃO VISUAL
    logo_url TEXT,
    primary_color VARCHAR(20) DEFAULT '#00a896',
    stamp_icon VARCHAR(50) DEFAULT '🍳',
    card_skin VARCHAR(50) DEFAULT 'default', 
    theme_mode VARCHAR(20) DEFAULT 'dark',

    -- REGRAS DE NEGÓCIO
    bronze_threshold INTEGER DEFAULT 15,
    silver_threshold INTEGER DEFAULT 40,
    gold_threshold INTEGER DEFAULT 100,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- 3. Tabela de Cartões (Clientes)
CREATE TABLE loyalty_cards (
    id TEXT PRIMARY KEY,
    store_id TEXT REFERENCES stores(id) ON DELETE CASCADE,
    
    member_number INTEGER, 
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

-- 4. Tabela de Skins Globais
CREATE TABLE skins (
    id TEXT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, 
    
    image_data TEXT,
    color_bg VARCHAR(20) DEFAULT '#cbd5e1',
    color_text VARCHAR(20) DEFAULT '#ffffff',
    color_border VARCHAR(20) DEFAULT '#ffd166',
    
    is_global BOOLEAN DEFAULT FALSE,
    store_id TEXT, -- NOVA COLUNA: Se preenchido, é exclusivo desta loja
    
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY(store_id) REFERENCES stores(id) ON DELETE SET NULL
);

-- 5. Índices para performance
CREATE INDEX idx_store_slug ON stores(slug);
CREATE INDEX idx_store_login ON stores(admin_username, admin_email);
CREATE INDEX idx_card_lookup ON loyalty_cards(store_id, email, phone);

-- 6. SEED LOJA (Dados Iniciais)
INSERT INTO stores (
    id, name, slug, 
    admin_username, admin_email, admin_password, 
    primary_color, stamp_icon, card_skin, theme_mode
)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Brunch VIP Project',
    'brunch',
    'admin_brunch',      
    'admin@brunch.com',  
    'brunch2026vip',     
    '#00a896',         
    '🍳',
    'default',
    'dark'
)
ON CONFLICT (id) DO NOTHING; -- Sintaxe correta para Postgres

-- 7. SEED SKINS (Dados Iniciais)
INSERT INTO skins (id, name, type, color_bg, color_text, color_border, is_global) VALUES 
('default', 'Default Teal', 'standard', '#00a896', '#ffffff', '#ffd166', TRUE),
('black', 'Black Edition', 'standard', '#1a1a1a', '#ffffff', '#ffd166', TRUE),
('valentine', 'Valentines', 'seasonal', '#ff4d6d', '#ffffff', '#ffffff', TRUE)
ON CONFLICT (id) DO NOTHING;