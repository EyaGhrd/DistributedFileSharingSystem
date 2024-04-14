package p2pnet

import (
	"bufio"
	"fmt"
	"github.com/libp2p/go-libp2p/core/network"
	"io"
)

func HandleStream(stream network.Stream) {
	fmt.Println("Got a new stream!")

	// Creating a buffer stream for non-blocking read and write.

	go readDataFromStream(stream)
	go writeDataToStream(stream)
}

func readDataFromStream(stream network.Stream) {
	reader := bufio.NewReader(stream)

	// Read data from the stream
	data, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println("End of stream reached")
		} else {
			fmt.Println("Error reading from stream:", err)
		}
		return
	}

	// Print the received data
	fmt.Println("Received data:", data)
}

func writeDataToStream(stream network.Stream) {
	writer := bufio.NewWriter(stream)

	// Write data to the stream
	data := "Hello, world!\n"
	_, err := writer.WriteString(data)
	if err != nil {
		fmt.Println("Error writing to stream:", err)
		return
	}

	// Flush the writer to ensure data is sent
	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flushing writer:", err)
		return
	}
}
