package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type Handler struct {
	nc *nats.Conn
	db *pgx.Conn
}

type Message struct {
	ID   int    `json:"id"`
	Data string `json:"data"`
}

func initPostgres() *pgx.Conn {
	db, err := pgx.Connect(context.Background(), "postgres://admin:admin@localhost:5432/message_db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	return db
}

func initNats() *nats.Conn {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}

	return nc
}

func (h *Handler) GetAllMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")

	rows, err := h.db.Query(context.Background(), "SELECT id, data FROM message")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var msg Message
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

	var msg Message
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

func saveMessageToDB(db *pgx.Conn, message Message) (Message, error) {
	var msg Message

	err := db.
		QueryRow(context.Background(), `INSERT INTO message (data) VALUES ($1) RETURNING id, data`, message.Data).
		Scan(&msg.ID, &msg.Data)
	if err != nil {
		fmt.Println(err)
		return Message{}, err
	}

	return msg, nil
}

func main() {
	nc := initNats()
	db := initPostgres()

	handler := Handler{
		nc: nc,
		db: db,
	}

	handler.nc.Subscribe("orders", func(msg *nats.Msg) {
		var data Message
		_ = json.Unmarshal(msg.Data, &data)

		message, err := saveMessageToDB(db, data)
		if err != nil {
			return
		}

		response, err := json.Marshal(message)
		if err != nil {
			log.Println("Ошибка сериализации:", err)
			return
		}

		err = msg.Respond(response)
		if err != nil {
			return
		}

		fmt.Printf("Получено сообщение: %s\n", string(msg.Data))
	})

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/messages", handler.GetAllMessage)
	router.Get("/message/{id}", handler.getMessageByID)

	srv := http.Server{
		Addr:    ":8001",
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
