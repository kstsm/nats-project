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
	"log"
	"net/http"
	"os/signal"
	"syscall"
)

type Handler struct {
	nc *nats.Conn
}

func (h *Handler) getMessage(w http.ResponseWriter, r *http.Request) {

}

type MessageRequest struct {
	Data string `json:"data"`
}

func (h *Handler) publishMessage(w http.ResponseWriter, r *http.Request) {
	const op = "publisher.publishMessage"

	var req MessageRequest
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

func initNats() *nats.Conn {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}

	return nc
}

func main() {
	handler := Handler{
		nc: initNats(),
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/publish", handler.publishMessage)
	router.Get("/publish/{id}", handler.getMessage)

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
