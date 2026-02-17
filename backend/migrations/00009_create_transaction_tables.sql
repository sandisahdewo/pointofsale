-- +goose Up

CREATE TABLE purchase_orders (
    id                       BIGSERIAL PRIMARY KEY,
    po_number                VARCHAR(20) NOT NULL UNIQUE,
    supplier_id              BIGINT NOT NULL REFERENCES suppliers(id),
    date                     DATE NOT NULL,
    status                   VARCHAR(20) NOT NULL DEFAULT 'draft',
    notes                    TEXT,
    received_date            TIMESTAMPTZ,
    payment_method           VARCHAR(20),
    supplier_bank_account_id UUID REFERENCES supplier_bank_accounts(id),
    subtotal                 DECIMAL(15,2),
    total_items              INTEGER,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_purchase_orders_supplier_id ON purchase_orders(supplier_id);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);
CREATE INDEX idx_purchase_orders_po_number ON purchase_orders(po_number);
CREATE INDEX idx_purchase_orders_date ON purchase_orders(date DESC);

CREATE TABLE purchase_order_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id BIGINT NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id        BIGINT NOT NULL REFERENCES products(id),
    variant_id        UUID NOT NULL REFERENCES product_variants(id),
    unit_id           BIGINT NOT NULL REFERENCES product_units(id),
    unit_name         VARCHAR(100) NOT NULL,
    product_name      VARCHAR(255) NOT NULL,
    variant_label     VARCHAR(255) NOT NULL,
    sku               VARCHAR(100),
    current_stock     INTEGER NOT NULL DEFAULT 0,
    ordered_qty       INTEGER NOT NULL CHECK (ordered_qty > 0),
    price             DECIMAL(15,2) NOT NULL DEFAULT 0,
    received_qty      INTEGER,
    received_price    DECIMAL(15,2),
    is_verified       BOOLEAN DEFAULT false
);

CREATE INDEX idx_po_items_purchase_order_id ON purchase_order_items(purchase_order_id);
CREATE INDEX idx_po_items_variant_id ON purchase_order_items(variant_id);

CREATE TABLE sales_transactions (
    id                 BIGSERIAL PRIMARY KEY,
    transaction_number VARCHAR(30) NOT NULL UNIQUE,
    date               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    subtotal           DECIMAL(15,2) NOT NULL,
    grand_total        DECIMAL(15,2) NOT NULL,
    total_items        INTEGER NOT NULL,
    payment_method     VARCHAR(20) NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sales_transactions_date ON sales_transactions(date DESC);
CREATE INDEX idx_sales_transactions_number ON sales_transactions(transaction_number);

CREATE TABLE sales_transaction_items (
    id              BIGSERIAL PRIMARY KEY,
    transaction_id  BIGINT NOT NULL REFERENCES sales_transactions(id) ON DELETE CASCADE,
    product_id      BIGINT NOT NULL REFERENCES products(id),
    variant_id      UUID NOT NULL REFERENCES product_variants(id),
    unit_id         BIGINT NOT NULL REFERENCES product_units(id),
    product_name    VARCHAR(255) NOT NULL,
    variant_label   VARCHAR(255) NOT NULL,
    sku             VARCHAR(100),
    unit_name       VARCHAR(100) NOT NULL,
    quantity        INTEGER NOT NULL CHECK (quantity > 0),
    base_qty        INTEGER NOT NULL,
    unit_price      DECIMAL(15,2) NOT NULL,
    total_price     DECIMAL(15,2) NOT NULL
);

CREATE INDEX idx_sales_items_transaction_id ON sales_transaction_items(transaction_id);
CREATE INDEX idx_sales_items_variant_id ON sales_transaction_items(variant_id);

CREATE TABLE stock_movements (
    id              BIGSERIAL PRIMARY KEY,
    variant_id      UUID NOT NULL REFERENCES product_variants(id),
    movement_type   VARCHAR(20) NOT NULL,
    quantity        INTEGER NOT NULL,
    reference_type  VARCHAR(20),
    reference_id    BIGINT,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stock_movements_variant_id ON stock_movements(variant_id);
CREATE INDEX idx_stock_movements_type ON stock_movements(movement_type);

-- +goose Down
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS sales_transaction_items;
DROP TABLE IF EXISTS sales_transactions;
DROP TABLE IF EXISTS purchase_order_items;
DROP TABLE IF EXISTS purchase_orders;
