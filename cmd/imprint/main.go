package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kessler-frost/imprint/internal/mcp"
	"github.com/kessler-frost/imprint/internal/terminal"
)

var Version = "dev"

func main() {
	shell := flag.String("shell", getDefaultShell(), "Shell to run")
	rows := flag.Int("rows", 24, "Terminal rows")
	cols := flag.Int("cols", 80, "Terminal columns")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("imprint version %s\n", Version)
		os.Exit(0)
	}

	term, err := terminal.New(*shell, *rows, *cols)
	if err != nil {
		log.Fatalf("Failed to create terminal: %v", err)
	}

	if err := term.Start(); err != nil {
		log.Fatalf("Failed to start terminal: %v", err)
	}
	defer term.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	mcpServer := mcp.New(term)
	done := make(chan struct{})
	go func() {
		if err := mcpServer.Start(); err != nil {
			log.Printf("MCP server error: %v", err)
		}
		close(done)
	}()

	select {
	case <-sigChan:
	case <-done:
	}
	fmt.Fprintln(os.Stderr, "Shutting down...")
}

func getDefaultShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/bash"
}
