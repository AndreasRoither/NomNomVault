package httpapi

import (
	"bytes"
	"errors"
	"image"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/webp"
	_ "image/jpeg"
	_ "image/png"

	"github.com/AndreasRoither/NomNomVault/backend/internal/api/apicontract"
	authsvc "github.com/AndreasRoither/NomNomVault/backend/internal/auth"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/httpx"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/securitylog"
	recipesvc "github.com/AndreasRoither/NomNomVault/backend/internal/recipes"
)

const maxAltTextLength = 500

type handler struct {
	service         *recipesvc.Service
	maxUploadBytes  int64
	allowedMimeType map[string]struct{}
}

func newHandler(service *recipesvc.Service, maxUploadBytes int64, allowedMIMEs []string) *handler {
	allowed := make(map[string]struct{}, len(allowedMIMEs))
	for _, mimeType := range allowedMIMEs {
		allowed[mimeType] = struct{}{}
	}

	return &handler{
		service:         service,
		maxUploadBytes:  maxUploadBytes,
		allowedMimeType: allowed,
	}
}

// listRecipes godoc
// @Summary List recipes
// @Description Return the current household recipe list using a cursor pagination envelope.
// @Tags recipes
// @Produce json
// @Param cursor query string false "Cursor token"
// @Param limit query int false "Maximum number of recipes to return"
// @Success 200 {object} RecipeListResponse
// @Failure 401 {object} apicontract.ErrorResponse
// @Router /recipes [get]
func (h *handler) listRecipes(c *gin.Context) {
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

	recipes, err := h.service.ListRecipes(c.Request.Context(), session.ActiveHouseholdID, limit)
	if err != nil {
		httpx.WriteServiceError(c, err)
		return
	}

	items := make([]RecipeSummary, 0, len(recipes))
	for _, recipe := range recipes {
		items = append(items, mapRecipeSummary(recipe))
	}

	c.JSON(http.StatusOK, RecipeListResponse{
		Data: items,
		Page: apicontract.CursorPageInfo{
			NextCursor: nil,
			HasMore:    false,
		},
	})
}

// createRecipe godoc
// @Summary Create a recipe
// @Description Create a recipe for the active household.
// @Tags recipes
// @Accept json
// @Produce json
// @Param payload body CreateRecipeRequest true "Create recipe payload"
// @Success 201 {object} RecipeDetailResponse
// @Failure 400 {object} apicontract.ErrorResponse
// @Failure 401 {object} apicontract.ErrorResponse
// @Failure 403 {object} apicontract.ErrorResponse
// @Router /recipes [post]
func (h *handler) createRecipe(c *gin.Context) {
	var request CreateRecipeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.WriteValidationError(c, validationErrors("payload", err.Error()))
		return
	}

	session, ok := authsvc.SessionFromGin(c)
	if !ok {
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return
	}

	ingredients := make([]recipesvc.CreateRecipeIngredientInput, 0, len(request.Ingredients))
	for _, ingredient := range request.Ingredients {
		ingredients = append(ingredients, recipesvc.CreateRecipeIngredientInput{
			Name:        ingredient.Name,
			Quantity:    ingredient.Quantity,
			Unit:        ingredient.Unit,
			Preparation: ingredient.Preparation,
			SortOrder:   ingredient.SortOrder,
		})
	}

	result, err := h.service.CreateRecipe(c.Request.Context(), recipesvc.CreateRecipeInput{
		HouseholdID: session.ActiveHouseholdID,
		ActorUserID: session.UserID,
		ActorRole:   string(session.HouseholdRole),
		Title:       request.Title,
		Description: request.Description,
		SourceURL:   request.SourceURL,
		PrepMinutes: request.PrepMinutes,
		CookMinutes: request.CookMinutes,
		Servings:    request.Servings,
		Ingredients: ingredients,
	})
	if err != nil {
		httpx.WriteServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, mapRecipeDetailResponse(result.Recipe))
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
func (h *handler) getRecipe(c *gin.Context) {
	session, ok := authsvc.SessionFromGin(c)
	if !ok {
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return
	}

	detail, err := h.service.GetRecipeDetail(c.Request.Context(), session.ActiveHouseholdID, c.Param("recipeId"))
	if err != nil {
		httpx.WriteServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapRecipeDetailResponse(detail))
}

// uploadRecipeMedia godoc
// @Summary Upload recipe media
// @Description Upload an image and attach it to the requested recipe.
// @Tags recipes
// @Accept mpfd
// @Produce json
// @Param recipeId path string true "Recipe ID"
// @Param file formData file true "Recipe image file"
// @Param altText formData string false "Optional alt text"
// @Success 201 {object} RecipeMediaItem
// @Failure 400 {object} apicontract.ErrorResponse
// @Failure 401 {object} apicontract.ErrorResponse
// @Failure 403 {object} apicontract.ErrorResponse
// @Failure 404 {object} apicontract.ErrorResponse
// @Failure 415 {object} apicontract.ErrorResponse
// @Router /recipes/{recipeId}/media [post]
func (h *handler) uploadRecipeMedia(c *gin.Context) {
	session, ok := authsvc.SessionFromGin(c)
	if !ok {
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxUploadBytes)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		if isMaxBytesError(err) {
			securitylog.Log(c, "media.upload.rejected", map[string]string{"reason": "payload_too_large"})
			httpx.WriteError(c, http.StatusRequestEntityTooLarge, "payload_too_large", "The uploaded file is too large.", nil)
			return
		}
		securitylog.Log(c, "media.upload.rejected", map[string]string{"reason": "missing_file"})
		httpx.WriteError(c, http.StatusBadRequest, "validation_error", "A media file is required.", nil)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		securitylog.Log(c, "media.upload.rejected", map[string]string{"reason": "upload_open_error"})
		httpx.WriteError(c, http.StatusBadRequest, "upload_read_error", "The uploaded file could not be read.", nil)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		if isMaxBytesError(err) {
			securitylog.Log(c, "media.upload.rejected", map[string]string{"reason": "payload_too_large"})
			httpx.WriteError(c, http.StatusRequestEntityTooLarge, "payload_too_large", "The uploaded file is too large.", nil)
			return
		}
		securitylog.Log(c, "media.upload.rejected", map[string]string{"reason": "upload_read_error"})
		httpx.WriteError(c, http.StatusBadRequest, "upload_read_error", "The uploaded file could not be read.", nil)
		return
	}
	if len(content) == 0 {
		securitylog.Log(c, "media.upload.rejected", map[string]string{"reason": "empty_file"})
		httpx.WriteError(c, http.StatusBadRequest, "validation_error", "The uploaded file cannot be empty.", nil)
		return
	}

	mimeType, err := validateImageContent(content, h.allowedMimeType)
	if err != nil {
		reason := "unsupported_media_type"
		status := http.StatusUnsupportedMediaType
		if errors.Is(err, errMalformedImage) {
			reason = "malformed_image"
			status = http.StatusBadRequest
		}
		securitylog.Log(c, "media.upload.rejected", map[string]string{"reason": reason})
		httpx.WriteError(c, status, reason, "The uploaded file type is not supported.", nil)
		return
	}

	altText := strings.TrimSpace(c.PostForm("altText"))
	if len(altText) > maxAltTextLength {
		securitylog.Log(c, "media.upload.rejected", map[string]string{"reason": "alt_text_too_long"})
		httpx.WriteValidationError(c, validationErrors("altText", "Alt text must be 500 characters or fewer."))
		return
	}

	filename := sanitizeFilename(fileHeader.Filename)
	if _, ok := h.allowedMimeType[mimeType]; !ok {
		securitylog.Log(c, "media.upload.rejected", map[string]string{"reason": "unsupported_media_type"})
		httpx.WriteError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "The uploaded file type is not supported.", nil)
		return
	}

	media, err := h.service.AttachRecipeMedia(c.Request.Context(), recipesvc.AttachRecipeMediaInput{
		HouseholdID:      session.ActiveHouseholdID,
		ActorUserID:      session.UserID,
		ActorRole:        string(session.HouseholdRole),
		RecipeID:         c.Param("recipeId"),
		OriginalFilename: filename,
		MimeType:         mimeType,
		AltText:          altText,
		Content:          content,
	})
	if err != nil {
		var statusErr httpx.StatusError
		if errors.As(err, &statusErr) && statusErr.Status == http.StatusForbidden {
			securitylog.Log(c, "auth.authorization.denied", map[string]string{"reason": statusErr.Code})
		}
		httpx.WriteServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, mapMediaItem(media))
}

// getMediaOriginal godoc
// @Summary Fetch recipe media
// @Description Stream the original media bytes for the requested asset.
// @Tags recipes
// @Produce octet-stream
// @Param mediaId path string true "Media ID"
// @Success 200 {file} file
// @Failure 401 {object} apicontract.ErrorResponse
// @Failure 404 {object} apicontract.ErrorResponse
// @Router /media/{mediaId}/original [get]
func (h *handler) getMediaOriginal(c *gin.Context) {
	session, ok := authsvc.SessionFromGin(c)
	if !ok {
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return
	}

	media, err := h.service.GetMediaOriginal(c.Request.Context(), session.ActiveHouseholdID, c.Param("mediaId"))
	if err != nil {
		httpx.WriteServiceError(c, err)
		return
	}

	c.Header("Content-Type", media.MimeType)
	c.Header("Content-Length", intToString(media.SizeBytes))
	c.Header("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": sanitizeFilename(media.Filename)}))
	c.Data(http.StatusOK, media.MimeType, media.Content)
}

func mapRecipeDetailResponse(detail recipesvc.DetailView) RecipeDetailResponse {
	return RecipeDetailResponse{
		Recipe:           mapRecipeSummary(detail.Recipe),
		Ingredients:      mapIngredientItems(detail.Ingredients),
		Steps:            mapStepItems(detail.Steps),
		Tags:             mapTagItems(detail.Tags),
		MediaAssets:      mapMediaItems(detail.MediaAssets),
		NutritionEntries: mapNutritionItems(detail.NutritionEntries),
	}
}

func mapRecipeSummary(summary recipesvc.RecipeSummaryView) RecipeSummary {
	return RecipeSummary{
		ID:               summary.ID,
		Title:            summary.Title,
		Description:      summary.Description,
		SourceURL:        summary.SourceURL,
		SourceCapturedAt: summary.SourceCapturedAt,
		PrimaryMediaID:   summary.PrimaryMediaID,
		GalleryMediaIDs:  summary.GalleryMediaIDs,
		PrepMinutes:      summary.PrepMinutes,
		CookMinutes:      summary.CookMinutes,
		Servings:         summary.Servings,
		Region:           summary.Region,
		MealType:         summary.MealType,
		Difficulty:       summary.Difficulty,
		Cuisine:          summary.Cuisine,
		Allergens:        summary.Allergens,
		CaloriesPerServe: summary.CaloriesPerServe,
		Version:          summary.Version,
	}
}

func mapIngredientItems(ingredients []recipesvc.IngredientView) []RecipeIngredientItem {
	items := make([]RecipeIngredientItem, 0, len(ingredients))
	for _, ingredient := range ingredients {
		items = append(items, RecipeIngredientItem{
			ID:          ingredient.ID,
			Name:        ingredient.Name,
			Quantity:    ingredient.Quantity,
			Unit:        ingredient.Unit,
			Preparation: ingredient.Preparation,
			SortOrder:   ingredient.SortOrder,
		})
	}
	return items
}

func mapStepItems(steps []recipesvc.StepView) []RecipeStepItem {
	items := make([]RecipeStepItem, 0, len(steps))
	for _, step := range steps {
		items = append(items, RecipeStepItem{
			ID:              step.ID,
			SortOrder:       step.SortOrder,
			Instruction:     step.Instruction,
			DurationMinutes: step.DurationMinutes,
			Tip:             step.Tip,
		})
	}
	return items
}

func mapTagItems(tags []recipesvc.TagView) []RecipeTagItem {
	items := make([]RecipeTagItem, 0, len(tags))
	for _, tag := range tags {
		items = append(items, RecipeTagItem{
			ID:     tag.ID,
			Name:   tag.Name,
			Slug:   tag.Slug,
			Color:  tag.Color,
			System: tag.System,
		})
	}
	return items
}

func mapMediaItems(media []recipesvc.MediaView) []RecipeMediaItem {
	items := make([]RecipeMediaItem, 0, len(media))
	for _, item := range media {
		items = append(items, mapMediaItem(item))
	}
	return items
}

func mapMediaItem(item recipesvc.MediaView) RecipeMediaItem {
	return RecipeMediaItem{
		ID:               item.ID,
		OriginalFilename: item.OriginalFilename,
		MimeType:         item.MimeType,
		MediaType:        item.MediaType,
		SizeBytes:        item.SizeBytes,
		Checksum:         item.Checksum,
		StoredAt:         item.StoredAt,
		URL:              item.URL,
		ThumbnailURL:     item.ThumbnailURL,
		AltText:          item.AltText,
		SortOrder:        item.SortOrder,
	}
}

func mapNutritionItems(entries []recipesvc.NutritionView) []RecipeNutritionItem {
	items := make([]RecipeNutritionItem, 0, len(entries))
	for _, entry := range entries {
		items = append(items, RecipeNutritionItem{
			ID:                entry.ID,
			ReferenceQuantity: entry.ReferenceQuantity,
			EnergyKcal:        entry.EnergyKcal,
			Protein:           entry.Protein,
			Carbohydrates:     entry.Carbohydrates,
			Fat:               entry.Fat,
			SaturatedFat:      entry.SaturatedFat,
			Fiber:             entry.Fiber,
			Sugars:            entry.Sugars,
			Sodium:            entry.Sodium,
			Salt:              entry.Salt,
		})
	}
	return items
}

func validationErrors(field string, message string) []apicontract.ValidationError {
	return []apicontract.ValidationError{{Field: field, Message: message}}
}

func sniffMimeType(content []byte) string {
	if len(content) == 0 {
		return "application/octet-stream"
	}
	if len(content) > 512 {
		return http.DetectContentType(content[:512])
	}
	return http.DetectContentType(content)
}

func intToString(value int64) string {
	return strconv.FormatInt(value, 10)
}

var errMalformedImage = errors.New("malformed image")

func validateImageContent(content []byte, allowedMimeTypes map[string]struct{}) (string, error) {
	mimeType := sniffMimeType(content)
	if _, ok := allowedMimeTypes[mimeType]; !ok {
		return "", http.ErrNotSupported
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil {
		return "", errMalformedImage
	}
	if config.Width <= 0 || config.Height <= 0 {
		return "", errMalformedImage
	}

	switch format {
	case "jpeg":
		if mimeType != "image/jpeg" {
			return "", errMalformedImage
		}
	case "png":
		if mimeType != "image/png" {
			return "", errMalformedImage
		}
	case "webp":
		if mimeType != "image/webp" {
			return "", errMalformedImage
		}
	default:
		return "", http.ErrNotSupported
	}

	return mimeType, nil
}

func sanitizeFilename(filename string) string {
	filename = strings.TrimSpace(filename)
	replacer := strings.NewReplacer("/", "_", "\\", "_", "\"", "_", "\r", "_", "\n", "_")
	filename = replacer.Replace(filename)

	var builder strings.Builder
	for _, r := range filename {
		if r < 32 || r == 127 {
			continue
		}
		builder.WriteRune(r)
	}

	cleaned := strings.Trim(strings.Join(strings.Fields(builder.String()), " "), " .")
	if cleaned == "" {
		return "download"
	}
	return cleaned
}

func isMaxBytesError(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}
