package handler

import (
	"net/http"

	"TaskService/internal/handler/task"
	"TaskService/internal/service"

	_ "TaskService/docs"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handler struct {
	srv    service.Service
	router chi.Router
}

func New(srv service.Service) http.Handler {
	handler := &Handler{
		srv:    srv,
		router: chi.NewRouter(),
	}

	taskHandler := task.New(srv)

	handler.router.Get("/swagger/*", httpSwagger.Handler())

	handler.router.Route("/tasks", func(r chi.Router) {
		r.Get("/", taskHandler.GetTaskListHandler)
		r.Post("/", taskHandler.CreateTaskHandler)
		r.Put("/", taskHandler.UpdateTaskHandler)
		r.Get("/{id}", taskHandler.GetTaskHandler)
	})

	return handler.router
}
