# SplitStream

A peer-to-peer file sharing system built with Go and libp2p.

## Features

- Decentralized peer-to-peer file sharing
- Content-addressed file storage
- DHT-based peer discovery
- Direct TCP file transfers
- Simple command-line interface

## Installation

```bash
go get github.com/rishi-2811/splitstream
```

## Usage

### Start a Server Node

```bash
./splitstream serve --port 8080 --bootstrap /ip4/1.2.3.4/tcp/4001/p2p/QmHash
```

### Upload a File

```bash
./splitstream upload --file myfile.txt --bootstrap /ip4/1.2.3.4/tcp/4001/p2p/QmHash
```

### Download a File

```bash
./splitstream download --filehash QmFileHash --out downloaded.txt --bootstrap /ip4/1.2.3.4/tcp/4001/p2p/QmHash
```

## Project Structure

- `main.go` - Main entry point and command handling
- `server/` - TCP server implementation for file transfers
- `client/` - File download client implementation
- `discovery/` - Peer discovery using libp2p DHT
- `storage/` - Local file storage management
- `background_services/` - Background services like RTT estimation
- `encryption/` - Encryption utilities

## Dependencies

- libp2p - For peer discovery and DHT
- Go standard library - For core functionality