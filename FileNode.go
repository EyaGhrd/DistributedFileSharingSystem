package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
)

func main() {
	listenAddr := "/ip4/0.0.0.0/tcp/9090"

	host, err := libp2p.New(
		libp2p.ListenAddrStrings(listenAddr),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Host ID: %s\n", host.ID())
	fmt.Printf("Listen addresses: %v\n", host.Addrs())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
