package notification

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Server manages WebSocket connections for real-time notifications and admin API.
type Server struct {
	port     string
	upgrader websocket.Upgrader
	conns    map[*websocket.Conn]struct{}
	mu       sync.RWMutex
	broadcast chan []byte
	shutdown   chan struct{}

	// strictFields is the current global strict set.
	strictFields map[string]struct{}
	strictMu     sync.RWMutex
}

// New creates a new DOIP server.
func New(port string) *Server {
	return &Server{
		port:     port,
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true}},
		conns:    make(map[*websocket.Conn]struct{}),
		broadcast: make(chan []byte, 1024),
		shutdown:  make(chan struct{}),
		strictFields: make(map[string]struct{}),
	}
}

// Start begins listening and serving notifications and admin API.
func (s *Server) Start() error {
	// Broadcast loop
	go s.runBroadcaster()

	// HTTP mux
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)
	mux.HandleFunc("/admin/strict-fields", s.handleAdminStrictFields)

	// Listen and serve in a goroutine
	go func() {
		log.Printf("DOIP server listening on %s", s.port)
		if err := http.ListenAndServe(s.port, mux); err != nil && err != http.ErrServerClosed {
			log.Printf("DOIP server error: %v", err)
		}
	}()

	return nil
}

// Stop shuts down the server and closes all connections.
func (s *Server) Stop() {
	close(s.shutdown)
	s.mu.Lock()
	for conn := range s.conns {
		conn.Close()
	}
	s.conns = nil
	s.mu.Unlock()
}

// Send broadcasts a JSON-serializable message to all connected clients.
func (s *Server) Send(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("DOIP: marshal error: %v", err)
		return
	}
	select {
	case s.broadcast <- data:
	default:
		log.Printf("DOIP: broadcast channel full, dropping message")
	}
}

func (s *Server) runBroadcaster() {
	for {
		select {
		case <-s.shutdown:
			return
		case data := <-s.broadcast:
			s.mu.Lock()
			for conn := range s.conns {
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					conn.Close()
					delete(s.conns, conn)
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("DOIP: WebSocket upgrade failed: %v", err)
		return
	}
	s.mu.Lock()
	s.conns[conn] = struct{}{}
	s.mu.Unlock()

	defer func() {
		conn.Close()
		s.mu.Lock()
		delete(s.conns, conn)
		s.mu.Unlock()
	}()

	// Read loop to detect disconnection (optional)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// handleAdminStrictFields manages GET/POST for strict fields configuration.
func (s *Server) handleAdminStrictFields(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.strictMu.RLock()
		fields := make([]string, 0, len(s.strictFields))
		for f := range s.strictFields {
			fields = append(fields, f)
		}
		s.strictMu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"strict_fields": fields})

	case http.MethodPost:
		// Expect JSON: {"strict_fields": ["field1","field2"]} or set to null to clear
		var payload struct {
			StrictFields []string `json:"strict_fields"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		// Update strict set
		s.setStrictFields(payload.StrictFields)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// setStrictFields updates the global strict fields and broadcasts the change.
func (s *Server) setStrictFields(fields []string) {
	s.strictMu.Lock()
	old := make(map[string]struct{}, len(s.strictFields))
	for k := range s.strictFields {
		old[k] = struct{}{}
	}
	newSet := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		newSet[f] = struct{}{}
	}
	s.strictFields = newSet
	s.strictMu.Unlock()

	// Notify all connected DOIP clients.
	msg := map[string]interface{}{
		"type":     "strictness_changed",
		"previous": old,
		"current":  newSet,
	}
	s.Send(msg)
}

// GetStrictFields returns the current strict set (copy).
func (s *Server) GetStrictFields() []string {
	s.strictMu.RLock()
	defer s.strictMu.RUnlock()
	fields := make([]string, 0, len(s.strictFields))
	for f := range s.strictFields {
		fields = append(fields, f)
	}
	return fields
}

// Global singleton instance.
var (
	globalServer *Server
	mu           sync.RWMutex
)

// Start initializes and starts the DOIP server if not already running.
func Start(port string) error {
	mu.Lock()
	defer mu.Unlock()
	if globalServer != nil {
		return nil // already started
	}
	globalServer = New(port)
	return globalServer.Start()
}

// Stop shuts down the global DOIP server.
func Stop() {
	mu.Lock()
	if globalServer != nil {
		globalServer.Stop()
		globalServer = nil
	}
	mu.Unlock()
}

// Send broadcasts a message through the global server if active.
func Send(msg interface{}) {
	mu.RLock()
	s := globalServer
	mu.RUnlock()
	if s != nil {
		s.Send(msg)
	}
}

// IsEnabled returns true if the DOIP server is running.
func IsEnabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return globalServer != nil
}
