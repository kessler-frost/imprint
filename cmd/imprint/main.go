package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
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

	lockFile, err := acquireLock()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer lockFile.Close()

	term, err := terminal.New(*shell, *rows, *cols)
	if err != nil {
		log.Fatalf("Failed to create terminal: %v", err)
	}

	if err := term.Start(); err != nil {
		log.Fatalf("Failed to start terminal: %v", err)
	}
	defer term.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	mcpServer := mcp.New(term)
	go func() {
		if err := mcpServer.Start(); err != nil {
			log.Fatalf("MCP server error: %v", err)
		}
	}()

	<-sigChan
	fmt.Fprintln(os.Stderr, "\nShutting down...")
}

func getDefaultShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/bash"
}

func getLockFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".imprint.lock")
}

func acquireLock() (*os.File, error) {
	lockPath := getLockFilePath()
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		// Check if another instance is running
		pidBytes := make([]byte, 32)
		n, _ := f.Read(pidBytes)
		f.Close()
		if n > 0 {
			return nil, fmt.Errorf("another imprint instance is already running (PID: %s)", string(pidBytes[:n]))
		}
		return nil, fmt.Errorf("another imprint instance is already running")
	}

	// Write our PID
	f.Truncate(0)
	f.Seek(0, 0)
	f.WriteString(strconv.Itoa(os.Getpid()))
	f.Sync()

	return f, nil
}
