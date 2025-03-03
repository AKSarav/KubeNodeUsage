package podmodel

import (
	"fmt"
	"kubenodeusage/k8s"
	"kubenodeusage/utils"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
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
}

// Init Bubble Tea podusage
func (m PodUsage) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update method for Bubble Tea - for constant update loop
func (m PodUsage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			fmt.Println("Ctrl+C pressed")
			return m, tea.Quit
		}
		//  check if Q or q is pressed
		if msg.Type == tea.KeyRunes && (msg.Runes[0] == 'Q' || msg.Runes[0] == 'q') {
			fmt.Println("Q or q pressed")
			return m, tea.Quit
		}
	case tickMsg:
		m.ClusterInfo = k8s.ClusterInfo()
		m.Podstats = k8s.Pods(m.Args)
		return m, tea.Batch(tickCmd())
	}
	return m, nil
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
	var output strings.Builder

	DebugView(m, &output) // If debug on this would print Pod and arg details

	MetricsHandler(m, &output)

	output.WriteString("\n" + helpStyle("Press Q or Ctrl+C to quit"))

	return output.String()
}
