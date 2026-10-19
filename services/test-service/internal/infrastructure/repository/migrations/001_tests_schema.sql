CREATE TABLE IF NOT EXISTS tests (
    id UUID PRIMARY KEY,
    course_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_visible BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    open_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
