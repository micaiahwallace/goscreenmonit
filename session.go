package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/micaiahwallace/goscreenmonit/uploadpb"
	"google.golang.org/protobuf/proto"
)

type Session struct {
	Address string
	Socket  net.Conn
	Ready   bool
}

// Create a new session that automatically connects to the server
func NewSession(address string) *Session {
	sess := &Session{Address: address, Ready: false}
	sess.Connect()
	return sess
}

// Start a retry loop to connect to the server
func (session *Session) Connect() {

	for {

		// Dial out to server
		conn, err := net.Dial("tcp4", session.Address)
		if err != nil {
			log.Printf("Unable to connect to server, retry in 5 seconds: %v\n", err)
			time.Sleep(time.Second * 5)
		}

		// store the connection
		session.Socket = conn

		// Start processing commands
		indata := make(chan []byte)
		go ReadCommand(session.Socket, indata)

		// Keep processing commands until socket closes
		for msgdata := range indata {
			response := &uploadpb.ServerResponse{}
			if err := proto.Unmarshal(msgdata, response); err != nil {
				log.Printf("Server message process error: %v\n", err)
				continue
			}
			session.ProcessMessage(response)
		}
	}
}

// Process inbound server messages
func (session *Session) ProcessMessage(response *uploadpb.ServerResponse) {
	switch response.Type {

	// Client is authenticated
	case uploadpb.ServerResponse_AUTHENTICATED:
		if !session.Ready {
			session.Ready = true
			go session.Record()
		} else {
			fmt.Println("Already started record session.")
		}

	// Client should quit now
	case uploadpb.ServerResponse_QUIT:
		log.Println("Quit command received, quitting now.")
		os.Exit(0)
	}

}

// Register self with the server
func (session *Session) Register(host, ip, user string) error {

	// Create registration
	cmd, err := CreateRegistration(host, ip, user)
	if err != nil {
		return err
	}

	// Send registration to server
	session.Socket.Write(cmd)
	return nil
}

// Begin recording screen and sending data to server
func (session *Session) Record() {

	// Get capturers
	caps, err := GetAllCapturers()
	if err != nil {
		log.Printf("Can't get all capturers: %v\n", err)
		return
	}

	// Take screenshots and send to server
	for {

		// Create image list
		images := make([][]byte, 0)

		// Take screenshot
		for _, cap := range caps {

			// get screen data
			img, err := CaptureScreen(cap)
			if err != nil {
				log.Printf("Unable to capture screen: %v\n", err)
				continue
			}

			// encode with zlib
			encimg := EncodeImage(img)

			// Add image to upload
			images = append(images, encimg)
		}

		// Create upload request
		msg, err := CreateUpload(images)
		if err != nil {
			log.Printf("Unable to create upload request: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Send message to server
		if err := session.SendMessage(msg); err != nil {
			log.Printf("Unable to send upload request: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}
	}
}

// Send a message by sending the message size before the message itself
func (session *Session) SendMessage(msg []byte) error {

	// First send the size of the message
	sizemsg := make([]byte, 1)
	sizemsg[0] = byte(len(msg))
	if _, err := session.Socket.Write(sizemsg); err != nil {
		return err
	}

	// Then send the actual message
	if _, err := session.Socket.Write(msg); err != nil {
		return err
	}

	return nil
}
