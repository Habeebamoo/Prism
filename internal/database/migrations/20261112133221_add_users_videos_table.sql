-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(email)
);

CREATE TABLE IF NOT EXISTS videos (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  title TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'READY',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS videos;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
