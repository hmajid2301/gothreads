-- +goose Up
ALTER TABLE calendar_events ADD COLUMN IF NOT EXISTS location VARCHAR(255);
CREATE UNIQUE INDEX IF NOT EXISTS idx_calendar_events_user_date ON calendar_events(user_id, event_date);

-- +goose Down
DROP INDEX IF EXISTS idx_calendar_events_user_date;
ALTER TABLE calendar_events DROP COLUMN IF EXISTS location;
