package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kessler-frost/imprint/internal/mcp"
	"github.com/kessler-frost/imprint/internal/rest"
	"github.com/kessler-frost/imprint/internal/terminal"
)

var Version = "dev"

func main() {
	port := flag.Int("port", 8080, "REST API port")
	shell := flag.String("shell", getDefaultShell(), "Shell to run")
	rows := flag.Int("rows", 24, "Terminal rows")
	cols := flag.Int("cols", 80, "Terminal columns")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("imprint version %s\n", Version)
		os.Exit(0)
	}

	// Create terminal
	term, err := terminal.New(*shell, *rows, *cols)
	if err != nil {
		log.Fatalf("Failed to create terminal: %v", err)
	}

	if err := term.Start(); err != nil {
		log.Fatalf("Failed to start terminal: %v", err)
	}
	defer term.Close()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start REST server in goroutine
	restServer := rest.New(term, *port)
	go func() {
		fmt.Printf("Starting REST server on port %d\n", *port)
		if err := restServer.Start(); err != nil {
			log.Printf("REST server error: %v", err)
		}
	}()

	// Start MCP server (stdio) in goroutine
	mcpServer := mcp.New(term)
	go func() {
		fmt.Println("Starting MCP server on stdio")
		if err := mcpServer.Start(); err != nil {
			log.Printf("MCP server error: %v", err)
		}
	}()

	fmt.Println("Imprint is running. Press Ctrl+C to exit.")

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down...")
}

func getDefaultShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "/bin/bash"
	}
	return shell
}
