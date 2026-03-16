package httpapi

import (
	"github.com/gin-gonic/gin"

	importsvc "github.com/AndreasRoither/NomNomVault/backend/internal/imports"
)

// RegisterRoutes binds the import job endpoints.
func RegisterRoutes(
	api *gin.RouterGroup,
	service *importsvc.Service,
	authMiddleware gin.HandlerFunc,
	csrfMiddleware gin.HandlerFunc,
) {
	h := newHandler(service)

	reads := api.Group("")
	reads.Use(authMiddleware)
	reads.GET("/imports", h.listImportJobs)
	reads.GET("/imports/:jobId", h.getImportJob)

	writes := api.Group("")
	writes.Use(authMiddleware, csrfMiddleware)
	writes.POST("/imports/url", h.createURLImport)
	writes.POST("/imports/:jobId/cancel", h.cancelImportJob)
	writes.POST("/imports/:jobId/retry", h.retryImportJob)
}
