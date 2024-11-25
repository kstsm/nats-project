package service

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/kstsm/nats-projetn/subscriber/internal/models"
	"github.com/kstsm/nats-projetn/subscriber/internal/repository"
	"github.com/nats-io/nats.go"
	"log"
)

func SubscribeOrders(db *pgx.Conn, nc *nats.Conn) {
	nc.Subscribe("orders", func(msg *nats.Msg) {
		var data models.Message
		_ = json.Unmarshal(msg.Data, &data)

		message, err := repository.SaveMessageToDB(db, data)
		if err != nil {
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
