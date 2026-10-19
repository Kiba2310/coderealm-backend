-- Add is_equipped column to user_inventory for cosmetic toggling
ALTER TABLE user_inventory ADD COLUMN IF NOT EXISTS is_equipped BOOLEAN DEFAULT true;

-- Update existing records to be equipped by default
UPDATE user_inventory SET is_equipped = true WHERE is_equipped IS NULL;

-- Update store items with detailed, high-quality descriptions
UPDATE store_items SET 
    name = 'Tűz Avatár Keret',
    description = 'Lángoló, animált vörös-narancs tűzaura és keret a profilképed körül, amely a ranglistán és a profilodon is ragyog.'
WHERE id = 'fire_avatar';

UPDATE store_items SET 
    name = 'Sárkány Címer',
    description = 'Legendás arany ragyogás a neveden és a profilképeden, amely tekintélyt parancsol a dicsőségfalon.'
WHERE id = 'dragon_crest';

UPDATE store_items SET 
    name = 'Neon Kiber Téma',
    description = 'Futurisztikus kiber-neon ciánkék fénykeret és modern stílus a profilodhoz.'
WHERE id = 'neon_theme';

UPDATE store_items SET 
    name = 'Fagyasztó Rúna',
    description = 'Jégkristályos fagyos aura a profilodhoz, amely megóvja a napi sorozatodat.'
WHERE id = 'freeze_rune';

UPDATE store_items SET 
    name = 'XP Amulett',
    description = '+10% bónusz pont és XP minden sikeresen megoldott feladatért és leckéért egy héten át.'
WHERE id = 'xp_amulet';

UPDATE store_items SET 
    name = 'Helper Segítő Tárgy',
    description = 'Egy ingyenes segítségkérés programozási feladatokhoz pontlevonás nélkül.'
WHERE id = 'helper';
