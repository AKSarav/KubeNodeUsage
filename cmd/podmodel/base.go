package podmodel

import (
	"fmt"
	"kubenodeusage/k8s"
	"kubenodeusage/utils"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render
	searchStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)
	highlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("#ab2770"))
)

type tickMsg time.Time

// podusage is the Bubble Tea model.
type PodUsage struct {
	ClusterInfo k8s.Cluster
	Podstats    []k8s.Pod
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

// NewPodUsage creates a new PodUsage model
func NewPodUsage(args *utils.Inputs) PodUsage {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 156
	ti.Width = 20

	return PodUsage{
		Args:        args,
		searchInput: ti,
		ClusterInfo: k8s.ClusterInfo(),
		Podstats:    k8s.Pods(args),
	}
}

// Init Bubble Tea podusage
func (m PodUsage) Init() tea.Cmd {
	return tea.Batch(tickCmd(), tea.EnterAltScreen)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update method for Bubble Tea - for constant update loop
func (m PodUsage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case tickMsg:
		m.ClusterInfo = k8s.ClusterInfo()
		m.Podstats = k8s.Pods(m.Args)
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

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
func (m PodUsage) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Apply horizontal scrolling and search highlighting
	lines := strings.Split(m.content, "\n")
	var scrolledLines []string

	searchTerm := strings.ToLower(m.searchInput.Value())
	for _, line := range lines {
		if len(line) > m.xOffset {
			displayLine := line[m.xOffset:]
			// If searching and line contains search term, highlight it
			if m.searching && searchTerm != "" && strings.Contains(strings.ToLower(displayLine), searchTerm) {
				displayLine = highlightStyle.Render(displayLine)
			}
			scrolledLines = append(scrolledLines, displayLine)
		} else {
			scrolledLines = append(scrolledLines, "")
		}
	}

	viewportContent := strings.Join(scrolledLines, "\n")
	m.viewport.SetContent(viewportContent)

	var helpText string
	if m.searching {
		helpText = fmt.Sprintf("\n%s %s (ESC to exit search)", searchStyle.Render("Search:"), m.searchInput.View())
	} else {
		helpText = helpStyle("\nUse ← and → to scroll horizontally, S to search, Q or Ctrl+C to quit")
	}

	return fmt.Sprintf("%s%s", m.viewport.View(), helpText)
}
