# Bubble Tea Demo (Text-Only Version)

A minimal [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI for testing imprint's terminal control using text-based screen capture.

## Build

```bash
go build -o demo
```

## Run

```bash
./demo
```

## Controls

- `j` / `down` - Move cursor down
- `k` / `up` - Move cursor up
- `space` / `enter` - Toggle selection
- `q` / `ctrl+c` - Quit

## Testing with Imprint

```python
import requests

# Type and run the demo
requests.post("http://localhost:8080/type", json={"text": "./demo"})
requests.post("http://localhost:8080/keystroke", json={"key": "enter"})

# Navigate and select
requests.post("http://localhost:8080/keystroke", json={"key": "j"})
requests.post("http://localhost:8080/keystroke", json={"key": "space"})

# Get screen as text (use /screen for PNG screenshot)
print(requests.get("http://localhost:8080/screen/text").text)
```

This example uses `/screen/text` since the TUI is text-based and doesn't rely on colors for meaning. Use `/screen` (PNG) when you need to verify colors or visual styling.
