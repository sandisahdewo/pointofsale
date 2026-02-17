-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE products (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    description   TEXT,
    category_id   BIGINT NOT NULL REFERENCES categories(id),
    price_setting VARCHAR(20) NOT NULL DEFAULT 'fixed',
    markup_type   VARCHAR(20),
    has_variants  BOOLEAN NOT NULL DEFAULT false,
    status        VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_status ON products(status);

CREATE TABLE product_images (
    id         BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url  TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_product_images_product_id ON product_images(product_id);

CREATE TABLE product_suppliers (
    product_id  BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    supplier_id BIGINT NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, supplier_id)
);

CREATE INDEX idx_product_suppliers_supplier_id ON product_suppliers(supplier_id);

CREATE TABLE product_units (
    id                BIGSERIAL PRIMARY KEY,
    product_id        BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name              VARCHAR(100) NOT NULL,
    conversion_factor DECIMAL(15,4) NOT NULL DEFAULT 1,
    converts_to_id    BIGINT REFERENCES product_units(id) ON DELETE SET NULL,
    to_base_unit      DECIMAL(15,4) NOT NULL DEFAULT 1,
    is_base           BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_product_units_product_id ON product_units(product_id);
CREATE UNIQUE INDEX idx_product_units_name_per_product ON product_units(product_id, LOWER(name));

CREATE TABLE product_variants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id    BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku           VARCHAR(100),
    barcode       VARCHAR(100),
    current_stock INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX idx_product_variants_sku ON product_variants(sku) WHERE sku IS NOT NULL;
CREATE INDEX idx_product_variants_barcode ON product_variants(barcode) WHERE barcode IS NOT NULL;

CREATE TABLE variant_attributes (
    id              BIGSERIAL PRIMARY KEY,
    variant_id      UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    attribute_name  VARCHAR(100) NOT NULL,
    attribute_value VARCHAR(255) NOT NULL
);

CREATE INDEX idx_variant_attributes_variant_id ON variant_attributes(variant_id);

CREATE TABLE variant_images (
    id         BIGSERIAL PRIMARY KEY,
    variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    image_url  TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_variant_images_variant_id ON variant_images(variant_id);

CREATE TABLE variant_pricing_tiers (
    id         BIGSERIAL PRIMARY KEY,
    variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    min_qty    INTEGER NOT NULL CHECK (min_qty > 0),
    value      DECIMAL(15,2) NOT NULL CHECK (value >= 0)
);

CREATE INDEX idx_variant_pricing_tiers_variant_id ON variant_pricing_tiers(variant_id);

CREATE TABLE variant_racks (
    variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    rack_id    BIGINT NOT NULL REFERENCES racks(id) ON DELETE CASCADE,
    PRIMARY KEY (variant_id, rack_id)
);

CREATE INDEX idx_variant_racks_rack_id ON variant_racks(rack_id);

-- +goose Down
DROP TABLE IF EXISTS variant_racks;
DROP TABLE IF EXISTS variant_pricing_tiers;
DROP TABLE IF EXISTS variant_images;
DROP TABLE IF EXISTS variant_attributes;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS product_units;
DROP TABLE IF EXISTS product_suppliers;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS products;
