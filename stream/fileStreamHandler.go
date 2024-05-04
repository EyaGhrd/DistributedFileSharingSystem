package stream


import (
	"fmt"
	"github.com/libp2p/go-libp2p/core/network"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)


var path string = "/home/eya/FileSharingNetwork/log/"
var filename = ""
var fileContent []byte


func SendToStream(str network.Stream) {
	defer str.Close() // Close the stream after sending the file
	fmt.Println("name of file recieved",filename)
	fmt.Println("Sending file to", str.Conn().RemotePeer())
	filename_send := path + filename
	file, err := os.Open(filename_send)
	if err != nil {
		fmt.Println("Error while opening the sending file:", err)
		return
	}
	defer file.Close()


	// Create a buffer to read file content
	buffer := make([]byte, 1024)


	// Loop until the end of file (EOF) is reached
	for {
		// Read from the file into the buffer
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("Error while reading from the file:", err)
			return
		}
		if bytesRead == 0 {
			break
		}


		// Write the buffer data to the stream
		_, err = str.Write(buffer[:bytesRead])
		if err != nil {
			fmt.Println("Error while sending to the stream:", err)
			return
		}
	}
	fmt.Println("File sent successfully")
}


func ReceivedFromStream(str network.Stream, filename string, filetype string, size int) {
	defer str.Close() // Close the stream after receiving the file


	fmt.Println("Receiving file from", str.Conn().RemotePeer())
	file, err := os.Create(filename  )
	if err != nil {
		fmt.Println("Error while creating the receiving file:", err)
		return
	}
	defer file.Close()


	// Create a buffer to receive file content
	buffer := make([]byte, 1024)


	// Loop until the end of file (EOF) is reached
	for {
		// Read from the stream into the buffer
		bytesRead, err := str.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("Error while reading from stream:", err)
			return
		}
		if bytesRead == 0 {
			// End of file reached
			break
		}
		//fileContent = buffer[]
		// Write the buffer data to the file
		_, err = file.Write(buffer[:bytesRead])


		if err != nil {
			fmt.Println("Error while writing to the file:", err)
			return
		}
	}


	fmt.Println("Completed reading from the stream")
}


func getByteSize(filename string) (int, error) {
	file, err := os.Stat(filename)
	if err != nil {
		fmt.Println("Error while walking the file - file doesn't exist")
		return 0, err
	}
	return int(file.Size()), nil
}
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*") // Caution, setting "*" is insecure
}

func HandleFileRequest(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not supported", http.StatusNotFound)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	}

	filename = string(body)
	fmt.Println("name of file recieved",filename)
	log.Printf("File requested: %s\n", filename)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Received file name: " + filename))
}
