package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
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
	daemon := flag.Bool("daemon", false, "Run in background (writes PID to ~/.imprint.pid)")
	flag.BoolVar(daemon, "d", false, "Run in background (shorthand)")
	stop := flag.Bool("stop", false, "Stop a running daemon")
	flag.Parse()

	if *version {
		fmt.Printf("imprint version %s\n", Version)
		os.Exit(0)
	}

	pidFile := getPidFilePath()

	if *stop {
		stopDaemon(pidFile)
		return
	}

	if *daemon {
		startDaemon(pidFile)
		return
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

func getPidFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/imprint.pid"
	}
	return filepath.Join(home, ".imprint.pid")
}

func startDaemon(pidFile string) {
	// Check if already running
	if pid, err := readPidFile(pidFile); err == nil {
		if processExists(pid) {
			fmt.Printf("Imprint is already running (PID %d)\n", pid)
			os.Exit(1)
		}
	}

	// Re-exec without -daemon/-d flag
	args := []string{}
	for _, arg := range os.Args[1:] {
		if arg != "-daemon" && arg != "--daemon" && arg != "-d" {
			args = append(args, arg)
		}
	}

	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start daemon: %v", err)
	}

	// Write PID file
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		log.Printf("Warning: failed to write PID file: %v", err)
	}

	fmt.Printf("Imprint started in background (PID %d)\n", cmd.Process.Pid)
}

func stopDaemon(pidFile string) {
	pid, err := readPidFile(pidFile)
	if err != nil {
		fmt.Println("Imprint is not running (no PID file)")
		return
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Process %d not found\n", pid)
		os.Remove(pidFile)
		return
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		fmt.Printf("Failed to stop process: %v\n", err)
		return
	}

	os.Remove(pidFile)
	fmt.Printf("Imprint stopped (PID %d)\n", pid)
}

func readPidFile(pidFile string) (int, error) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

func processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds; send signal 0 to check if process exists
	return process.Signal(syscall.Signal(0)) == nil
}
