package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateToken() (string, error) {
	// Keep 256 bits of randomness after the application prefix.
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", err
	}
	return "agentmesh_" + hex.EncodeToString(token[:]), nil
}

// HashToken hashes high-entropy credentials before persistence.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
