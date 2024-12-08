package handler

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gookit/slog"
	"github.com/kstsm/nats-projetn/subscriber/internal/helper"
	"github.com/kstsm/nats-projetn/subscriber/internal/models"
	"github.com/kstsm/nats-projetn/subscriber/internal/repository"
	"net/http"
)

type HandlerI interface {
	GetAllOrders(w http.ResponseWriter, r *http.Request)
	GetOrderByID(w http.ResponseWriter, r *http.Request)
	GetRouter() *chi.Mux
}

type Handler struct {
	//nc     *nats.Conn
	repo   repository.RepositoryI
	Router *chi.Mux
}

func NewHandler(repository repository.RepositoryI) HandlerI {
	h := &Handler{
		//nc:     nc,
		repo:   repository,
		Router: chi.NewRouter(),
	}

	h.Router.Use(middleware.Logger)
	h.Router.Get("/messages", h.GetAllOrders)
	h.Router.Get("/message/{id}", h.GetOrderByID)

	return h
}

func (h *Handler) GetRouter() *chi.Mux {
	return h.Router
}

func (h *Handler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orders, err := h.repo.GetAllOrders(ctx)
	if err != nil {
		http.Error(w, "Не удалось выполнить запрос к базе данных", http.StatusInternalServerError)
		slog.Error("Ошибка выполнения запроса", err)
		return
	}

	helper.ResponseJson(w, http.StatusOK, orders)
}

func (h *Handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bytes, err := h.repo.GetOrderByID(ctx, id)
	if err != nil {
		http.Error(w, "Не удалось выполнить запрос к базе данных", http.StatusInternalServerError)
		slog.Error("Ошибка выполнения запроса", err)
		return
	}

	var order models.Order
	err = json.Unmarshal(bytes, &order)
	if err != nil {
		http.Error(w, "Не удалось выполнить запрос к базе данных", http.StatusInternalServerError)
		slog.Error("Ошибка выполнения запроса", err)
		return
	}

	helper.ResponseJson(w, http.StatusOK, order)
}
