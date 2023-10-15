package main

import (
	"flag"
	"fmt"
	"kubenodeusage/k8s"
	"kubenodeusage/utils"
	"os"
	"reflect"
	"sort"
	"strconv"
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
	fmt.Printf(displayfmt, "  --help", "to display help")
	fmt.Printf(displayfmt, "  --sortby", "sort by memory, usage, color, name")
	fmt.Printf(displayfmt, "  --filternodes", "filter nodes based on their name")
	fmt.Printf(displayfmt, "  --filtercolor", "filter nodes based on their color")
	fmt.Printf(displayfmt, "  --filterlabels", "filter nodes based on their labels")
	fmt.Printf(displayfmt, "  --desc", "to enable reverse sort")
	fmt.Printf(displayfmt, "  --debug", "enable debug mode")
	fmt.Printf(displayfmt, "  --metrics", "choose which metrics to display (memory, usage, disk, all)")
	os.Exit(1)
}

func PrintArgs(args Inputs){
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

}

func DebugView(m model, output *strings.Builder){
	if m.args.debug {
		fmt.Fprintf(output,"\nDebug mode enabled")
		fmt.Println(output,"\nArgs: ", m.args)
		fmt.Println(output,"\nNodes: ", m.nodestats)
	}
}


func RightMetric(m model,index int)(float32){
	switch m.args.metrics{
		case "memory":
			if m.args.sortby == "used" { 
				return float32(m.nodestats[index].Usage_memory)
			}else if m.args.sortby == "capacity" {
				return float32(m.nodestats[index].Capacity_memory)
			} else if m.args.sortby == "color" {
				return m.nodestats[index].Usage_memory_percent
			}
		case "cpu":
			if m.args.sortby == "used" { 
				return float32(m.nodestats[index].Usage_cpu)
			}else if m.args.sortby == "capacity" {
				return float32(m.nodestats[index].Capacity_cpu)
			}else if m.args.sortby == "color" {
				return m.nodestats[index].Usage_cpu_percent
			}
		case "disk":
			if m.args.sortby == "used" { 
			return float32(m.nodestats[index].Usage_disk)
			} else if m.args.sortby == "capacity" {
				return float32(m.nodestats[index].Capacity_disk)
			} else if m.args.sortby == "color" {
				return m.nodestats[index].Usage_disk_percent
			}
		default:
			return m.nodestats[index].Usage_memory_percent
	}
	return m.nodestats[index].Usage_memory_percent
}


func SortByHandler(m model){

	if m.args.sortby != ""{
		if !m.args.reverseFlag{
			sort.Slice(m.nodestats, func (i, j int) bool {
				return RightMetric(m,i) < RightMetric(m,j)
			})
		} else{
			sort.Slice(m.nodestats, func(i, j int) bool {
				return RightMetric(m,i) > RightMetric(m,j)
			})
		}
	} 

	// // sort based on the sortby flag
	// if m.args.sortby == "usage" || m.args.sortby == "color"  {
	// 	if !m.args.reverseFlag{
	// 		sort.Slice(m.nodestats, func(i, j int) bool {
	// 			return RightMetric(m,i) < RightMetric(m,j)
	// 		})
	// 	} else{
	// 		sort.Slice(m.nodestats, func(i, j int) bool {
	// 			return RightMetric(m,i) > RightMetric(m,j)
	// 		})
	// 	}
		
	// } else if m.args.sortby == "name" {
	// 	if !m.args.reverseFlag{
	// 		sort.Slice(m.nodestats, func(i, j int) bool {
	// 			return m.nodestats[i].Name < m.nodestats[j].Name
	// 		})
	// 	} else{
	// 		sort.Slice(m.nodestats, func(i, j int) bool {
	// 			return m.nodestats[i].Name > m.nodestats[j].Name
	// 		})
	// 	}
	// } else if m.args.sortby == "capacity" {
	// }
}

func FilterForColor(m *model) []k8s.Node {
	if m.args.debug{
		fmt.Println("Filter For Color called")
	}
	var filteredNodes []k8s.Node
	var thresholdMin, thresholdMax float64

	// Define the color threshold values
	switch m.args.filtercolor {
	case "red":
		thresholdMin = 70
		thresholdMax = 100
	case "orange":
		thresholdMin = 30
		thresholdMax = 70
	case "green":
		thresholdMin = 0
		thresholdMax = 30
	default:
		thresholdMin = 0
		thresholdMax = 100
	}

	// fmt.Printf("Final Min %f and Max %f",thresholdMin,thresholdMax)

	// Filter nodes based on metric and threshold values
	for _, node := range m.nodestats {
		// fmt.Printf("Checking node %s and selected metric %s",string(node.Name), string(m.args.metrics))
		var usagepercent float64
		switch m.args.metrics {
		case "memory":
			usagepercent = float64(node.Usage_memory_percent) / 100.0
		case "cpu":
			usagepercent = float64(node.Usage_cpu_percent) / 100.0
		case "disk":
			usagepercent = float64(node.Usage_disk_percent) / 100.0
		default:
			if m.args.debug{
				fmt.Println("No Matching Metric",m.args.metrics)
			}
		}

		if (usagepercent*100) >= thresholdMin && (usagepercent*100) < thresholdMax {
			filteredNodes = append(filteredNodes, node)
		}
	}
	if m.args.debug{
		fmt.Println("Filter For Color result:",filteredNodes)
	}
	return filteredNodes
}

func PrintDesign(output *strings.Builder,maxNameWidth int){
	lines := strings.Repeat("-", maxNameWidth+12+12+20)
	fmt.Fprintf(output, lines)
	fmt.Fprintf(output,"\n")
}

func MetricsHandler(m model, output *strings.Builder){

	

	// Nodes Filtering based on filters
	filteredNodes := FilterForColor(&m)

	// decide formatting
	maxNameWidth := 35
	for _, node := range filteredNodes{
		if maxNameWidth < len(node.Name){
			maxNameWidth = len(node.Name)
		}
	}
	format := "%-"+strconv.Itoa(maxNameWidth)+"s %-12s %-12s %s\n"
	fmt.Fprintf(output,"\n# KubeNodeUsage\n# Version: 3\n# https://github.com/AKSarav/Kube-Node-Usage\n\n")
	if m.args.metrics == "memory" {
		fmt.Fprintf(output,"Memory Metrics\n\n")
		fmt.Fprintf(output, format, "Name", "Used(GB)", "Max(GB)", "Usage %")
		PrintDesign(output, maxNameWidth)
	
		for _, node := range filteredNodes {
			prog := GetBar(float64(node.Usage_memory_percent)/100.0)
			fmt.Fprintf(output, format,
				node.Name, strconv.Itoa(node.Usage_memory/1024/1024), strconv.Itoa(node.Capacity_memory/1024/1024), prog.ViewAs(float64(node.Usage_memory_percent)/100.0))
			}
	} else if m.args.metrics == "cpu" {
		fmt.Fprintf(output,"CPU Metrics\n\n")
		fmt.Fprintf(output, format,"Name", "Used(Cores)", "Max(Cores)", "Usage %")
		PrintDesign(output, maxNameWidth)
		for _, node := range filteredNodes {
			prog := GetBar(float64(node.Usage_cpu_percent)/100.0)
			fmt.Fprintf(output, format,
				node.Name, strconv.Itoa(int(node.Usage_cpu)), strconv.Itoa(node.Capacity_cpu), prog.ViewAs(float64(node.Usage_cpu_percent)/100.0))
			}
	} else if m.args.metrics == "disk" {
		fmt.Fprintf(output,"Disk Metrics\n\n")
		fmt.Fprintf(output, format,"Name", "Used(GB)", "Max(GB)", "Usage %")
		PrintDesign(output, maxNameWidth)
		for _, node := range filteredNodes {
			prog := GetBar(float64(node.Usage_disk_percent)/100.0)
			fmt.Fprintf(output, format,
				node.Name, strconv.Itoa(node.Usage_disk/1024/1024), strconv.Itoa(node.Capacity_disk/1024/1024), prog.ViewAs(float64(node.Usage_disk_percent)/100.0))
			}

	} else if m.args.metrics == "all" {
		fmt.Println("All Metrics")
	}
}

func checkinputs(args *Inputs){

	if args.filtercolor != "" {
		if !utils.IsValidColor(args.filtercolor){
			fmt.Println("Not a valid color please choose one of the following colors",utils.PrintValidColors())
			os.Exit(2)
		}
	}

	if args.metrics != ""{
		if !utils.IsValidMetric(args.metrics){
			fmt.Println("Not a valid Metric please choose one of",utils.PrintValidMetrics())
			os.Exit(2)
		}
	}

	if args.sortby != "" {
		if !utils.IsValidSort(args.sortby){
			fmt.Println("Not a valid Sort by option please choose one of",utils.PrintValidSorts())
			os.Exit(2)
		}
	}
}
type Inputs struct {
	helpFlag bool
	reverseFlag bool
	debug bool
	sortby string
	filternodes string
	filtercolor string
	filterlabels string
	metrics string
}

/*
	Function: main
	Description: main function
*/
func main() {


	// clearScreen()
	// parse command line arguments
	var (
		helpFlag bool
		reverseFlag bool
		debug bool
		sortby string
		filternodes string
		filtercolor string
		filterlabels string
		metrics string
	)

	flag.BoolVar(&helpFlag, "help", false, "to display help")
	flag.BoolVar(&reverseFlag, "desc", false, "to display sort in descending order")
	flag.BoolVar(&debug, "debug", false, "enable debug mode")
	flag.StringVar(&sortby, "sortby", "", "sort by capacity, usage, color, name")
	flag.StringVar(&filternodes, "filternodes", "", "filter nodes based on their name")
	flag.StringVar(&filtercolor, "filtercolor", "", "filter nodes based on their color")
	flag.StringVar(&filterlabels, "filterlabels", "", "filter nodes based on their labels")
	flag.StringVar(&metrics, "metrics", "all", "choose which metrics to display (memory, usage, disk, all)")
	flag.Parse()


	if helpFlag {
		usage()
	}

	args := Inputs{
		helpFlag: helpFlag,
		reverseFlag: reverseFlag,
		debug: debug,
		sortby: sortby,
		filternodes: filternodes,
		filtercolor: filtercolor,
		filterlabels: filterlabels,
		metrics: metrics,
	}

	
	checkinputs(&args) // sending the args using Address of Operator
	
	if debug {
		PrintArgs(args)
	}

	mdl := model{}
	mdl.args = &args
	mdl.nodestats = k8s.Nodes(metrics)

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
	args *Inputs
}

// Init Bubble Tea model
func (m model) Init() tea.Cmd {
	return tickCmd()
}

// Update method for Bubble Tea - for constant update loop
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
			m.nodestats = k8s.Nodes(m.args.metrics)
			return m, tea.ClearScreen
		}
	case tickMsg:
		m.nodestats = k8s.Nodes(m.args.metrics)
		return m, tea.Batch(tickCmd())

	}
	return m, nil

}


func GetBar(decider float64) (progress.Model) {

	decider = decider * 100

	var prog progress.Model
	// decide which color to use based on the usage percentage below 30% is green, above 70% is red, else yellow
	if  decider < 30 {
		prog = progress.New(progress.WithScaledGradient("#13B013","#1FE51F"))
	} else if decider > 70 {
		prog = progress.New(progress.WithScaledGradient("#13B013","#F11658"))
	} else {
		prog = progress.New(progress.WithScaledGradient("#13B013","#F18016"))
	}
	return prog
}

// View renders bubble tea
func (m model) View() string {
	
	var output strings.Builder
    
	DebugView(m, &output) // If debug on this would print Node and arg details

	SortByHandler(m)

	MetricsHandler(m, &output)
   
	output.WriteString("\n"+helpStyle("Press any key to quit"))
	
	return output.String()

}


// tickCmd returns a command that sends a tick every second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second * 1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}