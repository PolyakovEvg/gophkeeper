package app_test

import (
	"context"
	"encoding/base32"
	"log/slog"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/PolyakovEvg/gophkeeper/internal/auth"
	"github.com/PolyakovEvg/gophkeeper/internal/client"
	"github.com/PolyakovEvg/gophkeeper/internal/client/app"
	"github.com/PolyakovEvg/gophkeeper/internal/client/localstore"
	"github.com/PolyakovEvg/gophkeeper/internal/models"
	"github.com/PolyakovEvg/gophkeeper/internal/server/handlers"
	"github.com/PolyakovEvg/gophkeeper/internal/server/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppRegisterAddSyncOTP(t *testing.T) {
	srvStore := storage.OpenTest(t)

	authSvc := auth.NewService("secret")
	router := handlers.NewRouter(srvStore, authSvc, slog.Default())
	server := httptest.NewServer(router.Routes())
	t.Cleanup(server.Close)

	store, err := localstore.Open(filepath.Join(t.TempDir(), "client.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	api := client.New(server.URL, server.Client())
	application := app.New(api, store, t.TempDir())

	ctx := context.Background()
	require.NoError(t, application.Register(ctx, "carol", "pass"))

	item, err := application.AddItem(models.ItemPayload{
		Type: models.ItemText, Title: "note",
		Text:     &models.TextData{Content: "secret note"},
		Metadata: map[string]string{"site": "example.com"},
	})
	require.NoError(t, err)

	otpItem, err := application.AddItem(models.ItemPayload{
		Type: models.ItemOTP, Title: "gh",
		// ggignore - demo TOTP seed ("Hello!"), not a real secret
		OTP: &models.OTPData{
			Secret: base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("Hello!")),
			Period: 30,
			Digits: 6,
		},
	})
	require.NoError(t, err)

	require.NoError(t, application.Sync(ctx, false))
	require.NoError(t, application.Sync(ctx, true))

	list, err := application.ListItems()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 2)

	got, err := application.GetItem(item.ID)
	require.NoError(t, err)
	assert.Equal(t, "secret note", got.Payload.Text.Content)

	code, err := application.OTPCode(otpItem.ID)
	require.NoError(t, err)
	assert.Len(t, code, 6)

	require.NoError(t, application.DeleteItem(item.ID))
	require.NoError(t, application.Sync(ctx, false))
}

func TestAppUpdateAndLogin(t *testing.T) {
	srvStore := storage.OpenTest(t)

	server := httptest.NewServer(handlers.NewRouter(srvStore, auth.NewService("s"), slog.Default()).Routes())
	t.Cleanup(server.Close)

	store, err := localstore.Open(filepath.Join(t.TempDir(), "c.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	api := client.New(server.URL, server.Client())
	application := app.New(api, store, t.TempDir())
	ctx := context.Background()

	require.NoError(t, application.Register(ctx, "u", "p"))

	store2, err := localstore.Open(filepath.Join(t.TempDir(), "c2.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store2.Close() })
	app2 := app.New(client.New(server.URL, server.Client()), store2, t.TempDir())
	require.NoError(t, app2.Login(ctx, "u", "p"))

	item, err := application.AddItem(models.ItemPayload{
		Type: models.ItemCredentials, Title: "mail",
		Credentials: &models.CredentialsData{Login: "a", Password: "b"},
	})
	require.NoError(t, err)

	require.NoError(t, application.UpdateItem(item.ID, models.ItemPayload{
		Type: models.ItemCredentials, Title: "mail2",
		Credentials: &models.CredentialsData{Login: "a2", Password: "b2"},
	}))

	require.NoError(t, application.Sync(ctx, false))
	require.NoError(t, app2.Sync(ctx, true))

	list, err := app2.ListItems()
	require.NoError(t, err)
	require.NotEmpty(t, list)

	_, err = app.DefaultDataDir()
	require.NoError(t, err)
}

func TestAppLogoutAndGetLogin(t *testing.T) {
	srvStore := storage.OpenTest(t)

	server := httptest.NewServer(handlers.NewRouter(srvStore, auth.NewService("s"), slog.Default()).Routes())
	t.Cleanup(server.Close)

	store, err := localstore.Open(filepath.Join(t.TempDir(), "c.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	application := app.New(client.New(server.URL, server.Client()), store, t.TempDir())
	ctx := context.Background()

	require.NoError(t, application.Register(ctx, "u", "p"))
	require.Equal(t, "u", application.GetLogin())

	require.NoError(t, application.Logout())
	require.Empty(t, application.GetLogin())
}

func TestAppRestoreSessionNotLoggedIn(t *testing.T) {
	store, err := localstore.Open(filepath.Join(t.TempDir(), "c.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	application := app.New(client.New("http://localhost", nil), store, t.TempDir())
	err = application.RestoreSession()
	require.Error(t, err)
}

func TestAppAddItemInvalidType(t *testing.T) {
	store, err := localstore.Open(filepath.Join(t.TempDir(), "c.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	application := app.New(client.New("http://localhost", nil), store, t.TempDir())
	store.SetPassword("master")
	_, err = application.AddItem(models.ItemPayload{Type: models.ItemType("bogus")})
	require.Error(t, err)
}

func TestAppUpdateItemNotFound(t *testing.T) {
	store, err := localstore.Open(filepath.Join(t.TempDir(), "c.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	application := app.New(client.New("http://localhost", nil), store, t.TempDir())
	err = application.UpdateItem(uuid.New(), models.ItemPayload{Type: models.ItemText, Title: "x"})
	require.Error(t, err)
}

func TestAppDeleteItemNotFound(t *testing.T) {
	store, err := localstore.Open(filepath.Join(t.TempDir(), "c.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	application := app.New(client.New("http://localhost", nil), store, t.TempDir())
	err = application.DeleteItem(uuid.New())
	require.Error(t, err)
}

func TestAppOTPCodeNotOTP(t *testing.T) {
	store, err := localstore.Open(filepath.Join(t.TempDir(), "c.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	application := app.New(client.New("http://localhost", nil), store, t.TempDir())
	store.SetPassword("master")
	item, err := application.AddItem(models.ItemPayload{Type: models.ItemText, Title: "x"})
	require.NoError(t, err)

	_, err = application.OTPCode(item.ID)
	require.Error(t, err)
}

func TestAppOTPCodeNotFound(t *testing.T) {
	store, err := localstore.Open(filepath.Join(t.TempDir(), "c.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	application := app.New(client.New("http://localhost", nil), store, t.TempDir())
	_, err = application.OTPCode(uuid.New())
	require.Error(t, err)
}

func TestAppHasSession(t *testing.T) {
	store, err := localstore.Open(filepath.Join(t.TempDir(), "c.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	application := app.New(client.New("http://localhost", nil), store, t.TempDir())
	assert.False(t, application.HasSession())

	require.NoError(t, store.SetMeta("token", "test-token"))
	assert.True(t, application.HasSession())
}

func TestAppRestoreSessionWithPassword(t *testing.T) {
	srvStore := storage.OpenTest(t)
	server := httptest.NewServer(handlers.NewRouter(srvStore, auth.NewService("s"), slog.Default()).Routes())
	t.Cleanup(server.Close)

	store, err := localstore.Open(filepath.Join(t.TempDir(), "c.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	application := app.New(client.New(server.URL, server.Client()), store, t.TempDir())
	ctx := context.Background()

	require.NoError(t, application.Register(ctx, "restoreuser", "pass"))
	require.NoError(t, application.Logout())

	require.NoError(t, store.SetMeta("token", "test-token"))
	require.NoError(t, store.SetMeta("login", "restoreuser"))

	err = application.RestoreSessionWithPassword("newpass")
	require.NoError(t, err)
}
