package postgres

import (
	"context"
	"errors"

	"github.com/Cristoff-dev/keyforge/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClientRepository struct {
	db *pgxpool.Pool
}

func NewClientRepository(db *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{
		db: db,
	}
}

func (r *ClientRepository) Create(ctx context.Context, client *domain.Client) error {
	query := `
		INSERT INTO clients (id, name, rate_limit_req, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		client.ID,
		client.Name,
		client.RateLimitReq,
		client.CreatedAt,
		client.UpdatedAt,
	)
	return err
}

func (r *ClientRepository) GetByID(ctx context.Context, id string) (*domain.Client, error) {
	query := `
		SELECT id, name, rate_limit_req, created_at, updated_at
		FROM clients
		WHERE id = $1
	`
	var client domain.Client
	err := r.db.QueryRow(ctx, query, id).Scan(
		&client.ID,
		&client.Name,
		&client.RateLimitReq,
		&client.CreatedAt,
		&client.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("client not found")
	}
	if err != nil {
		return nil, err
	}
	return &client, nil
}
