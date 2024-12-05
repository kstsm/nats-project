package service

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/slog"
	"github.com/jackc/pgx/v5"
	"github.com/kstsm/nats-projetn/subscriber/internal/models"
	"github.com/kstsm/nats-projetn/subscriber/internal/repository"
	"github.com/nats-io/nats.go"
	"log"
)

func SubscribeOrders(db *pgx.Conn, nc *nats.Conn) {
	nc.Subscribe("orders", func(msg *nats.Msg) {
		var data models.Order

		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			slog.Error("Unmarshal", err)
		}

		message, err := repository.SaveMessageToDB(db, data)
		if err != nil {
			slog.Error("Ошибка!", err)
			return
		}

		response, err := json.Marshal(message)
		if err != nil {
			log.Println("Ошибка сериализации:", err)
			return
		}

		err = msg.Respond(response)
		if err != nil {
			return
		}

		fmt.Printf("Получено сообщение: %s\n", string(msg.Data))
	})
}
