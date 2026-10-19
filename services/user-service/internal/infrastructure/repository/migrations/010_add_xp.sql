-- Add xp column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS xp INT NOT NULL DEFAULT 0;

-- Copy existing points to xp so users don't lose their progress
UPDATE users SET xp = points;
