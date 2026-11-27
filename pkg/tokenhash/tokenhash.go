package tokenhash

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// Mint returns a raw token (shown once), a short prefix, and the SHA-256 hex hash to store.
func Mint() (raw, prefix, hash string, err error) {
	buf := make([]byte, 24)
	if _, err = rand.Read(buf); err != nil {
		return "", "", "", err
	}
	raw = "rd_" + hex.EncodeToString(buf)
	prefix = raw[:10]
	sum := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(sum[:])
	return raw, prefix, hash, nil
}

func Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
