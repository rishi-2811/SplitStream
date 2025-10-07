package encryption

import (
	"fmt"
	"crypto/sha256"
	"os"
)

func HashData(data []byte) string {
	hash :=sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

func HashFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}