-- Основная таблица заказов
CREATE TABLE orders (
    order_uid VARCHAR PRIMARY KEY,
    track_number VARCHAR NOT NULL,
    entry VARCHAR NOT NULL,
    locale VARCHAR NOT NULL,
    internal_signature VARCHAR,
    customer_id VARCHAR NOT NULL,
    delivery_service VARCHAR NOT NULL,
    shardkey VARCHAR NOT NULL,
    sm_id INT NOT NULL,
    date_created TIMESTAMP NOT NULL,
    oof_shard VARCHAR NOT NULL
);

-- Таблица доставки
CREATE TABLE delivery (
    order_uid VARCHAR PRIMARY KEY REFERENCES orders(order_uid),
    name VARCHAR NOT NULL,
    phone VARCHAR NOT NULL,
    zip VARCHAR NOT NULL,
    city VARCHAR NOT NULL,
    address VARCHAR NOT NULL,
    region VARCHAR NOT NULL,
    email VARCHAR NOT NULL
);

-- Таблица оплаты
CREATE TABLE payment (
    transaction VARCHAR PRIMARY KEY,
    order_uid VARCHAR NOT NULL REFERENCES orders(order_uid),
    request_id VARCHAR,
    currency VARCHAR NOT NULL,
    provider VARCHAR NOT NULL,
    amount INT NOT NULL,
    payment_dt BIGINT NOT NULL,
    bank VARCHAR NOT NULL,
    delivery_cost INT NOT NULL,
    goods_total INT NOT NULL,
    custom_fee INT NOT NULL
);

-- Таблица товаров в заказе
CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR NOT NULL REFERENCES orders(order_uid),
    chrt_id BIGINT NOT NULL,
    track_number VARCHAR NOT NULL,
    price INT NOT NULL,
    rid VARCHAR NOT NULL,
    name VARCHAR NOT NULL,
    sale INT NOT NULL,
    size VARCHAR NOT NULL,
    total_price INT NOT NULL,
    nm_id BIGINT NOT NULL,
    brand VARCHAR NOT NULL,
    status INT NOT NULL
);