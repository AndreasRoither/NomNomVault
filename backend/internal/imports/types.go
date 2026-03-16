package imports

import "time"

// SourceRecordView is the public import provenance representation.
type SourceRecordView struct {
	ID             string
	SourceType     string
	ImportKind     string
	SubmittedURL   *string
	NormalizedURL  *string
	CanonicalURL   *string
	TitleHint      *string
	RetentionState string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ImportJobView is the persisted import job representation exposed to the API.
type ImportJobView struct {
	ID                string
	ImportKind        string
	Status            string
	NormalizedPayload map[string]any
	DraftRecipeID     *string
	MatchRecipeID     *string
	ConflictState     string
	Warnings          []string
	ConfidenceScore   *float64
	ErrorCode         *string
	ErrorMessage      *string
	AttemptCount      int
	StartedAt         *time.Time
	FinishedAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Source            SourceRecordView
}

// ImportJobListResult wraps one import job list page.
type ImportJobListResult struct {
	Items      []ImportJobView
	NextCursor *string
	HasMore    bool
}

// ListImportJobsInput defines the supported import job filters.
type ListImportJobsInput struct {
	HouseholdID string
	Cursor      *string
	Limit       int
	Status      *string
}

// CreateURLImportInput creates a queued URL import job and source record.
type CreateURLImportInput struct {
	HouseholdID    string
	ActorUserID    string
	ActorRole      string
	URL            string
	TitleHint      string
	IdempotencyKey string
	ConfirmRestart bool
}

// CreateImportJobResult wraps the job returned by create semantics.
type CreateImportJobResult struct {
	Job      ImportJobView
	Existing bool
}

// CancelImportJobInput identifies one job to cancel.
type CancelImportJobInput struct {
	HouseholdID string
	ActorUserID string
	ActorRole   string
	JobID       string
}

// RetryImportJobInput identifies one finished job to requeue explicitly.
type RetryImportJobInput struct {
	HouseholdID     string
	ActorUserID     string
	ActorRole       string
	JobID           string
	ConfirmFinished bool
}
