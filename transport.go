package goscreenmonit

import (
	"encoding/binary"
	"log"
	"net"
)

// Helper function to read a certain amount of data into a buffer
func ReadConnBytes(count uint64, conn net.Conn) ([]byte, error) {

	final := make([]byte, 0)
	tmp := make([]byte, count)

	// Read until all bytes are retrieved
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			return nil, err
		}
		final = append(final, tmp[:n]...)

		// All data is captured
		if len(final) >= len(tmp) {
			break
		}
	}

	return final, nil
}

// Accept and process requests from other end of socket
func ReadCommand(conn net.Conn, datapipe chan []byte) {

	for {

		// Read first uint64 as the length of the message
		lengthdata, lerr := ReadConnBytes(8, conn)
		if lerr != nil {
			close(datapipe)
			return
		}
		msglen := binary.LittleEndian.Uint64(lengthdata)

		// Check the length to make sure it's a valid message
		if msglen > 1e9 {
			log.Println("Stream sync broken, resetting connection.")
			close(datapipe)
			return
		}

		// Read the full message based on previous length
		msgdata, merr := ReadConnBytes(msglen, conn)
		if merr != nil {
			close(datapipe)
			return
		}

		// Send data to channel
		datapipe <- msgdata
	}
}

// Send a message by sending the message size before the message itself
func SendMessage(msg []byte, conn net.Conn) error {

	// Create the message size data
	sizemsg := make([]byte, 8)
	msglen := uint64(len(msg))
	binary.LittleEndian.PutUint64(sizemsg, msglen)

	// First send the size of the message
	if _, err := conn.Write(sizemsg); err != nil {
		return err
	}

	// Then send the actual message
	if _, err := conn.Write(msg); err != nil {
		return err
	}

	return nil
}
