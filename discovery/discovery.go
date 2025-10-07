package discovery

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

type DNode struct {
	Host host.Host
	DHT  *dht.IpfsDHT
	Ctx  context.Context
}

func SetupNode(bootstrap string) (*DNode, error) {
	ctx := context.Background()

	// Create libp2p host
	h, err := libp2p.New()
	if err != nil {
		return nil, err
	}

	// Create Kademlia DHT
	kdht, err := dht.New(ctx, h)
	if err != nil {
		return nil, err
	}

	// Bootstrap connection
	if bootstrap != "" {
		maddr, err := multiaddr.NewMultiaddr(bootstrap)
		if err != nil {
			return nil, err
		}
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, err
		}
		if err := h.Connect(ctx, *info); err != nil {
			return nil, err
		}
		fmt.Println("Connected to bootstrap:", *info)
	}

	if err := kdht.Bootstrap(ctx); err != nil {
		return nil, err
	}

	return &DNode{Host: h, DHT: kdht, Ctx: ctx}, nil
}

// Announce that we provide this file
func Announce(d *DNode, filehash string) error {
	// FIXED: Convert the filehash string into a multihash, then into a CID.
	mh, err := multihash.Sum([]byte(filehash), multihash.SHA2_256, -1)
	if err != nil {
		return err
	}
	c := cid.NewCidV1(cid.Raw, mh)

	// The DHT.Provide method requires a CID as the key.
	return d.DHT.Provide(d.Ctx, c, true)
}

// Find providers for a file
func FindProviders(d *DNode, filehash string) ([]peer.AddrInfo, error) {
	// Convert the filehash string into a multihash, then into a CID.
	mh, err := multihash.Sum([]byte(filehash), multihash.SHA2_256, -1)
	if err != nil {
		return nil, err
	}
	c := cid.NewCidV1(cid.Raw, mh)

	// Use FindProvidersAsync, which returns only a channel.
	provsCh := d.DHT.FindProvidersAsync(d.Ctx, c, 10) // 10 is the max number of providers to find

	var providers []peer.AddrInfo
	for p := range provsCh {
		providers = append(providers, p)
	}

	return providers, nil
}