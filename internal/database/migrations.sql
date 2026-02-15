-- 1. Limpar tudo para começar a arquitetura SaaS limpa
DROP TABLE IF EXISTS brunch_cards;
DROP TABLE IF EXISTS system_settings; -- Já não vamos usar esta tabela antiga
DROP TABLE IF EXISTS stores;

-- 2. Tabela de LOJAS (O teu inventário de clientes pagantes)
CREATE TABLE stores (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL, -- ex: 'brunch', 'padaria-ze'. Será usado no URL.
    
    -- Configuração Visual (O que tu controlas centralmente)
    logo_url TEXT,
    primary_color VARCHAR(7) DEFAULT '#00a896', -- ex: #FF0000
    theme_mode VARCHAR(10) DEFAULT 'dark', -- 'light' ou 'dark'
    stamp_icon VARCHAR(50) DEFAULT '🍳', -- Emoji ou URL de imagem
    
    -- Regras de Negócio da Loja
    bronze_threshold INTEGER DEFAULT 15,
    silver_threshold INTEGER DEFAULT 40,
    gold_threshold INTEGER DEFAULT 100,
    
    -- Acesso do Gerente da Loja
    admin_password TEXT NOT NULL, 
    
    created_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

-- 3. Tabela de Cartões (Agora ligada a UMA loja específica)
CREATE TABLE loyalty_cards (
    id UUID PRIMARY KEY,
    store_id UUID REFERENCES stores(id) ON DELETE CASCADE, -- A magia do Multi-tenant
    
    member_number SERIAL, -- Este número será único globalmente ou por loja (depende da implementação do Go)
    customer_id TEXT NOT NULL, -- Nome
    last_name TEXT,
    email TEXT,
    phone TEXT,
    nif TEXT,
    
    -- Game Logic
    stamps_count INTEGER DEFAULT 0,
    total_stamps INTEGER DEFAULT 0,
    total_redeemed_bonuses INTEGER DEFAULT 0,
    is_reward_ready BOOLEAN DEFAULT FALSE,
    
    -- RGPD
    rgpd_accepted BOOLEAN DEFAULT FALSE,
    marketing_accepted BOOLEAN DEFAULT FALSE,
    consent_date TIMESTAMP DEFAULT NOW(),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Garante que um email não se repete DENTRO da mesma loja, mas pode existir noutras
    UNIQUE(store_id, email)
);

-- Índices para performance
CREATE INDEX idx_store_lookup ON stores(slug);
CREATE INDEX idx_card_lookup ON loyalty_cards(store_id, email, phone);