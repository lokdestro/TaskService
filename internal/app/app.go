package app

import (
	"TaskService/pkg/kafka"
	"TaskService/pkg/logger"
	"context"
	"errors"
	"fmt"
	"net/http"

	"TaskService/config"
	"TaskService/internal/handler"
	"TaskService/internal/service"
	"TaskService/internal/storage"
)

type App struct {
	server *http.Server
	kc     kafka.Kafka
}

func New() (*App, error) {
	db, err := storage.New(config.Psql())
	if err != nil {
		return nil, err
	}

	kc, err := kafka.NewKafkaClient(config.Kfk())
	if err != nil {
		return nil, err
	}

	srv := service.New(db, kc)

	go srv.Task().ProcessTasks()

	result := &App{
		server: &http.Server{
			Addr:    config.Srv(),
			Handler: handler.New(srv),
		},
		kc: kc,
	}

	return result, nil
}

func (a *App) Run(ctx context.Context) error {
	log := logger.Get()

	go func() {
		select {
		case <-ctx.Done():
			fmt.Println("Shutting down the server...")

			err := a.server.Shutdown(context.Background())
			if err != nil {
				fmt.Println(err)
			}

			err = a.kc.Close()
			if err != nil {
				log.Info().Err(err).Msg("Kafka client connection closed.")
				return
			}

			fmt.Println("Server shutting down successfully")

			return
		}
	}()

	log.Info().Msg("Server started")

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	log.Info().Msg("Server stopped")

	return nil
}
