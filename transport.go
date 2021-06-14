package main

import (
	"io"
	"log"
	"net"
)

// Accept and process requests from other end of socket
func ReadCommand(conn net.Conn, datapipe chan []byte) {

	for {

		// Get size of inbound message
		lengthdata := make([]byte, 1)
		if _, err := conn.Read(lengthdata); err != nil {
			if err != io.EOF {
				log.Printf("Unable to retrieve message: %v\n", err)
				close(datapipe)
				continue
			}
			break
		}

		// Attempt to read a message from the server
		msgdata := make([]byte, lengthdata[0])

		// Read the message data
		if _, err := conn.Read(msgdata); err != nil {
			if err != io.EOF {
				log.Printf("Unable to retrieve message: %v\n", err)
				continue
			}
			close(datapipe)
			break
		}

		// Send data to channel
		datapipe <- msgdata
	}
}
