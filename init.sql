CREATE TABLE IF NOT EXISTS inventory (
    id          SERIAL PRIMARY KEY,
    product_id  VARCHAR(64)  NOT NULL UNIQUE,
    total_stock BIGINT       NOT NULL DEFAULT 0,
    sold_stock  BIGINT       NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO inventory (product_id, total_stock, sold_stock)
VALUES ('flash-sale-item-001', 1000000, 0)
ON CONFLICT (product_id) DO NOTHING;


