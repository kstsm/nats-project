package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gookit/slog"
	"github.com/kstsm/nats-project/pablisher/configs"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

var cfg = configs.Config

type Message struct {
	ID   int    `json:"id"`
	Data string `validate:"required"`
}

type Handler struct {
	nc *nats.Conn
}

func setMessage(message Message) {
	mu.Lock()
	storageMap[message.ID] = message
	mu.Unlock()
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

var (
	storageMap = make(map[int]Message)
	mu         = sync.Mutex{}
)

func cacheMessages(messages []Message) {
	mu.Lock()
	for _, message := range messages {
		storageMap[message.ID] = message
	}
	mu.Unlock()
	slog.Info("Кэш полностью загрузился", len(storageMap))
}

// TODO: gRPC
func getMessages() ([]Message, error) {
	var messages []Message

	// Перенести в конфиг ссылку
	resp, err := http.Get(fmt.Sprintf("%s/messages", cfg.SubscriberAddr))
	if err != nil {
		slog.Error("Ошибка при выполнении GET-запроса", "url", "http://localhost:8001/messages", "error", err)
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&messages)
	if err != nil {
		slog.Error("Ошибка при декодировании JSON", "error", err)
		return nil, err
	}

	return messages, nil
}

func (h *Handler) getMessageByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	messageID := chi.URLParam(r, "id")

	id, err := strconv.Atoi(messageID)
	if err != nil {
		slog.Error("Ошибка конвертации строки в число", "Входящее", messageID, "error", err)
		http.Error(w, "Недопустимый формат идентификатора", http.StatusBadRequest)
		return
	}

	data, exists := storageMap[id]
	if exists {
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

	var message Message
	if err = json.NewDecoder(resp.Body).Decode(&message); err != nil {
		slog.Error("Ошибка при декодировании JSON", "url", url, "error", err)
		http.Error(w, "Не удалось выполнить декодирование ответа", http.StatusInternalServerError)
		return
	}

	setMessage(message)

	response, err := json.Marshal(message)
	if err != nil {
		slog.Error("getMessageByID: json.Marshal")
		http.Error(w, "Не удалось выполнить декодирование ответа", http.StatusInternalServerError)
		return
	}

	slog.Info("Данные успешно загружены и добавлены в кэш", "id", message.ID, "data", message.Data)
	w.Write(response)
}

func (h *Handler) publishMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request Message

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("Ошибка декодирования JSON", "error", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(request)
	if err != nil {
		slog.Error("getMessageByID: json.Marshal")
		http.Error(w, "Не удалось выполнить декодирование ответа", http.StatusInternalServerError)
		return
	}

	msg, err := h.nc.Request("orders", data, nats.DefaultTimeout)
	if err != nil {
		slog.Error("Ошибка при запросе NATS", data)
		http.Error(w, "Не удалось получить сообщение", http.StatusInternalServerError)
		return
	}

	var message Message
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		slog.Error("Ошибка декодирования ответа NATS", "error", err)
		http.Error(w, "Не удалось декодировать ответ", http.StatusInternalServerError)
		return
	}

	setMessage(message)
	slog.Info("Кэш успешно обновлён", "id", message.ID, "data", message.Data)

	response := msg.Data

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func main() {
	nc := initNats()

	handler := Handler{
		nc: nc,
	}

	messages, err := getMessages()
	if err != nil {
		slog.Error("Ошибка при получении сообщений", "error", err)
		return
	}

	cacheMessages(messages)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/publish", handler.publishMessage)
	router.Get("/publish/{id}", handler.getMessageByID)

	srv := http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("Starting server...")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve returned err: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("got interruption signal")
	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Printf("server shutdown returned an err: %v\n", err)
	}

	log.Println("final")
}
