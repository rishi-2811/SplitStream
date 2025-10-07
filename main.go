package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"peerVault/client"
	"peerVault/discovery"
	"peerVault/server"
	"peerVault/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expected 'serve', 'upload', or 'download' subcommands.")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		port := serveCmd.Int("port", 8080, "Port to listen on for file transfers")
		bootstrap := serveCmd.String("bootstrap", "", "Bootstrap peer multiaddr")
		serveCmd.Parse(os.Args[2:])

		// Initialize storage
		store := storage.NewStorage("peerVault_storage")

		// Start libp2p discovery node
		dnode, err := discovery.SetupNode(*bootstrap)
		if err != nil {
			fmt.Println("Discovery setup error:", err)
			os.Exit(1)
		}
		fmt.Println("Libp2p Peer listening on:", dnode.Host.Addrs())
		fmt.Println("Libp2p Peer ID:", dnode.Host.ID())


		// Start TCP file server
		s := server.NewServer(*port, "tcp", store, dnode)
		if err := s.ListenAndServe(); err != nil {
			fmt.Println("Server error:", err)
		}

	case "upload":
		uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
		file := uploadCmd.String("file", "", "File to upload locally and announce")
		bootstrap := uploadCmd.String("bootstrap", "", "Bootstrap peer multiaddr")
		uploadCmd.Parse(os.Args[2:])

		if *file == "" {
			fmt.Println("Usage: upload --file <file> --bootstrap <addr>")
			os.Exit(1)
		}

		// Store file locally
		store := storage.NewStorage("peerVault_storage")
		// FIXED: Changed to use the new SaveFileFromPath function
		hash, err := store.SaveFileFromPath(*file)
		if err != nil {
			fmt.Println("Save error:", err)
			os.Exit(1)
		}
		fmt.Println("File stored locally with hash:", hash)

		// Announce file on DHT
		dnode, err := discovery.SetupNode(*bootstrap)
		if err != nil {
			fmt.Println("Discovery setup error:", err)
			os.Exit(1)
		}
		if err := discovery.Announce(dnode, hash); err != nil {
			fmt.Println("Announce error:", err)
		} else {
			fmt.Println("File hash announced on DHT:", hash)
		}

	case "download":
		getCmd := flag.NewFlagSet("download", flag.ExitOnError)
		hash := getCmd.String("filehash", "", "Hash of file to retrieve")
		bootstrap := getCmd.String("bootstrap", "", "Bootstrap peer multiaddr")
		output := getCmd.String("out", "downloaded_file", "Output filename")
		getCmd.Parse(os.Args[2:])

		if *hash == "" {
			fmt.Println("Usage: download --filehash <hash> --bootstrap <addr>")
			os.Exit(1)
		}

		// Connect to DHT
		dnode, err := discovery.SetupNode(*bootstrap)
		if err != nil {
			fmt.Println("Discovery setup error:", err)
			os.Exit(1)
		}

		// Lookup peers with the file
		peers, err := discovery.FindProviders(dnode, *hash)
		if err != nil {
			fmt.Println("FindProviders error:", err)
			os.Exit(1)
		}
		if len(peers) == 0 {
			fmt.Println("No peers found for file:", *hash)
			os.Exit(1)
		}

		fmt.Printf("Found %d peer(s) with the file.\n", len(peers))
		// Pick the first provider
		provider := peers[0]
		
		// FIXED: Correctly parse the multiaddr to get a TCP address for net.Dial
		var tcpAddr string
		for _, addr := range provider.Addrs {
			if strings.Contains(addr.String(), "/tcp/") {
				// Example: /ip4/127.0.0.1/tcp/8080 -> 127.0.0.1:8080
				parts := strings.Split(addr.String(), "/")
				tcpAddr = fmt.Sprintf("%s:%s", parts[2], parts[4])
				break
			}
		}
		
		if tcpAddr == "" {
			fmt.Println("Could not find a TCP address for peer:", provider.ID)
			os.Exit(1)
		}
		
		fmt.Println("Attempting to download from peer at:", tcpAddr)

		// Fetch via TCP
		// FIXED: Changed to call the new client.Download function
		if err := client.Download(*hash, tcpAddr, *output); err != nil {
			fmt.Println("Download error:", err)
			os.Exit(1)
		}
		fmt.Println("File downloaded successfully:", *output)

	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}