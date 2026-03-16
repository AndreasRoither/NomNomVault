package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AndreasRoither/NomNomVault/backend/internal/api/apicontract"
	authsvc "github.com/AndreasRoither/NomNomVault/backend/internal/auth"
	importsvc "github.com/AndreasRoither/NomNomVault/backend/internal/imports"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/httpx"
)

type handler struct {
	service *importsvc.Service
}

func newHandler(service *importsvc.Service) *handler {
	return &handler{service: service}
}

// listImportJobs godoc
// @Summary List import jobs
// @Description Return the current household import jobs using cursor pagination and an optional status filter.
// @Tags imports
// @Produce json
// @Param cursor query string false "Cursor token"
// @Param limit query int false "Maximum number of import jobs to return"
// @Param status query string false "Filter by import job status"
// @Success 200 {object} ImportJobListResponse
// @Failure 400 {object} apicontract.ErrorResponse
// @Failure 401 {object} apicontract.ErrorResponse
// @Router /imports [get]
func (h *handler) listImportJobs(c *gin.Context) {
	session, ok := authsvc.SessionFromGin(c)
	if !ok {
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return
	}

	limit := 20
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			httpx.WriteValidationError(c, validationErrors("limit", "Limit must be a positive integer."))
			return
		}
		if parsed > 100 {
			parsed = 100
		}
		limit = parsed
	}

	var cursor *string
	if rawCursor := strings.TrimSpace(c.Query("cursor")); rawCursor != "" {
		cursor = &rawCursor
	}

	var status *string
	if rawStatus := strings.TrimSpace(c.Query("status")); rawStatus != "" {
		status = &rawStatus
	}

	result, err := h.service.ListImportJobs(c.Request.Context(), importsvc.ListImportJobsInput{
		HouseholdID: session.ActiveHouseholdID,
		Cursor:      cursor,
		Limit:       limit,
		Status:      status,
	})
	if err != nil {
		httpx.WriteServiceError(c, err)
		return
	}

	items := make([]ImportJobResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, mapImportJob(item))
	}

	c.JSON(http.StatusOK, ImportJobListResponse{
		Data: items,
		Page: apicontract.CursorPageInfo{
			NextCursor: result.NextCursor,
			HasMore:    result.HasMore,
		},
	})
}

// getImportJob godoc
// @Summary Fetch an import job
// @Description Return the detailed import job payload for the requested job ID.
// @Tags imports
// @Produce json
// @Param jobId path string true "Import job ID"
// @Success 200 {object} ImportJobResponse
// @Failure 401 {object} apicontract.ErrorResponse
// @Failure 404 {object} apicontract.ErrorResponse
// @Router /imports/{jobId} [get]
func (h *handler) getImportJob(c *gin.Context) {
	session, ok := authsvc.SessionFromGin(c)
	if !ok {
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return
	}

	job, err := h.service.GetImportJob(c.Request.Context(), session.ActiveHouseholdID, c.Param("jobId"))
	if err != nil {
		httpx.WriteServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapImportJob(job))
}

// createURLImport godoc
// @Summary Queue a URL import
// @Description Create a source record and a queued URL import job for the active household.
// @Tags imports
// @Accept json
// @Produce json
// @Param Idempotency-Key header string false "Client-supplied idempotency key"
// @Param payload body CreateURLImportRequest true "URL import payload"
// @Success 200 {object} ImportJobResponse
// @Success 202 {object} ImportJobResponse
// @Failure 400 {object} apicontract.ErrorResponse
// @Failure 401 {object} apicontract.ErrorResponse
// @Failure 403 {object} apicontract.ErrorResponse
// @Failure 409 {object} apicontract.ErrorResponse
// @Router /imports/url [post]
func (h *handler) createURLImport(c *gin.Context) {
	var request CreateURLImportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.WriteValidationError(c, validationErrors("payload", err.Error()))
		return
	}

	session, ok := authsvc.SessionFromGin(c)
	if !ok {
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return
	}

	result, err := h.service.CreateURLImport(c.Request.Context(), importsvc.CreateURLImportInput{
		HouseholdID:    session.ActiveHouseholdID,
		ActorUserID:    session.UserID,
		ActorRole:      string(session.HouseholdRole),
		URL:            request.URL,
		TitleHint:      request.TitleHint,
		IdempotencyKey: strings.TrimSpace(c.GetHeader("Idempotency-Key")),
		ConfirmRestart: request.ConfirmRestart,
	})
	if err != nil {
		httpx.WriteServiceError(c, err)
		return
	}

	status := http.StatusAccepted
	if result.Existing {
		status = http.StatusOK
	}
	c.JSON(status, mapImportJob(result.Job))
}

// cancelImportJob godoc
// @Summary Cancel an import job
// @Description Cancel a queued import job for the active household before worker execution starts.
// @Tags imports
// @Produce json
// @Param jobId path string true "Import job ID"
// @Success 200 {object} ImportJobResponse
// @Failure 401 {object} apicontract.ErrorResponse
// @Failure 403 {object} apicontract.ErrorResponse
// @Failure 404 {object} apicontract.ErrorResponse
// @Failure 409 {object} apicontract.ErrorResponse
// @Router /imports/{jobId}/cancel [post]
func (h *handler) cancelImportJob(c *gin.Context) {
	session, ok := authsvc.SessionFromGin(c)
	if !ok {
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return
	}

	job, err := h.service.CancelImportJob(c.Request.Context(), importsvc.CancelImportJobInput{
		HouseholdID: session.ActiveHouseholdID,
		ActorUserID: session.UserID,
		ActorRole:   string(session.HouseholdRole),
		JobID:       c.Param("jobId"),
	})
	if err != nil {
		httpx.WriteServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapImportJob(job))
}

// retryImportJob godoc
// @Summary Retry an import job
// @Description Queue a new attempt for a finished import job after explicit confirmation.
// @Tags imports
// @Accept json
// @Produce json
// @Param jobId path string true "Import job ID"
// @Param payload body RetryImportJobRequest true "Retry confirmation payload"
// @Success 202 {object} ImportJobResponse
// @Failure 400 {object} apicontract.ErrorResponse
// @Failure 401 {object} apicontract.ErrorResponse
// @Failure 403 {object} apicontract.ErrorResponse
// @Failure 404 {object} apicontract.ErrorResponse
// @Failure 409 {object} apicontract.ErrorResponse
// @Router /imports/{jobId}/retry [post]
func (h *handler) retryImportJob(c *gin.Context) {
	var request RetryImportJobRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.WriteValidationError(c, validationErrors("payload", err.Error()))
		return
	}

	session, ok := authsvc.SessionFromGin(c)
	if !ok {
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return
	}

	job, err := h.service.RetryImportJob(c.Request.Context(), importsvc.RetryImportJobInput{
		HouseholdID:     session.ActiveHouseholdID,
		ActorUserID:     session.UserID,
		ActorRole:       string(session.HouseholdRole),
		JobID:           c.Param("jobId"),
		ConfirmFinished: request.ConfirmFinished,
	})
	if err != nil {
		httpx.WriteServiceError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, mapImportJob(job))
}

func mapImportJob(view importsvc.ImportJobView) ImportJobResponse {
	return ImportJobResponse{
		ID:                view.ID,
		ImportKind:        view.ImportKind,
		Status:            view.Status,
		NormalizedPayload: view.NormalizedPayload,
		DraftRecipeID:     view.DraftRecipeID,
		MatchRecipeID:     view.MatchRecipeID,
		ConflictState:     view.ConflictState,
		Warnings:          view.Warnings,
		ConfidenceScore:   view.ConfidenceScore,
		ErrorCode:         view.ErrorCode,
		ErrorMessage:      view.ErrorMessage,
		AttemptCount:      view.AttemptCount,
		StartedAt:         view.StartedAt,
		FinishedAt:        view.FinishedAt,
		CreatedAt:         view.CreatedAt,
		UpdatedAt:         view.UpdatedAt,
		Source:            mapSourceRecord(view.Source),
	}
}

func mapSourceRecord(view importsvc.SourceRecordView) SourceRecordSummary {
	return SourceRecordSummary{
		ID:             view.ID,
		SourceType:     view.SourceType,
		ImportKind:     view.ImportKind,
		SubmittedURL:   view.SubmittedURL,
		NormalizedURL:  view.NormalizedURL,
		CanonicalURL:   view.CanonicalURL,
		TitleHint:      view.TitleHint,
		RetentionState: view.RetentionState,
		CreatedAt:      view.CreatedAt,
		UpdatedAt:      view.UpdatedAt,
	}
}

func validationErrors(field string, message string) []apicontract.ValidationError {
	return []apicontract.ValidationError{{Field: field, Message: message}}
}
