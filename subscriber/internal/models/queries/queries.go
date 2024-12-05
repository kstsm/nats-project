package queries

const (
	GetOrderByID = `
SELECT row_to_json(row)
FROM (SELECT order_uid,
             track_number,
             entry,
             "locale",
             internal_signature,
             customer_id,
             delivery_service,
             shardkey,
             sm_id,
             date_created AT TIME ZONE 'UTC',
             oof_shard,
             (SELECT row_to_json(row)
              FROM (SELECT id,
                           order_uid,
                           "name",
                           phone,
                           zip,
                           city,
                           address,
                           region,
                           email
                    FROM delivery AS d
                    WHERE d.order_uid = o.order_uid) AS row)  AS delivery,
             (SELECT json_agg(rows)
              FROM (SELECT id,
                           order_uid,
                           chrt_id,
                           track_number,
                           price,
                           rid,
                           name,
                           sale,
                           size,
                           total_price,
                           nm_id,
                           brand,
                           status
                    FROM items AS i
                    WHERE i.order_uid = o.order_uid) AS rows) AS items,
             (SELECT row_to_json(row)
              FROM (SELECT id,
                           order_uid,
                           "transaction",
                           currency,
                           "provider",
                           amount,
                           payment_dt AT TIME ZONE 'UTC',
                           bank,
                           delivery_cost,
                           goods_total,
                           custom_fee
                    FROM payment AS p
                    WHERE p.order_uid = o.order_uid) AS row)  AS payment
      FROM orders AS o) AS row
WHERE order_uid = $1;
`

	GetAllOrders = `
SELECT json_agg(rows)
FROM (SELECT order_uid,
       track_number,
       entry,
       "locale",
       internal_signature,
       customer_id,
       delivery_service,
       shardkey,
       sm_id,
       date_created AT TIME ZONE 'UTC',
       oof_shard,
       (SELECT row_to_json(row)
        FROM (SELECT id,
                     order_uid,
                     "name",
                     phone,
                     zip,
                     city,
                     address,
                     region,
                     email
              FROM delivery AS d
              WHERE d.order_uid = o.order_uid) AS row)  AS delivery,
       (SELECT json_agg(rows)
        FROM (SELECT id,
                     order_uid,
                     chrt_id,
                     track_number,
                     price,
                     rid,
                     name,
                     sale,
                     size,
                     total_price,
                     nm_id,
                     brand,
                     status
              FROM items AS i
              WHERE i.order_uid = o.order_uid) AS rows) AS items,
       (SELECT row_to_json(row)
        FROM (SELECT id,
                     order_uid,
                     "transaction",
                     currency,
                     "provider",
                     amount,
                     payment_dt AT TIME ZONE 'UTC',
                     bank,
                     delivery_cost,
                     goods_total,
                     custom_fee
              FROM payment AS p
              WHERE p.order_uid = o.order_uid) AS row)  AS payment
FROM orders AS o) rows;
`
	SaveOrdersToDB = `INSERT INTO orders (
    track_number, entry, locale, internal_signature, 
    customer_id, delivery_service, shardKey, sm_id, 
    date_created, oof_shard
) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
RETURNING order_uid, track_number, entry, locale, internal_signature, 
          customer_id, delivery_service, shardKey, sm_id, date_created, oof_shard;`

	SavedDeliveryToDB = `INSERT INTO delivery 
    (order_uid,name, phone, zip, city, address, region, email)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING name, phone, zip, city, address, region, email`

	SavePaymentToDB = `INSERT INTO payment 
    (order_uid, transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8,$9, $10)
RETURNING transaction,  currency, provider, 
    amount, payment_dt, bank, delivery_cost, 
    goods_total, custom_fee;`

	SaveItemsToDB = `INSERT INTO items
    (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8,$9, $10,$11,$12)
    RETURNING chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status;`
)
