package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestGenerateRandomPart(t *testing.T) {
	part1, err := GenerateRandomPart()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	part2, err := GenerateRandomPart()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if part1 == part2 {
		t.Error("expected generated parts to be unique")
	}
}

func TestCalculateChecksum(t *testing.T) {
	keyPart := "prod_live_abc123xyz"

	checksum1 := CalculateChecksum(keyPart)
	checksum2 := CalculateChecksum(keyPart)

	if len(checksum1) == 0 {
		t.Error("expected non-empty checksum")
	}

	if checksum1 != checksum2 {
		t.Error("expected deterministic checksum for identical input")
	}

	differentChecksum := CalculateChecksum("different_input_string")
	if checksum1 == differentChecksum {
		t.Error("expected different inputs to produce different checksums")
	}
}

func TestHashKey(t *testing.T) {
	secret := "test-secret-key"
	plainKey := "prod_live_test_api_key_12345"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(plainKey))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	actualHash := HashKey(plainKey, secret)

	if actualHash != expectedHash {
		t.Errorf("expected hash %s, got %s", expectedHash, actualHash)
	}
}

func TestVerifyKey(t *testing.T) {
	secret := "test-secret-key"
	plainKey := "prod_live_test_api_key_12345"
	expectedHash := HashKey(plainKey, secret)

	tests := []struct {
		name     string
		plainKey string
		secret   string
		hash     string
		expected bool
	}{
		{
			name:     "ValidCredentials",
			plainKey: plainKey,
			secret:   secret,
			hash:     expectedHash,
			expected: true,
		},
		{
			name:     "InvalidKey",
			plainKey: "prod_live_wrong_key",
			secret:   secret,
			hash:     expectedHash,
			expected: false,
		},
		{
			name:     "InvalidSecret",
			plainKey: plainKey,
			secret:   "wrong-secret",
			hash:     expectedHash,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := VerifyKey(tc.plainKey, tc.secret, tc.hash)
			if valid != tc.expected {
				t.Errorf("expected verification %v, got %v", tc.expected, valid)
			}
		})
	}
}
