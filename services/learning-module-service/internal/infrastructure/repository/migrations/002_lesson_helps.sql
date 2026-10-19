CREATE TABLE IF NOT EXISTS lesson_helps (
    user_id UUID NOT NULL,
    lesson_id UUID NOT NULL,
    help_count INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, lesson_id)
);
