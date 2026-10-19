-- +goose Up
CREATE TABLE IF NOT EXISTS user_achievements (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    achievement_id VARCHAR(255) NOT NULL,
    earned_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY(user_id, achievement_id)
);

-- Migrate existing JSONB array data into the new table
INSERT INTO user_achievements (user_id, achievement_id, earned_at)
SELECT id, value, NOW()
FROM users, jsonb_array_elements_text(achievements) AS value
WHERE jsonb_typeof(achievements) = 'array'
ON CONFLICT DO NOTHING;
