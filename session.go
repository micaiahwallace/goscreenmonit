package goscreenmonit

import (
	"bytes"
	"crypto/tls"
	"image/png"
	"math"
	"net"
	"os"
	"time"

	"github.com/kardianos/service"
	"github.com/micaiahwallace/goscreenmonit/uploadpb"
	"google.golang.org/protobuf/proto"
)

type Session struct {
	logger       service.Logger
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
func NewSession(logger service.Logger, address string, fps int, registration Registration) *Session {

	if fps <= 0 {
		logger.Errorf("FPS must be greater than 0")
		os.Exit(1)
	}

	sess := &Session{
		logger:       logger,
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
		session.logger.Warning("Already started session.")
	}
	session.running = true
	session.quit = quit
	go session.connect()
}

// Stop the session
func (session *Session) Stop() {
	session.running = false
	session.quit <- 0
}

// Start a retry loop to connect to the server
func (session *Session) connect() {

	for {

		// Dial out to server
		tlsconf := &tls.Config{

			// @TODO: Fix security here
			InsecureSkipVerify: true,
		}
		conn, err := tls.Dial("tcp4", session.address, tlsconf)
		if err != nil {
			session.logger.Warningf("Unable to connect to server, retry in 5 seconds: %v\n", err)
			time.Sleep(time.Second * 5)
			continue
		}

		// store the connection
		session.socket = conn
		session.running = true

		// register with the server
		session.register()

		// Start processing commands
		indata := make(chan []byte)
		go ReadCommand(session.socket, indata)

		// Keep processing commands until socket closes
		for msgdata := range indata {
			response := &uploadpb.ServerResponse{}
			if err := proto.Unmarshal(msgdata, response); err != nil {
				session.logger.Warningf("Server message process error: %v\n", err)
				continue
			}
			session.processResponse(response)
		}

		// connection was closed
		session.running = false
		session.logger.Warning("Connection to server closed. Connecting in 3 seconds.")
		time.Sleep(time.Second * 3)
	}
}

// Process inbound server messages
func (session *Session) processResponse(response *uploadpb.ServerResponse) {
	switch response.Type {

	// Client is authenticated
	case uploadpb.ServerResponse_AUTHENTICATED:
		session.logger.Info("Server registration successful, begin recording.")
		go session.record()

	// Client should quit now
	case uploadpb.ServerResponse_QUIT:
		session.logger.Info("Quit command received, quitting now.")
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
	session.logger.Info("Registering with the server.")
	SendMessage(cmd, session.socket)
	return nil
}

// Begin recording screen and sending data to server
func (session *Session) record() {

	// Check recording status
	if session.isRecording {
		session.logger.Warning("Already started record session.")
		return
	}

	// Setup recording session
	session.isRecording = true

	// Take screenshots and send to server per fps
	for {

		// Ensure we are running still
		if !session.running {
			session.isRecording = false
			session.logger.Warning("Cannot start, session not yet running.")
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
				session.logger.Errorf("Unable to capture screen: %v\n", err)
				continue
			}

			// encode with zlib
			pngbuff := new(bytes.Buffer)
			png.Encode(pngbuff, img)
			encimg, encerr := EncodeImage(pngbuff.Bytes())
			if encerr != nil {
				time.Sleep(2 * time.Second)
				session.logger.Errorf("Unable to encode bytes: %v\n", encerr)
				continue
			}

			// Add image to upload
			images = append(images, encimg)
		}

		// Create upload request
		msg, err := CreateUpload(images)
		if err != nil {
			session.logger.Errorf("Unable to create upload request: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Send image upload to server
		if err := SendMessage(msg, session.socket); err != nil {
			session.logger.Errorf("Unable to send upload request: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Check if we need to wait for fps compliance
		waitFor := session.getWaitTime()
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
