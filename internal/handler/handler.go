package handler

import (
	"TaskService/internal/handler/task"
	"TaskService/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
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

	handler.router.Route("/tasks", func(r chi.Router) {
		r.Get("/", taskHandler.GetTaskListHandler) // GET /tasks - получить список задач
		r.Post("/", taskHandler.CreateTaskHandler) // POST /tasks - создать задачу
		r.Put("/", taskHandler.UpdateTaskHandler)  // PUT /tasks - обновить задачу
		r.Get("/{id}", taskHandler.GetTaskHandler) // GET /tasks/{id} - получить задачу по ID
	})

	return handler.router
}
