package main

import (
	"fmt"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
)

func main() {
	// start a libp2p node that listens on a random local TCP port,
	// but without running the built-in ping protocol
	limits := rcmgr.InfiniteLimits
	rmgr, _ := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(limits))
	node, err := libp2p.New(
		libp2p.ResourceManager(rmgr),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0", "/ip4/0.0.0.0/tcp/0/ws"),
		libp2p.Ping(true),
	)
	if err != nil {
		panic(err)
	}

	// print the node's PeerInfo in multiaddr format
	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	fmt.Println("libp2p node address:", addrs)

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("Shutting down...")

	// shut the node down
	if err := node.Close(); err != nil {
		panic(err)
	}
}
