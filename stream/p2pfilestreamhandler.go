package stream


import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)


func HandleIncomingStreams(ctx context.Context, host host.Host) {
	fmt.Println("detect incoming stream")
	host.SetStreamHandler("/file/1.1.0", func(stream network.Stream) {
		defer stream.Close()

		filename := filename
		filetype := "txt"
		filesize := 2048
		ReceivedFromStream(stream, filename, filetype, filesize)
	})
}

