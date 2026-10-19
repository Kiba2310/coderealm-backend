-- Add points column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS points INT NOT NULL DEFAULT 0;

-- Track completed lessons (duplicate protection)
CREATE TABLE IF NOT EXISTS completed_lessons (
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id     TEXT NOT NULL,
    completed_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    points_earned INT NOT NULL DEFAULT 10,
    PRIMARY KEY (user_id, lesson_id)
);
