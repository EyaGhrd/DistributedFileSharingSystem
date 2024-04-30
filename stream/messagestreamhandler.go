package stream

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Mina218/FileSharingNetwork/fileshare"
	"github.com/Mina218/FileSharingNetwork/structure"
	"os"
	"strconv"
	"strings"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var pid string = "/file/1.1.0"
var IsBroadcaster bool = false
var isAlreadyRequested bool = false

type Chatmessage struct {
	Messagecontent string
	Messagefrom    peer.ID
	Authorname     string
}

type BroadcastMsg struct {
	MentorNode peer.ID
}

type Packet struct {
	Type         string
	InnerContent []byte
}

type BroadcastRely struct {
	To     peer.ID
	From   peer.ID
	status string
}

type filereceivereq struct {
	Filename string
	Type     string
	From     peer.ID
	Size     int
}

func (frq filereceivereq) newFileRecieveRequest(ctx context.Context, topic *pubsub.Topic) {
	frqbytes, err := json.Marshal(frq)
	if err != nil {
		fmt.Println("Error while marshalling the file recieve request")
	} else {
		pktbytes, err := json.Marshal(Packet{
			Type:         "frq",
			InnerContent: frqbytes,
		})
		if err != nil {
			fmt.Println("Error while marshalling the frq packet")
		} else {
			topic.Publish(ctx, pktbytes)
		}
	}

}

func composeMessage(msg string, host host.Host) *Chatmessage {
	return &Chatmessage{
		Messagecontent: msg,
		Messagefrom:    host.ID(),
		Authorname:     host.ID().String()[len(host.ID().String())-6:],
	}
}

func broadCastReply(ctx context.Context, host host.Host, topic *pubsub.Topic, brdpacket BroadcastMsg) {
	mentorPeerId := brdpacket.MentorNode
	replyPacket := BroadcastRely{
		To:     mentorPeerId,
		From:   host.ID(),
		status: "ready",
	}
	rplypacketbytes, err := json.Marshal(replyPacket)
	if err != nil {
		fmt.Println("Error while marhsalling the brd rply packet")
	} else {
		packet := Packet{
			Type:         "rpl",
			InnerContent: rplypacketbytes,
		}

		packetByte, err := json.Marshal(packet)
		if err != nil {
			fmt.Println("Error while marshalling rplypacket")
		} else {
			topic.Publish(ctx, packetByte)
		}
	}
}

func handleInputFromSubscription(ctx context.Context, host host.Host, sub *pubsub.Subscription, topic *pubsub.Topic) {
	inputPacket := &Packet{}
	fileList := make([]string, 0)
	for {
		inputMsg, err := sub.Next(ctx)
		if err != nil {
			fmt.Println("Error while getting message from subscription:", err)
			continue
		}
		err = json.Unmarshal(inputMsg.Data, inputPacket)
		if err != nil {
			fmt.Println("Error while unmarshalling the inputMsg from subscription:", err)
			continue
		}
		switch inputPacket.Type {
		case "filelist":
			// Unmarshal the file list
			err := json.Unmarshal(inputPacket.InnerContent, &fileList)
			if err != nil {
				fmt.Println("Error while unmarshalling file list packet:", err)
			}
			fmt.Println("Received file list from the network:")
			for i, file := range fileList {
				fmt.Printf("%d. %s\n", i+1, file)
			}
		case "msg":
			// Handle regular messages
			chatMsg := &Chatmessage{}
			err := json.Unmarshal(inputPacket.InnerContent, chatMsg)
			if err != nil {
				fmt.Println("Error while unmarshalling msg packet:", err)
				continue
			}
			fmt.Printf("[%s] %s\n", chatMsg.Authorname, chatMsg.Messagecontent)
		case "fileselect":
			// Handle file selection
			var selectedFileIndex int
			err := json.Unmarshal(inputPacket.InnerContent, &selectedFileIndex)
			if err != nil {
				fmt.Println("Error while unmarshalling file select packet:", err)
				continue
			}
			if selectedFileIndex < 1 || selectedFileIndex > len(fileList) {
				fmt.Println("Invalid file index")
				continue
			}
			selectedFile := fileList[selectedFileIndex-1]
			fmt.Printf("Selected file: %s\n", selectedFile)
			// Now you can proceed to request this file from the network
		}
	}
}

func requestFile(ctx context.Context, host host.Host, filename string, filetype string, size int, mentr peer.ID, proto protocol.ID) {
	file := fileshare.OpenFileStatusLog()
	time.Sleep(30 * time.Second)
	stream, err := host.NewStream(ctx, mentr, proto)
	if err != nil {
		fmt.Fprintln(file, "Error while handling file request stream")
	}
	ReceivedFromStream(stream, filename, filetype, size)
}

func writeToSubscription(ctx context.Context, host host.Host, pubSubTopic *pubsub.Topic) {
	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			messg, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error while reading from standard input:", err)
				continue
			}
			messg = strings.TrimSpace(messg)
			if messg == "list" {
				SendFileListToAllNodes(ctx, host, pubSubTopic)
			} else if strings.HasPrefix(messg, "select ") {
				parts := strings.Split(messg, " ")
				if len(parts) != 2 {
					fmt.Println("Invalid command. Usage: select <file index>")
					continue
				}
				fileIndexStr := parts[1]
				fileIndex, err := strconv.Atoi(fileIndexStr)
				if err != nil {
					fmt.Println("Invalid file index")
					continue
				}
				// Marshal the file selection packet
				fileSelectPacket, err := json.Marshal(fileIndex)
				if err != nil {
					fmt.Println("Error marshalling file selection packet:", err)
					continue
				}
				// Publish the file selection packet to the topic
				pktMsg, err := json.Marshal(Packet{
					Type:         "fileselect",
					InnerContent: fileSelectPacket,
				})
				if err != nil {
					fmt.Println("Error marshalling file select packet:", err)
					continue
				}
				pubSubTopic.Publish(ctx, pktMsg)
			} else {
				// Send regular message
				chatMsg := composeMessage(messg, host)
				inputCnt, err := json.Marshal(*chatMsg)
				if err != nil {
					fmt.Println("Error marshaling the chat message:", err)
					continue
				}
				pktMsg, err := json.Marshal(Packet{
					Type:         "msg",
					InnerContent: inputCnt,
				})
				if err != nil {
					fmt.Println("Error marshalling the packet:", err)
					continue
				}
				pubSubTopic.Publish(ctx, pktMsg)
			}
		}
	}
}

// f
func HandlePubSubMessages(ctx context.Context, host host.Host, sub *pubsub.Subscription, top *pubsub.Topic) {
	go handleInputFromSubscription(ctx, host, sub, top)
	writeToSubscription(ctx, host, top)
}
func SendFileListToAllNodes(ctx context.Context, host host.Host, pubSubTopic *pubsub.Topic) {
	fileList, err := structure.GetFileListYouHave("/home/amina/Desktop/FileSharingNetwork") // Change this to the path where your files are stored
	if err != nil {
		fmt.Println("Error getting file list:", err)
		return
	}
	// Marshal the file list
	fileListJSON, err := json.Marshal(fileList)
	if err != nil {
		fmt.Println("Error marshalling file list:", err)
		return
	}
	// Publish the file list to the topic
	pktMsg, err := json.Marshal(Packet{
		Type:         "filelist",
		InnerContent: fileListJSON,
	})
	if err != nil {
		fmt.Println("Error marshalling file list packet:", err)
		return
	}
	pubSubTopic.Publish(ctx, pktMsg)
	fmt.Println("File list sent to all nodes")
}
