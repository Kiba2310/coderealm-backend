-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS achievements JSONB DEFAULT '[]'::jsonb;
