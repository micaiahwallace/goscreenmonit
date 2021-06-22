package main

import (
	"flag"
	"log"
	"os"
	"os/user"
	"strconv"

	"github.com/micaiahwallace/goscreenmonit"
)

// Start running the program
func run() {

	// Parse cli arguments
	var server, fpsStr string
	flag.StringVar(&server, "server", "127.0.0.1:3000", "Specify server address")
	flag.StringVar(&fpsStr, "fps", "1", "Specify recording framerate")
	flag.Parse()

	// Get framerate int
	fps, fpserr := strconv.Atoi(fpsStr)
	if fpserr != nil {
		log.Fatalln("Please specify a valid numeric fps value")
	}

	// Log current settings
	log.Println("Current configuration:")
	log.Printf("FPS: %v\n", fpsStr)
	log.Printf("Server: %v\n", server)

	// Get system information
	hostName, userName, syserr := getSysInfo()
	if syserr != nil {
		log.Fatalf("System information cannot be retreived: %v\n", syserr)
	}

	// Create server registration data
	registration := goscreenmonit.Registration{
		Host: hostName,
		User: userName,
	}

	// Create and start a new session
	session := goscreenmonit.NewSession(server, fps, registration)
	quit := make(chan int)
	session.Start(quit)
	log.Println("Client agent running.")

	// Check for quit signal
	code := <-quit
	log.Printf("Received quit signal: %d\n", code)
	os.Exit(code)
}

// Get system information used for user registration
func getSysInfo() (string, string, error) {

	// Get hostname
	hostName, herr := os.Hostname()
	if herr != nil {
		return "", "", herr
	}

	// Get username
	user, uerr := user.Current()
	if uerr != nil {
		return "", "", uerr
	}

	return hostName, user.Username, nil
}
