-- Fix users table to match Keycloak integration
ALTER TABLE users ADD COLUMN IF NOT EXISTS keycloak_id VARCHAR(255);
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;

-- Ensure keycloak_id is unique if there are values
-- We can't immediately add NOT NULL if there are existing rows without it,
-- but this is a dev DB, so we can just update existing rows or accept nulls temporarily.
