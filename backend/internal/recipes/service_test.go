package recipes

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/url"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	"github.com/AndreasRoither/NomNomVault/backend/internal/ent"
	entgen "github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated"
	enthook "github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/hook"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/clock"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/httpx"
	platformmedia "github.com/AndreasRoither/NomNomVault/backend/internal/platform/media"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/storage"
)

func TestAttachRecipeMediaCleansUpCreatedObjectsOnMutationFailure(t *testing.T) {
	env := newServiceTestEnv(t)

	env.client.MediaAsset.Use(enthook.On(func(next entgen.Mutator) entgen.Mutator {
		return entgen.MutateFunc(func(ctx context.Context, m entgen.Mutation) (entgen.Value, error) {
			return nil, errors.New("forced media mutation failure")
		})
	}, entgen.OpCreate))

	_, err := env.service.AttachRecipeMedia(env.ctx, AttachRecipeMediaInput{
		HouseholdID:      env.household.ID,
		ActorRole:        "owner",
		RecipeID:         env.recipe.ID,
		OriginalFilename: "soup.png",
		MimeType:         "image/png",
		AltText:          "Soup",
		Content:          validOpaquePNG(),
	})
	if err == nil {
		t.Fatal("expected media attach to fail")
	}

	count, err := env.client.StoredObject.Query().Count(env.ctx)
	if err != nil {
		t.Fatalf("count stored objects: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected cleanup to remove created stored objects, got %d", count)
	}
}

func TestAttachRecipeMediaDoesNotDeleteReusedObjectsOnFailure(t *testing.T) {
	env := newServiceTestEnv(t)

	first, err := env.service.AttachRecipeMedia(env.ctx, AttachRecipeMediaInput{
		HouseholdID:      env.household.ID,
		ActorRole:        "owner",
		RecipeID:         env.recipe.ID,
		OriginalFilename: "first.png",
		MimeType:         "image/png",
		AltText:          "First",
		Content:          validOpaquePNG(),
	})
	if err != nil {
		t.Fatalf("attach first media: %v", err)
	}
	if first.ID == "" {
		t.Fatal("expected first media id")
	}

	secondRecipe, err := env.client.Recipe.Create().
		SetHouseholdID(env.household.ID).
		SetTitle("Second Recipe").
		Save(env.ctx)
	if err != nil {
		t.Fatalf("create second recipe: %v", err)
	}

	env.client.MediaAsset.Use(enthook.On(func(next entgen.Mutator) entgen.Mutator {
		return entgen.MutateFunc(func(ctx context.Context, m entgen.Mutation) (entgen.Value, error) {
			return nil, errors.New("forced media mutation failure")
		})
	}, entgen.OpCreate))

	_, err = env.service.AttachRecipeMedia(env.ctx, AttachRecipeMediaInput{
		HouseholdID:      env.household.ID,
		ActorRole:        "owner",
		RecipeID:         secondRecipe.ID,
		OriginalFilename: "second.png",
		MimeType:         "image/png",
		AltText:          "Second",
		Content:          validOpaquePNG(),
	})
	if err == nil {
		t.Fatal("expected second media attach to fail")
	}

	count, err := env.client.StoredObject.Query().Count(env.ctx)
	if err != nil {
		t.Fatalf("count stored objects: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected reused stored objects to remain, got %d", count)
	}
}

func TestGetMediaContentUsesMediaAssetFilenamesWithDedupedBlobs(t *testing.T) {
	env := newServiceTestEnv(t)

	first, err := env.service.AttachRecipeMedia(env.ctx, AttachRecipeMediaInput{
		HouseholdID:      env.household.ID,
		ActorRole:        "owner",
		RecipeID:         env.recipe.ID,
		OriginalFilename: "first.png",
		MimeType:         "image/png",
		AltText:          "First",
		Content:          validOpaquePNG(),
	})
	if err != nil {
		t.Fatalf("attach first media: %v", err)
	}

	secondRecipe, err := env.client.Recipe.Create().
		SetHouseholdID(env.household.ID).
		SetTitle("Second Recipe").
		Save(env.ctx)
	if err != nil {
		t.Fatalf("create second recipe: %v", err)
	}

	second, err := env.service.AttachRecipeMedia(env.ctx, AttachRecipeMediaInput{
		HouseholdID:      env.household.ID,
		ActorRole:        "owner",
		RecipeID:         secondRecipe.ID,
		OriginalFilename: "second.png",
		MimeType:         "image/png",
		AltText:          "Second",
		Content:          validOpaquePNG(),
	})
	if err != nil {
		t.Fatalf("attach second media: %v", err)
	}
	if first.ID == second.ID {
		t.Fatal("expected distinct media assets")
	}

	original, err := env.service.GetMediaOriginal(env.ctx, env.household.ID, second.ID)
	if err != nil {
		t.Fatalf("get original media: %v", err)
	}
	if original.Filename != "second.png" {
		t.Fatalf("expected second asset filename, got %q", original.Filename)
	}
	if original.MimeType != "image/png" {
		t.Fatalf("expected original mime image/png, got %q", original.MimeType)
	}

	thumbnail, err := env.service.GetMediaThumbnail(env.ctx, env.household.ID, second.ID)
	if err != nil {
		t.Fatalf("get thumbnail media: %v", err)
	}
	expectedThumbnailName := platformmedia.ThumbnailFilename("second.png", thumbnail.MimeType)
	if thumbnail.Filename != expectedThumbnailName {
		t.Fatalf("expected thumbnail filename %q, got %q", expectedThumbnailName, thumbnail.Filename)
	}
}

func TestGetMediaThumbnailReturnsVariantNotFoundForLegacyMedia(t *testing.T) {
	env := newServiceTestEnv(t)

	object, err := env.store.Put(env.ctx, storage.PutInput{
		HouseholdID:      env.household.ID,
		OriginalFilename: "legacy.png",
		MimeType:         "image/png",
		Checksum:         checksum(validOpaquePNG()),
		Content:          validOpaquePNG(),
	})
	if err != nil {
		t.Fatalf("store legacy original: %v", err)
	}

	legacyMedia, err := env.client.MediaAsset.Create().
		SetHouseholdID(env.household.ID).
		SetRecipeID(env.recipe.ID).
		SetStorageObjectID(object.ID).
		SetOriginalFilename("legacy.png").
		SetMimeType("image/png").
		SetMediaType("image").
		SetSizeBytes(object.SizeBytes).
		SetChecksum(object.Checksum).
		SetStoredAt(env.clock.Now()).
		SetSortOrder(1).
		Save(env.ctx)
	if err != nil {
		t.Fatalf("create legacy media asset: %v", err)
	}

	_, err = env.service.GetMediaThumbnail(env.ctx, env.household.ID, legacyMedia.ID)
	if err == nil {
		t.Fatal("expected missing thumbnail variant error")
	}

	var statusErr httpx.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected status error, got %T", err)
	}
	if statusErr.Status != 404 || statusErr.Code != "media_variant_not_found" {
		t.Fatalf("expected 404 media_variant_not_found, got %d %q", statusErr.Status, statusErr.Code)
	}
}

type serviceTestEnv struct {
	ctx       context.Context
	client    *ent.Client
	store     *storage.PostgresStore
	service   *Service
	household *entgen.Household
	recipe    *entgen.Recipe
	clock     clock.Clock
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
	household, err := client.Household.Create().SetName("Test Household").SetSlug("test-household").Save(ctx)
	if err != nil {
		t.Fatalf("create household: %v", err)
	}
	recipe, err := client.Recipe.Create().SetHouseholdID(household.ID).SetTitle("Soup").Save(ctx)
	if err != nil {
		t.Fatalf("create recipe: %v", err)
	}

	store := storage.NewPostgresStore(client)
	testClock := clock.RealClock{}

	return serviceTestEnv{
		ctx:       ctx,
		client:    client,
		store:     store,
		service:   NewService(client, store, testClock),
		household: household,
		recipe:    recipe,
		clock:     testClock,
	}
}

func validOpaquePNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 220, G: 64, B: 64, A: 255})
	img.Set(1, 0, color.RGBA{R: 64, G: 140, B: 220, A: 255})
	img.Set(0, 1, color.RGBA{R: 64, G: 220, B: 96, A: 255})
	img.Set(1, 1, color.RGBA{R: 240, G: 220, B: 64, A: 255})

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		panic(err)
	}
	return encoded.Bytes()
}
