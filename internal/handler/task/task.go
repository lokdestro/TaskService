package task

import (
	"TaskService/internal/dto"
	"TaskService/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service service.Service
}

func New(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetTaskHandler возвращает задачу по ID
// @Summary Get task by ID
// @Description Get task details by task ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} dto.GetTaskResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tasks/{id} [get]
func (h *Handler) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := h.service.Task().Get(r.Context(), id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// GetTaskListHandler возвращает список всех задач
// @Summary Get all tasks
// @Description Get list of all tasks
// @Tags tasks
// @Accept json
// @Produce json
// @Success 200 {object} dto.GetTaskListResponse
// @Failure 500 {object} map[string]string
// @Router /tasks [get]
func (h *Handler) GetTaskListHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.Task().GetList(r.Context())
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// CreateTaskHandler создает новую задачу
// @Summary Create a new task
// @Description Create a new task with title and description
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body dto.CreateTaskRequest true "Task creation data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tasks [post]
func (h *Handler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if err := h.service.Task().Create(r.Context(), req); err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task created successfully"})
}

// UpdateTaskHandler обновляет существующую задачу
// @Summary Update a task
// @Description Update an existing task
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body dto.UpdateTaskRequest true "Task update data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tasks [put]
func (h *Handler) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == 0 {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if err := h.service.Task().Update(r.Context(), req); err != nil {
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Task updated successfully"})
}
