package goscreenmonit

import (
	"crypto/tls"
	"errors"
	"log"
	"net"

	"github.com/micaiahwallace/goscreenmonit/uploadpb"
	"google.golang.org/protobuf/proto"
)

type RegisteredClient struct {
	Address      string
	Conn         net.Conn
	Register     *uploadpb.Register
	LatestUpload *uploadpb.ImageUpload
	Listeners    []*func()
}

type Server struct {
	address  string
	certPath string
	keyPath  string
	running  bool
	quit     chan int
	clients  map[string]*RegisteredClient
}

// Create and start a new server
func NewServer(address, certPath, keyPath string) *Server {
	server := &Server{
		address:  address,
		certPath: certPath,
		keyPath:  keyPath,
		running:  false,
		clients:  make(map[string]*RegisteredClient),
	}
	return server
}

// Start checks if already running before running listen
func (server *Server) Start(quit chan int) {
	if server.running {
		log.Println("Server already running.")
		return
	}
	server.quit = quit
	server.running = true
	go server.listen()
}

// Provide access to client list
func (server *Server) GetClients() map[string]*RegisteredClient {
	return server.clients
}

// Access a single client
func (server *Server) GetClient(address string) *RegisteredClient {
	client, ok := server.clients[address]
	if !ok {
		return nil
	}
	return client
}

// Register a listener for client image updates
func (server *Server) AddClientListener(address string, ln *func()) error {
	client, ok := server.clients[address]
	if !ok {
		return errors.New("client doesn't exist")
	}
	client.Listeners = append(client.Listeners, ln)
	return nil
}

// Remove a registered listener to cleanup
func (server *Server) RemoveClientListener(address string, ln *func()) error {
	client, ok := server.clients[address]
	if !ok {
		return errors.New("client doesn't exist")
	}
	for i, listener := range client.Listeners {
		if listener == ln {
			client.Listeners = append(client.Listeners[:i], client.Listeners[i+1:]...)
			break
		}
	}
	return nil
}

// Start listening on the address
func (server *Server) listen() {

	// Load tls keypair
	cert, certerr := tls.LoadX509KeyPair(server.certPath, server.keyPath)
	if certerr != nil {
		log.Printf("Unable to load server keypair: %v\n", certerr)
		server.quit <- 1
		return
	}
	tlsconfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	// Create the socket listener
	listener, err := tls.Listen("tcp4", server.address, tlsconfig)
	if err != nil {
		log.Printf("Unable to start server: %v\n", err)
		server.quit <- 1
		return
	}

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Incoming connection error: %v\n", err)
			continue
		}
		go server.handleClient(conn)
	}
}

// Handle client connection
func (server *Server) handleClient(conn net.Conn) {

	// New connection being processed
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	log.Printf("New connection: %s\n", addr)

	// Start processing requests
	indata := make(chan []byte)
	go ReadCommand(conn, indata)

	// Keep processing commands until socket closes
	for msgdata := range indata {
		req := &uploadpb.ClientRequest{}
		if err := proto.Unmarshal(msgdata, req); err != nil {
			log.Printf("Client request process error: %v\n", err)
			continue
		}
		server.processRequest(req, conn)
	}

	// Termination of connection
	server.deregister(addr)
	log.Printf("Connection closed %s\n", addr)
}

// Process client request
func (server *Server) processRequest(req *uploadpb.ClientRequest, conn net.Conn) {
	switch req.Type {

	// Parse registration and register connection
	case uploadpb.ClientRequest_REGISTER:
		regreq := &uploadpb.Register{}
		proto.Unmarshal(req.GetRequest(), regreq)
		server.register(regreq, conn)

	// Parse image upload request and process images
	case uploadpb.ClientRequest_UPLOAD:
		uploadreq := &uploadpb.ImageUpload{}
		proto.Unmarshal(req.GetRequest(), uploadreq)
		server.uploadImages(uploadreq, conn)
	}
}

// Register a new client
func (server *Server) register(req *uploadpb.Register, conn net.Conn) {

	address := conn.RemoteAddr().String()

	// Check if client registration exists
	if _, ok := server.clients[address]; ok {
		log.Printf("Client already registered: %v\n", address)
		server.quitConn(address, conn)
		return
	}

	// Add connection to registered clients
	log.Printf("Registering client: (%s) %s\n", req.GetUser(), address)
	server.clients[address] = &RegisteredClient{
		Address:   address,
		Conn:      conn,
		Register:  req,
		Listeners: make([]*func(), 0),
	}

	// Send auth response
	authresp, err := CreateResponse(uploadpb.ServerResponse_AUTHENTICATED)
	if err != nil {
		log.Printf("Unable to create auth response, quitting connection: %v\n", err)
		server.quitConn(address, conn)
		return
	}
	SendMessage(authresp, conn)
}

// Deregister a client
func (server *Server) deregister(address string) {
	if c, ok := server.clients[address]; ok {
		log.Printf("Deregistering client: (%s) %s\n", c.Register.GetUser(), address)
		delete(server.clients, address)
	}
}

// Send quit message to connection
func (server *Server) quitConn(address string, conn net.Conn) {

	// Send quit response
	quitres, qerr := CreateResponse(uploadpb.ServerResponse_QUIT)
	if qerr != nil {
		log.Printf("Unable to create quit response. %v\n", qerr)
		conn.Close()
	} else {
		SendMessage(quitres, conn)
	}

	// Delete registration if available
	server.deregister(address)
}

// Process image uploads
func (server *Server) uploadImages(req *uploadpb.ImageUpload, conn net.Conn) {

	// Get reference to client
	address := conn.RemoteAddr().String()
	client, ok := server.clients[address]
	if !ok {
		log.Printf("Received image upload from unregistered client: %v\n", address)
	}

	// Decode images with zlib
	for i, encim := range req.Images {

		decim, err := DecodeImage(encim)
		if err != nil {
			log.Printf("Unable to decode images: %v\n", err)
			return
		}

		req.Images[i] = decim
	}

	// Store image for later retrieval
	client.LatestUpload = req

	// Notify listeners of latest image
	for _, listener := range client.Listeners {
		(*listener)()
	}
}
