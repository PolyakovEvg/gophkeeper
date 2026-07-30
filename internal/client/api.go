// Package client реализует HTTP/бинарный API-клиент GophKeeper.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PolyakovEvg/gophkeeper/internal/protocol"

	"github.com/google/uuid"
)

const (
	maxResponseBodySize = 32 << 20
)

// Item - DTO записи сейфа на проводе.
type Item struct {
	ID        string    `json:"id"`
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `json:"deleted"`
	Payload   []byte    `json:"payload"`
}

// SyncRequest - JSON-запрос синхронизации.
type SyncRequest struct {
	Since time.Time `json:"since"`
	Items []Item    `json:"items"`
}

// SyncResponse - JSON-ответ синхронизации.
type SyncResponse struct {
	ServerTime time.Time `json:"server_time"`
	Items      []Item    `json:"items"`
}

// API - клиент удалённого сервера.
type API struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// New создаёт API-клиент.
func New(baseURL string, httpClient *http.Client) *API {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &API{baseURL: baseURL, httpClient: httpClient}
}

// SetToken сохраняет JWT для авторизованных запросов.
func (a *API) SetToken(token string) {
	a.token = token
}

// Token возвращает текущий JWT.
func (a *API) Token() string {
	return a.token
}

// Register регистрирует пользователя и сохраняет токен.
func (a *API) Register(ctx context.Context, login, password string) error {
	token, err := a.auth(ctx, "/api/v1/register", login, password)
	if err != nil {
		return err
	}
	a.token = token
	return nil
}

// Login выполняет вход и сохраняет токен.
func (a *API) Login(ctx context.Context, login, password string) error {
	token, err := a.auth(ctx, "/api/v1/login", login, password)
	if err != nil {
		return err
	}
	a.token = token
	return nil
}

// jsonRequest выполняет JSON-запрос к серверу и декодирует ответ в T.
// body может быть nil (например, для GET-запросов).
func jsonRequest[T any](ctx context.Context, a *API, method, path string, body any, authorize bool) (T, error) {
	var zero T

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return zero, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, reader)
	if err != nil {
		return zero, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authorize {
		a.setAuth(req)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
	if err != nil {
		return zero, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("%s: %s", resp.Status, string(data))
	}

	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return out, nil
}

func (a *API) auth(ctx context.Context, path, login, password string) (string, error) {
	out, err := jsonRequest[tokenResponse](ctx, a, http.MethodPost, path,
		map[string]string{"login": login, "password": password}, false)
	if err != nil {
		return "", err
	}
	return out.Token, nil
}

type tokenResponse struct {
	Token string `json:"token"`
}

// SyncJSON синхронизирует записи через JSON API.
func (a *API) SyncJSON(ctx context.Context, req SyncRequest) (*SyncResponse, error) {
	out, err := jsonRequest[SyncResponse](ctx, a, http.MethodPost, "/api/v1/sync", req, true)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SyncBinary синхронизирует записи через gob-протокол.
func (a *API) SyncBinary(ctx context.Context, since time.Time, items []Item) (*SyncResponse, error) {
	req := protocol.SyncRequest{Since: since}
	for _, it := range items {
		id, err := uuid.Parse(it.ID)
		if err != nil {
			continue
		}
		req.Items = append(req.Items, protocol.ItemEnvelope{
			ID: id, Version: it.Version, UpdatedAt: it.UpdatedAt, Deleted: it.Deleted, Payload: it.Payload,
		})
	}
	body, err := protocol.Encode(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/api/v1/sync/binary", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	a.setAuth(httpReq)
	httpReq.Header.Set("Content-Type", "application/octet-stream")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", resp.Status, string(data))
	}

	var bin protocol.SyncResponse
	if err := protocol.Decode(data, &bin); err != nil {
		return nil, err
	}
	out := &SyncResponse{ServerTime: bin.ServerTime}
	for _, e := range bin.Items {
		out.Items = append(out.Items, Item{
			ID: e.ID.String(), Version: e.Version, UpdatedAt: e.UpdatedAt, Deleted: e.Deleted, Payload: e.Payload,
		})
	}
	return out, nil
}

// ListItems возвращает все записи с сервера.
func (a *API) ListItems(ctx context.Context) ([]Item, error) {
	return jsonRequest[[]Item](ctx, a, http.MethodGet, "/api/v1/items", nil, true)
}

func (a *API) setAuth(req *http.Request) {
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}
}
