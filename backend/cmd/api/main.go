// @title NomNomVault API
// @version 0.1.0
// @BasePath /api/v1
// @schemes http https
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	authhttpapi "github.com/AndreasRoither/NomNomVault/backend/internal/api/httpapi/auth"
	recipeshttpapi "github.com/AndreasRoither/NomNomVault/backend/internal/api/httpapi/recipes"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/healthz", healthz)

	apiV1 := router.Group("/api/v1")
	apiV1.GET("/healthz", healthz)
	authhttpapi.RegisterRoutes(apiV1)
	recipeshttpapi.RegisterRoutes(apiV1)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		<-shutdownCtx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("api shutdown error: %v", err)
		}
	}()

	log.Printf("api listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("api server failed: %v", err)
	}
}

func healthz(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"status": "ok"})
}
