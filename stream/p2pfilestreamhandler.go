package stream

import (
	"context"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"os"
)

func HandleIncomingStreams(ctx context.Context, host host.Host, logfile *os.File) {
	// Set up stream handler
	host.SetStreamHandler("/file/1.1.0", func(stream network.Stream) {
		defer stream.Close()

		// Extract filename and other information from stream metadata if needed
		filename := "peerconnlog"
		filetype := "txt"
		filesize := 1024 // example file size

		// Call function to receive file data
		ReceivedFromStream(stream, filename, filetype, logfile, filesize)
	})
}
