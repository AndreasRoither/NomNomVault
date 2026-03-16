package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AndreasRoither/NomNomVault/backend/internal/api/apicontract"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/requestid"
)

// StatusError maps a service error to an HTTP status and stable code.
type StatusError struct {
	Status  int
	Code    string
	Message string
}

func (e StatusError) Error() string {
	return e.Message
}

// WriteValidationError returns a standard request validation failure.
func WriteValidationError(c *gin.Context, details []apicontract.ValidationError) {
	WriteError(c, http.StatusBadRequest, "validation_error", "Request payload is invalid.", details)
}

// WriteError writes the standard error envelope.
func WriteError(c *gin.Context, status int, code string, message string, details []apicontract.ValidationError) {
	apicontract.WriteError(c, status, code, message, requestid.FromContext(c), details)
}

// WriteServiceError converts typed service errors to the HTTP envelope.
func WriteServiceError(c *gin.Context, err error) {
	var statusErr StatusError
	if errors.As(err, &statusErr) {
		WriteError(c, statusErr.Status, statusErr.Code, statusErr.Message, nil)
		return
	}

	WriteError(c, http.StatusInternalServerError, "internal_error", "The request could not be completed.", nil)
}
