package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	authsvc "github.com/AndreasRoither/NomNomVault/backend/internal/auth"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/householdmember"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/recipe"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/user"
)

func TestRecipeListHidesDraftsUnlessRequested(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := registerUser(t, client, env.server.URL, "Draft User", "drafts@example.com")
	published := createRecipe(t, client, env.server.URL, csrfToken, "Published Soup")

	ctx := context.Background()
	householdID := activeHouseholdIDForEmail(t, env, "drafts@example.com")
	draftEntity, err := env.db.Recipe.Create().
		SetHouseholdID(householdID).
		SetTitle("Draft Soup").
		SetDescription("Pending review").
		SetStatus(recipe.StatusDraft).
		Save(ctx)
	if err != nil {
		t.Fatalf("create draft recipe: %v", err)
	}

	response := doGET(t, client, env.server.URL+"/api/v1/recipes")
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected recipe list success, got %d body=%s", response.StatusCode, string(body))
	}
	var listBody struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode default recipe list: %v", err)
	}
	response.Body.Close()
	if len(listBody.Data) != 1 || listBody.Data[0].ID != published.ID {
		t.Fatalf("expected default list to hide draft recipes")
	}

	response = doGET(t, client, env.server.URL+"/api/v1/recipes?draft=true")
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected draft-inclusive recipe list success, got %d body=%s", response.StatusCode, string(body))
	}
	var draftListBody struct {
		Data []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&draftListBody); err != nil {
		t.Fatalf("decode draft recipe list: %v", err)
	}
	response.Body.Close()
	if len(draftListBody.Data) != 2 {
		t.Fatalf("expected draft-inclusive list to return both recipes, got %d", len(draftListBody.Data))
	}

	detailResponse := doGET(t, client, env.server.URL+"/api/v1/recipes/"+draftEntity.ID)
	if detailResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(detailResponse.Body)
		t.Fatalf("expected draft recipe detail success, got %d body=%s", detailResponse.StatusCode, string(body))
	}
	var detailBody struct {
		Recipe struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"recipe"`
	}
	if err := json.NewDecoder(detailResponse.Body).Decode(&detailBody); err != nil {
		t.Fatalf("decode draft recipe detail: %v", err)
	}
	detailResponse.Body.Close()
	if detailBody.Recipe.ID != draftEntity.ID || detailBody.Recipe.Status != "draft" {
		t.Fatalf("expected draft recipe detail to remain addressable by id")
	}
}

func TestURLImportCreateListGetAndIdempotentReuse(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := registerUser(t, client, env.server.URL, "Import User", "import-user@example.com")

	first := doJSONWithHeaders(t, client, http.MethodPost, env.server.URL+"/api/v1/imports/url", csrfToken, map[string]string{
		"Idempotency-Key": "url-key-1",
	}, map[string]any{
		"url":       "https://example.com/recipe?utm_source=newsletter&a=1",
		"titleHint": "Weeknight Pasta",
	})
	if first.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(first.Body)
		t.Fatalf("expected import create accepted, got %d body=%s", first.StatusCode, string(body))
	}
	var firstBody struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Source struct {
			NormalizedURL  *string `json:"normalizedUrl"`
			RetentionState string  `json:"retentionState"`
		} `json:"source"`
	}
	var rawFirstBody map[string]any
	bodyBytes, err := io.ReadAll(first.Body)
	if err != nil {
		t.Fatalf("read import create body: %v", err)
	}
	if err := json.Unmarshal(bodyBytes, &firstBody); err != nil {
		t.Fatalf("decode import create body: %v", err)
	}
	if err := json.Unmarshal(bodyBytes, &rawFirstBody); err != nil {
		t.Fatalf("decode raw import create body: %v", err)
	}
	first.Body.Close()
	if firstBody.ID == "" || firstBody.Status != "queued" {
		t.Fatalf("expected queued import job with id")
	}
	if firstBody.Source.NormalizedURL == nil || *firstBody.Source.NormalizedURL != "https://example.com/recipe?a=1" {
		t.Fatalf("expected normalized URL without tracking params, got %v", firstBody.Source.NormalizedURL)
	}
	if firstBody.Source.RetentionState != "retained" {
		t.Fatalf("expected retained source, got %q", firstBody.Source.RetentionState)
	}
	if _, ok := rawFirstBody["idempotencyKey"]; ok {
		t.Fatalf("expected idempotencyKey to stay internal")
	}
	if _, ok := rawFirstBody["fallbackFingerprint"]; ok {
		t.Fatalf("expected fallbackFingerprint to stay internal")
	}
	sourceBody, ok := rawFirstBody["source"].(map[string]any)
	if !ok {
		t.Fatalf("expected source object in import response")
	}
	if _, ok := sourceBody["contentHash"]; ok {
		t.Fatalf("expected contentHash to stay internal")
	}
	if _, ok := sourceBody["metadata"]; ok {
		t.Fatalf("expected metadata to stay internal")
	}

	second := doJSONWithHeaders(t, client, http.MethodPost, env.server.URL+"/api/v1/imports/url", csrfToken, map[string]string{
		"Idempotency-Key": "url-key-1",
	}, map[string]any{
		"url":       "https://example.com/recipe?a=1&utm_source=another",
		"titleHint": "Weeknight Pasta",
	})
	if second.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(second.Body)
		t.Fatalf("expected active idempotent import reuse, got %d body=%s", second.StatusCode, string(body))
	}
	var secondBody struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(second.Body).Decode(&secondBody); err != nil {
		t.Fatalf("decode idempotent import body: %v", err)
	}
	second.Body.Close()
	if secondBody.ID != firstBody.ID {
		t.Fatalf("expected idempotent import to return the existing active job")
	}

	listResponse := doGET(t, client, env.server.URL+"/api/v1/imports?status=queued")
	if listResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(listResponse.Body)
		t.Fatalf("expected import list success, got %d body=%s", listResponse.StatusCode, string(body))
	}
	var listBody struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(listResponse.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode import list body: %v", err)
	}
	listResponse.Body.Close()
	if len(listBody.Data) != 1 || listBody.Data[0].ID != firstBody.ID {
		t.Fatalf("expected queued import to appear in list")
	}

	getResponse := doGET(t, client, env.server.URL+"/api/v1/imports/"+firstBody.ID)
	if getResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(getResponse.Body)
		t.Fatalf("expected import detail success, got %d body=%s", getResponse.StatusCode, string(body))
	}
	getResponse.Body.Close()
}

func TestURLImportRejectsUnsupportedScheme(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	client := newHTTPClient(t)
	csrfToken := registerUser(t, client, env.server.URL, "Bad Scheme User", "bad-scheme@example.com")

	response := doJSON(t, client, http.MethodPost, env.server.URL+"/api/v1/imports/url", csrfToken, map[string]any{
		"url": "ftp://example.com/recipe.txt",
	})
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected unsupported scheme rejection, got %d body=%s", response.StatusCode, string(body))
	}
}

func TestImportCancelRetryAndCrossHouseholdIsolation(t *testing.T) {
	env := newTestEnv(t)
	defer env.server.Close()

	ownerClient := newHTTPClient(t)
	otherClient := newHTTPClient(t)

	csrfToken := registerUser(t, ownerClient, env.server.URL, "Owner", "owner-imports@example.com")
	createResponse := doJSONWithHeaders(t, ownerClient, http.MethodPost, env.server.URL+"/api/v1/imports/url", csrfToken, map[string]string{
		"Idempotency-Key": "url-key-2",
	}, map[string]any{
		"url": "https://example.com/recipe/two",
	})
	if createResponse.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(createResponse.Body)
		t.Fatalf("expected import create accepted, got %d body=%s", createResponse.StatusCode, string(body))
	}
	var createBody struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createResponse.Body).Decode(&createBody); err != nil {
		t.Fatalf("decode cancel/retry create body: %v", err)
	}
	createResponse.Body.Close()

	csrfToken = currentCookieValue(t, ownerClient, env.server.URL, "/", authsvc.CSRFCookieName)
	cancelResponse := doJSON(t, ownerClient, http.MethodPost, env.server.URL+"/api/v1/imports/"+createBody.ID+"/cancel", csrfToken, map[string]any{})
	if cancelResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(cancelResponse.Body)
		t.Fatalf("expected cancel success, got %d body=%s", cancelResponse.StatusCode, string(body))
	}
	var cancelBody struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(cancelResponse.Body).Decode(&cancelBody); err != nil {
		t.Fatalf("decode cancel body: %v", err)
	}
	cancelResponse.Body.Close()
	if cancelBody.Status != "cancelled" {
		t.Fatalf("expected cancelled status, got %q", cancelBody.Status)
	}

	csrfToken = currentCookieValue(t, ownerClient, env.server.URL, "/", authsvc.CSRFCookieName)
	retryWithoutConfirm := doJSON(t, ownerClient, http.MethodPost, env.server.URL+"/api/v1/imports/"+createBody.ID+"/retry", csrfToken, map[string]any{
		"confirmFinished": false,
	})
	if retryWithoutConfirm.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(retryWithoutConfirm.Body)
		t.Fatalf("expected retry confirmation conflict, got %d body=%s", retryWithoutConfirm.StatusCode, string(body))
	}
	retryWithoutConfirm.Body.Close()

	csrfToken = currentCookieValue(t, ownerClient, env.server.URL, "/", authsvc.CSRFCookieName)
	retryResponse := doJSON(t, ownerClient, http.MethodPost, env.server.URL+"/api/v1/imports/"+createBody.ID+"/retry", csrfToken, map[string]any{
		"confirmFinished": true,
	})
	if retryResponse.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(retryResponse.Body)
		t.Fatalf("expected retry accepted, got %d body=%s", retryResponse.StatusCode, string(body))
	}
	var retryBody struct {
		ID           string `json:"id"`
		AttemptCount int    `json:"attemptCount"`
		Status       string `json:"status"`
	}
	if err := json.NewDecoder(retryResponse.Body).Decode(&retryBody); err != nil {
		t.Fatalf("decode retry body: %v", err)
	}
	retryResponse.Body.Close()
	if retryBody.ID == createBody.ID || retryBody.AttemptCount != 2 || retryBody.Status != "queued" {
		t.Fatalf("expected retry to create a new queued attempt")
	}

	_ = registerUser(t, otherClient, env.server.URL, "Other", "other-imports@example.com")
	crossHousehold := doGET(t, otherClient, env.server.URL+"/api/v1/imports/"+createBody.ID)
	if crossHousehold.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(crossHousehold.Body)
		t.Fatalf("expected cross-household import lookup to fail, got %d body=%s", crossHousehold.StatusCode, string(body))
	}
	crossHousehold.Body.Close()
}

func doJSONWithHeaders(t *testing.T, client *http.Client, method string, rawURL string, csrfToken string, headers map[string]string, payload any) *http.Response {
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
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("perform json request: %v", err)
	}
	return response
}

func activeHouseholdIDForEmail(t *testing.T, env *testEnv, email string) string {
	t.Helper()

	ctx := context.Background()
	sessionUser, err := env.db.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err != nil {
		t.Fatalf("query user: %v", err)
	}
	membership, err := env.db.HouseholdMember.Query().
		Where(householdmember.UserIDEQ(sessionUser.ID)).
		Only(ctx)
	if err != nil {
		t.Fatalf("query membership: %v", err)
	}
	return membership.HouseholdID
}
