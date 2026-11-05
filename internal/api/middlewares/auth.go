package middlewares

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Cristoff-dev/keyforge/internal/core/domain"
	"github.com/Cristoff-dev/keyforge/internal/core/ports"
	"github.com/Cristoff-dev/keyforge/internal/security"
)

type contextKey string

const ClientIDKey contextKey = "clientID"

func Auth(cache ports.CacheRepository, keyRepo ports.KeyRepository, secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			plainKey := strings.TrimPrefix(authHeader, "Bearer ")
			parts := strings.Split(plainKey, "_")
			if len(parts) < 3 {
				http.Error(w, "invalid format", http.StatusUnauthorized)
				return
			}

			baseStr := strings.Join(parts[:len(parts)-1], "_")
			if security.CalculateChecksum(baseStr) != parts[len(parts)-1] {
				http.Error(w, "invalid checksum", http.StatusUnauthorized)
				return
			}

			var clientID string
			cachedClientID, err := cache.GetHashValidation(r.Context(), plainKey)

			if err == nil && cachedClientID != "" {
				clientID = cachedClientID
			} else {
				hash := security.HashKey(plainKey, secret)
				key, err := keyRepo.GetByHash(r.Context(), hash)
				if err != nil || key.Status != domain.StatusActive || key.ExpiresAt.Before(time.Now().UTC()) {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				clientID = key.ClientID
				_ = cache.SetHashValidation(r.Context(), plainKey, clientID, 5*time.Minute)
			}

			ctx := context.WithValue(r.Context(), ClientIDKey, clientID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

