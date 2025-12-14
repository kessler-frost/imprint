# What Changed?

A visual memory game designed to showcase imprint's screenshot capabilities. The game displays a grid of colored cells, then changes one cell - the player must identify which cell changed.

## Why This Demo?

This game **requires visual perception** - there's no way to "cheat" by reading the source code because:
- Cell colors are randomly generated at runtime
- The changed cell position is randomly selected
- The new color is randomly chosen

An AI agent **must use screenshots** to see the grid state and identify the change.

## Building

```bash
cd examples/what-changed
go build -o what-changed .
```

## Playing

```bash
./what-changed
```

1. **BEFORE phase**: Memorize the grid colors
2. Press any key to advance
3. **AFTER phase**: One cell has changed - find it!
4. Navigate with arrow keys
5. Press `enter` to submit your answer
6. Press `q` to quit

## Controls

| Key | Action |
|-----|--------|
| `left` | Move cursor left |
| `down` | Move cursor down |
| `up` | Move cursor up |
| `right` | Move cursor right |
| `enter` / `space` | Submit answer (or advance phase) |
| `q` / `ctrl+c` | Quit |

## AI Agent Usage (via imprint)

This game is designed to be played by an AI agent using imprint's MCP tools:

```
1. type_text("./what-changed") + send_keystroke("enter")  # Launch game
2. get_screenshot()                                        # See BEFORE grid
3. send_keystroke("space")                                 # Advance to AFTER
4. get_screenshot()                                        # See AFTER grid
5. Compare the two screenshots to find the changed cell
6. Navigate to the changed cell using send_keystroke("up/down/left/right")
7. send_keystroke("enter")                                 # Submit answer
8. get_screenshot() or get_screen_text()                   # Check result (SUCCESS/FAIL)
9. send_keystroke("q")                                     # Quit
```

**Total screenshots needed: 2** (minimum for any visual comparison task)

## Imprint Capabilities Demonstrated

| Capability | Usage |
|------------|-------|
| `get_screenshot` | See grid colors (2 screenshots total) |
| `send_keystroke` | Navigate, submit, advance phase |
| `get_screen_text` | Parse result message |
| `type_text` | Launch the game |
