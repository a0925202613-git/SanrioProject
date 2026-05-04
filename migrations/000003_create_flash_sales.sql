CREATE TABLE IF NOT EXISTS flash_sales (
    id              BIGSERIAL PRIMARY KEY,
    product_id      BIGINT NOT NULL REFERENCES products(id),
    sale_price      DECIMAL(10,2) NOT NULL,
    total_stock     INT NOT NULL,
    remaining_stock INT NOT NULL,
    start_time      TIMESTAMPTZ NOT NULL,
    end_time        TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
