package recipes

import "time"

// RecipeSummaryView is the persisted recipe list/detail header representation.
type RecipeSummaryView struct {
	ID               string
	Title            string
	Description      string
	SourceURL        string
	SourceCapturedAt *time.Time
	PrimaryMediaID   *string
	GalleryMediaIDs  []string
	PrepMinutes      *int
	CookMinutes      *int
	Servings         *int
	Region           *string
	MealType         *string
	Difficulty       *string
	Cuisine          *string
	Allergens        []string
	CaloriesPerServe *int
	Version          int
}

// RecipeListResult wraps a list page plus pagination metadata.
type RecipeListResult struct {
	Items      []RecipeSummaryView
	NextCursor *string
	HasMore    bool
}

// IngredientView is the persisted ingredient representation.
type IngredientView struct {
	ID          string
	Name        string
	Quantity    *float64
	Unit        *string
	Preparation *string
	SortOrder   int
}

// StepView is the persisted step representation.
type StepView struct {
	ID              string
	SortOrder       int
	Instruction     string
	DurationMinutes *int
	Tip             *string
}

// TagView is the persisted tag representation.
type TagView struct {
	ID     string
	Name   string
	Slug   string
	Color  string
	System bool
}

// MediaView is the recipe media representation.
type MediaView struct {
	ID               string
	OriginalFilename string
	MimeType         string
	MediaType        string
	SizeBytes        int64
	Checksum         string
	StoredAt         time.Time
	URL              string
	ThumbnailURL     string
	AltText          string
	SortOrder        int
}

// NutritionView is the persisted nutrition representation.
type NutritionView struct {
	ID                string
	ReferenceQuantity *string
	EnergyKcal        *int
	Protein           *float64
	Carbohydrates     *float64
	Fat               *float64
	SaturatedFat      *float64
	Fiber             *float64
	Sugars            *float64
	Sodium            *float64
	Salt              *float64
}

// DetailView is the full persisted recipe representation.
type DetailView struct {
	Recipe           RecipeSummaryView
	Ingredients      []IngredientView
	Steps            []StepView
	Tags             []TagView
	MediaAssets      []MediaView
	NutritionEntries []NutritionView
}

// CreateRecipeIngredientInput is the create payload for one ingredient.
type CreateRecipeIngredientInput struct {
	Name        string
	Quantity    *float64
	Unit        *string
	Preparation *string
	SortOrder   int
}

// UpdateRecipeStepInput is the replace payload for one recipe step.
type UpdateRecipeStepInput struct {
	Instruction     string
	SortOrder       int
	DurationMinutes *int
	Tip             *string
}

// UpdateRecipeNutritionInput is the replace payload for one nutrition entry.
type UpdateRecipeNutritionInput struct {
	ReferenceQuantity *string
	EnergyKcal        *int
	Protein           *float64
	Carbohydrates     *float64
	Fat               *float64
	SaturatedFat      *float64
	Fiber             *float64
	Sugars            *float64
	Sodium            *float64
	Salt              *float64
}

// CreateRecipeInput creates a household-scoped recipe.
type CreateRecipeInput struct {
	HouseholdID string
	ActorUserID string
	ActorRole   string
	Title       string
	Description string
	SourceURL   string
	PrepMinutes *int
	CookMinutes *int
	Servings    *int
	Ingredients []CreateRecipeIngredientInput
}

// CreateRecipeResult wraps a persisted recipe detail.
type CreateRecipeResult struct {
	Recipe DetailView
}

// ListRecipesInput defines the available list and search filters.
type ListRecipesInput struct {
	HouseholdID string
	Cursor      *string
	Limit       int
	Query       string
	TagSlugs    []string
}

// UpdateRecipeInput replaces the editable fields on a recipe.
type UpdateRecipeInput struct {
	HouseholdID      string
	ActorUserID      string
	ActorRole        string
	RecipeID         string
	ExpectedVersion  int
	Title            string
	Description      string
	SourceURL        string
	PrepMinutes      *int
	CookMinutes      *int
	Servings         *int
	Ingredients      []CreateRecipeIngredientInput
	Steps            []UpdateRecipeStepInput
	NutritionEntries []UpdateRecipeNutritionInput
	TagIDs           []string
}

// ArchiveRecipeInput toggles recipe archive state.
type ArchiveRecipeInput struct {
	HouseholdID string
	ActorUserID string
	ActorRole   string
	RecipeID    string
}

// CreateTagInput creates one household-scoped tag.
type CreateTagInput struct {
	HouseholdID string
	ActorUserID string
	ActorRole   string
	Name        string
	Color       string
}

// AttachRecipeMediaInput uploads and attaches one recipe image.
type AttachRecipeMediaInput struct {
	HouseholdID      string
	ActorUserID      string
	ActorRole        string
	RecipeID         string
	OriginalFilename string
	MimeType         string
	AltText          string
	Content          []byte
}

// MediaContentView is the streamed blob payload.
type MediaContentView struct {
	Filename  string
	MimeType  string
	SizeBytes int64
	Content   []byte
}
