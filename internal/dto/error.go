package dto

type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid task ID"`
	Message string `json:"message,omitempty" example:"Detailed error description"`
}

type SuccessResponse struct {
	Message string `json:"message" example:"Task created successfully"`
}

func NewErrorResponse(err string) *ErrorResponse {
	return &ErrorResponse{Error: err}
}

func NewSuccessResponse(message string) *SuccessResponse {
	return &SuccessResponse{Message: message}
}
