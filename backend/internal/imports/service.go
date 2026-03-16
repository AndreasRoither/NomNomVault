package imports

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"sort"
	"strings"

	"github.com/AndreasRoither/NomNomVault/backend/internal/auth"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent"
	entgen "github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/importjob"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/recipe"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/sourcerecord"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/clock"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/httpx"
)

type importCursor struct {
	Offset int `json:"offset"`
}

// Service orchestrates persisted import job flows.
type Service struct {
	db    *ent.Client
	clock clock.Clock
}

// NewService creates an import service.
func NewService(db *ent.Client, clock clock.Clock) *Service {
	return &Service{db: db, clock: clock}
}

// ListImportJobs loads one household-scoped import job page.
func (s *Service) ListImportJobs(ctx context.Context, in ListImportJobsInput) (ImportJobListResult, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := s.db.ImportJob.Query().
		Where(importjob.HouseholdIDEQ(in.HouseholdID)).
		WithSourceRecord().
		Order(entgen.Desc(importjob.FieldCreatedAt), entgen.Desc(importjob.FieldID))

	if in.Status != nil && strings.TrimSpace(*in.Status) != "" {
		statusValue := strings.TrimSpace(*in.Status)
		if !isValidImportJobStatus(statusValue) {
			return ImportJobListResult{}, httpx.StatusError{
				Status:  http.StatusBadRequest,
				Code:    "validation_error",
				Message: "Import status filter is invalid.",
			}
		}
		query.Where(importjob.StatusEQ(importjob.Status(statusValue)))
	}

	if in.Cursor != nil && strings.TrimSpace(*in.Cursor) != "" {
		cursor, err := decodeImportCursor(*in.Cursor)
		if err != nil {
			return ImportJobListResult{}, httpx.StatusError{
				Status:  http.StatusBadRequest,
				Code:    "validation_error",
				Message: "Cursor is invalid.",
			}
		}
		query.Offset(cursor.Offset)
	}

	entities, err := query.Limit(limit + 1).All(ctx)
	if err != nil {
		return ImportJobListResult{}, fmt.Errorf("query import jobs: %w", err)
	}

	result := ImportJobListResult{
		Items: make([]ImportJobView, 0, min(limit, len(entities))),
	}
	for _, entity := range entities[:min(limit, len(entities))] {
		result.Items = append(result.Items, mapImportJob(entity))
	}

	if len(entities) > limit {
		result.HasMore = true
		nextCursor, err := encodeImportCursor(importCursor{Offset: offsetFromImportCursor(in.Cursor) + limit})
		if err != nil {
			return ImportJobListResult{}, fmt.Errorf("encode import cursor: %w", err)
		}
		result.NextCursor = &nextCursor
	}

	return result, nil
}

// GetImportJob loads one household-scoped import job.
func (s *Service) GetImportJob(ctx context.Context, householdID string, jobID string) (ImportJobView, error) {
	entity, err := s.db.ImportJob.Query().
		Where(importjob.IDEQ(jobID), importjob.HouseholdIDEQ(householdID)).
		WithSourceRecord().
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return ImportJobView{}, httpx.StatusError{Status: http.StatusNotFound, Code: "import_job_not_found", Message: "Import job was not found."}
		}
		return ImportJobView{}, fmt.Errorf("query import job: %w", err)
	}

	return mapImportJob(entity), nil
}

// CreateURLImport queues one URL import job and source record.
func (s *Service) CreateURLImport(ctx context.Context, in CreateURLImportInput) (CreateImportJobResult, error) {
	if !canManageImports(in.ActorRole) {
		return CreateImportJobResult{}, httpx.StatusError{Status: http.StatusForbidden, Code: "forbidden", Message: "The current account cannot manage imports."}
	}

	normalizedURL, err := normalizeImportURL(in.URL)
	if err != nil {
		return CreateImportJobResult{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Import URL is invalid."}
	}

	titleHint := normalizeTitleHint(in.TitleHint)
	idempotencyKey := normalizeOptionalString(in.IdempotencyKey)
	fallbackFingerprint := fingerprintURLImport(normalizedURL, titleHint)
	activeIdempotencyKey := buildActiveIdempotencyKey(in.HouseholdID, in.ActorUserID, importjob.ImportKindURL, idempotencyKey)
	activeFingerprintKey := buildActiveFingerprintKey(in.HouseholdID, importjob.ImportKindURL, fallbackFingerprint)

	activeMatch, err := s.findActiveMatchingJob(ctx, in.HouseholdID, activeIdempotencyKey, activeFingerprintKey)
	if err != nil {
		return CreateImportJobResult{}, err
	}
	if activeMatch != nil {
		return CreateImportJobResult{Job: mapImportJob(activeMatch), Existing: true}, nil
	}

	latestMatch, err := s.findLatestMatchingJob(ctx, in.HouseholdID, in.ActorUserID, importjob.ImportKindURL, idempotencyKey, fallbackFingerprint)
	if err != nil {
		return CreateImportJobResult{}, err
	}
	if latestMatch != nil {
		if locked, err := s.retryLockedByRecipeState(ctx, in.HouseholdID, latestMatch.DraftRecipeID); err != nil {
			return CreateImportJobResult{}, err
		} else if locked {
			return CreateImportJobResult{}, httpx.StatusError{
				Status:  http.StatusConflict,
				Code:    "import_job_locked",
				Message: "This import can no longer be restarted because its draft is no longer editable.",
			}
		}
		if !in.ConfirmRestart {
			return CreateImportJobResult{}, httpx.StatusError{
				Status:  http.StatusConflict,
				Code:    "import_restart_confirmation_required",
				Message: "The matching import job has already finished. Confirm restart to queue a new import attempt.",
			}
		}
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return CreateImportJobResult{}, fmt.Errorf("start import create tx: %w", err)
	}

	sourceRecordID := ""
	attemptCount := 1
	if latestMatch != nil {
		sourceRecordID = latestMatch.SourceRecordID
		attemptCount = latestMatch.AttemptCount + 1
	} else {
		sourceCreate := tx.SourceRecord.Create().
			SetHouseholdID(in.HouseholdID).
			SetSourceType(sourcerecord.SourceTypeURL).
			SetImportKind(sourcerecord.ImportKindURL).
			SetSubmittedURL(strings.TrimSpace(in.URL)).
			SetNormalizedURL(normalizedURL)
		if titleHint != "" {
			sourceCreate.SetTitleHint(titleHint)
		}

		sourceEntity, err := sourceCreate.Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return CreateImportJobResult{}, fmt.Errorf("create source record: %w", err)
		}
		sourceRecordID = sourceEntity.ID
	}

	jobCreate := tx.ImportJob.Create().
		SetHouseholdID(in.HouseholdID).
		SetRequestedByUserID(in.ActorUserID).
		SetSourceRecordID(sourceRecordID).
		SetImportKind(importjob.ImportKindURL).
		SetStatus(importjob.StatusQueued).
		SetFallbackFingerprint(fallbackFingerprint).
		SetActiveFingerprintKey(activeFingerprintKey).
		SetAttemptCount(attemptCount)
	if idempotencyKey != nil {
		jobCreate.SetIdempotencyKey(*idempotencyKey)
	}
	if activeIdempotencyKey != nil {
		jobCreate.SetActiveIdempotencyKey(*activeIdempotencyKey)
	}

	jobEntity, err := jobCreate.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		if entgen.IsConstraintError(err) {
			existing, loadErr := s.findActiveMatchingJob(ctx, in.HouseholdID, activeIdempotencyKey, activeFingerprintKey)
			if loadErr != nil {
				return CreateImportJobResult{}, loadErr
			}
			if existing != nil {
				return CreateImportJobResult{Job: mapImportJob(existing), Existing: true}, nil
			}
		}
		return CreateImportJobResult{}, fmt.Errorf("create import job: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return CreateImportJobResult{}, fmt.Errorf("commit import create tx: %w", err)
	}

	jobView, err := s.GetImportJob(ctx, in.HouseholdID, jobEntity.ID)
	if err != nil {
		return CreateImportJobResult{}, err
	}
	return CreateImportJobResult{Job: jobView}, nil
}

// CancelImportJob cancels one queued import job before worker execution starts.
func (s *Service) CancelImportJob(ctx context.Context, in CancelImportJobInput) (ImportJobView, error) {
	if !canManageImports(in.ActorRole) {
		return ImportJobView{}, httpx.StatusError{Status: http.StatusForbidden, Code: "forbidden", Message: "The current account cannot manage imports."}
	}

	entity, err := s.db.ImportJob.Query().
		Where(importjob.IDEQ(in.JobID), importjob.HouseholdIDEQ(in.HouseholdID)).
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return ImportJobView{}, httpx.StatusError{Status: http.StatusNotFound, Code: "import_job_not_found", Message: "Import job was not found."}
		}
		return ImportJobView{}, fmt.Errorf("query import job for cancel: %w", err)
	}

	if entity.Status != importjob.StatusQueued {
		return ImportJobView{}, httpx.StatusError{
			Status:  http.StatusConflict,
			Code:    "import_job_not_cancellable",
			Message: "Only queued import jobs can be cancelled from the API.",
		}
	}

	now := s.clock.Now()
	if _, err := s.db.ImportJob.UpdateOneID(entity.ID).
		SetStatus(importjob.StatusCancelled).
		SetFinishedAt(now).
		SetErrorCode("cancelled_by_user").
		SetErrorMessage("The import was cancelled by the user.").
		ClearActiveIdempotencyKey().
		ClearActiveFingerprintKey().
		Save(ctx); err != nil {
		return ImportJobView{}, fmt.Errorf("cancel import job: %w", err)
	}

	return s.GetImportJob(ctx, in.HouseholdID, entity.ID)
}

// RetryImportJob creates a new queued attempt for one finished import job.
func (s *Service) RetryImportJob(ctx context.Context, in RetryImportJobInput) (ImportJobView, error) {
	if !canManageImports(in.ActorRole) {
		return ImportJobView{}, httpx.StatusError{Status: http.StatusForbidden, Code: "forbidden", Message: "The current account cannot manage imports."}
	}

	entity, err := s.db.ImportJob.Query().
		Where(importjob.IDEQ(in.JobID), importjob.HouseholdIDEQ(in.HouseholdID)).
		WithSourceRecord().
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return ImportJobView{}, httpx.StatusError{Status: http.StatusNotFound, Code: "import_job_not_found", Message: "Import job was not found."}
		}
		return ImportJobView{}, fmt.Errorf("query import job for retry: %w", err)
	}

	if isActiveImportJobStatus(entity.Status) {
		return ImportJobView{}, httpx.StatusError{
			Status:  http.StatusConflict,
			Code:    "import_job_not_retryable",
			Message: "Active import jobs cannot be retried.",
		}
	}
	if !in.ConfirmFinished {
		return ImportJobView{}, httpx.StatusError{
			Status:  http.StatusConflict,
			Code:    "import_restart_confirmation_required",
			Message: "Confirm restart to queue a new import attempt for this finished job.",
		}
	}

	locked, err := s.retryLockedByRecipeState(ctx, in.HouseholdID, entity.DraftRecipeID)
	if err != nil {
		return ImportJobView{}, err
	}
	if locked {
		return ImportJobView{}, httpx.StatusError{
			Status:  http.StatusConflict,
			Code:    "import_job_locked",
			Message: "This import can no longer be restarted because its draft is no longer editable.",
		}
	}

	activeFingerprintKey := buildActiveFingerprintKey(entity.HouseholdID, entity.ImportKind, entity.FallbackFingerprint)
	activeMatch, err := s.findActiveMatchingJob(ctx, in.HouseholdID, nil, activeFingerprintKey)
	if err != nil {
		return ImportJobView{}, err
	}
	if activeMatch != nil {
		return ImportJobView{}, httpx.StatusError{
			Status:  http.StatusConflict,
			Code:    "import_job_already_active",
			Message: "A matching active import job already exists.",
		}
	}

	retryEntity, err := s.db.ImportJob.Create().
		SetHouseholdID(entity.HouseholdID).
		SetRequestedByUserID(in.ActorUserID).
		SetSourceRecordID(entity.SourceRecordID).
		SetImportKind(entity.ImportKind).
		SetStatus(importjob.StatusQueued).
		SetFallbackFingerprint(entity.FallbackFingerprint).
		SetActiveFingerprintKey(activeFingerprintKey).
		SetAttemptCount(entity.AttemptCount + 1).
		Save(ctx)
	if err != nil {
		if entgen.IsConstraintError(err) {
			return ImportJobView{}, httpx.StatusError{
				Status:  http.StatusConflict,
				Code:    "import_job_already_active",
				Message: "A matching active import job already exists.",
			}
		}
		return ImportJobView{}, fmt.Errorf("create retry import job: %w", err)
	}

	return s.GetImportJob(ctx, in.HouseholdID, retryEntity.ID)
}

func (s *Service) findLatestMatchingJob(
	ctx context.Context,
	householdID string,
	actorUserID string,
	kind importjob.ImportKind,
	idempotencyKey *string,
	fallbackFingerprint string,
) (*entgen.ImportJob, error) {
	if idempotencyKey != nil {
		entity, err := s.db.ImportJob.Query().
			Where(
				importjob.HouseholdIDEQ(householdID),
				importjob.RequestedByUserIDEQ(actorUserID),
				importjob.ImportKindEQ(kind),
				importjob.IdempotencyKeyEQ(*idempotencyKey),
			).
			WithSourceRecord().
			Order(entgen.Desc(importjob.FieldCreatedAt), entgen.Desc(importjob.FieldID)).
			First(ctx)
		if err == nil {
			return entity, nil
		}
		if !entgen.IsNotFound(err) {
			return nil, fmt.Errorf("query idempotent import job: %w", err)
		}
	}

	entity, err := s.db.ImportJob.Query().
		Where(
			importjob.HouseholdIDEQ(householdID),
			importjob.ImportKindEQ(kind),
			importjob.FallbackFingerprintEQ(fallbackFingerprint),
		).
		WithSourceRecord().
		Order(entgen.Desc(importjob.FieldCreatedAt), entgen.Desc(importjob.FieldID)).
		First(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("query fallback import job: %w", err)
	}
	return entity, nil
}

func (s *Service) findActiveMatchingJob(
	ctx context.Context,
	householdID string,
	activeIdempotencyKey *string,
	activeFingerprintKey string,
) (*entgen.ImportJob, error) {
	if activeIdempotencyKey != nil {
		entity, err := s.db.ImportJob.Query().
			Where(
				importjob.HouseholdIDEQ(householdID),
				importjob.ActiveIdempotencyKeyEQ(*activeIdempotencyKey),
			).
			WithSourceRecord().
			Only(ctx)
		if err == nil {
			return entity, nil
		}
		if !entgen.IsNotFound(err) {
			return nil, fmt.Errorf("query active idempotent import job: %w", err)
		}
	}

	entity, err := s.db.ImportJob.Query().
		Where(
			importjob.HouseholdIDEQ(householdID),
			importjob.ActiveFingerprintKeyEQ(activeFingerprintKey),
		).
		WithSourceRecord().
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("query active fingerprint import job: %w", err)
	}
	return entity, nil
}

func (s *Service) retryLockedByRecipeState(ctx context.Context, householdID string, draftRecipeID *string) (bool, error) {
	if draftRecipeID == nil || strings.TrimSpace(*draftRecipeID) == "" {
		return false, nil
	}

	entity, err := s.db.Recipe.Query().
		Where(recipe.IDEQ(*draftRecipeID), recipe.HouseholdIDEQ(householdID)).
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("query linked recipe state: %w", err)
	}

	return entity.Status != recipe.StatusDraft, nil
}

func mapImportJob(entity *entgen.ImportJob) ImportJobView {
	return ImportJobView{
		ID:                entity.ID,
		ImportKind:        string(entity.ImportKind),
		Status:            string(entity.Status),
		NormalizedPayload: cloneMap(entity.NormalizedPayloadJSON),
		DraftRecipeID:     entity.DraftRecipeID,
		MatchRecipeID:     entity.MatchRecipeID,
		ConflictState:     string(entity.ConflictState),
		Warnings:          append([]string{}, entity.WarningsJSON...),
		ConfidenceScore:   entity.ConfidenceScore,
		ErrorCode:         entity.ErrorCode,
		ErrorMessage:      entity.ErrorMessage,
		AttemptCount:      entity.AttemptCount,
		StartedAt:         entity.StartedAt,
		FinishedAt:        entity.FinishedAt,
		CreatedAt:         entity.CreatedAt,
		UpdatedAt:         entity.UpdatedAt,
		Source:            mapSourceRecord(entity.Edges.SourceRecord),
	}
}

func mapSourceRecord(entity *entgen.SourceRecord) SourceRecordView {
	if entity == nil {
		return SourceRecordView{}
	}

	return SourceRecordView{
		ID:             entity.ID,
		SourceType:     string(entity.SourceType),
		ImportKind:     string(entity.ImportKind),
		SubmittedURL:   entity.SubmittedURL,
		NormalizedURL:  entity.NormalizedURL,
		CanonicalURL:   entity.CanonicalURL,
		TitleHint:      entity.TitleHint,
		RetentionState: string(entity.RetentionState),
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      entity.UpdatedAt,
	}
}

func cloneMap(source map[string]any) map[string]any {
	if len(source) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func canManageImports(role string) bool {
	return role == string(auth.HouseholdRoleOwner) || role == string(auth.HouseholdRoleEditor)
}

func isValidImportJobStatus(value string) bool {
	switch importjob.Status(value) {
	case importjob.StatusQueued,
		importjob.StatusFetching,
		importjob.StatusParsing,
		importjob.StatusNeedsReview,
		importjob.StatusConflictDetected,
		importjob.StatusCompleted,
		importjob.StatusFailed,
		importjob.StatusCancelled:
		return true
	default:
		return false
	}
}

func isActiveImportJobStatus(status importjob.Status) bool {
	switch status {
	case importjob.StatusQueued, importjob.StatusFetching, importjob.StatusParsing:
		return true
	default:
		return false
	}
}

func normalizeOptionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeTitleHint(value string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(value)))
	return strings.Join(fields, " ")
}

func normalizeImportURL(raw string) (string, error) {
	parsed, err := neturl.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("scheme and host are required")
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme")
	}
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	if host := parsed.Hostname(); host != "" {
		port := parsed.Port()
		if (parsed.Scheme == "http" && port == "80") || (parsed.Scheme == "https" && port == "443") {
			parsed.Host = host
		}
	}

	filteredQuery := neturl.Values{}
	for key, values := range parsed.Query() {
		lowerKey := strings.ToLower(key)
		if strings.HasPrefix(lowerKey, "utm_") || lowerKey == "fbclid" || lowerKey == "gclid" || lowerKey == "mc_cid" || lowerKey == "mc_eid" {
			continue
		}
		sortedValues := append([]string{}, values...)
		sort.Strings(sortedValues)
		for _, value := range sortedValues {
			filteredQuery.Add(key, value)
		}
	}
	parsed.RawQuery = filteredQuery.Encode()

	return parsed.String(), nil
}

func fingerprintURLImport(normalizedURL string, titleHint string) string {
	sum := sha256.Sum256([]byte("url|" + normalizedURL + "|" + titleHint))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func buildActiveIdempotencyKey(householdID string, actorUserID string, kind importjob.ImportKind, idempotencyKey *string) *string {
	if idempotencyKey == nil {
		return nil
	}
	key := hashScopeKey("idempotency", householdID, actorUserID, string(kind), *idempotencyKey)
	return &key
}

func buildActiveFingerprintKey(householdID string, kind importjob.ImportKind, fallbackFingerprint string) string {
	return hashScopeKey("fingerprint", householdID, string(kind), fallbackFingerprint)
}

func hashScopeKey(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func encodeImportCursor(cursor importCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeImportCursor(raw string) (importCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return importCursor{}, err
	}

	var cursor importCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return importCursor{}, err
	}
	if cursor.Offset < 0 {
		return importCursor{}, fmt.Errorf("cursor missing fields")
	}
	return cursor, nil
}

func offsetFromImportCursor(raw *string) int {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return 0
	}
	cursor, err := decodeImportCursor(*raw)
	if err != nil {
		return 0
	}
	return cursor.Offset
}

func min(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
