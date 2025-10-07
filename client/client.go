package client

import (
	"fmt"
	"io"
	"net"
	"os"
	"peerVault/encryption"
	"strings"
)

func UploadFile(filePath string, addr string) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		return
	}
	defer f.Close()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	defer conn.Close()

	filenameSlices := strings.Split(filePath, ".")
	fileType := filenameSlices[len(filenameSlices)-1]

	conn.Write([]byte("UPLOAD\n"))
	conn.Write([]byte(fileType + "\n"))

	n, err := io.Copy(conn, f)
	if err != nil {
		fmt.Println("Failed to upload file:", err)
		return
	}

	filehash, err := encryption.HashFile(filePath)
	if err != nil {
		fmt.Println("Failed to hash file:", err)
		return
	}

	fmt.Printf("File uploaded, %d bytes sent. File hash: %s\n", n, filehash)
}

// FIXED: Added the missing Download function
func Download(hash string, addr string, outputPath string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to provider: %w", err)
	}
	defer conn.Close()

	// Send download command and hash to the server
	if _, err := conn.Write([]byte("DOWNLOAD\n")); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}
	if _, err := conn.Write([]byte(hash + "\n")); err != nil {
		return fmt.Errorf("failed to send hash: %w", err)
	}

	// Create the output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Copy the data from the connection to the file
	_, err = io.Copy(outFile, conn)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}