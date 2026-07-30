// Package app содержит use-case логику CLI-клиента.
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/PolyakovEvg/gophkeeper/internal/client"
	"github.com/PolyakovEvg/gophkeeper/internal/client/keyring"
	"github.com/PolyakovEvg/gophkeeper/internal/client/localstore"
	"github.com/PolyakovEvg/gophkeeper/internal/crypto"
	models "github.com/PolyakovEvg/gophkeeper/internal/models"
	"github.com/PolyakovEvg/gophkeeper/internal/otp"

	"github.com/google/uuid"
)

// SecretStore абстрагирует защищённое хранилище сессионных секретов (обычно OS keychain).
// Единственный источник истины для auth-токена и мастер-пароля.
type SecretStore interface {
	SetToken(login, token string) error
	GetToken(login string) (string, error)
	SetMasterPassword(login, password string) error
	GetMasterPassword(login string) (string, error)
	ClearSession(login string) error
}

// App - фасад клиентского приложения.
type App struct {
	API     *client.API
	Store   *localstore.Store
	DataDir string
	Keyring SecretStore
}

// New создаёт приложение с указанным API и локальным хранилищем.
func New(api *client.API, store *localstore.Store, dataDir string) *App {
	return &App{API: api, Store: store, DataDir: dataDir, Keyring: keyring.New()}
}

// DefaultDataDir возвращает ~/.gophkeeper.
func DefaultDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gophkeeper"), nil
}

// Register регистрирует пользователя и сохраняет сессию.
func (a *App) Register(ctx context.Context, login, password string) error {
	if err := a.API.Register(ctx, login, password); err != nil {
		return err
	}
	a.Store.SetPassword(password)
	return a.persistSession(login, password)
}

// Login выполняет вход и сохраняет сессию.
func (a *App) Login(ctx context.Context, login, password string) error {
	if err := a.API.Login(ctx, login, password); err != nil {
		return err
	}
	a.Store.SetPassword(password)
	return a.persistSession(login, password)
}

func (a *App) persistSession(login, password string) error {
	if err := a.Store.SetMeta("login", login); err != nil {
		return err
	}
	if err := a.Keyring.SetToken(login, a.API.Token()); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to store token in keychain: %v\n", err)
	}
	if err := a.Keyring.SetMasterPassword(login, password); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to store password in keychain: %v\n", err)
	}
	return nil
}

// RestoreSession восстанавливает сессию из keychain (токен и мастер-пароль).
// Если пароль не найден в keychain, возвращает ошибку ErrPasswordRequired.
var ErrPasswordRequired = fmt.Errorf("master password required")

func (a *App) RestoreSession() error {
	login, err := a.Store.GetMeta("login")
	if err != nil || login == "" {
		return fmt.Errorf("not logged in: %w", err)
	}
	token, err := a.Keyring.GetToken(login)
	if err != nil || token == "" {
		return fmt.Errorf("not logged in: %w", err)
	}
	password, err := a.Keyring.GetMasterPassword(login)
	if err != nil {
		return ErrPasswordRequired
	}
	a.API.SetToken(token)
	a.Store.SetPassword(password)
	return nil
}

// RestoreSessionWithPassword восстанавливает сессию с явно указанным паролем.
func (a *App) RestoreSessionWithPassword(password string) error {
	login, err := a.Store.GetMeta("login")
	if err != nil || login == "" {
		return fmt.Errorf("not logged in: %w", err)
	}
	token, err := a.Keyring.GetToken(login)
	if err != nil || token == "" {
		return fmt.Errorf("not logged in: %w", err)
	}
	a.API.SetToken(token)
	a.Store.SetPassword(password)
	if err := a.Keyring.SetMasterPassword(login, password); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to store password in keychain: %v\n", err)
	}
	return nil
}

// HasSession проверяет, есть ли сохранённая сессия (токен в keychain).
func (a *App) HasSession() bool {
	login, err := a.Store.GetMeta("login")
	if err != nil || login == "" {
		return false
	}
	token, err := a.Keyring.GetToken(login)
	return err == nil && token != ""
}

// AddItem добавляет новую запись локально.
func (a *App) AddItem(payload models.ItemPayload) (*models.LocalItem, error) {
	if !payload.Type.Valid() {
		return nil, fmt.Errorf("unsupported item type: %s", payload.Type)
	}
	item := models.LocalItem{
		ID:        uuid.New(),
		Version:   1,
		UpdatedAt: time.Now().UTC(),
		Deleted:   false,
		Dirty:     true,
		Payload:   payload,
	}
	if err := a.Store.SaveItem(item); err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateItem обновляет существующую запись.
func (a *App) UpdateItem(id uuid.UUID, payload models.ItemPayload) error {
	existing, err := a.Store.GetItem(id)
	if err != nil {
		return err
	}
	existing.Payload = payload
	existing.Version++
	existing.UpdatedAt = time.Now().UTC()
	existing.Dirty = true
	existing.Deleted = false
	return a.Store.SaveItem(*existing)
}

// DeleteItem помечает запись удалённой.
func (a *App) DeleteItem(id uuid.UUID) error {
	existing, err := a.Store.GetItem(id)
	if err != nil {
		return err
	}
	existing.Deleted = true
	existing.Dirty = true
	existing.Version++
	existing.UpdatedAt = time.Now().UTC()
	return a.Store.SaveItem(*existing)
}

// GetItem возвращает локальную запись.
func (a *App) GetItem(id uuid.UUID) (*models.LocalItem, error) {
	return a.Store.GetItem(id)
}

// ListItems возвращает локальные записи.
func (a *App) ListItems() ([]models.LocalItem, error) {
	return a.Store.ListItems()
}

// Sync выполняет синхронизацию. binary=true использует gob-протокол.
func (a *App) Sync(ctx context.Context, binary bool) error {
	if a.Store.Password() == "" {
		return fmt.Errorf("master password not set - please login first")
	}
	if a.API.Token() == "" {
		return fmt.Errorf("not logged in: please login first")
	}

	var since time.Time
	if raw, err := a.Store.GetMeta("last_sync"); err == nil {
		since, _ = time.Parse(time.RFC3339Nano, raw)
	}

	// Записи читаются из локальной БД потоково (без буферизации всего dirty-набора
	// в памяти) и сразу шифруются в исходящий wireItems.
	pushedVersions := make(map[uuid.UUID]int64)
	var dirtyIDs []uuid.UUID
	var wireItems []client.Item
	for it, err := range a.Store.ListDirtySeq() {
		if err != nil {
			return err
		}
		raw, err := json.Marshal(it.Payload)
		if err != nil {
			return err
		}
		enc, err := crypto.EncryptWithPassword(raw, a.Store.Password())
		if err != nil {
			return err
		}
		pushedVersions[it.ID] = it.Version
		dirtyIDs = append(dirtyIDs, it.ID)
		wireItems = append(wireItems, client.Item{
			ID: it.ID.String(), Version: it.Version, UpdatedAt: it.UpdatedAt,
			Deleted: it.Deleted, Payload: enc,
		})
	}

	var err error
	var resp *client.SyncResponse
	if binary {
		resp, err = a.API.SyncBinary(ctx, since, wireItems)
	} else {
		resp, err = a.API.SyncJSON(ctx, client.SyncRequest{Since: since, Items: wireItems})
	}
	if err != nil {
		return err
	}

	conflicts := make(map[uuid.UUID]struct{})
	incomplete := false
	for _, remote := range resp.Items {
		id, err := uuid.Parse(remote.ID)
		if err != nil {
			incomplete = true
			continue
		}
		raw, err := crypto.DecryptWithPassword(remote.Payload, a.Store.Password())
		if err != nil {
			incomplete = true
			continue
		}
		var payload models.ItemPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			incomplete = true
			continue
		}

		if pushedVersion, ok := pushedVersions[id]; ok && remote.Version > pushedVersion {
			conflicts[id] = struct{}{}
			continue
		}

		local, err := a.Store.GetItem(id)
		if err == nil && local.Version >= remote.Version && !local.Dirty {
			continue
		}

		item := models.LocalItem{
			ID: id, Version: remote.Version, UpdatedAt: remote.UpdatedAt,
			Deleted: remote.Deleted, Dirty: false, Payload: payload,
		}
		if err := a.Store.SaveItem(item); err != nil {
			return err
		}
	}

	for _, id := range dirtyIDs {
		if _, conflicted := conflicts[id]; conflicted {
			continue
		}
		_ = a.Store.MarkClean(id)
	}

	if !incomplete {
		if err := a.Store.SetMeta("last_sync", resp.ServerTime.UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("sync completed with %d conflicting item(s): local changes were newer on the server and were not applied; resolve and sync again", len(conflicts))
	}
	return nil
}

// Logout очищает сессию и разлогинивает пользователя.
func (a *App) Logout() error {
	login, _ := a.Store.GetMeta("login")
	a.API.SetToken("")
	a.Store.SetPassword("")
	if login != "" {
		if err := a.Keyring.ClearSession(login); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to clear keychain session: %v\n", err)
		}
	}
	if err := a.Store.SetMeta("login", ""); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to clear login: %v\n", err)
	}
	if err := a.Store.SetMeta("last_sync", ""); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to clear last_sync: %v\n", err)
	}
	return nil
}

// GetLogin возвращает логин текущего пользователя или пустую строку.
func (a *App) GetLogin() string {
	login, err := a.Store.GetMeta("login")
	if err != nil {
		return ""
	}
	return login
}

// OTPCode генерирует текущий TOTP для OTP-записи.
func (a *App) OTPCode(id uuid.UUID) (string, error) {
	item, err := a.Store.GetItem(id)
	if err != nil {
		return "", err
	}
	if item.Payload.Type != models.ItemOTP || item.Payload.OTP == nil {
		return "", fmt.Errorf("item is not otp")
	}
	period := item.Payload.OTP.Period
	digits := item.Payload.OTP.Digits
	return otp.Generate(item.Payload.OTP.Secret, time.Now(), uint(period), digits)
}
