package main

import (
	"flag"
	"fmt"
	"kubenodeusage/k8s"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding  = 2
	maxWidth = 80
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render
// map of keys of string type and values of interface type
// Keys are strings.
// Values can be of any type.



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

type inputs struct {
	helpFlag bool
	reverseFlag bool
	debug bool
	sortby string
	filternodes bool
	filtercolor bool
	filterlabels bool
	metrics string
}


/*
	Function: main
	Description: main function
*/
func main() {

	// parse command line arguments
	var (
		helpFlag bool
		reverseFlag bool
		debug bool
		sortby string
		filternodes bool
		filtercolor bool
		filterlabels bool
		metrics string
	)

	flag.BoolVar(&helpFlag, "help", false, "to display help")
	flag.BoolVar(&reverseFlag, "reverse", false, "to enable reverse sort")
	flag.BoolVar(&debug, "debug", false, "enable debug mode")
	flag.StringVar(&sortby, "sortby", "memory", "sort by memory, usage, color, name")
	flag.BoolVar(&filternodes, "filternodes", false, "filter nodes based on their name")
	flag.BoolVar(&filtercolor, "filtercolor", false, "filter nodes based on their color")
	flag.BoolVar(&filterlabels, "filterlabels", false, "filter nodes based on their labels")
	flag.StringVar(&metrics, "metrics", "all", "choose which metrics to display (memory, usage, disk, all)")
	flag.Parse()


	if helpFlag {
		usage()
	}

	args := inputs{
		helpFlag: helpFlag,
		reverseFlag: reverseFlag,
		debug: debug,
		sortby: sortby,
		filternodes: filternodes,
		filtercolor: filtercolor,
		filterlabels: filterlabels,
		metrics: metrics,
	}
	
	// print key value pairs
	t := reflect.TypeOf(args)
	v := reflect.ValueOf(args)

	if args.debug {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)
			fmt.Printf("%s: %v\n", field.Name, value)
		}
	}


	// prog := progress.New(progress.WithScaledGradient("#a4ebac", "#f266b3"))

	mdl := model{}
	mdl.args = args
	mdl.nodestats = k8s.Nodes()

	if _, err := tea.NewProgram(mdl).Run(); err != nil {
		fmt.Println("Oh no!", err)
		os.Exit(1)
	}
}

// tickMsg is the message we'll send to the update loop every second.
type tickMsg time.Time

// model is the Bubble Tea model.
type model struct {
	nodestats []k8s.Node
	args inputs
}

type ResetCursorCmd struct{}

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

		// check if R or R is pressed
		if msg.Type == tea.KeyRunes && (msg.Runes[0] == 'R' || msg.Runes[0] == 'r') {
			fmt.Println("R or r pressed")
			m.nodestats = k8s.Nodes()
			return m, tea.ClearScreen
		}
	case tickMsg:
		m.nodestats = k8s.Nodes()
		return m, tea.Batch(tickCmd(), tea.ClearScreen)

	}
	return m, nil
}


func GetBar(decider float64) (progress.Model) {

	decider = decider * 100

	var prog progress.Model
	// decide which color to use based on the usage percentage below 30% is green, above 70% is red, else yellow
	if  decider < 30 {
		prog = progress.New(progress.WithScaledGradient("#1ecb85","#2db43e"))
	} else if decider > 70 {
		prog = progress.New(progress.WithScaledGradient("#c38822","#f13d25"))
	} else {
		prog = progress.New(progress.WithScaledGradient("#b7c322","#fdbb2d"))
	}
	return prog
}
// View renders the progress bar.
func (m model) View() string {
	// pad := strings.Repeat(" ", padding)

	var output strings.Builder

	if m.args.debug {
		fmt.Fprintf(&output,"Debug mode enabled")
		fmt.Println(&output,"Args: ", m.args)
		fmt.Println(&output,"Nodes: ", m.nodestats)
	}

	// sort the nodestats based on the sortby flag
	if m.args.sortby == "usage" || m.args.sortby == "color" {
		sort.Slice(m.nodestats, func(i, j int) bool {
			return m.nodestats[i].Usage_memory_percent < m.nodestats[j].Usage_memory_percent
		})
	} else if m.args.sortby == "name" {
		sort.Slice(m.nodestats, func(i, j int) bool {
			return m.nodestats[i].Name < m.nodestats[j].Name
		})
	}


	
	
	if m.args.metrics == "memory" {
		fmt.Fprintf(&output,"Memory Metrics\n")
		fmt.Fprintf(&output, "%-30s %10s %10s %s\n","Name", "Usage(KB)", "Capacity(KB)", "Usage %")
	
		for _, node := range m.nodestats {
			prog := GetBar(float64(node.Usage_memory_percent)/100.0)
			fmt.Fprintf(&output, "%-30s %10d %10d %s\n",
				node.Name, node.Usage_memory, node.Capacity_memory, prog.ViewAs(float64(node.Usage_memory_percent)/100.0))
			}
	} else if m.args.metrics == "cpu" {
		fmt.Fprintf(&output,"CPU Metrics\n")
		fmt.Fprintf(&output, "%-30s %10s %10s %s\n","Name", "Usage(Cores)", "Capacity(Cores)", "Usage %")
		for _, node := range m.nodestats {
			prog := GetBar(float64(node.Usage_cpu_percent)/100.0)
			fmt.Fprintf(&output, "%-30s %10f %10d %s\n",
				node.Name, node.Usage_cpu, node.Capacity_cpu, prog.ViewAs(float64(node.Usage_cpu_percent)/100.0))
			}
	} else if m.args.metrics == "disk" {
		fmt.Fprintf(&output,"Disk Metrics\n")
		fmt.Fprintf(&output, "%-30s %10s %10s %s\n","Name", "Usage(KB)", "Capacity(KB)", "Usage %")
		for _, node := range m.nodestats {
			prog := GetBar(float64(node.Usage_disk_percent)/100.0)
			fmt.Fprintf(&output, "%-30s %10d %10d %s\n",
				node.Name, node.Usage_disk, node.Capacity_disk, prog.ViewAs(float64(node.Usage_disk_percent)/100.0))
			}

	} else if m.args.metrics == "all" {
		fmt.Println("All Metrics")
	}


	//Read Nodes from Nodes() function on the model and display it
	//output buffer

	


	output.WriteString("\n"+helpStyle("Press any key to quit"))
	
	return output.String()

	// return "\n" +
	// 	pad + m.progress.ViewAs(m.percent) + "\n\n" +
	// 	pad + helpStyle("Press any key to quit")
}

// tickCmd returns a command that sends a tick every second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second * 1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}