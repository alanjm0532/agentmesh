package util

import (
	"crypto/rand"
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
