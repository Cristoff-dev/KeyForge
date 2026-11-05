package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Cristoff-dev/keyforge/internal/storage/postgres"
	storageRedis "github.com/Cristoff-dev/keyforge/internal/storage/redis"
)

func RateLimiter(cache *storageRedis.Cache, clientRepo *postgres.ClientRepository, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID, ok := r.Context().Value(ClientIDKey).(string)
			if !ok || clientID == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			limitKey := "client_limit:" + clientID
			quotaStr, err := cache.Get(r.Context(), limitKey)

			var limit int
			if err == nil && quotaStr != "" {
				fmt.Sscanf(quotaStr, "%d", &limit)
			} else {
				clientData, err := clientRepo.GetByID(r.Context(), clientID)
				if err != nil {
					http.Error(w, "client not found", http.StatusUnauthorized)
					return
				}
				limit = clientData.RateLimitReq

				_ = cache.Set(r.Context(), limitKey, limit, 5*time.Minute)
			}

			key := "ratelimit:" + clientID
			allowed, err := cache.AllowRequest(r.Context(), key, limit, window)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

