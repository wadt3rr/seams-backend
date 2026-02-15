CREATE TABLE IF NOT EXISTS invoices (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id    UUID NOT NULL UNIQUE REFERENCES orders(id),
    number      TEXT NOT NULL UNIQUE,
    amount      BIGINT NOT NULL CHECK (amount >= 0),
    status      TEXT NOT NULL CHECK (status IN ('issued', 'sent', 'paid', 'cancelled')),
    issued_at   TIMESTAMP NOT NULL DEFAULT now(),
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now()
);
