package models

// base response envelope for all API responses, including errors, status code, success.
type BaseResponse struct {
	Data       any    `json:"data"`
	Message    string `json:"message,omitempty"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code" validate:"required,gte=100,lte=599"`
}
