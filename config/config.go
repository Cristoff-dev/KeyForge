package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Cristoff-dev/keyforge/internal/api/handlers"
	"github.com/Cristoff-dev/keyforge/internal/api/middlewares"
	"github.com/Cristoff-dev/keyforge/internal/core/services"
	"github.com/Cristoff-dev/keyforge/internal/storage/postgres"
	storageRedis "github.com/Cristoff-dev/keyforge/internal/storage/redis"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	dbURL := getEnv("DB_URL", "postgres://keyforge:forgepass@localhost:5432/keyforge_db?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "localhost:6379")
	hmacSecret := getEnv("HMAC_SECRET", "super-secret-keyforge-hmac-key")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbPool, err := postgres.NewDBPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer dbPool.Close()

	redisAddr := redisURL
	if strings.HasPrefix(redisURL, "redis://") {
		opts, err := redis.ParseURL(redisURL)
		if err == nil {
			redisAddr = opts.Addr
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer rdb.Close()

	keyRepo := postgres.NewKeyRepository(dbPool)
	clientRepo := postgres.NewClientRepository(dbPool)
	cacheRepo := storageRedis.NewCache(rdb)

	keyService := services.NewKeyService(keyRepo, clientRepo, hmacSecret)
	keyHandler := handlers.NewKeyHandler(keyService)

	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/admin/keys", func(r chi.Router) {
			r.Post("/", keyHandler.HandleCreateKey)
			r.Delete("/{id}", keyHandler.HandleRevokeKey)
		})

		r.Group(func(r chi.Router) {
			r.Use(middlewares.Auth(cacheRepo, keyRepo, hmacSecret))
			r.Use(middlewares.RateLimiter(cacheRepo, clientRepo, time.Minute))

			r.Get("/resource", func(w http.ResponseWriter, r *http.Request) {
				clientID, _ := r.Context().Value(middlewares.ClientIDKey).(string)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"message":"access granted","client_id":"` + clientID + `"}`))
			})
		})
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("KeyForge API listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("KeyForge API stopped clean")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

