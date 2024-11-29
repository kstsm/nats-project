package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gookit/slog"
	"github.com/jackc/pgx/v5"
	"github.com/kstsm/nats-projetn/subscriber/internal/models"
	"github.com/kstsm/nats-projetn/subscriber/internal/models/queries"
	"github.com/nats-io/nats.go"
	"net/http"
	"strconv"
)

type Handler struct {
	nc     *nats.Conn
	db     *pgx.Conn
	Router *chi.Mux
}

func NewHandler(nc *nats.Conn, db *pgx.Conn) *Handler {
	h := &Handler{
		nc:     nc,
		db:     db,
		Router: chi.NewRouter(),
	}

	h.Router.Use(middleware.Logger)
	h.Router.Get("/messages", h.GetAllMessage)
	h.Router.Get("/message/{id}", h.getMessageByID)

	return h
}

func (h *Handler) GetAllMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")

	rows, err := h.db.Query(context.Background(), queries.GetAllMessage)
	if err != nil {
		slog.Error("Не удалось выполнить запрос к базе данных", err)
		http.Error(w, "Не удалось выполнить запрос к базе данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []models.Order

	for rows.Next() {
		var msg models.Order
		if err = rows.Scan(
			&msg.TrackNumber, &msg.Entry, &msg.Locale, &msg.InternalSignature,
			&msg.CustomerID, &msg.DeliveryService, &msg.ShardKey, &msg.SmID,
			&msg.DateCreated, &msg.OofShard, &msg.Payment.Transaction,
			&msg.Payment.RequestID, &msg.Payment.Currency, &msg.Payment.Provider,
			&msg.Payment.Amount, &msg.Payment.PaymentDT, &msg.Payment.Bank,
			&msg.Payment.DeliveryCost, &msg.Payment.GoodsTotal,
			&msg.Payment.CustomFee, &msg.OrderUID, &msg.Delivery.Name, &msg.Delivery.Phone,
			&msg.Delivery.Zip, &msg.Delivery.City, &msg.Delivery.Address, &msg.Delivery.Region,
			&msg.Delivery.Email); err != nil {
			slog.Error("Sub.handler.GetAllMessage", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}
	if err = rows.Err(); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(messages)
	if err != nil {
		fmt.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}

func (h *Handler) getMessageByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")

	url := chi.URLParam(r, "id")
	id, err := strconv.Atoi(url)
	if err != nil {
		fmt.Println(err)
	}

	var msg models.Order
	err = h.db.QueryRow(context.Background(), "SELECT orders.order_uid,payment.order_uid, delivery.order_uid FROM orders,delivery,payment WHERE orders.order_uid = $1", id).Scan(&msg.OrderUID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}
