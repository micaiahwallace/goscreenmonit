package main

import (
	"log"
	"os"

	"github.com/micaiahwallace/goscreenmonit"
)

func main() {

	// Create and start a new session
	session := goscreenmonit.NewSession("127.0.0.1:3000", 1, goscreenmonit.Registration{Host: "linux-laptop", User: "micaiah"})
	quit := make(chan int)
	session.Start(quit)
	log.Println("Client agent running.")

	// Check for quit signal
	code := <-quit
	log.Printf("Received quit signal: %d\n", code)
	os.Exit(code)
}
