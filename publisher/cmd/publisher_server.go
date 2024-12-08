package cmd

import (
	"context"
	"errors"
	"github.com/gookit/slog"
	"github.com/kstsm/nats-project/pablisher/cache"
	"github.com/kstsm/nats-project/pablisher/internal/handler"
	"github.com/kstsm/nats-project/pablisher/internal/service"
	"github.com/nats-io/nats.go"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func initNats() *nats.Conn {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		slog.Fatal("Не удалось подключиться к NATS", err)
		os.Exit(1)
	}
	slog.Info("Успешное подключение к NATS")

	return nc
}

func Run() {
	nc := initNats()

	router := handler.NewHandler(nc)

	messages, err := service.GetOrders()
	if err != nil {
		slog.Error("Ошибка при получении сообщений", "error", err)
		return
	}

	cache.StorageOrders(messages)

	srv := http.Server{
		Addr:    ":8000",
		Handler: router.GetRouter(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("Starting server...")
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Fatal("listen and serve returned err:", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("got interruption signal")
	if err = srv.Shutdown(context.TODO()); err != nil {
		slog.Info("server shutdown returned an err:", err)
	}

	slog.Info("final")
}
