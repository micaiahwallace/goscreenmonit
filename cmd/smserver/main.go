package main

import (
	"log"
	"os"

	"github.com/micaiahwallace/goscreenmonit"
)

func main() {

	// Create a new monitor server and start it
	server := goscreenmonit.NewServer("127.0.0.1:3000")
	quit := make(chan int)
	server.Start(quit)
	log.Println("Monitor server running.")

	// Create a new webserver
	webServer := goscreenmonit.NewWebServer("127.0.0.1:8080", "server.pem", "server.key", server)
	go webServer.Start()
	log.Println("Web server is running.")

	// Check for quit signal
	code := <-quit
	log.Printf("Received quit signal: %d\n", code)
	os.Exit(code)
}
