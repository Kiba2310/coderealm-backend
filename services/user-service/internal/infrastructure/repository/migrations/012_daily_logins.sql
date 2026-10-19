CREATE TABLE IF NOT EXISTS daily_logins (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    login_date DATE NOT NULL,
    streak_day INT NOT NULL DEFAULT 1,
    points_awarded INT NOT NULL DEFAULT 5,
    PRIMARY KEY (user_id, login_date)
);
