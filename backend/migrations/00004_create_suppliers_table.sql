-- +goose Up
CREATE TABLE suppliers (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    address    TEXT NOT NULL,
    phone      VARCHAR(50),
    email      VARCHAR(255),
    website    VARCHAR(255),
    active     BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_suppliers_active ON suppliers(active);

-- +goose Down
DROP TABLE IF EXISTS suppliers;
