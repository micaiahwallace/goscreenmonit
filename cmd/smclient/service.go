package main

import (
	"log"
	"os"

	"github.com/kardianos/service"
)

// Run the service
func runService(prg *program, cfg *service.Config) {

	// Create the new service
	s, err := service.New(prg, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Get logger for service
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Run the service
	if err = s.Run(); err != nil {
		logger.Error(err)
	}
}

// Signal the OS to start the service
func startService(prg *program, cfg *service.Config) {

	// Create the service
	s, err := service.New(prg, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Trigger a start request
	if serr := s.Start(); serr != nil {
		log.Fatal(serr)
	}
}

// Signal the OS to stop the service
func stopService(prg *program, cfg *service.Config) {

	// Create the service
	s, err := service.New(prg, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Trigger a stop request
	if serr := s.Stop(); serr != nil {
		log.Fatal(serr)
	}
}

// Register the service with the operating system
func installService(prg *program, cfg *service.Config) {

	// Config the service
	cfg.Arguments = os.Args[2:]

	// Create the new service
	s, err := service.New(prg, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Install the service
	if err := s.Install(); err != nil {
		log.Fatal(err)
	}
}

// Unregister the service with the operating system
func uninstallService(prg *program, cfg *service.Config) {

	// Create the service
	s, err := service.New(prg, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Uninstall the service
	if err := s.Uninstall(); err != nil {
		log.Fatal(err)
	}
}

// Query service status
func serviceStatus(prg *program, cfg *service.Config) string {

	// Create the service
	s, err := service.New(prg, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Get service status
	stat, serr := s.Status()
	if serr != nil {
		log.Fatal(serr)
	}

	switch stat {
	case service.StatusRunning:
		return "running"
	case service.StatusStopped:
		return "stopped"
	default:
		return "unknown"
	}
}
