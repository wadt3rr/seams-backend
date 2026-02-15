CREATE TABLE IF NOT EXISTS orders (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id    UUID NOT NULL REFERENCES customers(id),
    status         TEXT NOT NULL CHECK (status IN ('new', 'confirmed', 'invoiced', 'paid', 'cancelled')),
    total          BIGINT NOT NULL CHECK (total >= 0),
    created_at     TIMESTAMP NOT NULL DEFAULT now(),
    updated_at     TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_customer_id ON orders(customer_id);
