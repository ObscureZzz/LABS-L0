package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// ConnectDB подключается к PostgreSQL и возвращает *sql.DB
func ConnectDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func SaveOrderToDB(db *sql.DB, order Order) (bool, error) {
	// Сначала проверяем, есть ли заказ в таблице orders
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM orders WHERE order_uid=$1)`, order.OrderUID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check existing order: %w", err)
	}
	if exists {
		return false, nil // заказ уже есть
	}

	// Если заказа нет — вставляем все таблицы в транзакции
	tx, err := db.Begin()
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}

	// orders
	_, err = tx.Exec(`
        INSERT INTO orders(
            order_uid, track_number, entry, locale, internal_signature, 
            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
    `, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("insert orders: %w", err)
	}

	// delivery
	_, err = tx.Exec(`
        INSERT INTO delivery(
            order_uid, name, phone, zip, city, address, region, email
        ) VALUES($1,$2,$3,$4,$5,$6,$7,$8)
    `, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("insert delivery: %w", err)
	}

	// payment
	_, err = tx.Exec(`
        INSERT INTO payment(
            order_uid, transaction, request_id, currency, provider, amount, payment_dt, 
            bank, delivery_cost, goods_total, custom_fee
        ) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
    `, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("insert payment: %w", err)
	}

	// items
	for _, item := range order.Items {
		_, err = tx.Exec(`
            INSERT INTO items(
                chrt_id, order_uid, track_number, price, rid, name, sale, size, 
                total_price, nm_id, brand, status
            ) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
        `, item.ChrtID, order.OrderUID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			tx.Rollback()
			return false, fmt.Errorf("insert item %d: %w", item.ChrtID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit tx: %w", err)
	}

	return true, nil
}

func LoadCacheFromDB(db *sql.DB) {
	rows, err := db.Query(`
        SELECT order_uid, track_number, entry, locale, internal_signature, 
               customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders
    `)
	if err != nil {
		log.Printf("Error loading orders from DB: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var order Order
		if err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
		); err != nil {
			log.Printf("Scan order error: %v", err)
			continue
		}

		// Загружаем delivery
		err = db.QueryRow(`
            SELECT name, phone, zip, city, address, region, email 
            FROM delivery WHERE order_uid=$1
        `, order.OrderUID).Scan(
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
		)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error loading delivery for order %s: %v", order.OrderUID, err)
		}

		// Загружаем payment
		err = db.QueryRow(`
            SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
            FROM payment WHERE order_uid=$1
        `, order.OrderUID).Scan(
			&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
			&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
			&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
		)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error loading payment for order %s: %v", order.OrderUID, err)
		}

		// Загружаем items
		itemRows, err := db.Query(`
            SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
            FROM items WHERE order_uid=$1
        `, order.OrderUID)
		if err != nil {
			log.Printf("Error loading items for order %s: %v", order.OrderUID, err)
		} else {
			for itemRows.Next() {
				var item Item
				if err := itemRows.Scan(
					&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
					&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
				); err != nil {
					log.Printf("Error scanning item for order %s: %v", order.OrderUID, err)
					continue
				}
				order.Items = append(order.Items, item)
			}
			itemRows.Close()
		}

		// Сохраняем полностью загруженный заказ в кэш
		SetCache(order)
		count++
	}
	log.Printf("Cache loaded with %d complete orders", count)
}
