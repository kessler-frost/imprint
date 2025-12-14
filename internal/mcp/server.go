package mcp

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/kessler-frost/imprint/internal/terminal"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server is the MCP server for Claude Code integration.
type Server struct {
	term *terminal.Terminal
}

// New creates a new MCP server.
func New(term *terminal.Terminal) *Server {
	return &Server{
		term: term,
	}
}

// Start begins the MCP server on stdio.
func (s *Server) Start() error {
	mcpServer := server.NewMCPServer(
		"imprint",
		"1.0.0",
		server.WithInstructions("AI-controllable terminal via MCP"),
	)

	s.registerTools(mcpServer)

	return server.ServeStdio(mcpServer)
}

// registerTools adds all terminal control tools to the MCP server.
func (s *Server) registerTools(mcpServer *server.MCPServer) {
	// Tool: send_keystroke
	sendKeyTool := mcp.NewTool(
		"send_keystroke",
		mcp.WithDescription("Send a key press to the terminal"),
		mcp.WithString("key",
			mcp.Description("Key to send (e.g., 'enter', 'ctrl+c', 'up')"),
			mcp.Required(),
		),
	)
	mcpServer.AddTool(sendKeyTool, s.handleSendKey)

	// Tool: type_text
	typeTextTool := mcp.NewTool(
		"type_text",
		mcp.WithDescription("Type a string of text into the terminal"),
		mcp.WithString("text",
			mcp.Description("Text to type into the terminal"),
			mcp.Required(),
		),
	)
	mcpServer.AddTool(typeTextTool, s.handleTypeText)

	// Tool: get_screenshot
	screenshotTool := mcp.NewTool(
		"get_screenshot",
		mcp.WithDescription("Get the current terminal screen as a base64-encoded JPEG image"),
		mcp.WithNumber("quality",
			mcp.Description("JPEG quality 0-100 (default: 70, lower = smaller file)"),
			mcp.Min(0),
			mcp.Max(100),
		),
	)
	mcpServer.AddTool(screenshotTool, s.handleGetScreenshot)

	// Tool: get_screen_text
	screenTextTool := mcp.NewTool(
		"get_screen_text",
		mcp.WithDescription("Get the current terminal screen content as plain text"),
	)
	mcpServer.AddTool(screenTextTool, s.handleGetScreenText)

	// Tool: get_status
	statusTool := mcp.NewTool(
		"get_status",
		mcp.WithDescription("Get terminal status information (rows, cols, ready)"),
	)
	mcpServer.AddTool(statusTool, s.handleGetStatus)

	// Tool: resize_terminal
	resizeTool := mcp.NewTool(
		"resize_terminal",
		mcp.WithDescription("Resize the terminal dimensions"),
		mcp.WithNumber("rows",
			mcp.Description("Number of rows"),
			mcp.Required(),
			mcp.Min(1),
		),
		mcp.WithNumber("cols",
			mcp.Description("Number of columns"),
			mcp.Required(),
			mcp.Min(1),
		),
	)
	mcpServer.AddTool(resizeTool, s.handleResize)
}

// handleSendKey handles the send_keystroke tool call.
func (s *Server) handleSendKey(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	key, err := request.RequireString("key")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = s.term.SendKey(key)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to send key: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Key '%s' sent successfully", key)), nil
}

// handleTypeText handles the type_text tool call.
func (s *Server) handleTypeText(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text, err := request.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = s.term.Type(text)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to type text: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Text typed successfully (%d characters)", len(text))), nil
}

// handleGetScreenshot handles the get_screenshot tool call.
func (s *Server) handleGetScreenshot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	quality := request.GetInt("quality", 70)

	jpegData, err := s.term.Screenshot(quality)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to capture screenshot: %v", err)), nil
	}

	encoded := base64.StdEncoding.EncodeToString(jpegData)
	return mcp.NewToolResultImage("Terminal screenshot", encoded, "image/jpeg"), nil
}

// handleGetScreenText handles the get_screen_text tool call.
func (s *Server) handleGetScreenText(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text, err := s.term.GetText()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get screen text: %v", err)), nil
	}

	return mcp.NewToolResultText(text), nil
}

// handleGetStatus handles the get_status tool call.
func (s *Server) handleGetStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rows, cols, ready := s.term.Status()

	status := fmt.Sprintf("Rows: %d\nCols: %d\nReady: %t", rows, cols, ready)
	return mcp.NewToolResultText(status), nil
}

// handleResize handles the resize_terminal tool call.
func (s *Server) handleResize(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rows, err := request.RequireInt("rows")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cols, err := request.RequireInt("cols")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = s.term.Resize(rows, cols)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to resize terminal: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Terminal resized to %dx%d", rows, cols)), nil
}
