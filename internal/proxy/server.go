package proxy

import (
	"fmt"
	"net"
	"sync"
)

// Server represents the proxy server
type Server struct {
	host       string
	port       int
	pass       string
	defaultHost string
	running    bool
	listener   net.Listener
	handlers   []*ConnectionHandler
	handlerMu  sync.Mutex
	logMu      sync.Mutex
}

// NewServer creates a new Server instance
func NewServer(host string, port int, pass string, defaultHost string) *Server {
	if defaultHost == "" {
		defaultHost = "127.0.0.1:143" // Default fallback value
	}
	
	return &Server{
		host:       host,
		port:       port,
		pass:       pass,
		defaultHost: defaultHost,
		running:    false,
		handlers:   make([]*ConnectionHandler, 0),
	}
}

// Run starts the server and begins accepting connections
func (s *Server) Run() error {
	var err error
	
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	
	s.running = true
	
	defer func() {
		s.running = false
		s.listener.Close()
	}()
	
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.running {
				// Server is shutting down
				return nil
			}
			s.PrintLog(fmt.Sprintf("Accept error: %v", err))
			continue
		}
		
		handler := NewConnectionHandler(conn, s)
		
		s.AddHandler(handler)
		go handler.Handle()
	}
	
	return nil
}

// PrintLog prints a log message with thread safety
func (s *Server) PrintLog(log string) {
	s.logMu.Lock()
	fmt.Println(log)
	s.logMu.Unlock()
}

// AddHandler adds a connection handler to the server's list
func (s *Server) AddHandler(handler *ConnectionHandler) {
	s.handlerMu.Lock()
	if s.running {
		s.handlers = append(s.handlers, handler)
	}
	s.handlerMu.Unlock()
}

// RemoveHandler removes a connection handler from the server's list
func (s *Server) RemoveHandler(handler *ConnectionHandler) {
	s.handlerMu.Lock()
	for i, h := range s.handlers {
		if h == handler {
			s.handlers = append(s.handlers[:i], s.handlers[i+1:]...)
			break
		}
	}
	s.handlerMu.Unlock()
}

// Close shuts down the server and all active connections
func (s *Server) Close() {
	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}
	
	s.handlerMu.Lock()
	handlers := make([]*ConnectionHandler, len(s.handlers))
	copy(handlers, s.handlers)
	s.handlerMu.Unlock()
	
	for _, h := range handlers {
		h.Close()
	}
}

// GetPass returns the server's password
func (s *Server) GetPass() string {
	return s.pass
}

// GetDefaultHost returns the default host to connect to if none specified
func (s *Server) GetDefaultHost() string {
	return s.defaultHost
}