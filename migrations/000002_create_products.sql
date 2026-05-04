DO $$ BEGIN
    CREATE TYPE sanrio_character AS ENUM (
        'hello_kitty', 'cinnamoroll', 'pompompurin', 'my_melody', 'kuromi', 'hangyodon', 'badtz_maru'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS products (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    character   sanrio_character NOT NULL,
    description TEXT,
    base_price  DECIMAL(10,2) NOT NULL,
    stock       INT NOT NULL DEFAULT 0,
    image_url   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
