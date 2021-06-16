package goscreenmonit

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

// Add any number of middlewares to execute before a route handler
func useMiddleware(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}

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

			// Decode the user:pass base64 encoded string
			b, err := base64.StdEncoding.DecodeString(s[1])
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// Separate the username and password
			pair := strings.SplitN(string(b), ":", 2)
			if len(pair) != 2 {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}

			// Check for credential set
			userPass, ok := credentials[pair[0]]
			if !ok {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}

			// Validate user password
			if pair[1] != userPass {
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
