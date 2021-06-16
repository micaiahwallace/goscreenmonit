package goscreenmonit

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

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
	server.router.HandleFunc("/monitors/{address}", server.handleScreenshot)
	server.router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./ui/build"))))
}

// Handle requests to main viewing page
// func (server *WebServer) handleMain(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFile(w, r, "./ui/server/index.html")
// }

// Handle retreiving screenshots
func (server *WebServer) handleScreenshot(w http.ResponseWriter, r *http.Request) {

	// Get address to retrieve screenshot for
	vars := mux.Vars(r)
	address := vars["address"]

	// Get client connection for address
	client := server.mserver.GetClient(address)

	// Send image response
	im := client.LatestUpload.GetImages()[0]
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(im)))
	if _, err := w.Write(im); err != nil {
		log.Println("unable to write image.")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
}

// Handle retreiving a list of available monitors
func (server *WebServer) handleGetMonitors(w http.ResponseWriter, r *http.Request) {

	// Get monitors from the monitor server
	clients := server.mserver.GetClients()

	// Convert to a list
	monitors := []map[string]string{}
	for _, client := range clients {
		monitors = append(monitors, map[string]string{
			"address": client.Address,
			"user":    client.Register.GetUser(),
			"host":    client.Register.GetHost(),
		})
	}

	// Send json back to user
	if err := json.NewEncoder(w).Encode(monitors); err != nil {
		http.Error(w, "Server Error", 500)
	}
}
