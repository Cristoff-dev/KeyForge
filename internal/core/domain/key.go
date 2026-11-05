package domain

import (
	"time"
)

type KeyStatus string

const (
	StatusActive  KeyStatus = "active"
	StatusRevoked KeyStatus = "revoked"
	StatusExpired KeyStatus = "expired"
)

type APIKey struct {
	ID        string
	ClientID  string
	Prefix    string
	Hash      string
	Status    KeyStatus
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Client struct {
	ID           string
	Name         string
	RateLimitReq int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PlainKeyResponse struct {
	ID        string    `json:"id"`
	PlainKey  string    `json:"plain_key"`
	ExpiresAt time.Time `json:"expires_at"`
}
