package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Cristoff-dev/keyforge/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type KeyRepository struct {
	db *pgxpool.Pool
}

func NewKeyRepository(db *pgxpool.Pool) *KeyRepository {
	return &KeyRepository{
		db: db,
	}
}

func (r *KeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	query := `
		INSERT INTO api_keys (id, client_id, prefix, hash, status, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		key.ID,
		key.ClientID,
		key.Prefix,
		key.Hash,
		key.Status,
		key.ExpiresAt,
		key.CreatedAt,
		key.UpdatedAt,
	)
	return err
}

func (r *KeyRepository) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	query := `
		SELECT id, client_id, prefix, hash, status, expires_at, created_at, updated_at
		FROM api_keys
		WHERE id = $1
	`
	return r.scanRow(r.db.QueryRow(ctx, query, id))
}

func (r *KeyRepository) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	query := `
		SELECT id, client_id, prefix, hash, status, expires_at, created_at, updated_at
		FROM api_keys
		WHERE hash = $1
	`
	return r.scanRow(r.db.QueryRow(ctx, query, hash))
}

func (r *KeyRepository) UpdateStatus(ctx context.Context, id string, status domain.KeyStatus) error {
	query := `
		UPDATE api_keys
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, status, time.Now().UTC(), id)
	return err
}

func (r *KeyRepository) GetActiveKeysByClient(ctx context.Context, clientID string) ([]*domain.APIKey, error) {
	query := `
		SELECT id, client_id, prefix, hash, status, expires_at, created_at, updated_at
		FROM api_keys
		WHERE client_id = $1 AND status = 'active'
	`
	rows, err := r.db.Query(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		key, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (r *KeyRepository) scanRow(row pgx.Row) (*domain.APIKey, error) {
	var key domain.APIKey
	err := row.Scan(
		&key.ID,
		&key.ClientID,
		&key.Prefix,
		&key.Hash,
		&key.Status,
		&key.ExpiresAt,
		&key.CreatedAt,
		&key.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("key not found")
	}
	if err != nil {
		return nil, err
	}
	return &key, nil
}

