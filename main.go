package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"kubenodeusage/cmd"
	"kubenodeusage/k8s"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pborman/getopt/v2"
)

const (
	padding  = 2
	maxWidth = 80
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render
// map of keys of string type and values of interface type
// Keys are strings.
// Values can be of any type.
var nodes map[string]interface{}



/*
	Function: usage
	Description: print usage
*/
func usage() {
	fmt.Println("Usage: go run main.go [options]")
	fmt.Println("Options:")
	// print in fine columns with fixed width
	displayfmt := "%-20s %-20s\n"
	fmt.Printf(displayfmt, "  -h, --help", "to display help")
	fmt.Printf(displayfmt, "  -s, --sortby", "sort by memory, usage, color, name")
	fmt.Printf(displayfmt, "  -n, --filternodes", "filter nodes based on their name")
	fmt.Printf(displayfmt, "  -c, --filtercolor", "filter nodes based on their color")
	fmt.Printf(displayfmt, "  -c, --filterlabels", "filter nodes based on their labels")
	fmt.Printf(displayfmt, "  -r, --reverse", "to enable reverse sort")
	fmt.Printf(displayfmt, "  -d, --debug", "enable debug mode")
	fmt.Printf(displayfmt, "  -m, --metrics", "choose which metrics to display (memory, usage, disk, all)")
	os.Exit(1)
}

/* 
	Function: parseinput
	Description: parse command line arguments
*/
func parseinput(args *getopt.Set) {
	fmt.Println("args:", args.Args())
}

/*
	Function: main
	Description: main function
*/
func main() {

	// parse command line arguments
	var helpFlag bool
	var reverseFlag bool
	var debug bool
	getopt.FlagLong(&helpFlag, "help", 'h', "display help")
	getopt.StringLong("sortby", 's', "sort order", "sort order")
	getopt.StringLong("filternodes", 'n', "filter nodes", "filter nodes")
	getopt.StringLong("filtercolor", 'c', "filter color", "filter color")
	getopt.StringLong("metrics", 'm', "metrics", "metrics")
	getopt.FlagLong(&reverseFlag, "reverse sort", 'r', "reverse sort")
	getopt.FlagLong(&debug, "debug", 'd', "debug")	
	getopt.Parse()
	args := getopt.CommandLine
	parseinput(args)


	cmd.NodeStats()
	
	prog := progress.New(progress.WithScaledGradient("#a4ebac", "#f266b3"))


	if _, err := tea.NewProgram(model{progress: prog}).Run(); err != nil {
		fmt.Println("Oh no!", err)
		os.Exit(1)
	}
}

// tickMsg is the message we'll send to the update loop every second.
type tickMsg time.Time

// model is the Bubble Tea model.
type model struct {
	nodestats []k8s.Node
}

// Init initializes the model.
func (m model) Init() tea.Cmd {
	return tickCmd()
}

// Update updates the model on each tick.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		return m, nil

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case tickMsg:
		m.percent += 0.25
		if m.percent > 1.0 {
			m.percent = 1.0
			return m, tea.Quit
		}
		return m, tickCmd()

	default:
		return m, nil
	}
}

// View renders the progress bar.
func (m model) View() string {
	pad := strings.Repeat(" ", padding)
	
	return "\n" +
		pad + m.progress.ViewAs(m.percent) + "\n\n" +
		pad + helpStyle("Press any key to quit")
}

// tickCmd returns a command that sends a tick every second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second * 5, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}