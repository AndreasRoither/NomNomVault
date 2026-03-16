package recipes

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/AndreasRoither/NomNomVault/backend/internal/auth"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent"
	entgen "github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/mediaasset"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/recipe"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/recipeingredient"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/recipenutrition"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/recipestep"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/tag"
	entschema "github.com/AndreasRoither/NomNomVault/backend/internal/ent/schema"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/clock"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/httpx"
	platformmedia "github.com/AndreasRoither/NomNomVault/backend/internal/platform/media"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/storage"
)

var (
	tagColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

type recipeCursor struct {
	Offset int `json:"offset"`
}

// Service implements the persisted recipe flows.
type Service struct {
	db    *ent.Client
	store storage.Store
	clock clock.Clock
}

// NewService creates a recipe service.
func NewService(db *ent.Client, store storage.Store, clock clock.Clock) *Service {
	return &Service{db: db, store: store, clock: clock}
}

// ListRecipes loads one household-scoped recipe page with optional filters.
func (s *Service) ListRecipes(ctx context.Context, in ListRecipesInput) (RecipeListResult, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := s.db.Recipe.Query().
		Where(
			recipe.HouseholdIDEQ(in.HouseholdID),
			recipe.ArchivedAtIsNil(),
		).
		Order(
			entgen.Desc(recipe.FieldUpdatedAt),
			entgen.Desc(recipe.FieldID),
		)

	if searchQuery := strings.TrimSpace(in.Query); searchQuery != "" {
		query.Where(recipe.Or(
			recipe.TitleContainsFold(searchQuery),
			recipe.DescriptionContainsFold(searchQuery),
		))
	}

	for _, slug := range normalizeTagSlugs(in.TagSlugs) {
		query.Where(recipe.HasTagsWith(
			tag.HouseholdIDEQ(in.HouseholdID),
			tag.SlugEQ(slug),
		))
	}

	if in.Cursor != nil && strings.TrimSpace(*in.Cursor) != "" {
		cursor, err := decodeRecipeCursor(*in.Cursor)
		if err != nil {
			return RecipeListResult{}, httpx.StatusError{
				Status:  http.StatusBadRequest,
				Code:    "validation_error",
				Message: "Cursor is invalid.",
			}
		}

		query.Offset(cursor.Offset)
	}

	entities, err := query.Limit(limit + 1).All(ctx)
	if err != nil {
		return RecipeListResult{}, fmt.Errorf("query recipe list: %w", err)
	}

	result := RecipeListResult{
		Items: make([]RecipeSummaryView, 0, min(limit, len(entities))),
	}

	for _, entity := range entities[:min(limit, len(entities))] {
		result.Items = append(result.Items, mapRecipeSummary(entity))
	}

	if len(entities) > limit {
		result.HasMore = true
		nextCursor, err := encodeRecipeCursor(recipeCursor{
			Offset: offsetFromCursor(in.Cursor) + limit,
		})
		if err != nil {
			return RecipeListResult{}, fmt.Errorf("encode recipe cursor: %w", err)
		}
		result.NextCursor = &nextCursor
	}

	return result, nil
}

// CreateRecipe persists a minimal household-scoped recipe and its ingredients.
func (s *Service) CreateRecipe(ctx context.Context, in CreateRecipeInput) (CreateRecipeResult, error) {
	if !canEdit(in.ActorRole) {
		return CreateRecipeResult{}, httpx.StatusError{Status: http.StatusForbidden, Code: "forbidden", Message: "The current account cannot modify recipes."}
	}
	if strings.TrimSpace(in.Title) == "" {
		return CreateRecipeResult{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Recipe title is required."}
	}
	if err := validateIngredients(in.Ingredients); err != nil {
		return CreateRecipeResult{}, err
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return CreateRecipeResult{}, fmt.Errorf("start tx: %w", err)
	}

	recipeCreate := tx.Recipe.Create().
		SetHouseholdID(in.HouseholdID).
		SetTitle(strings.TrimSpace(in.Title)).
		SetDescription(strings.TrimSpace(in.Description)).
		SetSourceURL(strings.TrimSpace(in.SourceURL))

	if in.PrepMinutes != nil {
		recipeCreate.SetPrepMinutes(*in.PrepMinutes)
	}
	if in.CookMinutes != nil {
		recipeCreate.SetCookMinutes(*in.CookMinutes)
	}
	if in.Servings != nil {
		recipeCreate.SetServings(*in.Servings)
	}

	recipeEntity, err := recipeCreate.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return CreateRecipeResult{}, fmt.Errorf("create recipe: %w", err)
	}

	if err := replaceIngredients(ctx, tx, recipeEntity.ID, in.Ingredients); err != nil {
		_ = tx.Rollback()
		return CreateRecipeResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return CreateRecipeResult{}, fmt.Errorf("commit recipe tx: %w", err)
	}

	detail, err := s.GetRecipeDetail(ctx, in.HouseholdID, recipeEntity.ID)
	if err != nil {
		return CreateRecipeResult{}, err
	}

	return CreateRecipeResult{Recipe: detail}, nil
}

// UpdateRecipe replaces the editable fields and child collections for a recipe.
func (s *Service) UpdateRecipe(ctx context.Context, in UpdateRecipeInput) (DetailView, error) {
	if !canEdit(in.ActorRole) {
		return DetailView{}, httpx.StatusError{Status: http.StatusForbidden, Code: "forbidden", Message: "The current account cannot modify recipes."}
	}
	if strings.TrimSpace(in.Title) == "" {
		return DetailView{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Recipe title is required."}
	}
	if err := validateIngredients(in.Ingredients); err != nil {
		return DetailView{}, err
	}
	if err := validateSteps(in.Steps); err != nil {
		return DetailView{}, err
	}
	if err := validateNutritionEntries(in.NutritionEntries); err != nil {
		return DetailView{}, err
	}

	uniqueTagIDs, err := uniqueIDs(in.TagIDs, "tagIds")
	if err != nil {
		return DetailView{}, err
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return DetailView{}, fmt.Errorf("start tx: %w", err)
	}

	recipeEntity, err := tx.Recipe.Query().
		Where(
			recipe.IDEQ(in.RecipeID),
			recipe.HouseholdIDEQ(in.HouseholdID),
		).
		Only(ctx)
	if err != nil {
		_ = tx.Rollback()
		if entgen.IsNotFound(err) {
			return DetailView{}, httpx.StatusError{Status: http.StatusNotFound, Code: "recipe_not_found", Message: "Recipe was not found."}
		}

		return DetailView{}, fmt.Errorf("query recipe for update: %w", err)
	}

	if recipeEntity.Version != in.ExpectedVersion {
		_ = tx.Rollback()
		return DetailView{}, httpx.StatusError{
			Status:  http.StatusConflict,
			Code:    "version_conflict",
			Message: "The recipe was modified by another request.",
		}
	}

	if _, err := loadTagsByID(ctx, tx, in.HouseholdID, uniqueTagIDs); err != nil {
		_ = tx.Rollback()
		return DetailView{}, err
	}

	updateBuilder := tx.Recipe.Update().
		Where(
			recipe.IDEQ(in.RecipeID),
			recipe.HouseholdIDEQ(in.HouseholdID),
			recipe.VersionEQ(in.ExpectedVersion),
		).
		SetTitle(strings.TrimSpace(in.Title)).
		SetDescription(strings.TrimSpace(in.Description)).
		SetSourceURL(strings.TrimSpace(in.SourceURL)).
		SetVersion(in.ExpectedVersion + 1)

	if in.PrepMinutes != nil {
		updateBuilder.SetPrepMinutes(*in.PrepMinutes)
	} else {
		updateBuilder.ClearPrepMinutes()
	}
	if in.CookMinutes != nil {
		updateBuilder.SetCookMinutes(*in.CookMinutes)
	} else {
		updateBuilder.ClearCookMinutes()
	}
	if in.Servings != nil {
		updateBuilder.SetServings(*in.Servings)
	} else {
		updateBuilder.ClearServings()
	}

	updated, err := updateBuilder.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return DetailView{}, fmt.Errorf("update recipe header: %w", err)
	}
	if updated == 0 {
		_ = tx.Rollback()
		return DetailView{}, httpx.StatusError{
			Status:  http.StatusConflict,
			Code:    "version_conflict",
			Message: "The recipe was modified by another request.",
		}
	}

	if err := replaceIngredients(ctx, tx, in.RecipeID, in.Ingredients); err != nil {
		_ = tx.Rollback()
		return DetailView{}, err
	}
	if err := replaceSteps(ctx, tx, in.RecipeID, in.Steps); err != nil {
		_ = tx.Rollback()
		return DetailView{}, err
	}
	if err := replaceNutritionEntries(ctx, tx, in.RecipeID, in.NutritionEntries); err != nil {
		_ = tx.Rollback()
		return DetailView{}, err
	}
	if err := replaceTagAssignments(ctx, tx, in.RecipeID, uniqueTagIDs); err != nil {
		_ = tx.Rollback()
		return DetailView{}, err
	}

	if err := tx.Commit(); err != nil {
		return DetailView{}, fmt.Errorf("commit update recipe tx: %w", err)
	}

	return s.GetRecipeDetail(ctx, in.HouseholdID, in.RecipeID)
}

// ArchiveRecipe marks a recipe as archived.
func (s *Service) ArchiveRecipe(ctx context.Context, in ArchiveRecipeInput) error {
	if !canEdit(in.ActorRole) {
		return httpx.StatusError{Status: http.StatusForbidden, Code: "forbidden", Message: "The current account cannot modify recipes."}
	}

	entity, err := s.db.Recipe.Query().
		Where(
			recipe.IDEQ(in.RecipeID),
			recipe.HouseholdIDEQ(in.HouseholdID),
		).
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return httpx.StatusError{Status: http.StatusNotFound, Code: "recipe_not_found", Message: "Recipe was not found."}
		}

		return fmt.Errorf("query recipe for archive: %w", err)
	}

	if entity.ArchivedAt != nil {
		return nil
	}

	if _, err := s.db.Recipe.UpdateOneID(entity.ID).
		SetArchivedAt(s.clock.Now()).
		SetVersion(entity.Version + 1).
		Save(ctx); err != nil {
		return fmt.Errorf("archive recipe: %w", err)
	}

	return nil
}

// UnarchiveRecipe marks a recipe as active again.
func (s *Service) UnarchiveRecipe(ctx context.Context, in ArchiveRecipeInput) error {
	if !canEdit(in.ActorRole) {
		return httpx.StatusError{Status: http.StatusForbidden, Code: "forbidden", Message: "The current account cannot modify recipes."}
	}

	entity, err := s.db.Recipe.Query().
		Where(
			recipe.IDEQ(in.RecipeID),
			recipe.HouseholdIDEQ(in.HouseholdID),
		).
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return httpx.StatusError{Status: http.StatusNotFound, Code: "recipe_not_found", Message: "Recipe was not found."}
		}

		return fmt.Errorf("query recipe for unarchive: %w", err)
	}

	if entity.ArchivedAt == nil {
		return nil
	}

	if _, err := s.db.Recipe.UpdateOneID(entity.ID).
		ClearArchivedAt().
		SetVersion(entity.Version + 1).
		Save(ctx); err != nil {
		return fmt.Errorf("unarchive recipe: %w", err)
	}

	return nil
}

// ListTags loads all household-scoped tags ordered by name.
func (s *Service) ListTags(ctx context.Context, householdID string) ([]TagView, error) {
	entities, err := s.db.Tag.Query().
		Where(tag.HouseholdIDEQ(householdID)).
		Order(entgen.Asc(tag.FieldName), entgen.Asc(tag.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}

	return mapTags(entities), nil
}

// CreateTag creates one household-scoped tag.
func (s *Service) CreateTag(ctx context.Context, in CreateTagInput) (TagView, error) {
	if !canEdit(in.ActorRole) {
		return TagView{}, httpx.StatusError{Status: http.StatusForbidden, Code: "forbidden", Message: "The current account cannot modify recipes."}
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return TagView{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Tag name is required."}
	}

	color := strings.TrimSpace(in.Color)
	if color != "" && !tagColorPattern.MatchString(color) {
		return TagView{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Tag color must be a hex value like #4FB8B2."}
	}

	slug := normalizeSlug(name)
	if slug == "" {
		return TagView{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Tag name must contain letters or digits."}
	}

	entity, err := s.db.Tag.Create().
		SetHouseholdID(in.HouseholdID).
		SetName(name).
		SetSlug(slug).
		SetColor(color).
		Save(ctx)
	if err != nil {
		if entgen.IsConstraintError(err) {
			return TagView{}, httpx.StatusError{Status: http.StatusConflict, Code: "tag_conflict", Message: "A tag with that name already exists."}
		}
		return TagView{}, fmt.Errorf("create tag: %w", err)
	}

	return mapTags([]*entgen.Tag{entity})[0], nil
}

// GetRecipeDetail loads a persisted recipe detail.
func (s *Service) GetRecipeDetail(ctx context.Context, householdID string, recipeID string) (DetailView, error) {
	recipeEntity, err := s.db.Recipe.Query().
		Where(
			recipe.IDEQ(recipeID),
			recipe.HouseholdIDEQ(householdID),
		).
		WithIngredients(func(query *entgen.RecipeIngredientQuery) {
			query.Order(entgen.Asc(recipeingredient.FieldSortOrder))
		}).
		WithSteps(func(query *entgen.RecipeStepQuery) {
			query.Order(entgen.Asc(recipestep.FieldSortOrder))
		}).
		WithTags(func(query *entgen.TagQuery) {
			query.Order(entgen.Asc(tag.FieldName))
		}).
		WithMediaAssets(func(query *entgen.MediaAssetQuery) {
			query.Order(entgen.Asc(mediaasset.FieldSortOrder))
		}).
		WithNutritionEntries(func(query *entgen.RecipeNutritionQuery) {
			query.Order(entgen.Asc(recipenutrition.FieldCreatedAt))
		}).
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return DetailView{}, httpx.StatusError{Status: http.StatusNotFound, Code: "recipe_not_found", Message: "Recipe was not found."}
		}

		return DetailView{}, fmt.Errorf("query recipe detail: %w", err)
	}

	return mapRecipeDetail(recipeEntity), nil
}

// AttachRecipeMedia stores and associates a media asset with a recipe.
func (s *Service) AttachRecipeMedia(ctx context.Context, in AttachRecipeMediaInput) (MediaView, error) {
	if !canEdit(in.ActorRole) {
		return MediaView{}, httpx.StatusError{Status: http.StatusForbidden, Code: "forbidden", Message: "The current account cannot modify recipes."}
	}

	recipeEntity, err := s.db.Recipe.Query().
		Where(recipe.IDEQ(in.RecipeID), recipe.HouseholdIDEQ(in.HouseholdID)).
		WithMediaAssets(func(query *entgen.MediaAssetQuery) {
			query.Order(entgen.Asc(mediaasset.FieldSortOrder))
		}).
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return MediaView{}, httpx.StatusError{Status: http.StatusNotFound, Code: "recipe_not_found", Message: "Recipe was not found."}
		}

		return MediaView{}, fmt.Errorf("query recipe for media upload: %w", err)
	}

	thumbnailContent, thumbnailMimeType, err := platformmedia.GenerateThumbnail(in.Content, platformmedia.DefaultThumbnailMaxDimension)
	if err != nil {
		return MediaView{}, fmt.Errorf("generate thumbnail: %w", err)
	}

	originalChecksum := checksum(in.Content)
	object, err := s.store.Put(ctx, storage.PutInput{
		HouseholdID:      in.HouseholdID,
		OriginalFilename: in.OriginalFilename,
		MimeType:         in.MimeType,
		Checksum:         originalChecksum,
		Content:          in.Content,
	})
	if err != nil {
		return MediaView{}, fmt.Errorf("store media object: %w", err)
	}

	thumbnailObject, err := s.store.Put(ctx, storage.PutInput{
		HouseholdID:      in.HouseholdID,
		OriginalFilename: platformmedia.ThumbnailFilename(in.OriginalFilename, thumbnailMimeType),
		MimeType:         thumbnailMimeType,
		Checksum:         checksum(thumbnailContent),
		Content:          thumbnailContent,
	})
	if err != nil {
		if cleanupErr := s.cleanupUnreferencedStoredObjects(ctx, in.HouseholdID, object); cleanupErr != nil {
			return MediaView{}, fmt.Errorf("store thumbnail object: %w; cleanup failed: %v", err, cleanupErr)
		}
		return MediaView{}, fmt.Errorf("store thumbnail object: %w", err)
	}

	nextSortOrder := len(recipeEntity.Edges.MediaAssets) + 1
	storedAt := s.clock.Now()

	tx, err := s.db.Tx(ctx)
	if err != nil {
		if cleanupErr := s.cleanupUnreferencedStoredObjects(ctx, in.HouseholdID, object, thumbnailObject); cleanupErr != nil {
			return MediaView{}, fmt.Errorf("start media tx: %w; cleanup failed: %v", err, cleanupErr)
		}
		return MediaView{}, fmt.Errorf("start media tx: %w", err)
	}

	mediaEntity, err := tx.MediaAsset.Create().
		SetHouseholdID(in.HouseholdID).
		SetRecipeID(recipeEntity.ID).
		SetStorageObjectID(object.ID).
		SetThumbnailStorageObjectID(thumbnailObject.ID).
		SetOriginalFilename(in.OriginalFilename).
		SetMimeType(in.MimeType).
		SetMediaType("image").
		SetSizeBytes(object.SizeBytes).
		SetChecksum(object.Checksum).
		SetStoredAt(storedAt).
		SetAltText(strings.TrimSpace(in.AltText)).
		SetSortOrder(nextSortOrder).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		if cleanupErr := s.cleanupUnreferencedStoredObjects(ctx, in.HouseholdID, object, thumbnailObject); cleanupErr != nil {
			return MediaView{}, fmt.Errorf("create media asset: %w; cleanup failed: %v", err, cleanupErr)
		}
		return MediaView{}, fmt.Errorf("create media asset: %w", err)
	}

	galleryMediaIDs := append(append([]string{}, recipeEntity.GalleryMediaIds...), mediaEntity.ID)
	updateRecipe := tx.Recipe.UpdateOneID(recipeEntity.ID).
		SetGalleryMediaIds(galleryMediaIDs)
	if recipeEntity.PrimaryMediaID == nil {
		updateRecipe.SetPrimaryMediaID(mediaEntity.ID)
	}
	if _, err := updateRecipe.Save(ctx); err != nil {
		_ = tx.Rollback()
		if cleanupErr := s.cleanupUnreferencedStoredObjects(ctx, in.HouseholdID, object, thumbnailObject); cleanupErr != nil {
			return MediaView{}, fmt.Errorf("update recipe media references: %w; cleanup failed: %v", err, cleanupErr)
		}
		return MediaView{}, fmt.Errorf("update recipe media references: %w", err)
	}

	if err := tx.Commit(); err != nil {
		if cleanupErr := s.cleanupUnreferencedStoredObjects(ctx, in.HouseholdID, object, thumbnailObject); cleanupErr != nil {
			return MediaView{}, fmt.Errorf("commit media tx: %w; cleanup failed: %v", err, cleanupErr)
		}
		return MediaView{}, fmt.Errorf("commit media tx: %w", err)
	}

	return mapMediaAsset(mediaEntity), nil
}

// GetMediaOriginal loads the original stored media content.
func (s *Service) GetMediaOriginal(ctx context.Context, householdID string, mediaID string) (MediaContentView, error) {
	return s.getMediaContent(ctx, householdID, mediaID, false)
}

// GetMediaThumbnail loads the thumbnail stored media content.
func (s *Service) GetMediaThumbnail(ctx context.Context, householdID string, mediaID string) (MediaContentView, error) {
	return s.getMediaContent(ctx, householdID, mediaID, true)
}

func (s *Service) getMediaContent(ctx context.Context, householdID string, mediaID string, thumbnail bool) (MediaContentView, error) {
	mediaEntity, err := s.db.MediaAsset.Query().
		Where(
			mediaasset.IDEQ(mediaID),
			mediaasset.HouseholdIDEQ(householdID),
		).
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return MediaContentView{}, httpx.StatusError{Status: http.StatusNotFound, Code: "media_not_found", Message: "Media was not found."}
		}

		return MediaContentView{}, fmt.Errorf("query media asset: %w", err)
	}

	objectID := mediaEntity.StorageObjectID
	filename := mediaEntity.OriginalFilename
	mimeType := mediaEntity.MimeType
	if thumbnail {
		if mediaEntity.ThumbnailStorageObjectID == nil {
			return MediaContentView{}, httpx.StatusError{
				Status:  http.StatusNotFound,
				Code:    "media_variant_not_found",
				Message: "Media variant was not found.",
			}
		}
		objectID = *mediaEntity.ThumbnailStorageObjectID
	}

	object, err := s.store.Get(ctx, householdID, objectID)
	if err != nil {
		return MediaContentView{}, fmt.Errorf("load stored object: %w", err)
	}
	if thumbnail {
		filename = platformmedia.ThumbnailFilename(mediaEntity.OriginalFilename, object.MimeType)
		mimeType = object.MimeType
	}

	return MediaContentView{
		Filename:  filename,
		MimeType:  mimeType,
		SizeBytes: object.SizeBytes,
		Content:   object.Content,
	}, nil
}

func canEdit(role string) bool {
	return role == string(auth.HouseholdRoleOwner) || role == string(auth.HouseholdRoleEditor)
}

func checksum(content []byte) string {
	sum := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func isValidUnit(unit string) bool {
	for _, candidate := range entschema.UnitValues {
		if candidate == strings.TrimSpace(unit) {
			return true
		}
	}

	return false
}

func encodeRecipeCursor(cursor recipeCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeRecipeCursor(raw string) (recipeCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return recipeCursor{}, err
	}

	var cursor recipeCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return recipeCursor{}, err
	}
	if cursor.Offset < 0 {
		return recipeCursor{}, fmt.Errorf("cursor missing fields")
	}

	return cursor, nil
}

func offsetFromCursor(raw *string) int {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return 0
	}
	cursor, err := decodeRecipeCursor(*raw)
	if err != nil {
		return 0
	}
	return cursor.Offset
}

func validateIngredients(ingredients []CreateRecipeIngredientInput) error {
	if len(ingredients) == 0 {
		return httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "At least one ingredient is required."}
	}

	seenSortOrder := make(map[int]struct{}, len(ingredients))
	for _, ingredient := range ingredients {
		if strings.TrimSpace(ingredient.Name) == "" {
			return httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Ingredient name is required."}
		}
		if ingredient.SortOrder <= 0 {
			return httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Ingredient sort order must be greater than zero."}
		}
		if _, ok := seenSortOrder[ingredient.SortOrder]; ok {
			return httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Ingredient sort order must be unique."}
		}
		seenSortOrder[ingredient.SortOrder] = struct{}{}

		if ingredient.Unit != nil && strings.TrimSpace(*ingredient.Unit) != "" && !isValidUnit(*ingredient.Unit) {
			return httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Ingredient unit is invalid."}
		}
	}

	return nil
}

func validateSteps(steps []UpdateRecipeStepInput) error {
	seenSortOrder := make(map[int]struct{}, len(steps))
	for _, step := range steps {
		if strings.TrimSpace(step.Instruction) == "" {
			return httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Recipe step instruction is required."}
		}
		if step.SortOrder <= 0 {
			return httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Recipe step sort order must be greater than zero."}
		}
		if _, ok := seenSortOrder[step.SortOrder]; ok {
			return httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Recipe step sort order must be unique."}
		}
		seenSortOrder[step.SortOrder] = struct{}{}
	}

	return nil
}

func validateNutritionEntries(entries []UpdateRecipeNutritionInput) error {
	seenReferenceQuantity := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		key := ""
		if entry.ReferenceQuantity != nil {
			key = strings.TrimSpace(*entry.ReferenceQuantity)
		}
		if _, ok := seenReferenceQuantity[key]; ok {
			return httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Nutrition reference quantity must be unique."}
		}
		seenReferenceQuantity[key] = struct{}{}
	}

	return nil
}

func uniqueIDs(values []string, field string) ([]string, error) {
	seen := make(map[string]struct{}, len(values))
	items := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: fmt.Sprintf("%s must not contain empty values.", field)}
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		items = append(items, trimmed)
	}
	return items, nil
}

func normalizeTagSlugs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	items := make([]string, 0, len(values))
	for _, value := range values {
		slug := normalizeSlug(value)
		if slug == "" {
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		items = append(items, slug)
	}
	return items
}

func normalizeSlug(value string) string {
	var builder strings.Builder
	lastDash := false
	for _, r := range strings.TrimSpace(strings.ToLower(value)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastDash = false
		case !lastDash:
			builder.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

func loadTagsByID(ctx context.Context, tx *ent.Tx, householdID string, tagIDs []string) ([]*entgen.Tag, error) {
	if len(tagIDs) == 0 {
		return nil, nil
	}

	entities, err := tx.Tag.Query().
		Where(
			tag.HouseholdIDEQ(householdID),
			tag.IDIn(tagIDs...),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	if len(entities) != len(tagIDs) {
		return nil, httpx.StatusError{
			Status:  http.StatusBadRequest,
			Code:    "validation_error",
			Message: "One or more tags are invalid for the active household.",
		}
	}

	return entities, nil
}

func replaceIngredients(ctx context.Context, tx *ent.Tx, recipeID string, ingredients []CreateRecipeIngredientInput) error {
	if _, err := tx.RecipeIngredient.Delete().
		Where(recipeingredient.RecipeIDEQ(recipeID)).
		Exec(ctx); err != nil {
		return fmt.Errorf("delete recipe ingredients: %w", err)
	}

	for _, ingredientInput := range ingredients {
		builder := tx.RecipeIngredient.Create().
			SetRecipeID(recipeID).
			SetName(strings.TrimSpace(ingredientInput.Name)).
			SetSortOrder(ingredientInput.SortOrder)

		if ingredientInput.Quantity != nil {
			builder.SetQuantity(*ingredientInput.Quantity)
		}
		if ingredientInput.Unit != nil && strings.TrimSpace(*ingredientInput.Unit) != "" {
			builder.SetUnit(recipeingredient.Unit(strings.TrimSpace(*ingredientInput.Unit)))
		}
		if ingredientInput.Preparation != nil && strings.TrimSpace(*ingredientInput.Preparation) != "" {
			builder.SetPreparation(strings.TrimSpace(*ingredientInput.Preparation))
		}

		if _, err := builder.Save(ctx); err != nil {
			return fmt.Errorf("create recipe ingredient: %w", err)
		}
	}

	return nil
}

func replaceSteps(ctx context.Context, tx *ent.Tx, recipeID string, steps []UpdateRecipeStepInput) error {
	if _, err := tx.RecipeStep.Delete().
		Where(recipestep.RecipeIDEQ(recipeID)).
		Exec(ctx); err != nil {
		return fmt.Errorf("delete recipe steps: %w", err)
	}

	for _, stepInput := range steps {
		builder := tx.RecipeStep.Create().
			SetRecipeID(recipeID).
			SetInstruction(strings.TrimSpace(stepInput.Instruction)).
			SetSortOrder(stepInput.SortOrder)

		if stepInput.DurationMinutes != nil {
			builder.SetDurationMinutes(*stepInput.DurationMinutes)
		}
		if stepInput.Tip != nil && strings.TrimSpace(*stepInput.Tip) != "" {
			builder.SetTip(strings.TrimSpace(*stepInput.Tip))
		}

		if _, err := builder.Save(ctx); err != nil {
			return fmt.Errorf("create recipe step: %w", err)
		}
	}

	return nil
}

func replaceNutritionEntries(ctx context.Context, tx *ent.Tx, recipeID string, entries []UpdateRecipeNutritionInput) error {
	if _, err := tx.RecipeNutrition.Delete().
		Where(recipenutrition.RecipeIDEQ(recipeID)).
		Exec(ctx); err != nil {
		return fmt.Errorf("delete recipe nutrition: %w", err)
	}

	for _, entryInput := range entries {
		builder := tx.RecipeNutrition.Create().
			SetRecipeID(recipeID)

		if entryInput.ReferenceQuantity != nil && strings.TrimSpace(*entryInput.ReferenceQuantity) != "" {
			builder.SetReferenceQuantity(strings.TrimSpace(*entryInput.ReferenceQuantity))
		}
		if entryInput.EnergyKcal != nil {
			builder.SetEnergyKcal(*entryInput.EnergyKcal)
		}
		if entryInput.Protein != nil {
			builder.SetProtein(*entryInput.Protein)
		}
		if entryInput.Carbohydrates != nil {
			builder.SetCarbohydrates(*entryInput.Carbohydrates)
		}
		if entryInput.Fat != nil {
			builder.SetFat(*entryInput.Fat)
		}
		if entryInput.SaturatedFat != nil {
			builder.SetSaturatedFat(*entryInput.SaturatedFat)
		}
		if entryInput.Fiber != nil {
			builder.SetFiber(*entryInput.Fiber)
		}
		if entryInput.Sugars != nil {
			builder.SetSugars(*entryInput.Sugars)
		}
		if entryInput.Sodium != nil {
			builder.SetSodium(*entryInput.Sodium)
		}
		if entryInput.Salt != nil {
			builder.SetSalt(*entryInput.Salt)
		}

		if _, err := builder.Save(ctx); err != nil {
			return fmt.Errorf("create recipe nutrition: %w", err)
		}
	}

	return nil
}

func replaceTagAssignments(ctx context.Context, tx *ent.Tx, recipeID string, tagIDs []string) error {
	update := tx.Recipe.UpdateOneID(recipeID).ClearTags()
	if len(tagIDs) > 0 {
		update.AddTagIDs(tagIDs...)
	}
	if _, err := update.Save(ctx); err != nil {
		return fmt.Errorf("replace recipe tags: %w", err)
	}
	return nil
}

func mapRecipeDetail(entity *entgen.Recipe) DetailView {
	return DetailView{
		Recipe:           mapRecipeSummary(entity),
		Ingredients:      mapIngredients(entity.Edges.Ingredients),
		Steps:            mapSteps(entity.Edges.Steps),
		Tags:             mapTags(entity.Edges.Tags),
		MediaAssets:      mapMediaAssets(entity.Edges.MediaAssets),
		NutritionEntries: mapNutrition(entity.Edges.NutritionEntries),
	}
}

func mapRecipeSummary(entity *entgen.Recipe) RecipeSummaryView {
	var caloriesPerServe *int
	if len(entity.Edges.NutritionEntries) > 0 {
		caloriesPerServe = entity.Edges.NutritionEntries[0].EnergyKcal
	}

	return RecipeSummaryView{
		ID:               entity.ID,
		Title:            entity.Title,
		Description:      entity.Description,
		SourceURL:        entity.SourceURL,
		SourceCapturedAt: entity.SourceCapturedAt,
		PrimaryMediaID:   entity.PrimaryMediaID,
		GalleryMediaIDs:  entity.GalleryMediaIds,
		PrepMinutes:      entity.PrepMinutes,
		CookMinutes:      entity.CookMinutes,
		Servings:         entity.Servings,
		Region:           recipeEnumPtrToString(entity.Region),
		MealType:         mealTypePtrToString(entity.MealType),
		Difficulty:       difficultyPtrToString(entity.Difficulty),
		Cuisine:          cuisinePtrToString(entity.Cuisine),
		Allergens:        entity.Allergens,
		CaloriesPerServe: caloriesPerServe,
		Version:          entity.Version,
	}
}

func mapIngredients(entities []*entgen.RecipeIngredient) []IngredientView {
	items := make([]IngredientView, 0, len(entities))
	for _, entity := range entities {
		items = append(items, IngredientView{
			ID:          entity.ID,
			Name:        entity.Name,
			Quantity:    entity.Quantity,
			Unit:        unitPtrToString(entity.Unit),
			Preparation: entity.Preparation,
			SortOrder:   entity.SortOrder,
		})
	}
	return items
}

func mapSteps(entities []*entgen.RecipeStep) []StepView {
	items := make([]StepView, 0, len(entities))
	for _, entity := range entities {
		items = append(items, StepView{
			ID:              entity.ID,
			SortOrder:       entity.SortOrder,
			Instruction:     entity.Instruction,
			DurationMinutes: entity.DurationMinutes,
			Tip:             entity.Tip,
		})
	}
	return items
}

func mapTags(entities []*entgen.Tag) []TagView {
	items := make([]TagView, 0, len(entities))
	for _, entity := range entities {
		items = append(items, TagView{
			ID:     entity.ID,
			Name:   entity.Name,
			Slug:   entity.Slug,
			Color:  entity.Color,
			System: entity.System,
		})
	}
	return items
}

func mapMediaAssets(entities []*entgen.MediaAsset) []MediaView {
	items := make([]MediaView, 0, len(entities))
	for _, entity := range entities {
		items = append(items, mapMediaAsset(entity))
	}
	return items
}

func mapMediaAsset(entity *entgen.MediaAsset) MediaView {
	url := fmt.Sprintf("/api/v1/media/%s/original", entity.ID)
	thumbnailURL := url
	if entity.ThumbnailStorageObjectID != nil {
		thumbnailURL = fmt.Sprintf("/api/v1/media/%s/thumbnail", entity.ID)
	}
	return MediaView{
		ID:               entity.ID,
		OriginalFilename: entity.OriginalFilename,
		MimeType:         entity.MimeType,
		MediaType:        string(entity.MediaType),
		SizeBytes:        entity.SizeBytes,
		Checksum:         entity.Checksum,
		StoredAt:         entity.StoredAt,
		URL:              url,
		ThumbnailURL:     thumbnailURL,
		AltText:          entity.AltText,
		SortOrder:        entity.SortOrder,
	}
}

func mapNutrition(entities []*entgen.RecipeNutrition) []NutritionView {
	items := make([]NutritionView, 0, len(entities))
	for _, entity := range entities {
		items = append(items, NutritionView{
			ID:                entity.ID,
			ReferenceQuantity: entity.ReferenceQuantity,
			EnergyKcal:        entity.EnergyKcal,
			Protein:           entity.Protein,
			Carbohydrates:     entity.Carbohydrates,
			Fat:               entity.Fat,
			SaturatedFat:      entity.SaturatedFat,
			Fiber:             entity.Fiber,
			Sugars:            entity.Sugars,
			Sodium:            entity.Sodium,
			Salt:              entity.Salt,
		})
	}
	return items
}

func unitPtrToString(value *recipeingredient.Unit) *string {
	if value == nil {
		return nil
	}

	stringValue := string(*value)
	return &stringValue
}

func recipeEnumPtrToString(value *recipe.Region) *string {
	if value == nil {
		return nil
	}

	stringValue := string(*value)
	return &stringValue
}

func mealTypePtrToString(value *recipe.MealType) *string {
	if value == nil {
		return nil
	}

	stringValue := string(*value)
	return &stringValue
}

func difficultyPtrToString(value *recipe.Difficulty) *string {
	if value == nil {
		return nil
	}

	stringValue := string(*value)
	return &stringValue
}

func cuisinePtrToString(value *recipe.Cuisine) *string {
	if value == nil {
		return nil
	}

	stringValue := string(*value)
	return &stringValue
}

func min(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func (s *Service) cleanupUnreferencedStoredObjects(ctx context.Context, householdID string, objects ...storage.Object) error {
	for _, object := range objects {
		if !object.Created || strings.TrimSpace(object.ID) == "" {
			continue
		}

		referenced, err := s.db.MediaAsset.Query().
			Where(mediaasset.Or(
				mediaasset.StorageObjectIDEQ(object.ID),
				mediaasset.ThumbnailStorageObjectIDEQ(object.ID),
			)).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("check stored object references: %w", err)
		}
		if referenced {
			continue
		}

		if err := s.store.Delete(ctx, householdID, object.ID); err != nil {
			return fmt.Errorf("delete stored object: %w", err)
		}
	}

	return nil
}
