package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Cristoff-dev/keyforge/internal/core/domain"
)

// --- Mocks de los Repositorios ---

type mockClientRepository struct {
	GetByIDFunc func(ctx context.Context, id string) (*domain.Client, error)
	CreateFunc  func(ctx context.Context, client *domain.Client) error
}

func (m *mockClientRepository) GetByID(ctx context.Context, id string) (*domain.Client, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("unimplemented GetByID")
}

func (m *mockClientRepository) Create(ctx context.Context, client *domain.Client) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, client)
	}
	return errors.New("unimplemented Create")
}

type mockKeyRepository struct {
	CreateFunc                func(ctx context.Context, key *domain.APIKey) error
	GetByIDFunc               func(ctx context.Context, id string) (*domain.APIKey, error)
	GetByHashFunc             func(ctx context.Context, hash string) (*domain.APIKey, error)
	UpdateStatusFunc          func(ctx context.Context, id string, status domain.KeyStatus) error
	GetActiveKeysByClientFunc func(ctx context.Context, clientID string) ([]*domain.APIKey, error)
}

func (m *mockKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, key)
	}
	return errors.New("unimplemented Create")
}

func (m *mockKeyRepository) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("unimplemented GetByID")
}

func (m *mockKeyRepository) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	if m.GetByHashFunc != nil {
		return m.GetByHashFunc(ctx, hash)
	}
	return nil, errors.New("unimplemented GetByHash")
}

func (m *mockKeyRepository) UpdateStatus(ctx context.Context, id string, status domain.KeyStatus) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status)
	}
	return errors.New("unimplemented UpdateStatus")
}

func (m *mockKeyRepository) GetActiveKeysByClient(ctx context.Context, clientID string) ([]*domain.APIKey, error) {
	if m.GetActiveKeysByClientFunc != nil {
		return m.GetActiveKeysByClientFunc(ctx, clientID)
	}
	return nil, errors.New("unimplemented GetActiveKeysByClient")
}

// --- Pruebas Unitarias del Servicio ---

func TestKeyService_CreateKey(t *testing.T) {
	hmacSecret := "test-secret-key-123"

	tests := []struct {
		name          string
		clientID      string
		prefix        string
		ttlHours      time.Duration
		mockClient    *domain.Client
		mockClientErr error
		mockRepoErr   error
		expectError   bool
	}{
		{
			name:          "Successful Key Creation",
			clientID:      "client_01",
			prefix:        "prod_live",
			ttlHours:      720 * time.Hour,
			mockClient:    &domain.Client{ID: "client_01", RateLimitReq: 100},
			mockClientErr: nil,
			mockRepoErr:   nil,
			expectError:   false,
		},
		{
			name:          "Client Not Found",
			clientID:      "non_existent",
			prefix:        "prod_live",
			ttlHours:      720 * time.Hour,
			mockClient:    nil,
			mockClientErr: ErrClientNotFound,
			mockRepoErr:   nil,
			expectError:   true,
		},
		{
			name:          "Database Error on Create",
			clientID:      "client_01",
			prefix:        "prod_live",
			ttlHours:      720 * time.Hour,
			mockClient:    &domain.Client{ID: "client_01", RateLimitReq: 100},
			mockClientErr: nil,
			mockRepoErr:   errors.New("db error"),
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clientRepo := &mockClientRepository{
				GetByIDFunc: func(ctx context.Context, id string) (*domain.Client, error) {
					return tc.mockClient, tc.mockClientErr
				},
			}

			keyRepo := &mockKeyRepository{
				CreateFunc: func(ctx context.Context, key *domain.APIKey) error {
					return tc.mockRepoErr
				},
			}

			service := NewKeyService(keyRepo, clientRepo, hmacSecret)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			response, err := service.CreateKey(ctx, tc.clientID, tc.prefix, tc.ttlHours)

			if tc.expectError {
				if err == nil {
					t.Error("expected an error but got nil")
				}
				if response != nil {
					t.Errorf("expected nil response on error, got %+v", response)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if response == nil {
					t.Fatal("expected a valid response, got nil")
				}
				if !strings.HasPrefix(response.PlainKey, tc.prefix+"_") {
					t.Errorf("expected key to start with prefix %s_, got %s", tc.prefix, response.PlainKey)
				}
			}
		})
	}
}
