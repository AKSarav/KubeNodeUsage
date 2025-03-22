package podmodel

import (
	"fmt"
	"kubenodeusage/k8s"
	"kubenodeusage/utils"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

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
}

// Init Bubble Tea podusage
func (m PodUsage) Init() tea.Cmd {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))
	m.viewport = vp
	m.xOffset = 0
	return tickCmd()
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
		if msg.Type == tea.KeyCtrlC {
			fmt.Println("Ctrl+C pressed")
			return m, tea.Quit
		}
		if msg.Type == tea.KeyRunes && (msg.Runes[0] == 'Q' || msg.Runes[0] == 'q') {
			fmt.Println("Q or q pressed")
			return m, tea.Quit
		}
		// Add horizontal scrolling with arrow keys
		switch msg.String() {
		case "left":
			if m.xOffset > 0 {
				m.xOffset -= 5
			}
		case "right":
			m.xOffset += 5
		}
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
	case tickMsg:
		m.ClusterInfo = k8s.ClusterInfo()
		m.Podstats = k8s.Pods(m.Args)
		var output strings.Builder
		MetricsHandler(m, &output)
		m.content = output.String()
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
func (m PodUsage) View() string {
	// Apply horizontal scrolling by splitting content into lines and shifting each line
	lines := strings.Split(m.viewport.View(), "\n")
	var scrolledLines []string

	for _, line := range lines {
		if len(line) > m.xOffset {
			scrolledLines = append(scrolledLines, line[m.xOffset:])
		} else {
			scrolledLines = append(scrolledLines, "")
		}
	}

	return strings.Join(scrolledLines, "\n") + "\n" + helpStyle("Use ← and → to scroll horizontally, Q or Ctrl+C to quit")
}
