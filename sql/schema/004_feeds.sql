-- +goose up
ALTER TABLE feeds ADD COLUMN last_fetched_at TIMESTAMP WITH TIME ZONE;

-- +goose down
ALTER TABLE feeds DROP COLUMN last_fetched_at;
