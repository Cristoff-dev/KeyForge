package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Cristoff-dev/keyforge/internal/api/middlewares"
	"github.com/go-chi/chi/v5"
)

func TestAuthMiddleware_UnauthorizedCases(t *testing.T) {
	hmacSecret := "super-secret-keyforge-hmac-key"

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Missing Authorization Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Header Format Without Bearer",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Empty Bearer Token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Malformed Token Format",
			authHeader:     "Bearer invalid_plain_text_without_prefix",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := chi.NewRouter()
			// Pasamos nil en repositorios ya que la validación falla en la inspección de la cabecera
			r.Use(middlewares.Auth(nil, nil, hmacSecret))
			r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status code %d, got %d", tc.expectedStatus, rec.Code)
			}
		})
	}
}
