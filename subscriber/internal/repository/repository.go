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
		slog.Error("Ошибка при начале транзакции:", err)
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
		slog.Error("Ошибка при сохранении сообщения в таблицу orders:", err)
		return models.Order{}, err
	}

	err = tx.QueryRow(
		context.Background(),
		queries.SavedDeliveryToDB,
		msg.OrderUID, message.Delivery.Name, message.Delivery.Phone, message.Delivery.Zip,
		message.Delivery.City, message.Delivery.Address, message.Delivery.Region,
		message.Delivery.Email,
	).Scan(
		&msg.Delivery.Name, &msg.Delivery.Phone, &msg.Delivery.Zip, &msg.Delivery.City,
		&msg.Delivery.Address, &msg.Delivery.Region, &msg.Delivery.Email,
	)
	if err != nil {
		slog.Error("Ошибка при сохранении сообщения в таблицу delivery:", err)
		return models.Order{}, err
	}

	err = tx.QueryRow(
		context.Background(),
		queries.SavePaymentToDB,
		msg.OrderUID, msg.OrderUID, message.Payment.Currency,
		message.Payment.Provider, message.Payment.Amount, message.Payment.PaymentDT,
		message.Payment.Bank, message.Payment.DeliveryCost, message.Payment.GoodsTotal,
		message.Payment.CustomFee,
	).Scan(
		&msg.Payment.Transaction, &msg.Payment.Currency,
		&msg.Payment.Provider, &msg.Payment.Amount, &msg.Payment.PaymentDT,
		&msg.Payment.Bank, &msg.Payment.DeliveryCost, &msg.Payment.GoodsTotal,
		&msg.Payment.CustomFee)
	if err != nil {
		slog.Error("Ошибка при сохранении сообщения в таблицу payment:", err)
		return models.Order{}, err
	}

	for _, item := range message.Items {
		err = tx.QueryRow(
			context.Background(),
			queries.SaveItemsToDB,
			msg.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NMID, item.Brand, item.Status,
		).Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NMID, &item.Brand, &item.Status,
		)
		if err != nil {
			slog.Error("Ошибка при сохранении сообщения в таблицу items:", err)
			return models.Order{}, err
		}
		msg.Items = append(msg.Items, item)
	}

	if err = tx.Commit(context.Background()); err != nil {
		slog.Error("Ошибка при коммите транзакции:", err)
		return models.Order{}, err
	}

	return msg, nil
}
