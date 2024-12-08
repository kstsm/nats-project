package cmd

import (
	"context"
	"errors"
	"github.com/gookit/slog"
	"github.com/kstsm/nats-projetn/subscriber/internal/handler"
	"github.com/kstsm/nats-projetn/subscriber/internal/repository"
	"github.com/kstsm/nats-projetn/subscriber/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	repo := repository.NewRepository()
	natsService := service.NewNats(repo)
	router := handler.NewHandler(repo)

	// Подписка на топик orders
	natsService.SubscribeOrders()

	srv := http.Server{
		Addr:    ":8001",
		Handler: router.GetRouter(),
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
