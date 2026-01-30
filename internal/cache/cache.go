package cache

import (
	"context"
	"log"
	"sync"
	"time"
	"wb-project/internal/metric"
	"wb-project/internal/models"
)

// Реализовать кэширование данных в сервисе: хранить последние полученные
// данные заказов в памяти (например, в map), чтобы быстро выдавать их по запросу.
type cacheItem struct {
	data      *models.Order
	expiresAt int64
}

type OrderCache struct {
	items             map[string]cacheItem
	defaultExpiration time.Duration //Это стандартное время жизни.
	cleanupInterval   time.Duration //Это частота работы нашего "уборщика", который чистит кеш
	sync.RWMutex
	ticker *time.Ticker
}

func NewOrderCache(defaultExpiration, cleanupInterval time.Duration) *OrderCache {
	c := &OrderCache{
		items:             make(map[string]cacheItem),
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		ticker:            time.NewTicker(cleanupInterval),
	}
	return c
}

func (ch *OrderCache) Set(uid string, order *models.Order) {
	ch.Lock()
	defer ch.Unlock()
	_, exists := ch.items[uid]
	//При сохранении указываем время жизни, когда нужно удалить объект
	expiration := time.Now().Add(ch.defaultExpiration).UnixNano()
	ch.items[uid] = cacheItem{
		data:      order,
		expiresAt: expiration,
	}
	if !exists {
		metric.CacheSize.Inc()
	}
	log.Printf("Добавли в кеш: %s", uid)
}

func (ch *OrderCache) Get(uid string) (*models.Order, bool) {
	ch.RLock()
	defer ch.RUnlock()

	res, ok := ch.items[uid]
	if !ok {
		return nil, false
	}

	// Если ключ есть, проверяем, не протух ли он
	if time.Now().UnixNano() > res.expiresAt {
		return nil, false
	}

	return res.data, true
}

func (ch *OrderCache) GC(ctx context.Context) error {
	log.Println("Начинаем проверку кеша")
	for {
		select {
		case <-ch.ticker.C:
			ch.Lock()
			// ... удаление просроченных ключей ...
			now := time.Now().UnixNano() //текущее время в UnixNano
			deletedCounter := 0
			for key, item := range ch.items { //
				if now > item.expiresAt { //проверка, что настало время очистки
					metric.CacheSize.Dec()
					delete(ch.items, key) //удаление данных их кеша
					deletedCounter++
				}
			}
			if deletedCounter > 0 {
				log.Printf("GC: удалено %d просроченных записей", deletedCounter)
			}
			ch.Unlock()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (ch *OrderCache) Stop() {
	defer ch.ticker.Stop()
}
