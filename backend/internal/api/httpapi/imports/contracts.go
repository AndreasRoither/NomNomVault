package httpapi

import (
	"time"

	"github.com/AndreasRoither/NomNomVault/backend/internal/api/apicontract"
)

// SourceRecordSummary is the public provenance representation for an import job.
type SourceRecordSummary struct {
	ID             string    `json:"id"`
	SourceType     string    `json:"sourceType"`
	ImportKind     string    `json:"importKind"`
	SubmittedURL   *string   `json:"submittedUrl,omitempty"`
	NormalizedURL  *string   `json:"normalizedUrl,omitempty"`
	CanonicalURL   *string   `json:"canonicalUrl,omitempty"`
	TitleHint      *string   `json:"titleHint,omitempty"`
	RetentionState string    `json:"retentionState"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// ImportJobResponse is the public import job payload.
type ImportJobResponse struct {
	ID                string              `json:"id"`
	ImportKind        string              `json:"importKind"`
	Status            string              `json:"status"`
	NormalizedPayload map[string]any      `json:"normalizedPayload"`
	DraftRecipeID     *string             `json:"draftRecipeId,omitempty"`
	MatchRecipeID     *string             `json:"matchRecipeId,omitempty"`
	ConflictState     string              `json:"conflictState"`
	Warnings          []string            `json:"warnings"`
	ConfidenceScore   *float64            `json:"confidenceScore,omitempty"`
	ErrorCode         *string             `json:"errorCode,omitempty"`
	ErrorMessage      *string             `json:"errorMessage,omitempty"`
	AttemptCount      int                 `json:"attemptCount"`
	StartedAt         *time.Time          `json:"startedAt,omitempty"`
	FinishedAt        *time.Time          `json:"finishedAt,omitempty"`
	CreatedAt         time.Time           `json:"createdAt"`
	UpdatedAt         time.Time           `json:"updatedAt"`
	Source            SourceRecordSummary `json:"source"`
}

// ImportJobListResponse is one import job page.
type ImportJobListResponse struct {
	Data []ImportJobResponse        `json:"data"`
	Page apicontract.CursorPageInfo `json:"page"`
}

// CreateURLImportRequest queues a new URL import job.
type CreateURLImportRequest struct {
	URL            string `json:"url" binding:"required"`
	TitleHint      string `json:"titleHint"`
	ConfirmRestart bool   `json:"confirmRestart"`
}

// RetryImportJobRequest confirms that a finished job should be requeued.
type RetryImportJobRequest struct {
	ConfirmFinished bool `json:"confirmFinished"`
}
