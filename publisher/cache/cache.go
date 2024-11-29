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

func StorageMessages(messages []models.Order) {
	mu.Lock()
	for _, message := range messages {
		StorageMap[message.OrderUID] = message
	}
	mu.Unlock()
	slog.Info("Кэш полностью загрузился", len(StorageMap))
}

func SetMessage(message models.Order) {
	mu.Lock()
	StorageMap[message.OrderUID] = message
	mu.Unlock()
}
