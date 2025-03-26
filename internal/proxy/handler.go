package proxy

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	BUFLEN   = 4096 * 4
	TIMEOUT  = 60 // seconds
	RESPONSE = "HTTP/1.1 101 Switching Protocol\r\nContent-Length: 1048576000000\r\n\r\n"
)

// ConnectionHandler handles an individual client connection
type ConnectionHandler struct {
	client       net.Conn
	target       net.Conn
	server       *Server
	clientClosed bool
	targetClosed bool
	clientBuffer []byte
	log          string
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(client net.Conn, server *Server) *ConnectionHandler {
	return &ConnectionHandler{
		client:       client,
		server:       server,
		clientClosed: false,
		targetClosed: true,
		log:          "Connection: " + client.RemoteAddr().String(),
	}
}

// Close closes both client and target connections
func (h *ConnectionHandler) Close() {
	if !h.clientClosed {
		h.client.Close()
		h.clientClosed = true
	}
	
	if !h.targetClosed {
		h.target.Close()
		h.targetClosed = true
	}
}

// FindHeader finds a header value in HTTP headers
func (h *ConnectionHandler) FindHeader(head []byte, header string) string {
	headStr := string(head)
	aux := strings.Index(headStr, header+": ")
	
	if aux == -1 {
		return ""
	}
	
	aux = strings.Index(headStr[aux:], ":") + aux
	headStr = headStr[aux+2:]
	aux = strings.Index(headStr, "\r\n")
	
	if aux == -1 {
		return ""
	}
	
	return headStr[:aux]
}

// ConnectTarget connects to the target server
func (h *ConnectionHandler) ConnectTarget(host string) error {
	i := strings.Index(host, ":")
	var hostStr string
	var portStr string
	
	if i != -1 {
		hostStr = host[:i]
		portStr = host[i+1:]
	} else {
		hostStr = host
		portStr = "443" // Default to HTTPS port
	}
	
	// Try to parse port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port format: %w", err)
	}
	
	// Connect to target
	targetAddr := fmt.Sprintf("%s:%d", hostStr, port)
	h.target, err = net.Dial("tcp", targetAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to target %s: %w", targetAddr, err)
	}
	
	h.targetClosed = false
	return nil
}

// MethodCONNECT handles HTTP CONNECT method
func (h *ConnectionHandler) MethodCONNECT(path string) {
	h.log += " - CONNECT " + path
	
	// Connect to target
	err := h.ConnectTarget(path)
	if err != nil {
		h.server.PrintLog(h.log + " - error: " + err.Error())
		return
	}
	
	// Send switching protocols response
	_, err = h.client.Write([]byte(RESPONSE))
	if err != nil {
		h.server.PrintLog(h.log + " - error: " + err.Error())
		return
	}
	
	h.clientBuffer = nil
	h.server.PrintLog(h.log)
	
	// Start proxying data
	h.DoCONNECT()
}

// DoCONNECT handles the bidirectional proxy between client and target
func (h *ConnectionHandler) DoCONNECT() {
	// Create channels for data
	clientChan := make(chan []byte)
	targetChan := make(chan []byte)
	errorChan := make(chan error)
	
	// Read from client, send to channel
	go func() {
		for {
			buf := make([]byte, BUFLEN)
			n, err := h.client.Read(buf)
			if err != nil {
				errorChan <- fmt.Errorf("client read error: %w", err)
				return
			}
			if n > 0 {
				clientChan <- buf[:n]
			}
		}
	}()
	
	// Read from target, send to channel
	go func() {
		for {
			buf := make([]byte, BUFLEN)
			n, err := h.target.Read(buf)
			if err != nil {
				errorChan <- fmt.Errorf("target read error: %w", err)
				return
			}
			if n > 0 {
				targetChan <- buf[:n]
			}
		}
	}()

	// Create a channel for the timeout
	timeoutChan := make(chan bool)
	
	// Create a function to restart the timeout
	resetTimeout := func() {
		// If there's an existing timeout goroutine, stop it
		if timeoutChan != nil {
			close(timeoutChan)
		}
		
		// Create a new timeout channel
		timeoutChan = make(chan bool)
		
		// Start a new timeout goroutine
		go func(done chan bool) {
			select {
			case <-time.After(time.Duration(TIMEOUT) * time.Second):
				// Send timeout signal to the main select
				errorChan <- fmt.Errorf("connection timeout")
			case <-done:
				// Timeout was reset, do nothing
				return
			}
		}(timeoutChan)
	}
	
	// Start the initial timeout
	resetTimeout()
	
	// Main proxy loop
	for {
		select {
		case data := <-clientChan:
			// Reset timeout
			resetTimeout()
			
			// Forward client data to target
			_, err := h.target.Write(data)
			if err != nil {
				h.server.PrintLog(fmt.Sprintf("%s - forward to target error: %v", h.log, err))
				return
			}
			
		case data := <-targetChan:
			// Reset timeout
			resetTimeout()
			
			// Forward target data to client
			_, err := h.client.Write(data)
			if err != nil {
				h.server.PrintLog(fmt.Sprintf("%s - forward to client error: %v", h.log, err))
				return
			}
			
		case err := <-errorChan:
			// An error occurred in one of the connections
			h.server.PrintLog(fmt.Sprintf("%s - %v", h.log, err))
			return
		}
	}
}

// Handle processes an incoming connection
func (h *ConnectionHandler) Handle() {
	defer func() {
		h.Close()
		h.server.RemoveHandler(h)
	}()
	
	// Read initial data from client
	buf := make([]byte, BUFLEN)
	n, err := h.client.Read(buf)
	if err != nil {
		h.log += " - error: " + err.Error()
		h.server.PrintLog(h.log)
		return
	}
	
	h.clientBuffer = buf[:n]
	
	// Extract headers
	hostPort := h.FindHeader(h.clientBuffer, "X-Real-Host")
	if hostPort == "" {
		hostPort = h.server.GetDefaultHost()
	}
	
	split := h.FindHeader(h.clientBuffer, "X-Split")
	if split != "" {
		// Receive additional data if split header is present
		_, err = h.client.Read(buf)
		if err != nil {
			h.log += " - error: " + err.Error()
			h.server.PrintLog(h.log)
			return
		}
	}
	
	if hostPort != "" {
		passwd := h.FindHeader(h.clientBuffer, "X-Pass")
		serverPass := h.server.GetPass()
		
		if len(serverPass) != 0 && passwd == serverPass {
			h.MethodCONNECT(hostPort)
		} else if len(serverPass) != 0 && passwd != serverPass {
			h.client.Write([]byte("HTTP/1.1 400 WrongPass!\r\n\r\n"))
		} else if strings.HasPrefix(hostPort, "127.0.0.1") || strings.HasPrefix(hostPort, "localhost") {
			h.MethodCONNECT(hostPort)
		} else {
			h.client.Write([]byte("HTTP/1.1 403 Forbidden!\r\n\r\n"))
		}
	} else {
		h.server.PrintLog("- No X-Real-Host!")
		h.client.Write([]byte("HTTP/1.1 400 NoXRealHost!\r\n\r\n"))
	}