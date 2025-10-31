package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	stan "github.com/nats-io/stan.go"
)

// Структуры заказа
type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string  `json:"transaction"`
	RequestID    string  `json:"request_id"`
	Currency     string  `json:"currency"`
	Provider     string  `json:"provider"`
	Amount       float64 `json:"amount"`
	PaymentDT    int64   `json:"payment_dt"`
	Bank         string  `json:"bank"`
	DeliveryCost float64 `json:"delivery_cost"`
	GoodsTotal   float64 `json:"goods_total"`
	CustomFee    float64 `json:"custom_fee"`
}

type Item struct {
	ChrtID      int     `json:"chrt_id"`
	TrackNumber string  `json:"track_number"`
	Price       float64 `json:"price"`
	RID         string  `json:"rid"`
	Name        string  `json:"name"`
	Sale        float64 `json:"sale"`
	Size        string  `json:"size"`
	TotalPrice  float64 `json:"total_price"`
	NmID        int     `json:"nm_id"`
	Brand       string  `json:"brand"`
	Status      int     `json:"status"`
}

type Order struct {
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

func main() {
	clusterID := os.Getenv("NATS_CLUSTER_ID")
	clientID := "publisher-client"
	natsURL := os.Getenv("NATS_URL")

	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer sc.Close()

	log.Println("Publisher HTTP server started on port 4000")

	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var order Order
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Генерация UID и TrackNumber, если не указаны
		if order.OrderUID == "" {
			order.OrderUID = "order-" + time.Now().Format("150405")
		}
		if order.TrackNumber == "" {
			order.TrackNumber = "TRACK-" + time.Now().Format("150405")
		}
		if order.DateCreated == "" {
			order.DateCreated = time.Now().UTC().Format(time.RFC3339)
		}

		data, _ := json.Marshal(order)

		if err := sc.Publish("orders", data); err != nil {
			http.Error(w, "Failed to publish order", http.StatusInternalServerError)
			log.Printf("Error publishing: %v", err)
			return
		}

		log.Printf("Order %s published", order.OrderUID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(order)
	})

	http.ListenAndServe(":4000", nil)
}
