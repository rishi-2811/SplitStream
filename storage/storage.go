package storage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"peerVault/encryption"
	"sync"
)

// Storage represents a simple file storage system.
type Storage struct {
	mu       sync.Mutex
	basePath string
}

// NewStorage creates a new Storage instance with the specified base path.
func NewStorage(basePath string) *Storage {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create base path: %v", err))
	}
	return &Storage{basePath: basePath}
}

// FIXED: Added missing SaveFileFromPath function for the 'upload' command
func (s *Storage) SaveFileFromPath(filePath string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file from path: %w", err)
	}
	
	hash := encryption.HashData(data)
	
	fileExt := filepath.Ext(filePath)
	newFilePath := filepath.Join(s.basePath, hash+fileExt)

	err = os.WriteFile(newFilePath, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save file to storage: %w", err)
	}

	return hash, nil
}


// SaveFile saves data from a reader to a file in the storage.
// The filename is the hash of the data.
// FIXED: Changed function to return the file hash and an error.
func (s *Storage) SaveFile(reader *bufio.Reader, filetype string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a temporary file to store the incoming data
	tempFile, err := os.CreateTemp(s.basePath, "upload-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempFileName := tempFile.Name()
	defer os.Remove(tempFileName) // Clean up the temp file

	// Copy data from the connection to the temp file
	if _, err := io.Copy(tempFile, reader); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("failed to save file content: %w", err)
	}
	tempFile.Close() // Close file so we can hash it

	// Hash the content of the temporary file
	filehash, err := encryption.HashFile(tempFileName)
	if err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	// Create the final path and rename the file
	newFilePath := filepath.Join(s.basePath, filehash) + "." + filetype
	if err = os.Rename(tempFileName, newFilePath); err != nil {
		return "", fmt.Errorf("failed to rename file: %w", err)
	}

	fmt.Printf("File saved successfully to path: %s\n", newFilePath)
	return filehash, nil
}

// ReadFile reads the content of a file from the storage.
// The filename is the hashed name of the file.
func (s *Storage) ReadFile(hash string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// FIXED: Find the file by hash prefix, ignoring the extension.
	// This is more robust than assuming ".txt".
	pattern := filepath.Join(s.basePath, hash+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("error searching for file: %w", err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("file with hash %s not found", hash)
	}

	filePath := matches[0] // Use the first match
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}