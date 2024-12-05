package handler

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gookit/slog"
	"github.com/kstsm/nats-project/pablisher/cache"
	"github.com/kstsm/nats-project/pablisher/internal/models"
	"github.com/nats-io/nats.go"
	"net/http"
)

type Handler struct {
	nc     *nats.Conn
	Router *chi.Mux
}

func NewHandler(nc *nats.Conn) *Handler {
	h := &Handler{
		nc:     nc,
		Router: chi.NewRouter(),
	}

	h.Router.Use(middleware.Logger)
	h.Router.Post("/publish", h.saveOrder)
	h.Router.Get("/publish/{id}", h.getOrderByID)

	return h
}

func (h *Handler) saveOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request models.Order

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("Ошибка декодирования JSON", "error", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(request)
	if err != nil {
		slog.Error("getOrderByID: json.Marshal")
		http.Error(w, "Не удалось выполнить декодирование ответа", http.StatusInternalServerError)
		return
	}

	msg, err := h.nc.Request("orders", data, nats.DefaultTimeout)
	if err != nil {
		slog.Error("Ошибка при запросе NATS", data)
		http.Error(w, "Не удалось получить сообщение", http.StatusInternalServerError)
		return
	}
	var message models.Order
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		slog.Error("Ошибка декодирования ответа NATS", "error", err)
		http.Error(w, "Не удалось декодировать ответ", http.StatusInternalServerError)
		return
	}

	cache.SetOrder(message)
	slog.Info("Кэш успешно обновлён", "id", message.OrderUID, "data", message)

	response := msg.Data

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *Handler) getOrderByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req := chi.URLParam(r, "id")

	id, err := uuid.Parse(req)
	if err != nil {
		slog.Error("Ошибка конвертации строки в UUID", err)
		http.Error(w, "Недопустимый формат идентификатора", http.StatusBadRequest)
		return
	}

	data, exists := cache.StorageMap[id]
	if exists {
		fmt.Println("qwe")

		json.NewEncoder(w).Encode(data)
		return
	}

	url := fmt.Sprintf("http://localhost:8001/message/%d", id)
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("Ошибка при выполнении GET-запроса", "url", url, "error", err)
		http.Error(w, "Не удалось получить данные из базы данных", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Неожиданный статус-код ответа", "url", url, "status", resp.StatusCode)
		http.Error(w, "Не удалось получить данные из базы данных", http.StatusInternalServerError)
		return
	}

	var message models.Order
	if err = json.NewDecoder(resp.Body).Decode(&message); err != nil {
		slog.Error("Ошибка при декодировании JSON", "url", url, "error", err)
		http.Error(w, "Не удалось выполнить декодирование ответа", http.StatusInternalServerError)
		return
	}

	cache.SetOrder(message)

	response, err := json.Marshal(message)
	if err != nil {
		slog.Error("getOrderByID: json.Marshal")
		http.Error(w, "Не удалось выполнить декодирование ответа", http.StatusInternalServerError)
		return
	}

	slog.Info("Данные успешно загружены и добавлены в кэш", "id", message.OrderUID, "data", message)
	w.Write(response)
}
