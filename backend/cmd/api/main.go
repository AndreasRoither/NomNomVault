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
	_ "github.com/jackc/pgx/v5/stdlib"

	authhttpapi "github.com/AndreasRoither/NomNomVault/backend/internal/api/httpapi/auth"
	importshttpapi "github.com/AndreasRoither/NomNomVault/backend/internal/api/httpapi/imports"
	recipeshttpapi "github.com/AndreasRoither/NomNomVault/backend/internal/api/httpapi/recipes"
	authsvc "github.com/AndreasRoither/NomNomVault/backend/internal/auth"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent"
	importsvc "github.com/AndreasRoither/NomNomVault/backend/internal/imports"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/clock"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/config"
	corsmiddleware "github.com/AndreasRoither/NomNomVault/backend/internal/platform/cors"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/ratelimit"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/requestid"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/storage"
	recipesvc "github.com/AndreasRoither/NomNomVault/backend/internal/recipes"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := ent.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestid.Middleware())
	router.Use(corsmiddleware.Middleware(cfg.AllowedCORSOrigins))
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Fatalf("configure trusted proxies: %v", err)
	}
	router.GET("/healthz", healthz)

	signer := authsvc.NewTokenSigner(cfg.AuthJWTSecret)
	csrf := authsvc.NewCSRFManager(cfg.AuthCSRFSecret)
	cookies := authsvc.NewCookieManager(cfg.CookieSecure)
	authService := authsvc.NewService(db, clock.RealClock{}, signer, csrf)
	importService := importsvc.NewService(db, clock.RealClock{})
	recipeService := recipesvc.NewService(db, storage.NewPostgresStore(db), clock.RealClock{})
	loginLimiter := ratelimit.New(cfg.AuthLoginRateLimitPerMinute, time.Minute)
	refreshLimiter := ratelimit.New(cfg.AuthRefreshRateLimitPerMinute, time.Minute)

	apiV1 := router.Group("/api/v1")
	apiV1.GET("/healthz", healthz)
	authhttpapi.RegisterRoutes(apiV1, authService, cookies, csrf, loginLimiter, refreshLimiter)
	importshttpapi.RegisterRoutes(
		apiV1,
		importService,
		authsvc.Middleware(signer, cookies),
		authsvc.CSRFMiddleware(csrf),
	)
	recipeshttpapi.RegisterRoutes(
		apiV1,
		recipeService,
		authsvc.Middleware(signer, cookies),
		authsvc.CSRFMiddleware(csrf),
		cfg.MaxUploadBytes,
		cfg.AllowedUploadMIMEs,
	)

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
