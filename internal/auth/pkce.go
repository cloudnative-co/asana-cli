package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func randomURLSafeString(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func NewState() (string, error) {
	return randomURLSafeString(18)
}

func NewCodeVerifier() (string, error) {
	return randomURLSafeString(32)
}

func NewCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
