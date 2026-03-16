package httpapi

import (
	"github.com/gin-gonic/gin"

	recipesvc "github.com/AndreasRoither/NomNomVault/backend/internal/recipes"
)

// RegisterRoutes binds the recipe endpoints for the persisted API implementation.
func RegisterRoutes(
	api *gin.RouterGroup,
	service *recipesvc.Service,
	authMiddleware gin.HandlerFunc,
	csrfMiddleware gin.HandlerFunc,
	maxUploadBytes int64,
	allowedMIMEs []string,
) {
	h := newHandler(service, maxUploadBytes, allowedMIMEs)

	reads := api.Group("")
	reads.Use(authMiddleware)
	reads.GET("/recipes", h.listRecipes)
	reads.GET("/recipes/:recipeId", h.getRecipe)
	reads.GET("/tags", h.listTags)
	reads.GET("/media/:mediaId/original", h.getMediaOriginal)
	reads.GET("/media/:mediaId/thumbnail", h.getMediaThumbnail)

	writes := api.Group("")
	writes.Use(authMiddleware, csrfMiddleware)
	writes.POST("/recipes", h.createRecipe)
	writes.PATCH("/recipes/:recipeId", h.patchRecipe)
	writes.POST("/recipes/:recipeId/archive", h.archiveRecipe)
	writes.POST("/recipes/:recipeId/unarchive", h.unarchiveRecipe)
	writes.POST("/recipes/:recipeId/media", h.uploadRecipeMedia)
	writes.POST("/tags", h.createTag)
}
