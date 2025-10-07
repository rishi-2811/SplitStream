package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"peerVault/discovery"
	"peerVault/storage"
	
	"github.com/ipfs/go-cid"
	
	mh "github.com/multiformats/go-multihash"
)

type Server struct {
	port     int
	protocol string
	Listener net.Listener
	Storage  *storage.Storage
	Disco    *discovery.DNode
}

// Updated constructor to accept disco
func NewServer(port int, protocol string, store *storage.Storage, disco *discovery.DNode) *Server {
	return &Server{
		port:     port,
		protocol: protocol,
		Storage:  store,
		Disco:    disco,
	}
}

func (s *Server) ListenAndServe() error {
	listener, err := net.Listen(s.protocol, fmt.Sprintf(":%d", s.port))

	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer listener.Close()

	s.Listener = listener
	fmt.Printf("File transfer server is listening on %s:%d...\n", s.protocol, s.port)

	s.StartAcceptConnectionsLoop()

	return nil
}

func (s *Server) StartAcceptConnectionsLoop() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	fmt.Println("new conn from", remote)

	reader := bufio.NewReader(conn)
	cmdLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("read cmd err:", err)
		return
	}
	cmd := strings.TrimSpace(cmdLine)
	switch cmd {
	case "UPLOAD":
		fileTypeLine, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("read filetype err:", err)
			return
		}
		fileType := strings.TrimSpace(fileTypeLine)
		
		// FIXED: Capture the hash returned by SaveFile
		hash, err := s.Storage.SaveFile(reader, fileType)
		if err != nil {
			fmt.Println("save file err:", err)
			return
		}
		fmt.Printf("Saved file with hash: %s\n", hash)

		// announce on DHT if discovery available
		if s.Disco != nil && s.Disco.DHT != nil {
			// FIXED: Use the actual hash from the saved file
			mhBytes, _ := mh.Encode([]byte(hash), mh.SHA2_256) // Use SHA2_256 to match hashing util
			c := cid.NewCidV1(cid.Raw, mhBytes)
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if err := s.Disco.DHT.Provide(ctx, c, true); err != nil {
				fmt.Println("dht provide error:", err)
			} else {
				fmt.Println("Announced provider on DHT for", c.String())
			}
		}
	case "DOWNLOAD":
		hashLine, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("read hash err:", err)
			return
		}
		hash := strings.TrimSpace(hashLine)
		fmt.Println("Download requested for", hash)
		data, err := s.Storage.ReadFile(hash)
		if err != nil {
			fmt.Println("read file err:", err)
			// Inform the client that the file was not found
			conn.Write([]byte("ERROR: File not found\n"))
			return
		}
		if _, err := conn.Write(data); err != nil {
			fmt.Println("write file err:", err)
			return
		}
		fmt.Println("Sent file to", remote)
	default:
		fmt.Println("unknown cmd:", cmd)
	}
}

// FIXED: Removed unused and redundant SendFileToClient function

func (s *Server) Close() error {
	if s.Listener != nil {
		if err := s.Listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
		fmt.Println("Server closed successfully")
	}
	return nil
}