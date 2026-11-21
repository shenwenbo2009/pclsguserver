package upload

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func computeExpectedChecksum(t *testing.T, data []byte) string {
	t.Helper()
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
