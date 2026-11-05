package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Cristoff-dev/keyforge/internal/core/services"
	"github.com/go-chi/chi/v5"
)

type KeyHandler struct {
	keyService *services.KeyService
}

func NewKeyHandler(ks *services.KeyService) *KeyHandler {
	return &KeyHandler{
		keyService: ks,
	}
}

type CreateKeyRequest struct {
	ClientID string `json:"client_id"`
	Prefix   string `json:"prefix"`
	TTLHours int    `json:"ttl_hours"`
}

func (h *KeyHandler) HandleCreateKey(w http.ResponseWriter, r *http.Request) {
	var req CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	if req.ClientID == "" || req.Prefix == "" || req.TTLHours <= 0 {
		http.Error(w, "missing or invalid required fields", http.StatusBadRequest)
		return
	}

	ttl := time.Duration(req.TTLHours) * time.Hour
	resp, err := h.keyService.CreateKey(r.Context(), req.ClientID, req.Prefix, ttl)
	if err != nil {
		if err == services.ErrClientNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *KeyHandler) HandleRevokeKey(w http.ResponseWriter, r *http.Request) {
	keyID := chi.URLParam(r, "id")
	if keyID == "" {
		http.Error(w, "missing key id", http.StatusBadRequest)
		return
	}

	err := h.keyService.RevokeKey(r.Context(), keyID)
	if err != nil {
		if err == services.ErrKeyNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

