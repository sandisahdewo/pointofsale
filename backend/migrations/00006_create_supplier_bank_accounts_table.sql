-- +goose Up
CREATE TABLE supplier_bank_accounts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id    BIGINT NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    account_name   VARCHAR(255) NOT NULL,
    account_number VARCHAR(100) NOT NULL
);

CREATE INDEX idx_supplier_bank_accounts_supplier_id ON supplier_bank_accounts(supplier_id);

-- +goose Down
DROP TABLE IF EXISTS supplier_bank_accounts;
