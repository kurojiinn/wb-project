-- +goose Up
-- +goose StatementBegin
    CREATE table orders (
        order_uid varchar primary key,
        track_number varchar not null,
        entry varchar not null,
        locale varchar not null,
        internal_signature varchar not null,
        customer_id varchar not null,
        delivery_service varchar not null,
        shard_key varchar not null,
        sm_id INT not null,
        date_created TIMESTAMP not null,
        oof_shard varchar not null
    );

    CREATE TABLE deliveries (
        order_uid varchar primary key,
        name varchar not null,
        phone varchar not null,
        zip varchar not null,
        city varchar not null,
        address varchar not null,
        region varchar not null,
        email varchar not null,

        CONSTRAINT fk_deliveries_order
            FOREIGN KEY (order_uid)
            REFERENCES orders(order_uid)
            ON DELETE CASCADE
    );


    CREATE TABLE payments (
        order_uid varchar primary key,
        "transaction" varchar not null,
        request_id varchar not null,
        currency varchar not null,
        provider varchar not null,
        amount INT not null,
        payment_dt INT not null,
        bank varchar not null,
        delivery_cost INT not null,
        goods_total INT not null,
        custom_fee INT not null,

        CONSTRAINT fk_payments_order
            FOREIGN KEY (order_uid)
            REFERENCES orders(order_uid)
            ON DELETE CASCADE
    );

    CREATE TABLE items (
        id serial primary key,
        order_uid varchar not null,
        chrt_id INT not null,
        track_number varchar not null,
        price INT not null,
        rid varchar not null,
        name varchar not null,
        sale INT not null,
        size varchar not null,
        total_price INT not null,
        nm_id INT not null,
        brand varchar not null,
        status INT not null,

        CONSTRAINT fk_items_order
                       FOREIGN KEY (order_uid)
                       REFERENCES orders(order_uid)
                       ON DELETE CASCADE
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table items;
drop table payments;
drop table deliveries;
Drop table orders;
-- +goose StatementEnd

