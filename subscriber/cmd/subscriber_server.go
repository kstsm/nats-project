package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/gookit/slog"
	"github.com/jackc/pgx/v5"
	"github.com/kstsm/nats-projetn/subscriber/internal/handler"
	"github.com/kstsm/nats-projetn/subscriber/internal/service"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

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

func Run() {
	nc := initNats()
	db := initPostgres()

	router := handler.NewHandler(nc, db)

	service.SubscribeOrders(db, nc)

	srv := http.Server{
		Addr:    ":8001",
		Handler: router.Router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("Starting server...")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Fatal("listen and serve returned err:", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("got interruption signal")
	if err := srv.Shutdown(context.TODO()); err != nil {
		slog.Info("server shutdown returned an err:", err)
	}

	slog.Info("final")
}
