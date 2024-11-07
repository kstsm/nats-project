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

func getMessages() map[int]string {
	var messages []Message

	resp, err := http.Get("http://localhost:8001/messages")
	if err != nil {
		fmt.Println(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(body, &messages)

	for _, message := range messages {
		storageMap[message.ID] = message.Data
	}
	return storageMap
}

func (h *Handler) getMessageID(w http.ResponseWriter, r *http.Request) {
	url := chi.URLParam(r, "id")
	id, err := strconv.Atoi(url)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Сообщение под номером:", storageMap[id])
}

func main() {

	nc := initNats()

	handler := Handler{
		nc: nc,
	}
	fmt.Println(getMessages())

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
