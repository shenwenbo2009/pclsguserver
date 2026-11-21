package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func ComputeSHA256(reader io.Reader) (string, int64, error) {
	h := sha256.New()
	n, err := io.Copy(h, reader)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func ValidateSHA256(path string, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sum, _, err := ComputeSHA256(f)
	if err != nil {
		return err
	}
	if sum != expected {
		return ErrChecksumMismatch{Expected: expected, Actual: sum}
	}
	return nil
}

type ErrChecksumMismatch struct {
	Expected string
	Actual   string
}

func (e ErrChecksumMismatch) Error() string {
	return "checksum mismatch: expected " + e.Expected + " got " + e.Actual
}
