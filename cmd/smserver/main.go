package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/micaiahwallace/goscreenmonit"
)

func main() {

	// Parse cli arguments
	var maddress, waddress, certPath, keyPath string
	flag.StringVar(&maddress, "mserver", "127.0.0.1:3000", "Specify listening address for monitor server")
	flag.StringVar(&waddress, "wserver", "127.0.0.1:8080", "Specify listening address for web server")
	flag.StringVar(&certPath, "cert", "server.crt", "Specify certificate file")
	flag.StringVar(&keyPath, "key", "server.key", "Specify private key file")
	flag.Parse()

	// Display settings
	fmt.Println("Current configuration:")
	fmt.Println("Monitor server: ", maddress)
	fmt.Println("Web server: ", waddress)
	fmt.Println("Cert: ", certPath)
	fmt.Println("Key: ", keyPath)

	// Create a new monitor server and start it
	server := goscreenmonit.NewServer(maddress, certPath, keyPath)
	quit := make(chan int)
	server.Start(quit)
	log.Println("Monitor server running.", maddress)

	// Create a new webserver and starts it
	webServer := goscreenmonit.NewWebServer(waddress, certPath, keyPath, server)
	go webServer.Start()
	log.Println("Web server is running.", waddress)

	// Check for quit signal
	code := <-quit
	log.Printf("Received quit signal: %d\n", code)
	os.Exit(code)
}
