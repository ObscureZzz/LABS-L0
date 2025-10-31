package main

import "sync"

// Кэш in-memory
var Cache = struct {
	sync.RWMutex
	Orders map[string]Order
}{Orders: make(map[string]Order)}

// Добавление заказа в кэш
func SetCache(order Order) {
	Cache.Lock()
	defer Cache.Unlock()
	Cache.Orders[order.OrderUID] = order
}

// Получение заказа из кэша
func GetCache(id string) (Order, bool) {
	Cache.RLock()
	defer Cache.RUnlock()
	order, ok := Cache.Orders[id]
	return order, ok
}
