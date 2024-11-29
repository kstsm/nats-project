package repository

import (
	"context"
	"github.com/gookit/slog"
	"github.com/jackc/pgx/v5"
	"github.com/kstsm/nats-projetn/subscriber/internal/models"
	"github.com/kstsm/nats-projetn/subscriber/internal/models/queries"
)

func SaveMessageToDB(db *pgx.Conn, message models.Order) (models.Order, error) {
	var msg models.Order

	tx, err := db.Begin(context.Background())
	if err != nil {
		slog.Error("Ошибка при начале транзакции", "error", err)
		return models.Order{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(context.Background())
		}
	}()

	err = tx.QueryRow(
		context.Background(),
		queries.SaveOrdersToDB,
		message.TrackNumber, message.Entry, message.Locale,
		message.InternalSignature, message.CustomerID, message.DeliveryService,
		message.ShardKey, message.SmID, message.DateCreated, message.OofShard,
	).Scan(
		&msg.OrderUID, &msg.TrackNumber, &msg.Entry, &msg.Locale,
		&msg.InternalSignature, &msg.CustomerID, &msg.DeliveryService,
		&msg.ShardKey, &msg.SmID, &msg.DateCreated, &msg.OofShard,
	)
	if err != nil {
		slog.Error("Ошибка при сохранении сообщения в таблицу orders", "error", err)
		return models.Order{}, err
	}

	_, err = tx.Exec(
		context.Background(),
		queries.SavePaymentToDB,
		message.Payment.Transaction, message.Payment.RequestID, message.Payment.Currency,
		message.Payment.Provider, message.Payment.Amount, message.Payment.PaymentDT,
		message.Payment.Bank, message.Payment.DeliveryCost, message.Payment.GoodsTotal,
		message.Payment.CustomFee, msg.OrderUID, // Используем OrderUID из вставки в orders
	)
	if err != nil {
		slog.Error("Ошибка при сохранении сообщения в таблицу payment", "order_uid", msg.OrderUID, "error", err)
		return models.Order{}, err
	}

	_, err = tx.Exec(
		context.Background(),
		queries.SavedDeliveryToDB,
		message.Delivery.Name, message.Delivery.Phone, message.Delivery.Zip,
		message.Delivery.City, message.Delivery.Address, message.Delivery.Region,
		message.Delivery.Email, msg.OrderUID,
	)
	if err != nil {
		slog.Error("Ошибка при сохранении сообщения в таблицу delivery", "order_uid", msg.OrderUID, "error", err)
		return models.Order{}, err
	}

	if err = tx.Commit(context.Background()); err != nil {
		slog.Error("Ошибка при коммите транзакции", "order_uid", msg.OrderUID, "error", err)
		return models.Order{}, err
	}

	return msg, nil
}
