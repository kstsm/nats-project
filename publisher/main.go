package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

type Message struct {
	ID   int    `json:"id"`
	Data string `validate:"required"`
}

type Handler struct {
	nc *nats.Conn
}

func initNats() *nats.Conn {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		slog.Error("Не удалось подключиться к NATS", err)
	}
	slog.Info("Успешное подключение к NATS")

	return nc
}

var (
	storageMap = make(map[int]string)
	mapMutex   sync.Mutex
	mapRWMutex sync.RWMutex
)

func cachedMessage(messages *[]Message) {
	mapMutex.Lock()
	for _, message := range *messages {
		storageMap[message.ID] = message.Data
		slog.Info("Загружено в кэш", "ID", message.ID, "Data", message.Data)
	}
	slog.Info("Кэш полностью загрузился", "Всего записей", len(storageMap))
	mapMutex.Unlock()
}

func getMessages() (*[]Message, error) {
	var messages []Message

	resp, err := http.Get("http://localhost:8001/messages")
	if err != nil {
		slog.Error("Ошибка при выполнении GET-запроса", "url", "http://localhost:8001/messages", "error", err)
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&messages)
	if err != nil {
		slog.Error("Ошибка при декодировании JSON", "error", err)
		return nil, err
	}

	slog.Info("GET-запрос успешно выполнен", "Количество сообщений получено", len(messages))
	return &messages, nil
}

func (h *Handler) getMessageID(w http.ResponseWriter, r *http.Request) {
	idURL := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idURL)
	if err != nil {
		slog.Error("Ошибка конвертации строки в число", "Входящее", idURL, "error", err)
		http.Error(w, "Недопустимый формат идентификатора", http.StatusBadRequest)
		return
	}

	mapRWMutex.RLock()
	data, exists := storageMap[id]
	mapRWMutex.RUnlock()

	if exists {
		slog.Info("Данные найдены в кэше", "id", id, "data", data)
		w.Write([]byte(fmt.Sprintf("Номер запроса: %d\nДанные из кэша: %s", id, data)))
		return
	}

	url := fmt.Sprintf("http://localhost:8001/message/%d", id)
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("Ошибка при выполнении GET-запроса", "url", url, "error", err)
		http.Error(w, "Не удалось получить данные из базы данных", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Неожиданный статус-код ответа", "url", url, "status", resp.StatusCode)
		http.Error(w, "Не удалось получить данные из базы данных", http.StatusInternalServerError)
		return
	}

	var message Message
	if err = json.NewDecoder(resp.Body).Decode(&message); err != nil {
		slog.Error("Ошибка при декодировании JSON!!!!", "url", url, "error", err)
		http.Error(w, "Не удалось выполнить декодирование ответа", http.StatusInternalServerError)
		return
	}

	mapRWMutex.Lock()
	storageMap[message.ID] = message.Data
	mapRWMutex.Unlock()

	slog.Info("Данные успешно загружены и добавлены в кэш", "id", message.ID, "data", message.Data)
	w.Write([]byte(fmt.Sprintf("Данные загружены из базы данных:\nНомер запроса: %d\nДанные из кэша: %s", message.ID, message.Data)))
}

func (h *Handler) publishMessage(w http.ResponseWriter, r *http.Request) {
	const op = "publisher.publishMessage"
	var req Message

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Ошибка декодирования JSON", "error", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	message, err := validateMessage(req)
	if err != nil {
		slog.Error("Ошибка валидации", "message", req, "error", err)
		http.Error(w, "Ошибка валидации", http.StatusBadRequest)
		return
	}

	slog.Info("Данные успешно прошли валидацию", "message", message.Data)

	msg, err := h.nc.Request("orders", []byte(message.Data), nats.DefaultTimeout)
	if err != nil {
		slog.Error("Ошибка при запросе NATS", "message", message)
		http.Error(w, "Не удалось получить сообщение", http.StatusInternalServerError)
		return
	}

	var response Message
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		slog.Error("Ошибка декодирования ответа NATS", "error", err)
		http.Error(w, "Не удалось декодировать ответ", http.StatusInternalServerError)
		return
	}

	mapRWMutex.Lock()
	defer mapRWMutex.Unlock()
	storageMap[response.ID] = response.Data
	slog.Info("Кэш успешно обновлён", "id", response.ID, "data", response.Data)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Сообщение успешно обработано"))
}

func validateMessage(req Message) (Message, error) {
	validate := validator.New()

	err := validate.Struct(req)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			fmt.Printf("Ошибка в поле %s: %s\n", err.Field(), err.Tag())
		}
		return Message{}, err
	}

	return req, nil
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

	cachedMessage(messages)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/publish", handler.publishMessage)
	router.Get("/publish/{id}", handler.getMessageID)

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
