package main

import (
	"context"
	"fmt"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	// create a background context (i.e. one that never cancels)
	ctx := context.Background()

	// start a libp2p node that listens on a random local TCP port,
	// but without running the built-in ping protocol
	limits := rcmgr.InfiniteLimits
	rmgr, _ := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(limits))
	node, err := libp2p.New(
		libp2p.ResourceManager(rmgr), // 禁用资源管理器
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0", "/ip4/0.0.0.0/tcp/0/ws"),
		libp2p.Ping(true),
	)
	if err != nil {
		panic(err)
	}

	// configure our own ping protocol
	pingService := &ping.PingService{Host: node}
	node.SetStreamHandler(ping.ID, pingService.PingHandler)

	// print the node's PeerInfo in multiaddr format
	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	fmt.Println("libp2p node address:", addrs)

	// if a remote peer has been passed on the command line, connect to it
	// and start ping it
	// otherwise wait for a signal to stop
	if len(os.Args) > 1 {
		go pingPeer(ctx, node, pingService, os.Args[1])
	}

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

func pingPeer(ctx context.Context, node host.Host, pingService *ping.PingService, peerAddress string) {
	ticker := time.NewTicker(2 * time.Second) // 每2秒触发一次
	defer ticker.Stop()                       // 程序结束时停止 Ticker

	addr, err := multiaddr.NewMultiaddr(peerAddress)
	if err != nil {
		panic(err)
	}
	peer, err := peerstore.AddrInfoFromP2pAddr(addr)
	if err != nil {
		panic(err)
	}
	if err := node.Connect(ctx, *peer); err != nil {
		panic(err)
	}
	fmt.Println("pinging peer at", addr)
	for range ticker.C {
		ch := pingService.Ping(ctx, peer.ID)
		res := <-ch
		fmt.Println("ping result:", res)
	}
}
