package goscreenmonit

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"math"
	"net"
	"time"

	"github.com/micaiahwallace/goscreenmonit/uploadpb"
	"google.golang.org/protobuf/proto"
)

type Session struct {
	address      string
	socket       net.Conn
	quit         chan int
	isRecording  bool
	fps          int
	lastImgStamp int64
	running      bool
	registration Registration
}

type Registration struct {
	Host string
	User string
}

// Create a new session that automatically connects to the server
func NewSession(address string, fps int, registration Registration) *Session {

	if fps <= 0 {
		log.Panicln("FPS must be greater than 0")
	}

	sess := &Session{
		address:      address,
		fps:          fps,
		isRecording:  false,
		running:      false,
		lastImgStamp: 0,
		registration: registration,
	}
	return sess
}

// Start a new session
func (session *Session) Start(quit chan int) {
	if session.running {
		log.Println("Already started session.")
	}
	session.running = true
	session.quit = quit
	go session.connect()
}

// Start a retry loop to connect to the server
func (session *Session) connect() {

	for {

		// Dial out to server
		conn, err := net.Dial("tcp4", session.address)
		if err != nil {
			log.Printf("Unable to connect to server, retry in 5 seconds: %v\n", err)
			time.Sleep(time.Second * 5)
			continue
		}

		// store the connection
		session.socket = conn

		// register with the server
		session.register()

		// Start processing commands
		indata := make(chan []byte)
		go ReadCommand(session.socket, indata)

		// Keep processing commands until socket closes
		for msgdata := range indata {
			response := &uploadpb.ServerResponse{}
			if err := proto.Unmarshal(msgdata, response); err != nil {
				log.Printf("Server message process error: %v\n", err)
				continue
			}
			session.processResponse(response)
		}

		// connection was closed
		session.running = false
		log.Println("Connection to server closed. Connecting in 3 seconds.")
		time.Sleep(time.Second * 3)
	}
}

// Process inbound server messages
func (session *Session) processResponse(response *uploadpb.ServerResponse) {
	switch response.Type {

	// Client is authenticated
	case uploadpb.ServerResponse_AUTHENTICATED:
		log.Println("Server registration successful, begin recording.")
		go session.record()

	// Client should quit now
	case uploadpb.ServerResponse_QUIT:
		log.Println("Quit command received, quitting now.")
		session.quit <- 0
	}

}

// Register user recording sessi5on with the server
func (session *Session) register() error {

	// Create registration
	cmd, err := CreateRegistration(session.registration.Host, session.registration.User)
	if err != nil {
		return err
	}

	// Send registration to server
	log.Println("Registering with the server.")
	SendMessage(cmd, session.socket)
	return nil
}

// Begin recording screen and sending data to server
func (session *Session) record() {

	// Check recording status
	if session.isRecording {
		fmt.Println("Already started record session.")
		return
	}

	// Setup recording session
	session.isRecording = true

	// Take screenshots and send to server per fps
	for {

		// Ensure we are running still
		if !session.running {
			session.isRecording = false
			return
		}

		// Get display count
		dcount := GetScreenCount()

		// Create image list
		images := make([][]byte, 0)

		// Take screenshot
		for i := 0; i < dcount; i++ {

			// get screen data
			img, err := CaptureScreen(i)
			if err != nil {
				time.Sleep(2 * time.Second)
				log.Printf("Unable to capture screen: %v\n", err)
				continue
			}

			// encode with zlib
			pngbuff := new(bytes.Buffer)
			png.Encode(pngbuff, img)
			encimg, encerr := EncodeImage(pngbuff.Bytes())
			if encerr != nil {
				time.Sleep(2 * time.Second)
				log.Printf("Unable to encode bytes: %v\n", encerr)
				continue
			}

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

		// Send image upload to server
		if err := SendMessage(msg, session.socket); err != nil {
			log.Printf("Unable to send upload request: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Check if we need to wait for fps compliance
		waitFor := session.getWaitTime()
		// log.Printf("Waiting for %v ms\n", waitFor)
		time.Sleep(waitFor)

		// Set the last screenshot timestamp
		session.lastImgStamp = time.Now().UnixNano() / int64(time.Millisecond)
	}
}

// Get waiting period required for next screenshot
func (session *Session) getWaitTime() time.Duration {

	// Time since last screenshot
	var timeDiff float64 = float64((time.Now().UnixNano() / int64(time.Millisecond)) - session.lastImgStamp)

	// Get number of ms per frame based on fps
	var msPerFrame float64 = 1000 / float64(session.fps)

	// Get required minimum time to wait for next frame to meet max fps
	waitForMs := math.Max(0, msPerFrame-float64(timeDiff))

	// Return ms duration as a time.Duration
	return time.Millisecond * time.Duration(waitForMs)
}
