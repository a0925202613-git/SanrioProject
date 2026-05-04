CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(50) UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    balance       DECIMAL(10,2) NOT NULL DEFAULT 0,
    total_spent   INT DEFAULT 0,
    vip_level     VARCHAR(20) DEFAULT 'normal',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
