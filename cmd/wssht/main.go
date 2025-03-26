package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/sauufi/wssht/internal/tunnel"
)

const (
	DEFAULT_ADDR = "0.0.0.0"
	DEFAULT_PORT = 80
	DEFAULT_HOST = "127.0.0.1:143"
)

func main() {
	// Parse command line flags
	hostPtr := flag.String("b", DEFAULT_ADDR, "Binding address")
	portPtr := flag.Int("p", DEFAULT_PORT, "Listening port")
	passPtr := flag.String("pass", "", "Password for authentication")
	defHostPtr := flag.String("t", DEFAULT_HOST, "Default target host:port")
	flag.Parse()

	// Override with positional argument if provided
	listeningPort := *portPtr
	if len(os.Args) > 1 && os.Args[1] != "-b" && os.Args[1] != "-p" && os.Args[1] != "-pass" && os.Args[1] != "-h" {
		port, err := strconv.Atoi(os.Args[1])
		if err == nil {
			listeningPort = port
		}
	}

	// Print banner
	fmt.Println("\n:-------WSSHTunnel-------:")
	fmt.Println("Listening addr:", *hostPtr)
	fmt.Println("Listening port:", listeningPort)
	fmt.Println("Default target:", *defHostPtr)
	fmt.Println(":----------------------:\n")

	// Create new server
	server := proxy.NewServer(*hostPtr, listeningPort, *passPtr, *defHostPtr)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run server in a goroutine
	go func() {
		if err := server.Run(); err != nil {
			fmt.Printf("Server error: %v\n", err)
			sigChan <- syscall.SIGTERM // Trigger shutdown
		}
	}()

	// Block until signal received
	<-sigChan
	fmt.Println("Stopping server...")
	server.Close()
}