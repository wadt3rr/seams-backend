CREATE TABLE IF NOT EXISTS products (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id  UUID NOT NULL REFERENCES categories(id),
    sku          TEXT NOT NULL UNIQUE,
    name         TEXT NOT NULL,
    description  TEXT,
    price        BIGINT NOT NULL CHECK (price >= 0),
    created_at   TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_category_id ON products(category_id);
