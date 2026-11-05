package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"hash/crc32"
)

const (
	randomBytesSize = 32
)

var ErrInvalidKeyFormat = errors.New("invalid key format or checksum")

func GenerateRandomPart() (string, error) {
	b := make([]byte, randomBytesSize)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func CalculateChecksum(data string) string {
	crc := crc32.ChecksumIEEE([]byte(data))
	checksumBytes := make([]byte, 4)
	checksumBytes[0] = byte(crc >> 24)
	checksumBytes[1] = byte(crc >> 16)
	checksumBytes[2] = byte(crc >> 8)
	checksumBytes[3] = byte(crc)
	return hex.EncodeToString(checksumBytes)
}

func HashKey(plainKey, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(plainKey))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifyKey(plainKey, secret, expectedHash string) bool {
	actualHash := HashKey(plainKey, secret)
	return hmac.Equal([]byte(actualHash), []byte(expectedHash))
}
