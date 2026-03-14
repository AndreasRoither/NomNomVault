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
	SourceURL        string     `json:"sourceUrl"`
	SourceCapturedAt *time.Time `json:"sourceCapturedAt,omitempty"`
	PrimaryMediaID   *string    `json:"primaryMediaId,omitempty"`
	PrepMinutes      *int       `json:"prepMinutes,omitempty"`
	CookMinutes      *int       `json:"cookMinutes,omitempty"`
	Servings         *int       `json:"servings,omitempty"`
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
}

// RecipeDetailResponse is the detail representation for a recipe.
type RecipeDetailResponse struct {
	Recipe      RecipeSummary          `json:"recipe"`
	Ingredients []RecipeIngredientItem `json:"ingredients"`
	Steps       []RecipeStepItem       `json:"steps"`
	Tags        []RecipeTagItem        `json:"tags"`
	MediaAssets []RecipeMediaItem      `json:"mediaAssets"`
}

// RecipeListResponse is the list representation for recipe browsing.
type RecipeListResponse struct {
	Data []RecipeSummary            `json:"data"`
	Page apicontract.CursorPageInfo `json:"page"`
}
