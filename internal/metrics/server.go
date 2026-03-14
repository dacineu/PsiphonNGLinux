package metrics

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

// Server manages the metrics HTTP server
type Server struct {
	metrics *Metrics
	addr    string
	server  *http.Server
}

// NewServer creates a new metrics server on the given address
func NewServer(metrics *Metrics, addr string) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Ready")
	})
	return &Server{
		metrics: metrics,
		addr:    addr,
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

// Start runs the server in a background goroutine
func (s *Server) Start() {
	go func() {
		log.Printf("Metrics server listening on %s", s.addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Shutting down metrics server")
	return s.server.Shutdown(ctx)
}
