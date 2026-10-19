CREATE TABLE IF NOT EXISTS lesson_time_tracking (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id VARCHAR(255) NOT NULL,
    time_spent_seconds INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, lesson_id)
);
