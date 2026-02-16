-- +goose Up
CREATE TABLE racks (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(50) NOT NULL UNIQUE,
    location    VARCHAR(255) NOT NULL,
    capacity    INTEGER NOT NULL CHECK (capacity > 0),
    description TEXT,
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_racks_code_lower ON racks(LOWER(code));
CREATE INDEX idx_racks_active ON racks(active);

-- +goose Down
DROP TABLE IF EXISTS racks;
