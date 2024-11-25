package cache

import (
	"github.com/gookit/slog"
	"github.com/kstsm/nats-project/pablisher/internal/models"
	"sync"
)

var (
	StorageMap = make(map[int]models.Message)
	mu         = sync.Mutex{}
)

func StorageMessages(messages []models.Message) {
	mu.Lock()
	for _, message := range messages {
		StorageMap[message.ID] = message
	}
	mu.Unlock()
	slog.Info("Кэш полностью загрузился", len(StorageMap))
}

func SetMessage(message models.Message) {
	mu.Lock()
	StorageMap[message.ID] = message
	mu.Unlock()
}
