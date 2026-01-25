package main

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type GamePhase int

const (
	PhaseBefore  GamePhase = iota // Showing "before" grid
	PhaseAfter                    // Showing "after" grid, player navigates
	PhaseSuccess                  // Correct answer
	PhaseFail                     // Wrong answer
)

var Colors = []lipgloss.Color{
	lipgloss.Color("#FF0000"), // Red
	lipgloss.Color("#00FF00"), // Green
	lipgloss.Color("#0000FF"), // Blue
	lipgloss.Color("#FFFF00"), // Yellow
	lipgloss.Color("#FF00FF"), // Magenta
	lipgloss.Color("#00FFFF"), // Cyan
}

type Cell struct {
	ColorIndex int // -1 = empty, 0-5 = color index
}

type model struct {
	phase      GamePhase
	grid       []Cell
	changedPos int
	cursorRow  int
	cursorCol  int
	gridSize   int
	width      int
	height     int
}

// NewModel creates a new game model with the given random seed.
// If seed is 0, the current time is used for randomization.
func NewModel(seed int64) model {
	if seed != 0 {
		rand.Seed(seed)
	} else {
		rand.Seed(time.Now().UnixNano())
	}

	gridSize := 4
	cells := make([]Cell, gridSize*gridSize)

	// Fill ~50% of cells with random colors
	for i := range cells {
		if rand.Float32() < 0.5 {
			cells[i].ColorIndex = rand.Intn(len(Colors))
		} else {
			cells[i].ColorIndex = -1
		}
	}

	return model{
		phase:      PhaseBefore,
		grid:       cells,
		gridSize:   gridSize,
		changedPos: -1,
		width:      60,
		height:     20,
	}
}

func (m *model) makeChange() {
	m.changedPos = rand.Intn(len(m.grid))

	oldColor := m.grid[m.changedPos].ColorIndex
	newColor := rand.Intn(len(Colors)+1) - 1 // -1 to len-1

	for newColor == oldColor {
		newColor = rand.Intn(len(Colors)+1) - 1
	}

	m.grid[m.changedPos].ColorIndex = newColor
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Universal quit
		switch key {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		// Phase-specific handling
		switch m.phase {
		case PhaseBefore:
			// Any key advances to "after" phase
			m.makeChange()
			m.phase = PhaseAfter
			return m, nil

		case PhaseAfter:
			return m.handleAfterPhase(key), nil

		case PhaseSuccess, PhaseFail:
			// Only q/ctrl+c quits (handled above)
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m model) handleAfterPhase(key string) model {
	// Navigation (arrow keys only)
	moves := map[string][2]int{
		"left":  {0, -1},
		"right": {0, 1},
		"up":    {-1, 0},
		"down":  {1, 0},
	}

	if delta, ok := moves[key]; ok {
		m.cursorRow = (m.cursorRow + delta[0] + m.gridSize) % m.gridSize
		m.cursorCol = (m.cursorCol + delta[1] + m.gridSize) % m.gridSize
		return m
	}

	// Submit answer
	if key == "enter" || key == "space" {
		cursorPos := m.cursorRow*m.gridSize + m.cursorCol
		if cursorPos == m.changedPos {
			m.phase = PhaseSuccess
		} else {
			m.phase = PhaseFail
		}
	}

	return m
}

func (m model) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Width(m.width).
		Align(lipgloss.Center)

	b.WriteString(titleStyle.Render("What Changed?"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", m.width))
	b.WriteString("\n\n")

	// Grid
	b.WriteString(m.renderGrid())
	b.WriteString("\n\n")

	// Status message
	statusStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)

	switch m.phase {
	case PhaseBefore:
		b.WriteString(statusStyle.Foreground(lipgloss.Color("#888888")).Render("Memorize this grid. Press any key to continue..."))
	case PhaseAfter:
		b.WriteString(statusStyle.Foreground(lipgloss.Color("#FFFF00")).Render("Find the cell that changed!"))
	case PhaseSuccess:
		b.WriteString(statusStyle.Foreground(lipgloss.Color("#00FF00")).Render("SUCCESS! You found it! Press q to quit."))
	case PhaseFail:
		b.WriteString(statusStyle.Foreground(lipgloss.Color("#FF0000")).Render("FAIL! Wrong cell. Press q to quit."))
	}

	b.WriteString("\n\n")
	b.WriteString(strings.Repeat("-", m.width))
	b.WriteString("\n")

	// Controls
	controlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Width(m.width).
		Align(lipgloss.Center)

	b.WriteString(controlStyle.Render("arrows: move | enter: submit | q: quit"))

	return b.String()
}

func (m model) renderGrid() string {
	var rows []string

	for r := 0; r < m.gridSize; r++ {
		var rowCells []string
		for c := 0; c < m.gridSize; c++ {
			idx := r*m.gridSize + c
			cell := m.grid[idx]
			isSelected := m.phase == PhaseAfter && r == m.cursorRow && c == m.cursorCol
			isChanged := (m.phase == PhaseSuccess || m.phase == PhaseFail) && idx == m.changedPos

			rowCells = append(rowCells, renderCell(cell, isSelected, isChanged))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
	}

	grid := lipgloss.JoinVertical(lipgloss.Center, rows...)

	// Center the grid
	gridStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)

	return gridStyle.Render(grid)
}

func renderCell(cell Cell, isSelected bool, isChanged bool) string {
	style := lipgloss.NewStyle().
		Width(5).
		Height(2).
		Align(lipgloss.Center, lipgloss.Center)

	content := "   "
	if cell.ColorIndex >= 0 {
		style = style.Background(Colors[cell.ColorIndex])
		content = "   "
	} else {
		style = style.Background(lipgloss.Color("#333333"))
	}

	borderColor := lipgloss.Color("#555555")
	borderStyle := lipgloss.NormalBorder()

	if isSelected {
		borderColor = lipgloss.Color("#FFD700")
		borderStyle = lipgloss.ThickBorder()
	}

	if isChanged {
		borderColor = lipgloss.Color("#FF00FF")
		borderStyle = lipgloss.DoubleBorder()
	}

	return style.
		Border(borderStyle).
		BorderForeground(borderColor).
		Render(content)
}
