# Bubble Tea Demo

A minimal [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI for testing imprint's terminal control.

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

# Get screen
print(requests.get("http://localhost:8080/screen/text").text)
```
