package dto

// APIResponse is the standard API response envelope.
type APIResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Success bool        `json:"success"`
}

// NewSuccessResponse creates a successful API response.
func NewSuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates an error API response.
func NewErrorResponse(message string, err string) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Error:   err,
	}
}

// HealthResponse contains service health status details.
type HealthResponse struct {
	Services map[string]string `json:"services"`
	Status   string            `json:"status"`
}

// RateLimitInfoResponse shows rate-limit status to the client.
type RateLimitInfoResponse struct {
	ResetAt   string `json:"reset_at"`
	Tier      string `json:"tier"`
	Algorithm string `json:"algorithm"`
	Limit     int    `json:"limit"`
	Remaining int    `json:"remaining"`
}
