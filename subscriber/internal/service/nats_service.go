package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gookit/slog"
	"github.com/kstsm/nats-projetn/subscriber/internal/models"
	"github.com/kstsm/nats-projetn/subscriber/internal/repository"
	"github.com/nats-io/nats.go"
	"log"
	"os"
)

type NatsI interface {
	SubscribeOrders()
}

type Nats struct {
	nc   *nats.Conn
	repo repository.RepositoryI
}

func NewNats(repository repository.RepositoryI) NatsI {
	n := &Nats{
		nc:   initNats(),
		repo: repository,
	}
	return n
}

func initNats() *nats.Conn {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		slog.Fatal("Не удалось подключиться к NATS", err)
		os.Exit(1)
	}
	slog.Info("Успешное подключение к NATS")

	return nc
}

func (n *Nats) SubscribeOrders() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n.nc.Subscribe("orders", func(msg *nats.Msg) {
		var data models.Order

		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			slog.Error("Unmarshal", err)
		}

		message, err := n.repo.SaveMessageToDB(ctx, data)
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
