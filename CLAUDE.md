# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Build imprint binary
go build -o bin/imprint ./cmd/imprint

# Run tests
go test ./...

# Run imprint (MCP server on stdio)
./bin/imprint
./bin/imprint --rows 30 --cols 120 --shell /bin/zsh
```

## Examples

Each example is a separate Go module:

```bash
# Build and run an example
cd examples/screenshot-demo && go build && ./screenshot-demo
cd examples/text-demo && go build && ./text-demo
cd examples/what-changed && go build && ./what-changed
```

## Architecture

Imprint lets AI agents control a real terminal via MCP. The stack:

```
Claude Code → MCP (stdio) → Terminal Manager → ttyd + headless Chrome/xterm.js
```

- **ttyd**: Web terminal daemon exposing a real PTY via WebSocket
- **go-rod**: Headless Chrome automation for keyboard input and screenshots
- **xterm.js**: Terminal emulator in Chrome for pixel-perfect rendering

Key packages:
- `cmd/imprint/main.go` - CLI entry point, flag parsing, signal handling
- `internal/terminal/terminal.go` - Terminal manager (ttyd process + go-rod browser control)
- `internal/mcp/server.go` - MCP server with all tool handlers

## MCP Tools

All tools are defined in `internal/mcp/server.go`:
- `send_keystrokes` - Batch key presses via go-rod
- `type_text` - Type text into terminal textarea
- `get_screenshot` - JPEG screenshot via Chrome
- `get_screen_text` - Extract text via xterm.js buffer API
- `wait_for_text` / `wait_for_stable` - Polling helpers for test synchronization

## Key Mappings

Key handling is in `internal/terminal/terminal.go`:
- Special keys: `keyMap` (enter, arrows, function keys, etc.)
- Letters: `letterKeyMap` (a-z)
- Modifiers: `ctrl+c`, `alt+f`, `shift+a` format

## Dependencies

- **ttyd** must be installed (`brew install ttyd` on macOS)
- Chrome is auto-downloaded by go-rod on first run
