package cache

import (
	"github.com/google/uuid"
	"github.com/gookit/slog"
	"github.com/kstsm/nats-project/pablisher/internal/models"
	"sync"
)

var (
	StorageMap = make(map[uuid.UUID]models.Order)
	mu         = sync.Mutex{}
)

func StorageOrders(orders []models.Order) {
	mu.Lock()
	for _, order := range orders {
		StorageMap[order.OrderUID] = order
	}
	mu.Unlock()
	slog.Info("Кэш полностью загрузился", len(StorageMap))
}

func SetOrder(order models.Order) {
	mu.Lock()
	StorageMap[order.OrderUID] = order
	mu.Unlock()
}
