package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/kstsm/nats-projetn/subscriber/internal/models"
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

	rows, err := h.db.Query(context.Background(), "SELECT id, data FROM message")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer rows.Close()

	var messages []models.Message

	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.Data); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
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

	var msg models.Message
	err = h.db.QueryRow(context.Background(), "SELECT id, data FROM message WHERE id = $1", id).Scan(&msg.ID, &msg.Data)
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
