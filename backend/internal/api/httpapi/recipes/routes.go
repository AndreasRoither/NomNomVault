package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AndreasRoither/NomNomVault/backend/internal/api/apicontract"
)

// RegisterRoutes binds the recipe endpoints for the first generated API contract.
func RegisterRoutes(api *gin.RouterGroup) {
	api.GET("/recipes", listRecipes)
	api.GET("/recipes/:recipeId", getRecipe)
}

// listRecipes godoc
// @Summary List recipes
// @Description Return the current household recipe list using a cursor pagination envelope.
// @Tags recipes
// @Produce json
// @Param cursor query string false "Cursor token"
// @Param limit query int false "Maximum number of recipes to return"
// @Success 200 {object} RecipeListResponse
// @Router /recipes [get]
func listRecipes(context *gin.Context) {
	context.JSON(http.StatusOK, RecipeListResponse{
		Data: []RecipeSummary{sampleRecipeSummary()},
		Page: apicontract.CursorPageInfo{
			NextCursor: nil,
			HasMore:    false,
		},
	})
}

// getRecipe godoc
// @Summary Fetch a recipe detail
// @Description Return the detailed recipe payload for the requested recipe ID.
// @Tags recipes
// @Produce json
// @Param recipeId path string true "Recipe ID"
// @Success 200 {object} RecipeDetailResponse
// @Failure 404 {object} apicontract.ErrorResponse
// @Router /recipes/{recipeId} [get]
func getRecipe(context *gin.Context) {
	if context.Param("recipeId") != "recipe_sample" {
		apicontract.WriteError(context, http.StatusNotFound, "recipe_not_found", "Recipe was not found.", nil)
		return
	}

	context.JSON(http.StatusOK, sampleRecipeDetail())
}

func sampleRecipeSummary() RecipeSummary {
	sourceCapturedAt := time.Date(2026, time.March, 12, 9, 0, 0, 0, time.UTC)
	primaryMediaID := "media_sample"
	prepMinutes := 15
	cookMinutes := 30
	servings := 4

	return RecipeSummary{
		ID:               "recipe_sample",
		Title:            "Tomato Basil Soup",
		Description:      "A sample recipe used to prove the generated contract.",
		SourceURL:        "https://example.com/tomato-basil-soup",
		SourceCapturedAt: &sourceCapturedAt,
		PrimaryMediaID:   &primaryMediaID,
		PrepMinutes:      &prepMinutes,
		CookMinutes:      &cookMinutes,
		Servings:         &servings,
		Version:          1,
	}
}

func sampleRecipeDetail() RecipeDetailResponse {
	quantityOne := 1.0
	quantityTwo := 2.0
	unitCan := "can"
	unitCup := "cup"
	tip := "Blend until smooth."
	stepDuration := 20
	storedAt := time.Date(2026, time.March, 12, 9, 15, 0, 0, time.UTC)

	return RecipeDetailResponse{
		Recipe: sampleRecipeSummary(),
		Ingredients: []RecipeIngredientItem{
			{ID: "ingredient_1", Name: "Tomatoes", Quantity: &quantityTwo, Unit: &unitCan, SortOrder: 1},
			{ID: "ingredient_2", Name: "Vegetable broth", Quantity: &quantityOne, Unit: &unitCup, SortOrder: 2},
		},
		Steps: []RecipeStepItem{
			{ID: "step_1", SortOrder: 1, Instruction: "Simmer the tomatoes and broth.", DurationMinutes: &stepDuration},
			{ID: "step_2", SortOrder: 2, Instruction: "Blend with basil and serve.", Tip: &tip},
		},
		Tags: []RecipeTagItem{
			{ID: "tag_1", Name: "Soup", Slug: "soup", Color: "#E76F51", System: false},
		},
		MediaAssets: []RecipeMediaItem{
			{
				ID:               "media_sample",
				OriginalFilename: "tomato-basil-soup.jpg",
				MimeType:         "image/jpeg",
				MediaType:        "image",
				SizeBytes:        248320,
				Checksum:         "sha256:samplechecksum",
				StoredAt:         storedAt,
			},
		},
	}
}
