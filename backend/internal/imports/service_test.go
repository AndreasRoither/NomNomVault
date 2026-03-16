package imports

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	"github.com/AndreasRoither/NomNomVault/backend/internal/ent"
	entgen "github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/clock"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/httpx"
)

func TestNormalizeImportURLRemovesTrackingParameters(t *testing.T) {
	t.Parallel()

	normalized, err := normalizeImportURL("HTTPS://Example.com/Recipe?b=2&utm_source=newsletter&a=1&fbclid=abc")
	if err != nil {
		t.Fatalf("normalize import url: %v", err)
	}

	expected := "https://example.com/Recipe?a=1&b=2"
	if normalized != expected {
		t.Fatalf("expected normalized URL %q, got %q", expected, normalized)
	}
}

func TestCreateURLImportReusesActiveJobAndRequiresConfirmationAfterFinish(t *testing.T) {
	env := newServiceTestEnv(t)

	first, err := env.service.CreateURLImport(env.ctx, CreateURLImportInput{
		HouseholdID:    env.household.ID,
		ActorUserID:    env.user.ID,
		ActorRole:      "owner",
		URL:            "https://example.com/recipe?utm_source=ad",
		TitleHint:      "Weeknight Pasta",
		IdempotencyKey: "same-key",
	})
	if err != nil {
		t.Fatalf("create first import: %v", err)
	}

	second, err := env.service.CreateURLImport(env.ctx, CreateURLImportInput{
		HouseholdID:    env.household.ID,
		ActorUserID:    env.user.ID,
		ActorRole:      "owner",
		URL:            "https://example.com/recipe?utm_source=ad",
		TitleHint:      "Weeknight Pasta",
		IdempotencyKey: "same-key",
	})
	if err != nil {
		t.Fatalf("create second import: %v", err)
	}
	if !second.Existing || second.Job.ID != first.Job.ID {
		t.Fatalf("expected active idempotent create to return existing job")
	}

	if _, err := env.service.CancelImportJob(env.ctx, CancelImportJobInput{
		HouseholdID: env.household.ID,
		ActorUserID: env.user.ID,
		ActorRole:   "owner",
		JobID:       first.Job.ID,
	}); err != nil {
		t.Fatalf("cancel first job: %v", err)
	}

	_, err = env.service.CreateURLImport(env.ctx, CreateURLImportInput{
		HouseholdID:    env.household.ID,
		ActorUserID:    env.user.ID,
		ActorRole:      "owner",
		URL:            "https://example.com/recipe?utm_source=ad",
		TitleHint:      "Weeknight Pasta",
		IdempotencyKey: "same-key",
	})
	if err == nil {
		t.Fatal("expected confirmation requirement for finished job")
	}

	var statusErr httpx.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected status error, got %T", err)
	}
	if statusErr.Status != 409 || statusErr.Code != "import_restart_confirmation_required" {
		t.Fatalf("expected restart confirmation conflict, got %d %q", statusErr.Status, statusErr.Code)
	}

	restarted, err := env.service.CreateURLImport(env.ctx, CreateURLImportInput{
		HouseholdID:    env.household.ID,
		ActorUserID:    env.user.ID,
		ActorRole:      "owner",
		URL:            "https://example.com/recipe?utm_source=ad",
		TitleHint:      "Weeknight Pasta",
		IdempotencyKey: "same-key",
		ConfirmRestart: true,
	})
	if err != nil {
		t.Fatalf("restart finished import: %v", err)
	}
	if restarted.Existing {
		t.Fatal("expected confirmed restart to create a new job")
	}
	if restarted.Job.ID == first.Job.ID {
		t.Fatal("expected confirmed restart to create a new job id")
	}
}

type serviceTestEnv struct {
	ctx       context.Context
	client    *ent.Client
	service   *Service
	household *entgen.Household
	user      *entgen.User
}

func newServiceTestEnv(t *testing.T) serviceTestEnv {
	t.Helper()

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
	t.Cleanup(func() {
		_ = client.Close()
	})
	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	ctx := context.Background()
	household, err := client.Household.Create().SetName("Imports Household").SetSlug("imports-household").Save(ctx)
	if err != nil {
		t.Fatalf("create household: %v", err)
	}
	user, err := client.User.Create().
		SetDisplayName("Importer").
		SetEmail("importer@example.com").
		SetPasswordHash("hashed").
		Save(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	return serviceTestEnv{
		ctx:       ctx,
		client:    client,
		service:   NewService(client, clock.RealClock{}),
		household: household,
		user:      user,
	}
}
