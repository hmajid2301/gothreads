-- +goose Up
ALTER TABLE items ADD COLUMN IF NOT EXISTS description TEXT;

-- +goose Down
ALTER TABLE items DROP COLUMN IF EXISTS description;
