package queries

const (
	getMessageByID = ``
	GetAllMessage  = `SELECT 
            o.track_number, o.entry, o.locale, o.internal_signature, 
            o.customer_id, o.delivery_service, o."shardkey", o.sm_id, 
            o.date_created, o.oof_shard, 
            p.transaction, p.request_id, p.currency, p.provider, 
            p.amount, p.payment_dt, p.bank, p.delivery_cost, 
            p.goods_total, p.custom_fee, p.order_uid,
            d.name, d.phone,d.zip, d.city, d.address, d.region, d.email
        FROM orders o
        INNER JOIN payment p ON o.order_uid = p.order_uid
        INNER JOIN delivery d ON o.order_uid = d.order_uid;
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
    (name, phone, zip, city, address, region, email, order_uid)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	SavePaymentToDB = `INSERT INTO payment (
    transaction, request_id, currency, provider, 
    amount, payment_dt, bank, delivery_cost, 
    goods_total, custom_fee, order_uid
) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`
)
