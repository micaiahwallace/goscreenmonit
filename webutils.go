package goscreenmonit

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

// An http basic auth middleware
func basicAuth(credentials map[string]string) mux.MiddlewareFunc {

	// Middleware function
	return func(h http.Handler) http.Handler {

		// Http handler
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

			// Check for authorization header existence
			s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(s) != 2 {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}

			// Get basic auth user and pass
			sentUser, sentPass, ok := r.BasicAuth()
			if !ok {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}

			// Check for credential set
			userPass, ok := credentials[sentUser]
			if !ok {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}

			// Validate user password
			if sentPass != userPass {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

// Read a credentials json file and return the resulting credentials map
func parseCredsFile(file string) (map[string]string, error) {

	// Open file from fs
	credsFile, rerr := os.Open(file)
	if rerr != nil {
		return nil, rerr
	}

	// Read bytes into byte slice
	credsBytes, rferr := ioutil.ReadAll(credsFile)
	if rferr != nil {
		return nil, rferr
	}

	// Parse json data from bytes
	creds := map[string]string{}
	if perr := json.Unmarshal(credsBytes, &creds); perr != nil {
		return nil, perr
	}

	return creds, nil
}
