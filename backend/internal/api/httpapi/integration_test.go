package httpapi_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"

	authhttpapi "github.com/AndreasRoither/NomNomVault/backend/internal/api/httpapi/auth"
	recipeshttpapi "github.com/AndreasRoither/NomNomVault/backend/internal/api/httpapi/recipes"
	authsvc "github.com/AndreasRoither/NomNomVault/backend/internal/auth"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/householdmember"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/user"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/clock"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/ratelimit"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/requestid"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/storage"
	recipesvc "github.com/AndreasRoither/NomNomVault/backend/internal/recipes"
)

type testEnv struct {
	server *httptest.Server
	db     *ent.Client
}

func TestAuthAndRecipeFirstSliceFlow(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)

	csrfToken := bootstrapCSRF(t, client, env.server.URL)

	registerPayload := map[string]any{
		"displayName": "Nom Nom",
		"email":       "cook@example.com",
		"password":    "supersecret123",
	}
	registerResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/auth/register", csrfToken, registerPayload)
	if registerResponse.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected register status: %d", registerResponse.StatusCode)
	}
	registerResponse.Body.Close()

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	if currentCookieValue(t, client, env.server.URL, "/", authsvc.AccessCookieName) == "" {
		t.Fatal("expected access cookie after register")
	}
	if currentCookieValue(t, client, env.server.URL, "/api/v1/auth/refresh", authsvc.RefreshCookieName) == "" {
		t.Fatal("expected refresh cookie after register")
	}

	createPayload := map[string]any{
		"title":       "Tomato Soup",
		"description": "Simple soup",
		"ingredients": []map[string]any{
			{"name": "Tomatoes", "quantity": 2, "unit": "pc", "sortOrder": 1},
			{"name": "Salt", "sortOrder": 2},
		},
	}
	createResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/recipes", csrfToken, createPayload)
	if createResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createResponse.Body)
		t.Fatalf("unexpected create recipe status: %d body=%s", createResponse.StatusCode, string(body))
	}

	var createBody struct {
		Recipe struct {
			ID string `json:"id"`
		} `json:"recipe"`
	}
	if err := json.NewDecoder(createResponse.Body).Decode(&createBody); err != nil {
		t.Fatalf("decode create recipe body: %v", err)
	}
	createResponse.Body.Close()

	if createBody.Recipe.ID == "" {
		t.Fatal("expected created recipe ID")
	}

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	uploadResponse := doMultipart(t, client, env.server.URL+"/api/v1/recipes/"+createBody.Recipe.ID+"/media", csrfToken)
	if uploadResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(uploadResponse.Body)
		t.Fatalf("unexpected media upload status: %d body=%s", uploadResponse.StatusCode, string(body))
	}
	uploadResponse.Body.Close()

	detailRequest, err := http.NewRequest(http.MethodGet, env.server.URL+"/api/v1/recipes/"+createBody.Recipe.ID, nil)
	if err != nil {
		t.Fatalf("new detail request: %v", err)
	}

	detailResponse, err := client.Do(detailRequest)
	if err != nil {
		t.Fatalf("perform detail request: %v", err)
	}
	defer detailResponse.Body.Close()

	if detailResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(detailResponse.Body)
		t.Fatalf("unexpected detail status: %d body=%s", detailResponse.StatusCode, string(body))
	}

	var detailBody struct {
		Recipe struct {
			ID              string   `json:"id"`
			PrimaryMediaID  *string  `json:"primaryMediaId"`
			GalleryMediaIDs []string `json:"galleryMediaIds"`
		} `json:"recipe"`
		Ingredients []any `json:"ingredients"`
		MediaAssets []struct {
			ID string `json:"id"`
		} `json:"mediaAssets"`
	}
	if err := json.NewDecoder(detailResponse.Body).Decode(&detailBody); err != nil {
		t.Fatalf("decode detail body: %v", err)
	}

	if detailBody.Recipe.ID != createBody.Recipe.ID {
		t.Fatalf("unexpected detail recipe id: %s", detailBody.Recipe.ID)
	}
	if len(detailBody.Ingredients) != 2 {
		t.Fatalf("expected 2 ingredients, got %d", len(detailBody.Ingredients))
	}
	if detailBody.Recipe.PrimaryMediaID == nil || len(detailBody.MediaAssets) != 1 {
		t.Fatalf("expected uploaded media to be attached to recipe")
	}
}

func TestRecipeMutationRequiresCSRF(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	_ = bootstrapCSRF(t, client, env.server.URL)

	registerPayload := map[string]any{
		"displayName": "Nom Nom",
		"email":       "cook2@example.com",
		"password":    "supersecret123",
	}
	requestBody, err := json.Marshal(registerPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, env.server.URL+"/api/v1/auth/register", bytes.NewReader(requestBody))
	if err != nil {
		t.Fatalf("new register request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("perform register request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected csrf failure, got %d body=%s", response.StatusCode, string(body))
	}
}

func TestRegisterDuplicateEmailReturnsGenericFailure(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	firstClient := newHTTPClient(t)
	secondClient := newHTTPClient(t)

	csrfToken := bootstrapCSRF(t, firstClient, env.server.URL)
	response := doJSON(t, firstClient, http.MethodPost, env.server.URL+"/api/v1/auth/register", csrfToken, map[string]any{
		"displayName": "Cook One",
		"email":       "duplicate@example.com",
		"password":    "supersecret123",
	})
	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected first register success, got %d body=%s", response.StatusCode, string(body))
	}
	response.Body.Close()

	csrfToken = bootstrapCSRF(t, secondClient, env.server.URL)
	response = doJSON(t, secondClient, http.MethodPost, env.server.URL+"/api/v1/auth/register", csrfToken, map[string]any{
		"displayName": "Cook Two",
		"email":       "duplicate@example.com",
		"password":    "supersecret123",
	})
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected duplicate register failure, got %d body=%s", response.StatusCode, string(body))
	}

	var errorBody struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(response.Body).Decode(&errorBody); err != nil {
		t.Fatalf("decode duplicate register response: %v", err)
	}
	if errorBody.Code != "registration_failed" {
		t.Fatalf("expected generic registration failure code, got %q", errorBody.Code)
	}
	if strings.Contains(strings.ToLower(errorBody.Message), "email") {
		t.Fatalf("expected non-enumerating message, got %q", errorBody.Message)
	}
}

func TestRefreshRotationRejectsReusedToken(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := bootstrapCSRF(t, client, env.server.URL)

	registerResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/auth/register", csrfToken, map[string]any{
		"displayName": "Refresh User",
		"email":       "refresh@example.com",
		"password":    "supersecret123",
	})
	if registerResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(registerResponse.Body)
		t.Fatalf("expected register success, got %d body=%s", registerResponse.StatusCode, string(body))
	}
	registerResponse.Body.Close()

	oldRefresh := currentCookieValue(t, client, env.server.URL, "/api/v1/auth/refresh", authsvc.RefreshCookieName)
	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)

	refreshResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/auth/refresh", csrfToken, map[string]any{})
	if refreshResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(refreshResponse.Body)
		t.Fatalf("expected refresh success, got %d body=%s", refreshResponse.StatusCode, string(body))
	}
	refreshResponse.Body.Close()

	newRefresh := currentCookieValue(t, client, env.server.URL, "/api/v1/auth/refresh", authsvc.RefreshCookieName)
	if newRefresh == "" || newRefresh == oldRefresh {
		t.Fatal("expected refresh token rotation")
	}

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	replayRequest, err := http.NewRequest(http.MethodPost, env.server.URL+"/api/v1/auth/refresh", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		t.Fatalf("new replay request: %v", err)
	}
	replayRequest.Header.Set("Content-Type", "application/json")
	replayRequest.Header.Set("X-CSRF-Token", csrfToken)
	replayRequest.AddCookie(&http.Cookie{Name: authsvc.CSRFCookieName, Value: csrfToken, Path: "/"})
	replayRequest.AddCookie(&http.Cookie{Name: authsvc.RefreshCookieName, Value: oldRefresh, Path: "/api/v1/auth"})

	replayResponse, err := (&http.Client{}).Do(replayRequest)
	if err != nil {
		t.Fatalf("perform replay request: %v", err)
	}
	defer replayResponse.Body.Close()

	if replayResponse.StatusCode != http.StatusUnauthorized {
		body, _ := io.ReadAll(replayResponse.Body)
		t.Fatalf("expected old refresh token to be rejected, got %d body=%s", replayResponse.StatusCode, string(body))
	}
}

func TestLogoutRevokesCurrentRefreshToken(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := bootstrapCSRF(t, client, env.server.URL)

	registerResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/auth/register", csrfToken, map[string]any{
		"displayName": "Logout User",
		"email":       "logout@example.com",
		"password":    "supersecret123",
	})
	if registerResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(registerResponse.Body)
		t.Fatalf("expected register success, got %d body=%s", registerResponse.StatusCode, string(body))
	}
	registerResponse.Body.Close()

	refreshToken := currentCookieValue(t, client, env.server.URL, "/api/v1/auth/refresh", authsvc.RefreshCookieName)
	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)

	logoutResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/auth/logout", csrfToken, map[string]any{})
	if logoutResponse.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(logoutResponse.Body)
		t.Fatalf("expected logout success, got %d body=%s", logoutResponse.StatusCode, string(body))
	}
	logoutResponse.Body.Close()

	replayRequest, err := http.NewRequest(http.MethodPost, env.server.URL+"/api/v1/auth/refresh", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		t.Fatalf("new revoked refresh request: %v", err)
	}
	replayRequest.Header.Set("Content-Type", "application/json")
	replayRequest.Header.Set("X-CSRF-Token", csrfToken)
	replayRequest.AddCookie(&http.Cookie{Name: authsvc.CSRFCookieName, Value: csrfToken, Path: "/"})
	replayRequest.AddCookie(&http.Cookie{Name: authsvc.RefreshCookieName, Value: refreshToken, Path: "/api/v1/auth"})

	replayResponse, err := (&http.Client{}).Do(replayRequest)
	if err != nil {
		t.Fatalf("perform revoked refresh request: %v", err)
	}
	defer replayResponse.Body.Close()

	if replayResponse.StatusCode != http.StatusUnauthorized {
		body, _ := io.ReadAll(replayResponse.Body)
		t.Fatalf("expected logged-out refresh token to be rejected, got %d body=%s", replayResponse.StatusCode, string(body))
	}
}

func TestCrossHouseholdRecipeAccessIsDenied(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	ownerClient := newHTTPClient(t)
	otherClient := newHTTPClient(t)

	csrfToken := bootstrapCSRF(t, ownerClient, env.server.URL)
	registerResponse := doJSON(t, ownerClient, http.MethodPost, env.server.URL+"/api/v1/auth/register", csrfToken, map[string]any{
		"displayName": "Owner User",
		"email":       "owner@example.com",
		"password":    "supersecret123",
	})
	if registerResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(registerResponse.Body)
		t.Fatalf("expected owner register success, got %d body=%s", registerResponse.StatusCode, string(body))
	}
	registerResponse.Body.Close()

	csrfToken = currentCookieValue(t, ownerClient, env.server.URL, "/", authsvc.CSRFCookieName)
	createResponse := doJSON(t, ownerClient, http.MethodPost, env.server.URL+"/api/v1/recipes", csrfToken, map[string]any{
		"title":       "Private Soup",
		"description": "Household scoped",
		"ingredients": []map[string]any{{"name": "Tomatoes", "sortOrder": 1}},
	})
	if createResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createResponse.Body)
		t.Fatalf("expected owner recipe create success, got %d body=%s", createResponse.StatusCode, string(body))
	}

	var createBody struct {
		Recipe struct {
			ID string `json:"id"`
		} `json:"recipe"`
	}
	if err := json.NewDecoder(createResponse.Body).Decode(&createBody); err != nil {
		t.Fatalf("decode cross-household create response: %v", err)
	}
	createResponse.Body.Close()

	csrfToken = bootstrapCSRF(t, otherClient, env.server.URL)
	registerResponse = doJSON(t, otherClient, http.MethodPost, env.server.URL+"/api/v1/auth/register", csrfToken, map[string]any{
		"displayName": "Other User",
		"email":       "other@example.com",
		"password":    "supersecret123",
	})
	if registerResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(registerResponse.Body)
		t.Fatalf("expected other register success, got %d body=%s", registerResponse.StatusCode, string(body))
	}
	registerResponse.Body.Close()

	detailRequest, err := http.NewRequest(http.MethodGet, env.server.URL+"/api/v1/recipes/"+createBody.Recipe.ID, nil)
	if err != nil {
		t.Fatalf("new cross-household detail request: %v", err)
	}
	detailResponse, err := otherClient.Do(detailRequest)
	if err != nil {
		t.Fatalf("perform cross-household detail request: %v", err)
	}
	defer detailResponse.Body.Close()

	if detailResponse.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(detailResponse.Body)
		t.Fatalf("expected cross-household recipe access to fail, got %d body=%s", detailResponse.StatusCode, string(body))
	}
}

func TestViewerCannotCreateRecipe(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := bootstrapCSRF(t, client, env.server.URL)

	registerResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/auth/register", csrfToken, map[string]any{
		"displayName": "Viewer User",
		"email":       "viewer@example.com",
		"password":    "supersecret123",
	})
	if registerResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(registerResponse.Body)
		t.Fatalf("expected register success, got %d body=%s", registerResponse.StatusCode, string(body))
	}
	registerResponse.Body.Close()

	ctx := context.Background()
	userEntity, err := env.db.User.Query().Where(user.EmailEQ("viewer@example.com")).Only(ctx)
	if err != nil {
		t.Fatalf("query viewer user: %v", err)
	}
	if _, err := env.db.HouseholdMember.Update().
		Where(householdmember.UserIDEQ(userEntity.ID)).
		SetRole(householdmember.RoleViewer).
		Save(ctx); err != nil {
		t.Fatalf("downgrade viewer role: %v", err)
	}

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	refreshResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/auth/refresh", csrfToken, map[string]any{})
	if refreshResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(refreshResponse.Body)
		t.Fatalf("expected refresh success after role change, got %d body=%s", refreshResponse.StatusCode, string(body))
	}
	refreshResponse.Body.Close()

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	createResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/recipes", csrfToken, map[string]any{
		"title":       "Viewer Soup",
		"description": "Should be blocked",
		"ingredients": []map[string]any{{"name": "Salt", "sortOrder": 1}},
	})
	defer createResponse.Body.Close()

	if createResponse.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(createResponse.Body)
		t.Fatalf("expected viewer create recipe to be rejected, got %d body=%s", createResponse.StatusCode, string(body))
	}
}

func TestMediaDownloadSanitizesFilenameHeader(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := bootstrapCSRF(t, client, env.server.URL)

	registerResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/auth/register", csrfToken, map[string]any{
		"displayName": "Media User",
		"email":       "media@example.com",
		"password":    "supersecret123",
	})
	if registerResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(registerResponse.Body)
		t.Fatalf("expected register success, got %d body=%s", registerResponse.StatusCode, string(body))
	}
	registerResponse.Body.Close()

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	createResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/recipes", csrfToken, map[string]any{
		"title":       "Media Soup",
		"description": "Header hardening",
		"ingredients": []map[string]any{{"name": "Salt", "sortOrder": 1}},
	})
	if createResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createResponse.Body)
		t.Fatalf("expected recipe create success, got %d body=%s", createResponse.StatusCode, string(body))
	}

	var createBody struct {
		Recipe struct {
			ID string `json:"id"`
		} `json:"recipe"`
	}
	if err := json.NewDecoder(createResponse.Body).Decode(&createBody); err != nil {
		t.Fatalf("decode media recipe create response: %v", err)
	}
	createResponse.Body.Close()

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	uploadResponse := doMultipartNamed(t, client, env.server.URL+"/api/v1/recipes/"+createBody.Recipe.ID+"/media", csrfToken, "bad\r\nname\\\".png")
	if uploadResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(uploadResponse.Body)
		t.Fatalf("expected media upload success, got %d body=%s", uploadResponse.StatusCode, string(body))
	}

	var mediaBody struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(uploadResponse.Body).Decode(&mediaBody); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	uploadResponse.Body.Close()

	mediaRequest, err := http.NewRequest(http.MethodGet, env.server.URL+"/api/v1/media/"+mediaBody.ID+"/original", nil)
	if err != nil {
		t.Fatalf("new media request: %v", err)
	}
	mediaResponse, err := client.Do(mediaRequest)
	if err != nil {
		t.Fatalf("perform media request: %v", err)
	}
	defer mediaResponse.Body.Close()

	contentDisposition := mediaResponse.Header.Get("Content-Disposition")
	if strings.Contains(contentDisposition, "\r") || strings.Contains(contentDisposition, "\n") {
		t.Fatalf("expected sanitized content-disposition header, got %q", contentDisposition)
	}
	if strings.Contains(contentDisposition, "\\") {
		t.Fatalf("expected path separators to be sanitized, got %q", contentDisposition)
	}
}

func TestPatchRecipeReplacesCollectionsAndTags(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := registerUser(t, client, env.server.URL, "Patch User", "patch@example.com")
	createdRecipe := createRecipe(t, client, env.server.URL, csrfToken, "Tomato Soup")
	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)

	quickTag := createTag(t, client, env.server.URL, csrfToken, "Quick Meals", "#4FB8B2")
	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	dinnerTag := createTag(t, client, env.server.URL, csrfToken, "Dinner", "#173A40")
	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)

	response := doJSON(t, client, http.MethodPatch, env.server.URL+"/api/v1/recipes/"+createdRecipe.ID, csrfToken, map[string]any{
		"version":     createdRecipe.Version,
		"title":       "Roasted Tomato Soup",
		"description": "Silky and rich",
		"sourceUrl":   "https://example.com/soup",
		"prepMinutes": 10,
		"cookMinutes": 35,
		"servings":    4,
		"ingredients": []map[string]any{
			{"name": "Tomatoes", "quantity": 6, "unit": "pc", "sortOrder": 1},
			{"name": "Basil", "quantity": 5, "unit": "g", "sortOrder": 2},
		},
		"steps": []map[string]any{
			{"instruction": "Roast the tomatoes.", "sortOrder": 1, "durationMinutes": 20},
			{"instruction": "Blend and simmer.", "sortOrder": 2},
		},
		"nutritionEntries": []map[string]any{
			{"referenceQuantity": "per serving", "energyKcal": 240, "protein": 6.5},
		},
		"tagIds": []string{quickTag.ID, dinnerTag.ID},
	})
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected recipe patch success, got %d body=%s", response.StatusCode, string(body))
	}
	defer response.Body.Close()

	var detailBody struct {
		Recipe struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Version     int    `json:"version"`
		} `json:"recipe"`
		Ingredients []any `json:"ingredients"`
		Steps       []any `json:"steps"`
		Tags        []struct {
			ID string `json:"id"`
		} `json:"tags"`
		NutritionEntries []any `json:"nutritionEntries"`
	}
	if err := json.NewDecoder(response.Body).Decode(&detailBody); err != nil {
		t.Fatalf("decode patch response: %v", err)
	}

	if detailBody.Recipe.Title != "Roasted Tomato Soup" {
		t.Fatalf("unexpected patched title %q", detailBody.Recipe.Title)
	}
	if detailBody.Recipe.Version != createdRecipe.Version+1 {
		t.Fatalf("expected version %d, got %d", createdRecipe.Version+1, detailBody.Recipe.Version)
	}
	if len(detailBody.Ingredients) != 2 || len(detailBody.Steps) != 2 || len(detailBody.Tags) != 2 || len(detailBody.NutritionEntries) != 1 {
		t.Fatalf("expected patched collections to be fully replaced")
	}
}

func TestPatchRecipeRejectsStaleVersion(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := registerUser(t, client, env.server.URL, "Conflict User", "conflict@example.com")
	createdRecipe := createRecipe(t, client, env.server.URL, csrfToken, "Conflict Soup")

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	firstPatch := doJSON(t, client, http.MethodPatch, env.server.URL+"/api/v1/recipes/"+createdRecipe.ID, csrfToken, map[string]any{
		"version":     createdRecipe.Version,
		"title":       "Conflict Soup Updated",
		"description": "one",
		"sourceUrl":   "",
		"ingredients": []map[string]any{{"name": "Salt", "sortOrder": 1}},
	})
	if firstPatch.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(firstPatch.Body)
		t.Fatalf("expected first patch success, got %d body=%s", firstPatch.StatusCode, string(body))
	}
	firstPatch.Body.Close()

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	secondPatch := doJSON(t, client, http.MethodPatch, env.server.URL+"/api/v1/recipes/"+createdRecipe.ID, csrfToken, map[string]any{
		"version":     createdRecipe.Version,
		"title":       "Conflict Soup Stale",
		"description": "two",
		"sourceUrl":   "",
		"ingredients": []map[string]any{{"name": "Salt", "sortOrder": 1}},
	})
	defer secondPatch.Body.Close()

	if secondPatch.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(secondPatch.Body)
		t.Fatalf("expected stale patch conflict, got %d body=%s", secondPatch.StatusCode, string(body))
	}

	var errorBody struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(secondPatch.Body).Decode(&errorBody); err != nil {
		t.Fatalf("decode conflict response: %v", err)
	}
	if errorBody.Code != "version_conflict" {
		t.Fatalf("expected version_conflict, got %q", errorBody.Code)
	}
}

func TestArchiveAndUnarchiveRecipeVisibility(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := registerUser(t, client, env.server.URL, "Archive User", "archive@example.com")
	createdRecipe := createRecipe(t, client, env.server.URL, csrfToken, "Archive Soup")

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	archiveResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/recipes/"+createdRecipe.ID+"/archive", csrfToken, map[string]any{})
	if archiveResponse.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(archiveResponse.Body)
		t.Fatalf("expected archive success, got %d body=%s", archiveResponse.StatusCode, string(body))
	}
	archiveResponse.Body.Close()

	listResponse := doGET(t, client, env.server.URL+"/api/v1/recipes")
	if listResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(listResponse.Body)
		t.Fatalf("expected list success after archive, got %d body=%s", listResponse.StatusCode, string(body))
	}
	var archivedList struct {
		Data []any `json:"data"`
	}
	if err := json.NewDecoder(listResponse.Body).Decode(&archivedList); err != nil {
		t.Fatalf("decode archived list: %v", err)
	}
	listResponse.Body.Close()
	if len(archivedList.Data) != 0 {
		t.Fatalf("expected archived recipe to be hidden from list")
	}

	detailResponse := doGET(t, client, env.server.URL+"/api/v1/recipes/"+createdRecipe.ID)
	if detailResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(detailResponse.Body)
		t.Fatalf("expected archived detail success, got %d body=%s", detailResponse.StatusCode, string(body))
	}
	detailResponse.Body.Close()

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	unarchiveResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/recipes/"+createdRecipe.ID+"/unarchive", csrfToken, map[string]any{})
	if unarchiveResponse.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(unarchiveResponse.Body)
		t.Fatalf("expected unarchive success, got %d body=%s", unarchiveResponse.StatusCode, string(body))
	}
	unarchiveResponse.Body.Close()

	listResponse = doGET(t, client, env.server.URL+"/api/v1/recipes")
	if listResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(listResponse.Body)
		t.Fatalf("expected list success after unarchive, got %d body=%s", listResponse.StatusCode, string(body))
	}
	var activeList struct {
		Data []any `json:"data"`
	}
	if err := json.NewDecoder(listResponse.Body).Decode(&activeList); err != nil {
		t.Fatalf("decode active list: %v", err)
	}
	listResponse.Body.Close()
	if len(activeList.Data) != 1 {
		t.Fatalf("expected unarchived recipe to be visible again, got %d items", len(activeList.Data))
	}
}

func TestArchiveBumpsVersionAndMakesStalePatchConflict(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := registerUser(t, client, env.server.URL, "Archive Conflict User", "archive-conflict@example.com")
	createdRecipe := createRecipe(t, client, env.server.URL, csrfToken, "Archive Conflict Soup")

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	archiveResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/recipes/"+createdRecipe.ID+"/archive", csrfToken, map[string]any{})
	if archiveResponse.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(archiveResponse.Body)
		t.Fatalf("expected archive success, got %d body=%s", archiveResponse.StatusCode, string(body))
	}
	archiveResponse.Body.Close()

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	patchResponse := doJSON(t, client, http.MethodPatch, env.server.URL+"/api/v1/recipes/"+createdRecipe.ID, csrfToken, map[string]any{
		"version":     createdRecipe.Version,
		"title":       "Archive Conflict Soup Updated",
		"description": "stale",
		"sourceUrl":   "",
		"ingredients": []map[string]any{{"name": "Salt", "sortOrder": 1}},
	})
	defer patchResponse.Body.Close()

	if patchResponse.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(patchResponse.Body)
		t.Fatalf("expected stale patch conflict after archive, got %d body=%s", patchResponse.StatusCode, string(body))
	}

	var errorBody struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(patchResponse.Body).Decode(&errorBody); err != nil {
		t.Fatalf("decode stale patch response: %v", err)
	}
	if errorBody.Code != "version_conflict" {
		t.Fatalf("expected version_conflict after archive, got %q", errorBody.Code)
	}
}

func TestCreateTagRejectsDuplicateNormalizedName(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := registerUser(t, client, env.server.URL, "Tag User", "tag@example.com")
	createTag(t, client, env.server.URL, csrfToken, "Quick Meals", "#4FB8B2")

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	response := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/tags", csrfToken, map[string]any{
		"name":  "  quick   meals ",
		"color": "#4FB8B2",
	})
	defer response.Body.Close()

	if response.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected duplicate tag conflict, got %d body=%s", response.StatusCode, string(body))
	}

	var errorBody struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(response.Body).Decode(&errorBody); err != nil {
		t.Fatalf("decode duplicate tag response: %v", err)
	}
	if errorBody.Code != "tag_conflict" {
		t.Fatalf("expected tag_conflict, got %q", errorBody.Code)
	}
}

func TestRecipeListSupportsSearchFilterAndCursor(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := registerUser(t, client, env.server.URL, "List User", "list@example.com")
	quickTag := createTag(t, client, env.server.URL, csrfToken, "Quick", "#4FB8B2")
	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	veganTag := createTag(t, client, env.server.URL, csrfToken, "Vegan", "#2F6A4A")
	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)

	soupRecipe := createRecipe(t, client, env.server.URL, csrfToken, "Tomato Soup")
	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	saladRecipe := createRecipe(t, client, env.server.URL, csrfToken, "Green Salad")
	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	pastaRecipe := createRecipe(t, client, env.server.URL, csrfToken, "Lemon Pasta")

	patchRecipeTags(t, client, env.server.URL, soupRecipe, "Tomato Soup", []string{quickTag.ID})
	patchRecipeTags(t, client, env.server.URL, saladRecipe, "Green Salad", []string{quickTag.ID, veganTag.ID})
	patchRecipeTags(t, client, env.server.URL, pastaRecipe, "Lemon Pasta", []string{veganTag.ID})

	response := doGET(t, client, env.server.URL+"/api/v1/recipes?q=salad")
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected search success, got %d body=%s", response.StatusCode, string(body))
	}
	var searchBody struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&searchBody); err != nil {
		t.Fatalf("decode search body: %v", err)
	}
	response.Body.Close()
	if len(searchBody.Data) != 1 || searchBody.Data[0].ID != saladRecipe.ID {
		t.Fatalf("expected salad search hit")
	}

	response = doGET(t, client, env.server.URL+"/api/v1/recipes?tag=quick")
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected tag filter success, got %d body=%s", response.StatusCode, string(body))
	}
	var quickBody struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&quickBody); err != nil {
		t.Fatalf("decode quick filter body: %v", err)
	}
	response.Body.Close()
	if len(quickBody.Data) != 2 {
		t.Fatalf("expected 2 quick-tag recipes, got %d", len(quickBody.Data))
	}

	response = doGET(t, client, env.server.URL+"/api/v1/recipes?tag=quick&tag=vegan")
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected multi-tag filter success, got %d body=%s", response.StatusCode, string(body))
	}
	var comboBody struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&comboBody); err != nil {
		t.Fatalf("decode combo filter body: %v", err)
	}
	response.Body.Close()
	if len(comboBody.Data) != 1 || comboBody.Data[0].ID != saladRecipe.ID {
		t.Fatalf("expected only salad recipe for combined tag filter")
	}

	response = doGET(t, client, env.server.URL+"/api/v1/recipes?limit=2")
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected paginated list success, got %d body=%s", response.StatusCode, string(body))
	}
	var pageOne struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Page struct {
			NextCursor *string `json:"nextCursor"`
			HasMore    bool    `json:"hasMore"`
		} `json:"page"`
	}
	if err := json.NewDecoder(response.Body).Decode(&pageOne); err != nil {
		t.Fatalf("decode first page: %v", err)
	}
	response.Body.Close()
	if len(pageOne.Data) != 2 || !pageOne.Page.HasMore || pageOne.Page.NextCursor == nil {
		t.Fatalf("expected first page with next cursor")
	}

	response = doGET(t, client, env.server.URL+"/api/v1/recipes?limit=2&cursor="+url.QueryEscape(*pageOne.Page.NextCursor))
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected second page success, got %d body=%s", response.StatusCode, string(body))
	}
	var pageTwo struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	}
	if err := json.NewDecoder(response.Body).Decode(&pageTwo); err != nil {
		t.Fatalf("decode second page: %v", err)
	}
	response.Body.Close()
	if len(pageTwo.Data) != 1 || pageTwo.Page.HasMore {
		t.Fatalf("expected final second page with one remaining item")
	}
}

func TestViewerCannotPatchArchiveOrCreateTag(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := registerUser(t, client, env.server.URL, "Viewer User", "viewer-phase1@example.com")
	createdRecipe := createRecipe(t, client, env.server.URL, csrfToken, "Viewer Soup")

	ctx := context.Background()
	userEntity, err := env.db.User.Query().Where(user.EmailEQ("viewer-phase1@example.com")).Only(ctx)
	if err != nil {
		t.Fatalf("query viewer user: %v", err)
	}
	if _, err := env.db.HouseholdMember.Update().
		Where(householdmember.UserIDEQ(userEntity.ID)).
		SetRole(householdmember.RoleViewer).
		Save(ctx); err != nil {
		t.Fatalf("downgrade viewer role: %v", err)
	}

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	refreshResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/auth/refresh", csrfToken, map[string]any{})
	if refreshResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(refreshResponse.Body)
		t.Fatalf("expected refresh success after role change, got %d body=%s", refreshResponse.StatusCode, string(body))
	}
	refreshResponse.Body.Close()

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	patchResponse := doJSON(t, client, http.MethodPatch, env.server.URL+"/api/v1/recipes/"+createdRecipe.ID, csrfToken, map[string]any{
		"version":     createdRecipe.Version,
		"title":       "Viewer Soup Updated",
		"description": "blocked",
		"sourceUrl":   "",
		"ingredients": []map[string]any{{"name": "Salt", "sortOrder": 1}},
	})
	if patchResponse.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(patchResponse.Body)
		t.Fatalf("expected viewer patch to be rejected, got %d body=%s", patchResponse.StatusCode, string(body))
	}
	patchResponse.Body.Close()

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	archiveResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/recipes/"+createdRecipe.ID+"/archive", csrfToken, map[string]any{})
	if archiveResponse.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(archiveResponse.Body)
		t.Fatalf("expected viewer archive to be rejected, got %d body=%s", archiveResponse.StatusCode, string(body))
	}
	archiveResponse.Body.Close()

	csrfToken = currentCookieValue(t, client, env.server.URL, "/", authsvc.CSRFCookieName)
	tagResponse := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/tags", csrfToken, map[string]any{
		"name": "Blocked",
	})
	if tagResponse.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(tagResponse.Body)
		t.Fatalf("expected viewer tag create to be rejected, got %d body=%s", tagResponse.StatusCode, string(body))
	}
	tagResponse.Body.Close()
}

func TestPatchRejectsCrossHouseholdTagAssignmentAndBadCursor(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	ownerClient := newHTTPClient(t)
	otherClient := newHTTPClient(t)

	csrfToken := registerUser(t, ownerClient, env.server.URL, "Owner", "owner-tags@example.com")
	foreignTag := createTag(t, ownerClient, env.server.URL, csrfToken, "Foreign", "#173A40")

	csrfToken = registerUser(t, otherClient, env.server.URL, "Other", "other-tags@example.com")
	recipe := createRecipe(t, otherClient, env.server.URL, csrfToken, "Other Soup")

	csrfToken = currentCookieValue(t, otherClient, env.server.URL, "/", authsvc.CSRFCookieName)
	patchResponse := doJSON(t, otherClient, http.MethodPatch, env.server.URL+"/api/v1/recipes/"+recipe.ID, csrfToken, map[string]any{
		"version":     recipe.Version,
		"title":       "Other Soup Updated",
		"description": "blocked",
		"sourceUrl":   "",
		"ingredients": []map[string]any{{"name": "Salt", "sortOrder": 1}},
		"tagIds":      []string{foreignTag.ID},
	})
	if patchResponse.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(patchResponse.Body)
		t.Fatalf("expected foreign tag assignment rejection, got %d body=%s", patchResponse.StatusCode, string(body))
	}
	patchResponse.Body.Close()

	cursorResponse := doGET(t, otherClient, env.server.URL+"/api/v1/recipes?cursor=not-a-real-cursor")
	defer cursorResponse.Body.Close()
	if cursorResponse.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(cursorResponse.Body)
		t.Fatalf("expected malformed cursor rejection, got %d body=%s", cursorResponse.StatusCode, string(body))
	}
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	rawDB, err := sql.Open("sqlite", "file:"+url.QueryEscape(t.Name())+"?mode=memory&cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	t.Cleanup(func() {
		_ = rawDB.Close()
	})
	if _, err := rawDB.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, rawDB)))
	if client == nil {
		t.Fatal("expected ent client")
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestid.Middleware())

	signer := authsvc.NewTokenSigner("test-jwt-secret")
	csrf := authsvc.NewCSRFManager("test-csrf-secret")
	cookies := authsvc.NewCookieManager(false)
	authService := authsvc.NewService(client, clock.RealClock{}, signer, csrf)
	store := storage.NewPostgresStore(client)
	recipeService := recipesvc.NewService(client, store, clock.RealClock{})
	loginLimiter := ratelimit.New(5, time.Minute)
	refreshLimiter := ratelimit.New(30, time.Minute)

	api := router.Group("/api/v1")
	authhttpapi.RegisterRoutes(api, authService, cookies, csrf, loginLimiter, refreshLimiter)
	recipeshttpapi.RegisterRoutes(
		api,
		recipeService,
		authsvc.Middleware(signer, cookies),
		authsvc.CSRFMiddleware(csrf),
		2<<20,
		[]string{"image/png", "image/jpeg", "image/webp"},
	)

	return &testEnv{
		server: httptest.NewServer(router),
		db:     client,
	}
}

func newHTTPClient(t *testing.T) *http.Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create cookie jar: %v", err)
	}

	return &http.Client{Jar: jar}
}

func bootstrapCSRF(t *testing.T, client *http.Client, serverURL string) string {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, serverURL+"/api/v1/auth/session", nil)
	if err != nil {
		t.Fatalf("new csrf bootstrap request: %v", err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("perform csrf bootstrap request: %v", err)
	}
	response.Body.Close()

	token := currentCookieValue(t, client, serverURL, "/", authsvc.CSRFCookieName)
	if token == "" {
		t.Fatal("expected csrf cookie after bootstrap")
	}
	return token
}

func currentCookieValue(t *testing.T, client *http.Client, rawURL string, path string, name string) string {
	t.Helper()

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	parsedURL.Path = path

	for _, cookie := range client.Jar.Cookies(parsedURL) {
		if cookie.Name == name {
			return cookie.Value
		}
	}

	return ""
}

func doJSON(t *testing.T, client *http.Client, method string, rawURL string, csrfToken string, payload any) *http.Response {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal json body: %v", err)
	}

	request, err := http.NewRequest(method, rawURL, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new json request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("perform json request: %v", err)
	}

	return response
}

func doGET(t *testing.T, client *http.Client, rawURL string) *http.Response {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("new get request: %v", err)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("perform get request: %v", err)
	}

	return response
}

type recipeResult struct {
	ID      string
	Version int
}

type tagResult struct {
	ID   string
	Slug string
}

func registerUser(t *testing.T, client *http.Client, serverURL string, displayName string, email string) string {
	t.Helper()

	csrfToken := bootstrapCSRF(t, client, serverURL)
	response := doJSON(t, client, http.MethodPost, serverURL+"/api/v1/auth/register", csrfToken, map[string]any{
		"displayName": displayName,
		"email":       email,
		"password":    "supersecret123",
	})
	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected register success, got %d body=%s", response.StatusCode, string(body))
	}
	response.Body.Close()

	return currentCookieValue(t, client, serverURL, "/", authsvc.CSRFCookieName)
}

func createRecipe(t *testing.T, client *http.Client, serverURL string, csrfToken string, title string) recipeResult {
	t.Helper()

	response := doJSON(t, client, http.MethodPost, serverURL+"/api/v1/recipes", csrfToken, map[string]any{
		"title":       title,
		"description": title + " description",
		"ingredients": []map[string]any{{"name": "Salt", "sortOrder": 1}},
	})
	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected recipe create success, got %d body=%s", response.StatusCode, string(body))
	}
	defer response.Body.Close()

	var responseBody struct {
		Recipe struct {
			ID      string `json:"id"`
			Version int    `json:"version"`
		} `json:"recipe"`
	}
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		t.Fatalf("decode create recipe body: %v", err)
	}

	return recipeResult{
		ID:      responseBody.Recipe.ID,
		Version: responseBody.Recipe.Version,
	}
}

func createTag(t *testing.T, client *http.Client, serverURL string, csrfToken string, name string, color string) tagResult {
	t.Helper()

	response := doJSON(t, client, http.MethodPost, serverURL+"/api/v1/tags", csrfToken, map[string]any{
		"name":  name,
		"color": color,
	})
	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected tag create success, got %d body=%s", response.StatusCode, string(body))
	}
	defer response.Body.Close()

	var responseBody struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		t.Fatalf("decode create tag body: %v", err)
	}

	return tagResult{
		ID:   responseBody.ID,
		Slug: responseBody.Slug,
	}
}

func patchRecipeTags(t *testing.T, client *http.Client, serverURL string, recipe recipeResult, title string, tagIDs []string) {
	t.Helper()

	csrfToken := currentCookieValue(t, client, serverURL, "/", authsvc.CSRFCookieName)
	response := doJSON(t, client, http.MethodPatch, serverURL+"/api/v1/recipes/"+recipe.ID, csrfToken, map[string]any{
		"version":     recipe.Version,
		"title":       title,
		"description": "tagged",
		"sourceUrl":   "",
		"ingredients": []map[string]any{{"name": "Salt", "sortOrder": 1}},
		"tagIds":      tagIDs,
	})
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected tag patch success, got %d body=%s", response.StatusCode, string(body))
	}
	response.Body.Close()
}

func doMultipart(t *testing.T, client *http.Client, rawURL string, csrfToken string) *http.Response {
	t.Helper()
	return doMultipartNamed(t, client, rawURL, csrfToken, "soup.png")
}

func doMultipartNamed(t *testing.T, client *http.Client, rawURL string, csrfToken string, filename string) *http.Response {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(validPNG()); err != nil {
		t.Fatalf("write png bytes: %v", err)
	}
	if err := writer.WriteField("altText", "Tomato soup in a bowl"); err != nil {
		t.Fatalf("write alt text field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, rawURL, &body)
	if err != nil {
		t.Fatalf("new multipart request: %v", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-CSRF-Token", csrfToken)

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("perform multipart request: %v", err)
	}

	return response
}

func validPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x03, 0x01, 0x01, 0x00, 0xc9, 0xfe, 0x92,
		0xef, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
