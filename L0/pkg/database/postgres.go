package database

import (
	"L0/pkg/models"
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	deliveriesTable = "deliveries"
	itemsTable      = "items"
	ordersTable     = "orders"
	paymentsTable   = "payments"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetOrders(db *sqlx.DB) ([]models.Order, error) {
	var orders []models.Order
	query_ := fmt.Sprintf(`
	SELECT ord.id, ord.order_uid, ord.track_number, ord.entry, ord.locale, 
	ord.internal_signature, ord.customer_id, ord.delivery_service, 
	ord.shardkey, ord.sm_id, ord.date_created, ord.oof_shard, 
	del.name, del.phone, del.zip, del.city, del.address, del.region, 
	del.email, 
	pay.transaction, pay.request_id, pay.currency, pay.provider, 
	pay.amount, pay.payment_dt,	pay.bank, pay.delivery_cost, 
	pay.goods_total, pay.custom_fee 
	FROM %s AS ord 
	JOIN %s AS del 
	ON del.order_id = ord.id 
	JOIN %s AS pay 
	ON pay.order_id = ord.id
	`, ordersTable, deliveriesTable, paymentsTable)

	rows, err := db.Query(query_)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var locOrder models.Order
		err := rows.Scan(&locOrder.Id, &locOrder.OrderUid, &locOrder.TrackNumber,
			&locOrder.Entry, &locOrder.Locale, &locOrder.InternalSignature,
			&locOrder.CustomerId, &locOrder.DeliveryService, &locOrder.Shardkey,
			&locOrder.SmId, &locOrder.DateCreated, &locOrder.OofShard,
			&locOrder.Delivery.Name, &locOrder.Delivery.Phone,
			&locOrder.Delivery.Zip, &locOrder.Delivery.City,
			&locOrder.Delivery.Address, &locOrder.Delivery.Region,
			&locOrder.Delivery.Email, &locOrder.Payment.Transaction,
			&locOrder.Payment.RequestId, &locOrder.Payment.Currency,
			&locOrder.Payment.Provider, &locOrder.Payment.Amount,
			&locOrder.Payment.PaymentDt, &locOrder.Payment.Bank,
			&locOrder.Payment.DeliveryCost, &locOrder.Payment.GoodsTotal,
			&locOrder.Payment.CustomFee)

		if err != nil {
			return nil, err
		}
		orders = append(orders, locOrder)
	}

	for i, orderVal := range orders {
		var locItemSlice []models.Items
		query_ := fmt.Sprintf(`SELECT chrt_id, 
		track_number, price, rid, name, sale, 
		size, total_price, nm_id, brand, status 
		FROM %s 
		WHERE order_id = $1`, itemsTable)

		rows, err = db.Query(query_, orderVal.Id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var locItem models.Items
			err := rows.Scan(&locItem.ChrtId, &locItem.TrackNumber, &locItem.Price,
				&locItem.Rid, &locItem.Name, &locItem.Sale,
				&locItem.Size, &locItem.TotalPrice, &locItem.NmId,
				&locItem.Brand, &locItem.Status)

			if err != nil {
				return nil, err
			}
		}
		orders[i].Items = locItemSlice

	}

	return orders, nil
}

func SaveOrder(db *sqlx.DB, order models.Order) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	//Insert into orders
	orderId := 0
	query_ := fmt.Sprintf(`
	INSERT INTO %s (order_uid, track_number, entry, locale, internal_signature,
		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id
	`, ordersTable)
	row := tx.QueryRow(query_, order.OrderUid,
		order.TrackNumber, order.Entry,
		order.Locale, order.InternalSignature,
		order.CustomerId, order.DeliveryService,
		order.Shardkey, order.SmId,
		order.DateCreated, order.OofShard)
	err = row.Scan(&orderId)

	if err != nil {
		return 0, err
	}
	fmt.Println(orderId)

	//Insert into delivery
	query_ = fmt.Sprintf(`
	INSERT INTO %s (name, phone, zip, city, address, region, email, order_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, deliveriesTable)

	_, err = tx.Exec(query_, order.Delivery.Name,
		order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email,
		orderId)

	if err != nil {
		return 0, err
	}

	//Insert into payment
	query_ = fmt.Sprintf(`
	INSERT INTO %s (transaction, request_id, currency, provider, amount, payment_dt,
	bank, delivery_cost, goods_total, custom_fee, order_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, paymentsTable)

	_, err = tx.Exec(query_, order.Payment.Transaction,
		order.Payment.RequestId, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDt, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal,
		order.Payment.CustomFee, orderId)
	if err != nil {
		return 0, err
	}

	//Insert into items
	for _, item := range order.Items {
		query_ = fmt.Sprintf(`
		INSERT INTO %s(chrt_id, track_number, price, rid, name, sale, size, total_price,
		nm_id, brand, status, order_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, itemsTable)
		_, err = tx.Exec(query_, item.ChrtId,
			item.TrackNumber, item.Price,
			item.Rid, item.Name,
			item.Sale, item.Size,
			item.TotalPrice, item.NmId,
			item.Brand, item.Status,
			orderId)
		if err != nil {
			return 0, err
		}
	}

	fmt.Println("Postgres")

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return orderId, nil
}
