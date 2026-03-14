package apicontract

import "github.com/gin-gonic/gin"

// ValidationError describes a single request field validation problem.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse standardizes API error payloads.
type ErrorResponse struct {
	Status  int               `json:"status"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
}

// CursorPageInfo describes a cursor-based page boundary.
type CursorPageInfo struct {
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
}

// WriteError writes the standard error envelope.
func WriteError(context *gin.Context, status int, code string, message string, details []ValidationError) {
	context.JSON(status, ErrorResponse{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
	})
}
