# Screenshot Demo

A [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI demonstrating visual elements that require screenshot analysis. Shows random colors and intentional visual bugs.

## Building

```bash
go build
./screenshot-demo
```

## Controls

- `j` / `down` - Move selection down
- `k` / `up` - Move selection up
- `r` - Regenerate random colors
- `q` / `ctrl+c` - Quit

## Testing with Imprint

Assumes imprint MCP server is configured (see [main README](../../README.md#mcp-server-claude-code)).

```
"Use type_text to run './screenshot-demo' and send_keystrokes ['enter']"
"Use get_screenshot to capture the visual output"
"Navigate with send_keystrokes ['j'] and press 'r' to regenerate colors"
```

## Features

### Random Color Display

- Shows 4 colored squares using randomly selected colors from a palette at startup
- Available colors: Red, Green, Blue, Yellow, Magenta, Cyan
- Labels show "Color1", "Color2", etc. - the actual colors are only visible via screenshot
- Press `r` to regenerate random colors

### Intentional Visual Bugs

These bugs look correct in the source code but render incorrectly in the terminal:

1. **Misaligned Title**: Raw ANSI escape codes are injected into the centered title text, breaking lipgloss's centering calculation. The title appears off-center visually.

2. **Color Bleed**: An incomplete ANSI reset sequence (`\x1b[` instead of `\x1b[0m`) causes magenta color to bleed into the block characters on the next line.

3. **Off-by-one Positioning**: Extra space added after "Misaligned text here" causes subtle positioning issues.

4. **Poor Contrast**: Yellow text (`#FFFF00`) on light gray background (`#F0F0F0`) is nearly impossible to read, but appears fine in the code.

## Why Screenshot Analysis?

Text-based analysis using `get_screen_text` would show:
- "Color1 Color2 Color3 Color4" (no actual color information)
- The text content appears structurally correct
- No indication of the visual bugs

Screenshot analysis reveals:
- The actual RGB colors of each square
- Visual alignment and spacing issues
- Color bleeding artifacts
- Contrast problems that make text hard to read
