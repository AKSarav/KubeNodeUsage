package nodemodel

import (
	"fmt"
	"strings"
	"time"

	"github.com/AKSarav/KubeNodeUsage/v3/k8s"
	"github.com/AKSarav/KubeNodeUsage/v3/utils"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render
	searchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)
	// highlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("#fff0f4"))
)

type tickMsg time.Time

// nodeusage is the Bubble Tea model.
type NodeUsage struct {
	ClusterInfo k8s.Cluster
	Nodestats   []k8s.Node
	Args        *utils.Inputs
	Format      string
	viewport    viewport.Model
	content     string
	xOffset     int // Track horizontal scroll position
	width       int // Terminal width
	height      int // Terminal height
	ready       bool
	maxWidth    int // Maximum content width
	searchInput textinput.Model
	searching   bool
}

// NewNodeUsage creates a new NodeUsage model
func NewNodeUsage(args *utils.Inputs) NodeUsage {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 156
	ti.Width = 20

	model := NodeUsage{
		Args:        args,
		searchInput: ti,
		ClusterInfo: k8s.ClusterInfo(),
		Nodestats:   k8s.Nodes(args),
		Format:      "table",
		content:     "",
		xOffset:     0,
		width:       0,
		height:      0,
		ready:       false,
		maxWidth:    0,
		searching:   false,
	}

	// Initialize content
	var output strings.Builder
	MetricsHandler(model, &output)
	model.content = output.String()

	// Calculate initial maxWidth
	for _, line := range strings.Split(model.content, "\n") {
		if len(line) > model.maxWidth {
			model.maxWidth = len(line)
		}
	}

	return model
}

// Init Bubble Tea nodeusage
func (m NodeUsage) Init() tea.Cmd {
	return tea.Batch(tickCmd(), tea.EnterAltScreen)
}

// Update method for Bubble Tea - for constant update loop
func (m NodeUsage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		case msg.Type == tea.KeyEsc && m.searching:
			// Exit search mode
			m.searching = false
			m.searchInput.Reset()
			m.searchInput.Blur()
		case msg.Type == tea.KeyRunes && (msg.Runes[0] == 'Q' || msg.Runes[0] == 'q') && !m.searching:
			return m, tea.Quit
		case msg.Type == tea.KeyRunes && (msg.Runes[0] == 'S' || msg.Runes[0] == 's') && !m.searching:
			// Enter search mode
			m.searching = true
			m.searchInput.Focus()
			return m, nil
		}

		if m.searching {
			m.searchInput, cmd = m.searchInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		// Handle horizontal scrolling only when not searching
		switch msg.String() {
		case "left":
			if m.xOffset > 0 {
				m.xOffset -= 5
			}
		case "right":
			maxScroll := m.maxWidth - m.width
			if maxScroll > 0 && m.xOffset < maxScroll {
				m.xOffset = min(m.xOffset+5, maxScroll)
			}
		}
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-1)
			m.ready = true
		}
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 1

		// Re-render content with new size
		var output strings.Builder
		MetricsHandler(m, &output)
		m.content = output.String()

		// Recalculate maxWidth
		m.maxWidth = 0
		for _, line := range strings.Split(m.content, "\n") {
			if len(line) > m.maxWidth {
				m.maxWidth = len(line)
			}
		}
	case tickMsg:
		m.ClusterInfo = k8s.ClusterInfo()
		m.Nodestats = k8s.Nodes(m.Args)
		var output strings.Builder
		MetricsHandler(m, &output)
		m.content = output.String()

		m.maxWidth = 0
		for _, line := range strings.Split(m.content, "\n") {
			if len(line) > m.maxWidth {
				m.maxWidth = len(line)
			}
		}

		m.viewport.SetContent(m.content)
		cmds = append(cmds, tickCmd())
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func GetBar(decider float64) progress.Model {

	decider = decider * 100

	var prog progress.Model
	// decide which color to use based on the usage percentage below 30% is green, above 70% is red, else yellow
	if decider < 30 {
		prog = progress.New(progress.WithScaledGradient("#0bad5d", "#74b03f"))
	} else if decider > 70 {
		prog = progress.New(progress.WithScaledGradient("#13B013", "#F11658"))
	} else {
		prog = progress.New(progress.WithScaledGradient("#13B013", "#F18016"))
	}
	return prog
}

// View renders bubble tea
func (m NodeUsage) View() string {
	if !m.ready {
		return "Initializing..."
	}

	lines := strings.Split(m.content, "\n")
	var displayLines []string

	searchTerm := strings.ToLower(m.searchInput.Value())

	// Always include the header lines (first 6 lines)
	headerLines := 14
	for i := 0; i < min(headerLines, len(lines)); i++ {
		if len(lines[i]) > m.xOffset {
			displayLines = append(displayLines, lines[i][m.xOffset:])
		} else {
			displayLines = append(displayLines, "")
		}
	}

	// For the rest of the content
	for i := headerLines; i < len(lines); i++ {
		line := lines[i]
		// If searching, only include lines that match the search term
		if m.searching && searchTerm != "" {
			if strings.Contains(strings.ToLower(line), searchTerm) {
				if len(line) > m.xOffset {
					displayLine := line[m.xOffset:]
					// displayLine = highlightStyle.Render(displayLine)
					displayLines = append(displayLines, displayLine)
				} else {
					displayLines = append(displayLines, "")
				}
			}
		} else {
			// If not searching, include all lines
			if len(line) > m.xOffset {
				displayLines = append(displayLines, line[m.xOffset:])
			} else {
				displayLines = append(displayLines, "")
			}
		}
	}

	viewportContent := strings.Join(displayLines, "\n")
	m.viewport.SetContent(viewportContent)

	var helpText string
	if m.searching {
		matchCount := len(displayLines) - headerLines // Subtract header lines
		helpText = fmt.Sprintf("\n%s %s (%d matches) (ESC to exit search)",
			searchStyle.Render("Search:"),
			m.searchInput.View(),
			matchCount)
	} else {
		helpText = helpStyle("\nUse ← and → to scroll horizontally, S to search, Q or Ctrl+C to quit")
	}

	return fmt.Sprintf("%s%s", m.viewport.View(), helpText)
}

// tickCmd returns a command that sends a tick every second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
