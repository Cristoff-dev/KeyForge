package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Cristoff-dev/keyforge/internal/core/domain"
	"github.com/Cristoff-dev/keyforge/internal/core/ports"
	"github.com/Cristoff-dev/keyforge/internal/security"
	"github.com/segmentio/ksuid"
)

var (
	ErrClientNotFound = errors.New("client not found")
	ErrKeyNotFound    = errors.New("key not found")
)

type KeyService struct {
	keyRepo    ports.KeyRepository
	clientRepo ports.ClientRepository
	hmacSecret string
}

func NewKeyService(kr ports.KeyRepository, cr ports.ClientRepository, secret string) *KeyService {
	return &KeyService{
		keyRepo:    kr,
		clientRepo: cr,
		hmacSecret: secret,
	}
}

func (s *KeyService) CreateKey(ctx context.Context, clientID, prefix string, ttl time.Duration) (*domain.PlainKeyResponse, error) {
	if _, err := s.clientRepo.GetByID(ctx, clientID); err != nil {
		return nil, ErrClientNotFound
	}

	randomPart, err := security.GenerateRandomPart()
	if err != nil {
		return nil, err
	}

	baseStr := fmt.Sprintf("%s_%s", prefix, randomPart)
	checksum := security.CalculateChecksum(baseStr)
	plainKey := fmt.Sprintf("%s_%s", baseStr, checksum)

	hash := security.HashKey(plainKey, s.hmacSecret)

	now := time.Now().UTC()

	apiKey := &domain.APIKey{
		ID:        ksuid.New().String(),
		ClientID:  clientID,
		Prefix:    prefix,
		Hash:      hash,
		Status:    domain.StatusActive,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.keyRepo.Create(ctx, apiKey); err != nil {
		return nil, err
	}

	return &domain.PlainKeyResponse{
		ID:        apiKey.ID,
		PlainKey:  plainKey,
		ExpiresAt: apiKey.ExpiresAt,
	}, nil
}

func (s *KeyService) ValidateChecksum(plainKey string) bool {
	parts := strings.Split(plainKey, "_")
	if len(parts) < 3 {
		return false
	}

	providedChecksum := parts[len(parts)-1]
	baseStr := strings.Join(parts[:len(parts)-1], "_")

	expectedChecksum := security.CalculateChecksum(baseStr)
	return providedChecksum == expectedChecksum
}

func (s *KeyService) RevokeKey(ctx context.Context, keyID string) error {
	key, err := s.keyRepo.GetByID(ctx, keyID)
	if err != nil {
		return ErrKeyNotFound
	}

	if key.Status == domain.StatusRevoked {
		return nil
	}

	return s.keyRepo.UpdateStatus(ctx, keyID, domain.StatusRevoked)
}

