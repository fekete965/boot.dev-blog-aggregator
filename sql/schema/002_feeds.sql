-- +goose up
CREATE TABLE feeds (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  url VARCHAR(255) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL
);

-- +goose down
DROP TABLE feeds;
