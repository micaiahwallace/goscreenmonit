package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/micaiahwallace/gowatchprog"
)

const PROG_NAME = "GoScreenMonit"
const LOG_FILE = "gsm-client.log"

func main() {

	// Parse installation command
	command := os.Args[1]

	// Log next to exe when running installation commands
	if command == "install" || command == "uninstall" {
		logToFile(LOG_FILE)
	}
	log.Printf("args: %v\n", os.Args[1:])

	switch command {

	case "install":
		/**
		Handle service installation in the AllUsers context
		*/
		prog, exeDir := makeCurrentProgram(os.Args[2:])
		log.Println("Installing service.")

		// Create path and command to install with user context for each user that signs in
		installBin, idErr := prog.InstallPathBin()
		if idErr != nil {
			log.Fatalf("Unable to get install binary path: %v\n", idErr)
		}
		prog.UserInstaller = createPathWithArgs(installBin, append([]string{"installuser"}, prog.Args...))

		// Install to path for all users
		if err := prog.Install(exeDir); err != nil {
			log.Fatalf("Install failed: %v\n", err)
		}

		// Register startup for all users
		if serr := prog.RegisterStartup(); serr != nil {
			log.Fatalf("Startup register failed: %v\n", serr)
		}

	case "installuser":
		/**
		Handle service installation in the CurrentUser context
		*/
		prog, _ := makeCurrentProgram(os.Args[2:])
		prog.StartupContext = gowatchprog.CurrentUser
		logToDataDir(prog, LOG_FILE)
		log.Println("Installing service for user.")

		// Register for startup
		if serr := prog.RegisterStartup(); serr != nil {
			log.Fatalf("Startup register failed: %v\n", serr)
		}

		// Start the watchdog as a separate process then quit the install
		installBinPath, ibpErr := prog.InstallPathBin()
		if ibpErr != nil {
			log.Fatalf("Launch after install failed: %v", ibpErr)
		}
		cmd := exec.Command(installBinPath, prog.Args...)
		cmd.Start()

	case "uninstall":
		/**
		Remove the installation from and unregister startup for all users
		*/
		log.Println("Removing service.")
		prog, _ := makeCurrentProgram([]string{})
		if err := prog.Uninstall(); err != nil {
			log.Fatalf("Uninstall failed: %v\n", err)
		}

	case "watch":
		/**
		Begin the watchdog to restart the service when it fails indefinitely
		*/
		startWatchdog(os.Args[2:])

	default:
		/**
		Finally, the actual logic to run the service
		*/
		prog, _ := makeCurrentProgram(os.Args[1:])
		prog.StartupContext = gowatchprog.CurrentUser
		logToDataDir(prog, LOG_FILE)
		log.Println("Gsm client running.")
		run()
		log.Println("Gsm client stopped.")
	}
}

// Run the watchdog service
func startWatchdog(args []string) {
	prog, _ := makeCurrentProgram(args)
	prog.StartupContext = gowatchprog.CurrentUser
	logToDataDir(prog, LOG_FILE)
	log.Println("Starting watchdog service.")

	// Setup communication channels
	errs := make(chan error)
	msgs := make(chan string)
	quit := make(chan interface{})

	// Run watchdog and wait for completion
	go prog.RunWatchdog(errs, msgs, quit)
	complete := false
	for !complete {
		select {
		case msg := <-msgs:
			log.Println(msg)
		case err := <-errs:
			log.Println(err)
		case <-quit:
			complete = true
		}
	}
	log.Println("Watchdog process quit")
}
