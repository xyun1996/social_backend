package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Token returns a random hex token with the requested byte length.
func Token(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return hex.EncodeToString(buf), nil
}
