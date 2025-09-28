package task

import (
	"TaskService/internal/dto"
	"TaskService/internal/service"
	"TaskService/internal/service/task"
	"encoding/json"
	"errors"
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
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks/{id} [get]
func (h *Handler) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := h.service.Task().Get(r.Context(), id)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "Task not found")
		return
	}

	writeJSONResponse(w, http.StatusOK, task)
}

// GetTaskListHandler возвращает список всех задач
// @Summary Get all tasks
// @Description Get list of all tasks
// @Tags tasks
// @Accept json
// @Produce json
// @Success 200 {object} dto.GetTaskListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks [get]
func (h *Handler) GetTaskListHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.Task().GetList(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get tasks")
		return
	}

	writeJSONResponse(w, http.StatusOK, tasks)
}

// CreateTaskHandler создает новую задачу
// @Summary Create a new task
// @Description Create a new task with title and description
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body dto.CreateTaskRequest true "Task creation data"
// @Success 201 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks [post]
func (h *Handler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Title is required")
		return
	}

	if err := h.service.Task().Create(r.Context(), req); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	writeJSONResponse(w, http.StatusCreated, dto.NewSuccessResponse("Task created successfully"))
}

// UpdateTaskHandler обновляет существующую задачу
// @Summary Update a task
// @Description Update an existing task
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body dto.UpdateTaskRequest true "Task update data"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks [put]
func (h *Handler) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ID == 0 {
		writeErrorResponse(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	if req.Title == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Title is required")
		return
	}

	err := h.service.Task().Update(r.Context(), req)
	if err != nil {
		if errors.Is(err, task.ErrInvalidStatus) {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid status")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to update task")
		return
	}

	writeJSONResponse(w, http.StatusOK, dto.NewSuccessResponse("Task updated successfully"))
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	writeJSONResponse(w, statusCode, dto.NewErrorResponse(message))
}
