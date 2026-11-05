package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Cristoff-dev/keyforge/internal/core/domain"
	"github.com/Cristoff-dev/keyforge/internal/storage/postgres"
)

func TestClientRepository_GetByID(t *testing.T) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://keyforge:forgepass@localhost:5432/keyforge_db?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := postgres.NewDBPool(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping PostgreSQL integration test (database unavailable): %v", err)
	}
	defer pool.Close()

	repo := postgres.NewClientRepository(pool)

	seedClient := &domain.Client{
		ID:           "client_01",
		Name:         "Test Client",
		RateLimitReq: 100,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_ = repo.Create(ctx, seedClient)

	t.Run("Get Existing Client", func(t *testing.T) {
		client, err := repo.GetByID(ctx, "client_01")
		if err != nil {
			t.Fatalf("unexpected error fetching client: %v", err)
		}

		if client.ID != "client_01" {
			t.Errorf("expected client_id 'client_01', got '%s'", client.ID)
		}

		if client.RateLimitReq <= 0 {
			t.Errorf("expected rate_limit_req > 0, got %d", client.RateLimitReq)
		}
	})

	t.Run("Get Non-Existing Client", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "non_existent_client_id_999")
		if err == nil {
			t.Error("expected error for non-existent client, got nil")
		}
	})
}
