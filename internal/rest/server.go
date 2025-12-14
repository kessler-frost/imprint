package rest

import (
	"fmt"
	"net/http"

	"github.com/kessler-frost/imprint/internal/terminal"
)

// Server is the REST API server.
type Server struct {
	term *terminal.Terminal
	port int
}

// New creates a new REST server.
func New(term *terminal.Terminal, port int) *Server {
	return &Server{
		term: term,
		port: port,
	}
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/keystroke", corsMiddleware(s.handleKeystroke))
	mux.HandleFunc("/type", corsMiddleware(s.handleType))
	mux.HandleFunc("/screen", corsMiddleware(s.handleScreen))
	mux.HandleFunc("/screen/text", corsMiddleware(s.handleScreenText))
	mux.HandleFunc("/status", corsMiddleware(s.handleStatus))
	mux.HandleFunc("/resize", corsMiddleware(s.handleResize))

	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, mux)
}

// corsMiddleware adds CORS headers to all responses.
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
