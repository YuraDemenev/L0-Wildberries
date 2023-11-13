CREATE TABLE orders(
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) NOT NULL,
    track_number VARCHAR(255) NOT NULL,
    entry VARCHAR(255) NOT NULL,
    locale VARCHAR(10) NOT NULL,
    internal_signature VARCHAR(255) NOT NULL,
    customer_id VARCHAR(255) NOT NULL,
    delivery_service VARCHAR(255) NOT NULL,
    shardkey VARCHAR(255) NOT NULL,
    sm_id INT NOT NULL,
    date_created TIMESTAMP,
    oof_shard VARCHAR(255) NOT NULL
);
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    transaction VARCHAR(255) NOT NULL,
    request_id VARCHAR(255) NOT NULL,
    currency VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    amount INT NOT NULL,
    payment_dt INT NOT NULL,
    bank VARCHAR(255) NOT NULL,
    delivery_cost INT NOT NULL,
    goods_total INT NOT NULL,
    custom_fee INT NOT NULL,
    order_id INT REFERENCES orders(id)
);
CREATE TABLE deliveries(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(255) NOT NULL,
    zip VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    region VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    order_id INT REFERENCES orders(id)
);
CREATE TABLE items(
    id SERIAL PRIMARY KEY,
    chrt_id INT NOT NULL,
    track_number VARCHAR(255) NOT NULL,
    price INT NOT NULL,
    rid VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sale INT NOT NULL,
    size VARCHAR(255) NOT NULL,
    total_price INT NOT NULL,
    nm_id INT NOT NULL,
    brand VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    order_id int REFERENCES orders(id)
);