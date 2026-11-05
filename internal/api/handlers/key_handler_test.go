package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Cristoff-dev/keyforge/internal/api/handlers"
	"github.com/Cristoff-dev/keyforge/internal/core/domain"
	"github.com/Cristoff-dev/keyforge/internal/core/services"
	"github.com/go-chi/chi/v5"
)

type mockClientRepo struct {
	GetByIDFunc func(ctx context.Context, id string) (*domain.Client, error)
}

func (m *mockClientRepo) GetByID(ctx context.Context, id string) (*domain.Client, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockClientRepo) Create(ctx context.Context, client *domain.Client) error { return nil }

type mockKeyRepo struct {
	CreateFunc func(ctx context.Context, key *domain.APIKey) error
}

func (m *mockKeyRepo) Create(ctx context.Context, key *domain.APIKey) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, key)
	}
	return nil
}
func (m *mockKeyRepo) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	return nil, nil
}
func (m *mockKeyRepo) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	return nil, nil
}
func (m *mockKeyRepo) UpdateStatus(ctx context.Context, id string, status domain.KeyStatus) error {
	return nil
}
func (m *mockKeyRepo) GetActiveKeysByClient(ctx context.Context, clientID string) ([]*domain.APIKey, error) {
	return nil, nil
}

func TestKeyHandler_HandleCreateKey(t *testing.T) {
	clientRepo := &mockClientRepo{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Client, error) {
			return &domain.Client{ID: "client_01", RateLimitReq: 100}, nil
		},
	}
	keyRepo := &mockKeyRepo{
		CreateFunc: func(ctx context.Context, key *domain.APIKey) error { return nil },
	}

	svc := services.NewKeyService(keyRepo, clientRepo, "test-secret")
	handler := handlers.NewKeyHandler(svc)

	r := chi.NewRouter()
	r.Post("/keys", handler.HandleCreateKey)

	payload := map[string]interface{}{
		"client_id": "client_01",
		"prefix":    "prod",
		"ttl_hours": 720,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Errorf("expected status 201 Created, got %d", rec.Code)
	}

	if !bytes.Contains(rec.Body.Bytes(), []byte("prod_")) {
		t.Errorf("expected response body to contain the generated key, got: %s", rec.Body.String())
	}
}
