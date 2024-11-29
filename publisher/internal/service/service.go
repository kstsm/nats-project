package service

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/slog"
	"github.com/kstsm/nats-project/pablisher/configs"
	"github.com/kstsm/nats-project/pablisher/internal/models"
	"net/http"
)

var cfg = configs.Config

func GetMessages() ([]models.Order, error) {
	var messages []models.Order

	resp, err := http.Get(fmt.Sprintf("%s/messages", cfg.SubscriberAddr))
	if err != nil {
		slog.Error("Ошибка при выполнении GET-запроса", "url", "http://localhost:8001/messages", "error", err)
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&messages)
	if err != nil {
		slog.Error("Ошибка при декодировании JSON", "error", err)
		return nil, err
	}

	return messages, nil
}
