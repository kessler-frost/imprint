package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var seed = flag.Int64("seed", 0, "Random seed (0=time-based)")

// Color palette for random selection
var colorPalette = []lipgloss.Color{
	lipgloss.Color("#FF0000"), // Red
	lipgloss.Color("#00FF00"), // Green
	lipgloss.Color("#0000FF"), // Blue
	lipgloss.Color("#FFFF00"), // Yellow
	lipgloss.Color("#FF00FF"), // Magenta
	lipgloss.Color("#00FFFF"), // Cyan
}

type model struct {
	colors   [4]lipgloss.Color
	selected int
	width    int
	height   int
}

func initialModel() model {
	if *seed != 0 {
		rand.Seed(*seed)
	} else {
		rand.Seed(time.Now().UnixNano())
	}
	m := model{
		selected: 0,
		width:    50,
		height:   20,
	}
	m.regenerateColors()
	return m
}

func (m *model) regenerateColors() {
	for i := 0; i < 4; i++ {
		m.colors[i] = colorPalette[rand.Intn(len(colorPalette))]
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			m.selected = (m.selected + 1) % 3
			return m, nil
		case "k", "up":
			m.selected = (m.selected - 1 + 3) % 3
			return m, nil
		case "r":
			m.regenerateColors()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	// Bug 1: Misaligned title due to ANSI escape codes in centered text
	// The lipgloss centering doesn't account for the raw ANSI codes we inject
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Width(m.width).
		Align(lipgloss.Center)

	// Inject raw ANSI code that breaks centering calculation
	buggyTitle := "\x1b[1m   Visual Demo for Imprint\x1b[0m"
	b.WriteString(titleStyle.Render(buggyTitle))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n\n")

	// Random Colors Section
	b.WriteString("  Random Colors (changes each run):\n  ")

	// Render colored squares
	for i, color := range m.colors {
		squareStyle := lipgloss.NewStyle().
			Background(color).
			Foreground(lipgloss.Color("#000000"))
		b.WriteString(squareStyle.Render("████"))
		if i < 3 {
			b.WriteString(" ")
		}
	}
	b.WriteString("\n  ")

	// Labels
	for i := 0; i < 4; i++ {
		label := fmt.Sprintf("Color%d", i+1)
		b.WriteString(label)
		if i < 3 {
			b.WriteString(" ")
		}
	}
	b.WriteString("\n\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n\n")

	// Visual Bugs Section
	b.WriteString("  Visual Bugs:\n")

	// Bug 2: Color bleed - missing reset escape code
	// This causes the color to bleed into the next line
	indicator := "  "
	if m.selected == 0 {
		indicator = "> "
	}
	// Using raw ANSI without proper reset
	b.WriteString(indicator)
	b.WriteString("\x1b[32mMisaligned text here\x1b[0m")
	// Off-by-one positioning bug - extra space added
	b.WriteString(" \n")

	indicator = "  "
	if m.selected == 1 {
		indicator = "> "
	}
	// Bug 3: Color bleed - incomplete reset sequence
	b.WriteString(indicator)
	b.WriteString("\x1b[35mColor bleed example\x1b[") // Missing the 0m part
	b.WriteString("████\n")

	indicator = "  "
	if m.selected == 2 {
		indicator = "> "
	}
	// Bug 4: Poor contrast - yellow text on light background
	poorContrastStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Background(lipgloss.Color("#F0F0F0"))
	b.WriteString(indicator)
	b.WriteString(poorContrastStyle.Render("Hard to read text"))
	b.WriteString("\n\n")

	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")

	// Controls
	controlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Align(lipgloss.Center).
		Width(m.width)
	b.WriteString(controlStyle.Render("j/k: navigate | r: regenerate | q: quit"))

	return b.String()
}

func main() {
	flag.Parse()
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
