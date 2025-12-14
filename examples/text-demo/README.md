# Text Demo

A minimal [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI for testing imprint's `get_screen_text` tool.

## Build

```bash
go build -o text-demo
```

## Run

```bash
./text-demo
```

## Controls

- `j` / `down` - Move cursor down
- `k` / `up` - Move cursor up
- `space` / `enter` - Toggle selection
- `q` / `ctrl+c` - Quit

## Testing with Imprint

Add imprint as an MCP server:

```bash
claude mcp add imprint -- imprint
```

Then use Claude Code with commands like:

```
"Use type_text to type './text-demo' and then send_keystrokes with ['enter']"
"Navigate down with send_keystrokes ['j'] and toggle with ['space']"
"Use get_screen_text to see the current terminal state"
```

This example uses `get_screen_text` since the TUI is text-based and doesn't rely on colors for meaning. Use `get_screenshot` when you need to verify colors or visual styling.
