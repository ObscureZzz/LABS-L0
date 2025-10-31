package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	stan "github.com/nats-io/stan.go"
)

func main() {
	db, err := ConnectDB()
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	LoadCacheFromDB(db)
	Cache.RLock()
	fmt.Printf("Cache content at start: %+v\n", Cache.Orders)
	Cache.RUnlock()

	clusterID := os.Getenv("NATS_CLUSTER_ID")
	clientID := os.Getenv("NATS_CLIENT_ID")
	natsURL := os.Getenv("NATS_URL")

	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer sc.Close()

	_, err = sc.Subscribe("orders", func(msg *stan.Msg) {
		var orderRaw struct {
			OrderUID          string   `json:"order_uid"`
			TrackNumber       string   `json:"track_number"`
			Entry             string   `json:"entry"`
			Delivery          Delivery `json:"delivery"`
			Payment           Payment  `json:"payment"`
			Items             []Item   `json:"items"`
			Locale            string   `json:"locale"`
			InternalSignature string   `json:"internal_signature"`
			CustomerID        string   `json:"customer_id"`
			DeliveryService   string   `json:"delivery_service"`
			Shardkey          string   `json:"shardkey"`
			SmID              int      `json:"sm_id"`
			DateCreated       string   `json:"date_created"`
			OofShard          string   `json:"oof_shard"`
		}

		if err := json.Unmarshal(msg.Data, &orderRaw); err != nil {
			log.Printf("Invalid order JSON: %v", err)
			return
		}

		// Парсим дату
		dateCreated, err := time.Parse(time.RFC3339, orderRaw.DateCreated)
		if err != nil || orderRaw.DateCreated == "" {
			dateCreated = time.Now()
		}

		order := Order{
			OrderUID:          orderRaw.OrderUID,
			TrackNumber:       orderRaw.TrackNumber,
			Entry:             orderRaw.Entry,
			Delivery:          orderRaw.Delivery,
			Payment:           orderRaw.Payment,
			Items:             orderRaw.Items,
			Locale:            orderRaw.Locale,
			InternalSignature: orderRaw.InternalSignature,
			CustomerID:        orderRaw.CustomerID,
			DeliveryService:   orderRaw.DeliveryService,
			Shardkey:          orderRaw.Shardkey,
			SmID:              orderRaw.SmID,
			DateCreated:       dateCreated,
			OofShard:          orderRaw.OofShard,
		}

		SetCache(order)

		inserted, err := SaveOrderToDB(db, order)
		if err != nil {
			log.Printf("Error saving order to DB: %v", err)
		} else if inserted {
			log.Printf("✅ Order %s успешно сохранён в базу", order.OrderUID)
		} else {
			log.Printf("⚠️ Order %s уже существует, сохранение пропущено", order.OrderUID)
		}

	}, stan.DeliverAllAvailable())
	if err != nil {
		log.Fatalf("Error subscribing to orders: %v", err)
	}

	Routes()

	port := os.Getenv("PORT")
	fmt.Printf("Service running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
