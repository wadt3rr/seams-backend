CREATE TABLE IF NOT EXISTS order_items (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id     UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id   UUID NOT NULL REFERENCES products(id),
    price        BIGINT NOT NULL CHECK (price >= 0),
    quantity     INTEGER NOT NULL CHECK (quantity > 0),
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
