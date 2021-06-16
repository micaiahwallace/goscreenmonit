package main

import (
	"flag"
	"log"
	"os"

	"github.com/kardianos/service"
)

var logger service.Logger

func main() {

	// Parse installation command
	flag.Parse()
	command := flag.Arg(0)

	// Create the program instance
	prg := &program{}

	// Config the service
	svcConfig := &service.Config{
		Name:        "GoScMonit",
		DisplayName: "Go Screen Monitor",
		Description: "Remotely provide screen support for enrolled devices.",
		Arguments:   os.Args[2:],
	}

	// Test if not in interactive mode to see if run from process manager
	if !service.ChosenSystem().Interactive() {
		runService(prg, svcConfig)
		return
	}

	switch command {
	case "install":
		log.Println("Installing service.", os.Args[2:])
		installService(prg, svcConfig)

	case "uninstall":
		log.Println("Removing service.")
		uninstallService(prg, svcConfig)

	case "start":
		log.Println("Starting service.")
		startService(prg, svcConfig)

	case "stop":
		log.Println("Stopping service.")
		stopService(prg, svcConfig)

	case "status":
		log.Printf("Status: %v\n", serviceStatus(prg, svcConfig))
	}

}
