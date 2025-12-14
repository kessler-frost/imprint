# Text Demo

A minimal [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI for testing imprint's `get_screen_text` tool.

## Building

```bash
go build
./text-demo
```

## Controls

- `j` / `down` - Move cursor down
- `k` / `up` - Move cursor up
- `space` / `enter` - Toggle selection
- `q` / `ctrl+c` - Quit

## Testing with Imprint

Assumes imprint MCP server is configured (see [main README](../../README.md#mcp-server-claude-code)).

```
"Use type_text to run './text-demo' and send_keystrokes ['enter']"
"Navigate down with send_keystrokes ['j'] and toggle with ['space']"
"Use get_screen_text to see the current terminal state"
```

This example uses `get_screen_text` since the TUI is text-based and doesn't rely on colors for meaning. Use `get_screenshot` when you need to verify colors or visual styling.
