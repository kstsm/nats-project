package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/kstsm/nats-projetn/subscriber/internal/models"
)

func SaveMessageToDB(db *pgx.Conn, message models.Message) (models.Message, error) {
	var msg models.Message

	err := db.
		QueryRow(context.Background(), `INSERT INTO message (data) VALUES ($1) RETURNING id, data`, message.Data).
		Scan(&msg.ID, &msg.Data)
	if err != nil {
		fmt.Println(err)
		return models.Message{}, err
	}

	return msg, nil
}
