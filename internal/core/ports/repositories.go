package ports

import (
	"context"
	"time"

	"github.com/Cristoff-dev/keyforge/internal/core/domain"
)

type KeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByID(ctx context.Context, id string) (*domain.APIKey, error)
	GetByHash(ctx context.Context, hash string) (*domain.APIKey, error)
	UpdateStatus(ctx context.Context, id string, status domain.KeyStatus) error
	GetActiveKeysByClient(ctx context.Context, clientID string) ([]*domain.APIKey, error)
}

type ClientRepository interface {
	Create(ctx context.Context, client *domain.Client) error
	GetByID(ctx context.Context, id string) (*domain.Client, error)
}

type CacheRepository interface {
	SetHashValidation(ctx context.Context, plainKey string, hash string, ttl time.Duration) error
	GetHashValidation(ctx context.Context, plainKey string) (string, error)
	IncrementRateLimit(ctx context.Context, clientID string, limit int, window time.Duration) (bool, error)
	InvalidateHash(ctx context.Context, plainKey string) error
}

