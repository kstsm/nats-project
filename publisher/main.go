package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"io"
	"log"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
)

type Message struct {
	ID   int    `json:"id"`
	Data string `json:"data"`
}

type Handler struct {
	nc *nats.Conn
}

func initNats() *nats.Conn {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}

	return nc
}

func (h *Handler) publishMessage(w http.ResponseWriter, r *http.Request) {
	const op = "publisher.publishMessage"

	var req Message
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Errorf("%s: %w", op, err)
		return
	}

	if err := h.nc.Publish("orders", []byte(req.Data)); err != nil {
		fmt.Errorf("%s: %w", op, err)
		return
	}

}

var storageMap = make(map[int]string)

func cachedMessage(messages *[]Message) {
	for _, message := range *messages {
		storageMap[message.ID] = message.Data
	}
	fmt.Println("Кэш загрузился:", storageMap)
}

func getMessages() (*[]Message, error) {
	var messages []Message

	resp, err := http.Get("http://localhost:8001/messages")
	if err != nil {
		fmt.Println(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(body, &messages)
	if err != nil {
		return nil, err
	}

	return &messages, nil
}

func (h *Handler) getMessageID(w http.ResponseWriter, r *http.Request) {
	idURL := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idURL)
	if err != nil {
		fmt.Println(err)
	}
	var message Message

	value, exists := storageMap[id]
	if exists {
		fmt.Println("Выгрузка данных из кэша ID:", id, "Data:", storageMap[id])

	} else {
		url := fmt.Sprintf("http://localhost:8001/message/%d", id)
		fmt.Println("Данные из кэша не удалось подгрузить", value)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		err = json.Unmarshal(body, &message)
		if err != nil {
			fmt.Println("Никаких данных из БД не пришло:", err)
			return
		}

		fmt.Printf("Подгрузка из БД ID: %d, Data: %s\n", message.ID, message.Data)
		storageMap[message.ID] = message.Data
		fmt.Println("Кэш обновлен:", storageMap)
	}
}

func main() {

	nc := initNats()

	handler := Handler{
		nc: nc,
	}

	messages, err := getMessages()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	cachedMessage(messages)

	handler.nc.Subscribe("orders", func(m *nats.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	})

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
