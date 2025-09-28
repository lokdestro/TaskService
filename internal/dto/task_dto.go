package dto

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GetTaskResponse struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type GetTaskListResponse struct {
	Tasks []GetTaskResponse `json:"tasks"`
}

type UpdateTaskRequest struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}
