-- Таблица orders
CREATE TABLE IF NOT EXISTS orders (
    order_uid TEXT PRIMARY KEY,
    track_number TEXT,
    entry TEXT,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard TEXT
);

-- Таблица delivery
CREATE TABLE IF NOT EXISTS delivery (
    order_uid TEXT REFERENCES orders(order_uid),
    name TEXT,
    phone TEXT,
    zip TEXT,
    city TEXT,
    address TEXT,
    region TEXT,
    email TEXT,
    PRIMARY KEY(order_uid)
);

-- Таблица payment
CREATE TABLE IF NOT EXISTS payment (
    order_uid TEXT REFERENCES orders(order_uid),
    transaction TEXT,
    request_id TEXT,
    currency TEXT,
    provider TEXT,
    amount INT,
    payment_dt BIGINT,
    bank TEXT,
    delivery_cost INT,
    goods_total INT,
    custom_fee INT,
    PRIMARY KEY(order_uid)
);

-- Таблица items
CREATE TABLE IF NOT EXISTS items (
    chrt_id BIGINT PRIMARY KEY,
    order_uid TEXT REFERENCES orders(order_uid),
    track_number TEXT,
    price INT,
    rid TEXT,
    name TEXT,
    sale INT,
    size TEXT,
    total_price INT,
    nm_id BIGINT,
    brand TEXT,
    status INT
);
