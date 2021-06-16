package main

import (
	"flag"
	"log"
	"os"
	"os/user"
	"strconv"

	"github.com/kardianos/service"
	"github.com/micaiahwallace/goscreenmonit"
)

type program struct {
	session *goscreenmonit.Session
}

// Handle starting the service
func (p *program) Start(s service.Service) error {

	// Get logger for service
	logger, err := s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Start program
	go p.run(logger)

	return nil
}

// Handle stopping the service
func (p *program) Stop(s service.Service) error {
	p.session.Stop()
	return nil
}

// Start running the program
func (p *program) run(logger service.Logger) {

	// Parse cli arguments
	var server, fpsStr string
	flag.StringVar(&server, "server", "127.0.0.1:3000", "Specify server address")
	flag.StringVar(&fpsStr, "fps", "1", "Specify recording framerate")
	flag.Parse()

	// Get framerate int
	fps, fpserr := strconv.Atoi(fpsStr)
	if fpserr != nil {
		logger.Error("Please specify a valid numeric fps value")
		os.Exit(1)
	}

	// Log current settings
	logger.Info("Current configuration:")
	logger.Infof("FPS: %v\n", fpsStr)
	logger.Infof("Server: %v\n", server)

	// Get system information
	hostName, userName, syserr := getSysInfo()
	if syserr != nil {
		logger.Errorf("System information cannot be retreived: %v\n", syserr)
		os.Exit(1)
	}

	registration := goscreenmonit.Registration{
		Host: hostName,
		User: userName,
	}

	// Create and start a new session
	session := goscreenmonit.NewSession(logger, server, fps, registration)
	quit := make(chan int)
	session.Start(quit)
	logger.Info("Client agent running.")
	p.session = session

	// Check for quit signal
	code := <-quit
	logger.Infof("Received quit signal: %d\n", code)
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
