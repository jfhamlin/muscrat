package webaudio

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed static/index.html
var indexHTML []byte

type (
	// Server manages the WebAudio HTTP and WebSocket server
	Server struct {
		port      int
		ngrokURL  string
		clients   map[*websocket.Conn]bool
		clientsMu sync.RWMutex
		broadcast chan Message
		upgrader  websocket.Upgrader
		httpSrv   *http.Server
		ngrokCmd  *exec.Cmd
		ctx       context.Context
		cancel    context.CancelFunc
	}

	// Message represents a message sent to/from clients
	Message struct {
		Type   string                 `json:"type"`
		Params map[string]float64     `json:"params,omitempty"`
		Action string                 `json:"action,omitempty"`
		Data   map[string]interface{} `json:"data,omitempty"`
	}
)

// NewServer creates a new WebAudio server
func NewServer(port int) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		port:      port,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan Message, 256),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the HTTP server and ngrok tunnel
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/ws", s.handleWebSocket)

	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	// Start HTTP server in background
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("WebAudio server error: %v", err)
		}
	}()

	// Start broadcast handler
	go s.handleBroadcasts()

	// Start ngrok tunnel
	if err := s.startNgrok(); err != nil {
		return fmt.Errorf("failed to start ngrok: %w", err)
	}

	log.Printf("WebAudio server started on port %d", s.port)
	log.Printf("Public URL: %s", s.ngrokURL)

	return nil
}

// Stop stops the server and ngrok tunnel
func (s *Server) Stop() error {
	s.cancel()

	// Close all WebSocket connections
	s.clientsMu.Lock()
	for client := range s.clients {
		client.Close()
	}
	s.clientsMu.Unlock()

	// Stop ngrok
	if s.ngrokCmd != nil && s.ngrokCmd.Process != nil {
		s.ngrokCmd.Process.Kill()
	}

	// Stop HTTP server
	if s.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.httpSrv.Shutdown(ctx)
	}

	return nil
}

// Broadcast sends a message to all connected clients
func (s *Server) Broadcast(msg Message) {
	select {
	case s.broadcast <- msg:
	default:
		// Channel full, drop message
	}
}

// GetURL returns the public ngrok URL
func (s *Server) GetURL() string {
	return s.ngrokURL
}

// GetClientCount returns the number of connected clients
func (s *Server) GetClientCount() int {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return len(s.clients)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(indexHTML)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	log.Printf("WebAudio client connected (total: %d)", s.GetClientCount())

	// Send initial connection message
	conn.WriteJSON(Message{
		Type:   "connected",
		Action: "ready",
	})

	// Handle client messages (for future bidirectional comms)
	go s.handleClient(conn)
}

func (s *Server) handleClient(conn *websocket.Conn) {
	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		conn.Close()
		log.Printf("WebAudio client disconnected (total: %d)", s.GetClientCount())
	}()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (ping, etc)
		if msg.Type == "ping" {
			conn.WriteJSON(Message{Type: "pong"})
		}
	}
}

func (s *Server) handleBroadcasts() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.broadcast:
			s.clientsMu.RLock()
			for client := range s.clients {
				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("WebSocket write error: %v", err)
					client.Close()
				}
			}
			s.clientsMu.RUnlock()
		}
	}
}

func (s *Server) startNgrok() error {
	// Start ngrok HTTP tunnel
	cmd := exec.CommandContext(s.ctx, "ngrok", "http", fmt.Sprintf("%d", s.port), "--log=stdout")
	s.ngrokCmd = cmd

	// Capture output to get the public URL
	output, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Parse ngrok output to get public URL
	// This is a simple approach - in production you'd use ngrok API
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := output.Read(buf)
			if err != nil {
				return
			}
			line := string(buf[:n])
			if strings.Contains(line, "url=") {
				// Extract URL from log line like: url=https://xxx.ngrok.io
				parts := strings.Split(line, "url=")
				if len(parts) > 1 {
					url := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if strings.HasPrefix(url, "https://") {
						s.ngrokURL = url
						log.Printf("ngrok tunnel established: %s", url)
					}
				}
			}
		}
	}()

	// Wait a bit for ngrok to start
	time.Sleep(2 * time.Second)

	// Fallback: try to get URL from ngrok API
	if s.ngrokURL == "" {
		if err := s.fetchNgrokURL(); err != nil {
			log.Printf("Warning: could not fetch ngrok URL: %v", err)
			s.ngrokURL = fmt.Sprintf("http://localhost:%d (ngrok URL pending)", s.port)
		}
	}

	return nil
}

func (s *Server) fetchNgrokURL() error {
	// Query ngrok's local API
	resp, err := http.Get("http://127.0.0.1:4040/api/tunnels")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Tunnels []struct {
			PublicURL string `json:"public_url"`
			Proto     string `json:"proto"`
		} `json:"tunnels"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	for _, tunnel := range result.Tunnels {
		if tunnel.Proto == "https" {
			s.ngrokURL = tunnel.PublicURL
			return nil
		}
	}

	return fmt.Errorf("no HTTPS tunnel found")
}
