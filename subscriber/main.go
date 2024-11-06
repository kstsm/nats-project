package main

import (
	"context"
	"database/sql"
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
	db *sql.DB
}

func initPostgres() *sql.DB {
	db, err := sql.Open("postgres", "user=admin_wb password=admin_wb dbname=message_db sslmode=disable")
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
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

func storageMessageToDB(db *sql.DB, storageMap map[string]string) {

	rows, _ := db.Query("SELECT id, data FROM message")

	for rows.Next() {
		var id string
		var data string

		if err := rows.Scan(&id, &data); err != nil {
			log.Fatal(err)
		}

		storageMap[id] = data

	}

	for id, name := range storageMap {
		fmt.Printf("ID: %s, Data: %s\n", id, name)
	}

}

func saveMessageToDB(db *sql.DB, message string) error {
	stmt, err := db.Prepare(`INSERT INTO message (data) VALUES ($1)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(message)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	handler := Handler{
		nc: initNats(),
		db: initPostgres(),
	}

	storageMap := make(map[string]string)

	handler.nc.Subscribe("orders", func(msg *nats.Msg) {

		err := saveMessageToDB(handler.db, string(msg.Data))
		if err != nil {
			return
		}
		storageMessageToDB(handler.db, storageMap)

		fmt.Printf("Получено сообщение: %s\n", string(msg.Data))

	})

	router := chi.NewRouter()
	router.Use(middleware.Logger)

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
