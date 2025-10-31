package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// Routes задаёт HTTP-маршруты для сервиса
func Routes() {
	// Получение конкретного заказа по ID
	http.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/orders/")
		if id == "" {
			http.Error(w, "Order ID is required", http.StatusBadRequest)
			return
		}

		order, ok := GetCache(id)
		if !ok {
			http.NotFound(w, r)
			log.Printf("Order %s not found in cache", id)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(order); err != nil {
			log.Printf("Failed to encode order %s: %v", id, err)
		}
	})

	// Новый endpoint для просмотра всего кэша
	http.HandleFunc("/cache", func(w http.ResponseWriter, r *http.Request) {
		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		Cache.RLock()
		defer Cache.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(Cache.Orders); err != nil {
			log.Printf("Failed to encode cache: %v", err)
		}
	})
}
