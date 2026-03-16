package recipes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

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
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/storage"
)

// Service implements the first-slice persisted recipe flows.
type Service struct {
	db    *ent.Client
	store storage.Store
	clock clock.Clock
}

// NewService creates a recipe service.
func NewService(db *ent.Client, store storage.Store, clock clock.Clock) *Service {
	return &Service{db: db, store: store, clock: clock}
}

// ListRecipes loads recipe summaries for one household.
func (s *Service) ListRecipes(ctx context.Context, householdID string, limit int) ([]RecipeSummaryView, error) {
	if limit <= 0 {
		limit = 20
	}

	entities, err := s.db.Recipe.Query().
		Where(recipe.HouseholdIDEQ(householdID)).
		Order(entgen.Desc(recipe.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query recipe list: %w", err)
	}

	items := make([]RecipeSummaryView, 0, len(entities))
	for _, entity := range entities {
		items = append(items, mapRecipeSummary(entity))
	}
	return items, nil
}

// CreateRecipe persists a minimal household-scoped recipe and its ingredients.
func (s *Service) CreateRecipe(ctx context.Context, in CreateRecipeInput) (CreateRecipeResult, error) {
	if !canEdit(in.ActorRole) {
		return CreateRecipeResult{}, httpx.StatusError{Status: http.StatusForbidden, Code: "forbidden", Message: "The current account cannot modify recipes."}
	}
	if strings.TrimSpace(in.Title) == "" {
		return CreateRecipeResult{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Recipe title is required."}
	}
	if len(in.Ingredients) == 0 {
		return CreateRecipeResult{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "At least one ingredient is required."}
	}

	for _, ingredient := range in.Ingredients {
		if strings.TrimSpace(ingredient.Name) == "" {
			return CreateRecipeResult{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Ingredient name is required."}
		}
		if ingredient.SortOrder <= 0 {
			return CreateRecipeResult{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Ingredient sort order must be greater than zero."}
		}
		if ingredient.Unit != nil && strings.TrimSpace(*ingredient.Unit) != "" && !isValidUnit(*ingredient.Unit) {
			return CreateRecipeResult{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Ingredient unit is invalid."}
		}
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

	for _, ingredientInput := range in.Ingredients {
		builder := tx.RecipeIngredient.Create().
			SetRecipeID(recipeEntity.ID).
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
			_ = tx.Rollback()
			return CreateRecipeResult{}, fmt.Errorf("create recipe ingredient: %w", err)
		}
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

	checksum := checksum(in.Content)
	object, err := s.store.Put(ctx, storage.PutInput{
		HouseholdID:      in.HouseholdID,
		OriginalFilename: in.OriginalFilename,
		MimeType:         in.MimeType,
		Checksum:         checksum,
		Content:          in.Content,
	})
	if err != nil {
		return MediaView{}, fmt.Errorf("store media object: %w", err)
	}

	nextSortOrder := len(recipeEntity.Edges.MediaAssets) + 1
	storedAt := s.clock.Now()

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return MediaView{}, fmt.Errorf("start media tx: %w", err)
	}

	mediaEntity, err := tx.MediaAsset.Create().
		SetHouseholdID(in.HouseholdID).
		SetRecipeID(recipeEntity.ID).
		SetStorageObjectID(object.ID).
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
		return MediaView{}, fmt.Errorf("update recipe media references: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return MediaView{}, fmt.Errorf("commit media tx: %w", err)
	}

	return mapMediaAsset(mediaEntity), nil
}

// GetMediaOriginal loads the original stored media content.
func (s *Service) GetMediaOriginal(ctx context.Context, householdID string, mediaID string) (MediaContentView, error) {
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

	object, err := s.store.Get(ctx, householdID, mediaEntity.StorageObjectID)
	if err != nil {
		return MediaContentView{}, fmt.Errorf("load stored object: %w", err)
	}

	return MediaContentView{
		Filename:  mediaEntity.OriginalFilename,
		MimeType:  mediaEntity.MimeType,
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
	return MediaView{
		ID:               entity.ID,
		OriginalFilename: entity.OriginalFilename,
		MimeType:         entity.MimeType,
		MediaType:        string(entity.MediaType),
		SizeBytes:        entity.SizeBytes,
		Checksum:         entity.Checksum,
		StoredAt:         entity.StoredAt,
		URL:              url,
		ThumbnailURL:     url,
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
