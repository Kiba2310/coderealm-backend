ALTER TABLE completed_lessons ADD COLUMN IF NOT EXISTS lesson_type VARCHAR(50) DEFAULT 'unknown';
