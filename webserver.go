package goscreenmonit

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/mux"
)

// Runs a web server front-end for the monitor server backend
type WebServer struct {
	address  string
	certPath string
	keyPath  string
	router   *mux.Router
	mserver  *Server
}

// Create a web server
func NewWebServer(address, cert, key string, monitorsrv *Server) *WebServer {
	return &WebServer{
		mserver:  monitorsrv,
		address:  address,
		certPath: cert,
		keyPath:  key,
	}
}

// Start running the web server
func (server *WebServer) Start() {
	server.setupRoutes()
	http.ListenAndServeTLS(server.address, server.certPath, server.keyPath, server.router)
}

// Configure router and all routes
func (server *WebServer) setupRoutes() {

	// Fetch credentials
	credPath := path.Join(path.Dir(os.Args[0]), "credentials.json")
	creds, err := parseCredsFile(credPath)
	if err != nil {
		log.Printf("Unable to parse credentials file. %v\n", err)
		return
	}

	// Setup routes
	server.router = mux.NewRouter()
	authMiddleware := basicAuth(creds)
	server.router.Use(authMiddleware)
	server.router.HandleFunc("/monitors", server.handleGetMonitors)
	server.router.HandleFunc("/ws/{address}/{screen}", server.handleWebsocket)
	// server.router.HandleFunc("/monitors/{address}/{screen}", server.handleScreenshot)
	server.router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./ui/build"))))
}

// Handle retreiving screenshots
// func (server *WebServer) handleScreenshot(w http.ResponseWriter, r *http.Request) {

// 	// Get address to retrieve screenshot for
// 	vars := mux.Vars(r)
// 	address := vars["address"]
// 	screennum, converr := strconv.Atoi(vars["screen"])
// 	if converr != nil {
// 		http.Error(w, "Bad Request", http.StatusBadRequest)
// 		return
// 	}

// 	// Get client connection for address
// 	client := server.mserver.GetClient(address)
// 	if client == nil {
// 		http.Error(w, "Not Found", http.StatusNotFound)
// 		return
// 	}

// 	// Get list of images
// 	images := client.LatestUpload.GetImages()

// 	// Verify image index is valid
// 	if screennum > len(images)-1 || screennum < 0 {
// 		http.Error(w, "Bad Request", http.StatusBadRequest)
// 		return
// 	}

// 	// Get requested image
// 	im := images[screennum]

// 	// Set image headers
// 	w.Header().Set("Content-Type", "image/png")
// 	w.Header().Set("Content-Length", strconv.Itoa(len(im)))

// 	// Write image data to http response
// 	if _, err := w.Write(im); err != nil {
// 		log.Printf("unable to write image. %v\n", err)
// 		http.Error(w, "Server Error", http.StatusInternalServerError)
// 		return
// 	}
// }

// Handle retreiving a list of available monitors
func (server *WebServer) handleGetMonitors(w http.ResponseWriter, r *http.Request) {

	// Get monitors from the monitor server
	clients := server.mserver.GetClients()

	// Convert to the format we want for json
	monitors := []map[string]string{}

	for _, client := range clients {
		monitors = append(monitors, map[string]string{
			"address":     client.Address,
			"user":        client.Register.GetUser(),
			"host":        client.Register.GetHost(),
			"screenCount": strconv.Itoa(len(client.LatestUpload.GetImages())),
		})
	}

	// Send json back to user
	if err := json.NewEncoder(w).Encode(monitors); err != nil {
		http.Error(w, "Server Error", 500)
	}
}

// Handle websocket connections
func (server *WebServer) handleWebsocket(w http.ResponseWriter, r *http.Request) {

	// Get address to retrieve screenshot for
	vars := mux.Vars(r)
	address := vars["address"]
	screennum, converr := strconv.Atoi(vars["screen"])
	if converr != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Get basic auth user and pass
	authUser, _, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Get client connection for address
	client := server.mserver.GetClient(address)
	if client == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Upgrade request to a websocket
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		http.Error(w, "Upgrade Error", http.StatusInternalServerError)
		return
	}

	// Handle sending images to client
	go func(conn net.Conn) {

		// Handle image updates from the client
		handler := func() {

			// Get list of images
			images := client.LatestUpload.GetImages()

			// Verify image index is valid
			if screennum > len(images)-1 || screennum < 0 {
				log.Printf("invalid screen number: %d\n", screennum)
				return
			}

			// Send requested image to websocket
			im := images[screennum]
			if err := wsutil.WriteServerBinary(conn, im); err != nil {
				log.Printf("Unable to write server binary: %v\n", err)
			}
		}

		// Cleanup after function ends
		defer func() {
			conn.Close()
			if err := server.mserver.RemoveClientListener(address, &handler); err != nil {
				log.Printf("Unable to remove listener: %v (%s->%s)\n", err, client.Register.GetUser(), client.Address)
			} else {
				log.Printf("Removed listener for user %s to %s -> %s\n", authUser, client.Register.GetUser(), client.Address)
			}
		}()

		// Add client listener
		if err := server.mserver.AddClientListener(address, &handler); err != nil {
			log.Printf("Unable to add client listener: %v\n", err)
		} else {
			log.Printf("Added listener for %s to %s -> %s\n", authUser, client.Register.GetUser(), client.Address)
		}

		// Listen for client messages
		for {
			_, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				return
			}
		}
	}(conn)
}
