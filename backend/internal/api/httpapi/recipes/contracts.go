package httpapi

import (
	"time"

	"github.com/AndreasRoither/NomNomVault/backend/internal/api/apicontract"
)

// RecipeSummary is the list representation for a recipe.
type RecipeSummary struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	SourceURL        string     `json:"sourceUrl"`
	SourceCapturedAt *time.Time `json:"sourceCapturedAt,omitempty"`
	PrimaryMediaID   *string    `json:"primaryMediaId,omitempty"`
	GalleryMediaIDs  []string   `json:"galleryMediaIds,omitempty"`
	PrepMinutes      *int       `json:"prepMinutes,omitempty"`
	CookMinutes      *int       `json:"cookMinutes,omitempty"`
	Servings         *int       `json:"servings,omitempty"`
	Region           *string    `json:"region,omitempty"`
	MealType         *string    `json:"mealType,omitempty"`
	Difficulty       *string    `json:"difficulty,omitempty"`
	Cuisine          *string    `json:"cuisine,omitempty"`
	Allergens        []string   `json:"allergens,omitempty"`
	CaloriesPerServe *int       `json:"caloriesPerServe,omitempty"`
	Version          int        `json:"version"`
}

// RecipeIngredientItem is the API representation for a recipe ingredient row.
type RecipeIngredientItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Quantity    *float64 `json:"quantity,omitempty"`
	Unit        *string  `json:"unit,omitempty"`
	Preparation *string  `json:"preparation,omitempty"`
	SortOrder   int      `json:"sortOrder"`
}

// RecipeStepItem is the API representation for a recipe instruction step.
type RecipeStepItem struct {
	ID              string  `json:"id"`
	SortOrder       int     `json:"sortOrder"`
	Instruction     string  `json:"instruction"`
	DurationMinutes *int    `json:"durationMinutes,omitempty"`
	Tip             *string `json:"tip,omitempty"`
}

// RecipeTagItem is the API representation for a recipe tag.
type RecipeTagItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Color  string `json:"color"`
	System bool   `json:"system"`
}

// RecipeMediaItem is the API representation for a recipe media asset.
type RecipeMediaItem struct {
	ID               string    `json:"id"`
	OriginalFilename string    `json:"originalFilename"`
	MimeType         string    `json:"mimeType"`
	MediaType        string    `json:"mediaType"`
	SizeBytes        int64     `json:"sizeBytes"`
	Checksum         string    `json:"checksum"`
	StoredAt         time.Time `json:"storedAt"`
	URL              string    `json:"url"`
	ThumbnailURL     string    `json:"thumbnailUrl"`
	AltText          string    `json:"altText"`
	SortOrder        int       `json:"sortOrder"`
}

// RecipeNutritionItem is the API representation for one nutrition row.
type RecipeNutritionItem struct {
	ID                string   `json:"id"`
	ReferenceQuantity *string  `json:"referenceQuantity,omitempty"`
	EnergyKcal        *int     `json:"energyKcal,omitempty"`
	Protein           *float64 `json:"protein,omitempty"`
	Carbohydrates     *float64 `json:"carbohydrates,omitempty"`
	Fat               *float64 `json:"fat,omitempty"`
	SaturatedFat      *float64 `json:"saturatedFat,omitempty"`
	Fiber             *float64 `json:"fiber,omitempty"`
	Sugars            *float64 `json:"sugars,omitempty"`
	Sodium            *float64 `json:"sodium,omitempty"`
	Salt              *float64 `json:"salt,omitempty"`
}

// RecipeDetailResponse is the detail representation for a recipe.
type RecipeDetailResponse struct {
	Recipe           RecipeSummary          `json:"recipe"`
	Ingredients      []RecipeIngredientItem `json:"ingredients"`
	Steps            []RecipeStepItem       `json:"steps"`
	Tags             []RecipeTagItem        `json:"tags"`
	MediaAssets      []RecipeMediaItem      `json:"mediaAssets"`
	NutritionEntries []RecipeNutritionItem  `json:"nutritionEntries"`
}

// RecipeListResponse is the list representation for recipe browsing.
type RecipeListResponse struct {
	Data []RecipeSummary            `json:"data"`
	Page apicontract.CursorPageInfo `json:"page"`
}

// TagListResponse is the household-scoped list representation for tags.
type TagListResponse struct {
	Data []RecipeTagItem `json:"data"`
}

// CreateRecipeIngredientRequest is the API payload for one ingredient row.
type CreateRecipeIngredientRequest struct {
	Name        string   `json:"name" binding:"required"`
	Quantity    *float64 `json:"quantity,omitempty"`
	Unit        *string  `json:"unit,omitempty"`
	Preparation *string  `json:"preparation,omitempty"`
	SortOrder   int      `json:"sortOrder" binding:"required"`
}

// CreateRecipeRequest is the API payload for creating a recipe.
type CreateRecipeRequest struct {
	Title       string                          `json:"title" binding:"required"`
	Description string                          `json:"description"`
	SourceURL   string                          `json:"sourceUrl"`
	PrepMinutes *int                            `json:"prepMinutes,omitempty"`
	CookMinutes *int                            `json:"cookMinutes,omitempty"`
	Servings    *int                            `json:"servings,omitempty"`
	Ingredients []CreateRecipeIngredientRequest `json:"ingredients" binding:"required"`
}

// UpdateRecipeStepRequest is the replace payload for one recipe step.
type UpdateRecipeStepRequest struct {
	Instruction     string  `json:"instruction" binding:"required"`
	SortOrder       int     `json:"sortOrder" binding:"required"`
	DurationMinutes *int    `json:"durationMinutes,omitempty"`
	Tip             *string `json:"tip,omitempty"`
}

// UpdateRecipeNutritionRequest is the replace payload for one recipe nutrition entry.
type UpdateRecipeNutritionRequest struct {
	ReferenceQuantity *string  `json:"referenceQuantity,omitempty"`
	EnergyKcal        *int     `json:"energyKcal,omitempty"`
	Protein           *float64 `json:"protein,omitempty"`
	Carbohydrates     *float64 `json:"carbohydrates,omitempty"`
	Fat               *float64 `json:"fat,omitempty"`
	SaturatedFat      *float64 `json:"saturatedFat,omitempty"`
	Fiber             *float64 `json:"fiber,omitempty"`
	Sugars            *float64 `json:"sugars,omitempty"`
	Sodium            *float64 `json:"sodium,omitempty"`
	Salt              *float64 `json:"salt,omitempty"`
}

// UpdateRecipeRequest replaces the editable recipe fields and collections.
type UpdateRecipeRequest struct {
	Version          int                             `json:"version" binding:"required"`
	Title            string                          `json:"title" binding:"required"`
	Description      string                          `json:"description"`
	SourceURL        string                          `json:"sourceUrl"`
	PrepMinutes      *int                            `json:"prepMinutes,omitempty"`
	CookMinutes      *int                            `json:"cookMinutes,omitempty"`
	Servings         *int                            `json:"servings,omitempty"`
	Ingredients      []CreateRecipeIngredientRequest `json:"ingredients" binding:"required"`
	Steps            []UpdateRecipeStepRequest       `json:"steps"`
	NutritionEntries []UpdateRecipeNutritionRequest  `json:"nutritionEntries"`
	TagIDs           []string                        `json:"tagIds"`
}

// CreateTagRequest creates one household-scoped tag.
type CreateTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}
