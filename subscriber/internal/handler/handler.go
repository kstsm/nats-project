package handler

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gookit/slog"
	"github.com/jackc/pgx/v5"
	"github.com/kstsm/nats-projetn/subscriber/internal/models"
	"github.com/kstsm/nats-projetn/subscriber/internal/models/queries"
	"github.com/nats-io/nats.go"
	"net/http"
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
	h.Router.Get("/messages", h.GetAllOrders)
	h.Router.Get("/message/{id}", h.getOrderByID)

	return h
}

func (h *Handler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := context.Background()
	var orders []models.Order
	var bytes []byte

	err := h.db.QueryRow(ctx, queries.GetAllOrders).Scan(&bytes)
	if err != nil {
		http.Error(w, "Не удалось выполнить запрос к базе данных", http.StatusInternalServerError)
		slog.Error("Ошибка выполнения запроса", err)
		return
	}

	err = json.Unmarshal(bytes, &orders)
	if err != nil {
		http.Error(w, "Не удалось выполнить запрос к базе данных", http.StatusInternalServerError)
		slog.Error("Ошибка выполнения запроса", err)
		return
	}

	if err = json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Ошибка при кодировании JSON", http.StatusInternalServerError)
		slog.Error("Ошибка JSON кодирования", err)
	}
}

func (h *Handler) getOrderByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")

	id := chi.URLParam(r, "id")

	ctx := context.Background()
	var order models.Order
	var bytes []byte

	err := h.db.QueryRow(ctx, queries.GetOrderByID, id).Scan(&bytes)
	if err != nil {
		http.Error(w, "Не удалось выполнить запрос к базе данных", http.StatusInternalServerError)
		slog.Error("Ошибка выполнения запроса", err)
		return
	}

	err = json.Unmarshal(bytes, &order)
	if err != nil {
		http.Error(w, "Не удалось выполнить запрос к базе данных", http.StatusInternalServerError)
		slog.Error("Ошибка выполнения запроса", err)
		return
	}

	if err = json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Ошибка при кодировании JSON", http.StatusInternalServerError)
		slog.Error("Ошибка JSON кодирования", err)
	}
}
