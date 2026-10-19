-- Store System Migration

CREATE TABLE IF NOT EXISTS store_items (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    cost INT NOT NULL,
    effect_type VARCHAR(50) NOT NULL,
    duration_hours INT NOT NULL DEFAULT 0,
    rarity VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS user_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id VARCHAR(50) NOT NULL REFERENCES store_items(id) ON DELETE CASCADE,
    purchased_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_used BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS user_store_state (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    items JSONB NOT NULL,
    last_refresh TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO store_items (id, name, description, cost, effect_type, duration_hours, rarity) VALUES
('dragon_crest', 'Sárkány címer', 'Ranglista név aranyszínű glow-val', 5200, 'cosmetic', 0, 'Legendás'),
('xp_amulet', 'XP amulett', '+10% pont 1 héten keresztül', 900, 'xp_boost', 168, 'Közönséges'),
('helper', 'Helper', '1 ingyenes segítség', 1500, 'free_help', 0, 'Ritka'),
('neon_theme', 'Neon terminál téma', 'Zöld scanline effekt a szerkesztőn', 1200, 'cosmetic', 0, 'Ritka'),
('fire_avatar', 'Tűz keretes avatár', 'Animált lángkeret a profilképen', 2400, 'cosmetic', 0, 'Epikus'),
('freeze_rune', 'Fagyasztó rúna', 'Megállítja a napi streak-et 1 napra', 1500, 'freeze', 24, 'Ritka')
ON CONFLICT (id) DO NOTHING;
