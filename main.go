package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/AKSarav/KubeNodeUsage/v3/cmd/nodemodel"
	"github.com/AKSarav/KubeNodeUsage/v3/cmd/podmodel"
	"github.com/AKSarav/KubeNodeUsage/v3/utils"

	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"

	tea "github.com/charmbracelet/bubbletea"
)

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
	fmt.Printf(displayfmt, "  --sortby", utils.PrintValidSorts())
	fmt.Printf(displayfmt, "  --filternodes", "filter based on node name")
	fmt.Printf(displayfmt, "  --filtercolor", "filter based on color category <30 Green, >30 <70 Orange, >70 Red")
	fmt.Printf(displayfmt, "  --filterlabel", "filter based on labels input should be key value pair in labelkey=labelvalue format")
	fmt.Printf(displayfmt, "  --desc", "to enable reverse sort")
	fmt.Printf(displayfmt, "  --debug", "enable debug mode")
	fmt.Printf(displayfmt, "  --metrics", utils.PrintValidMetrics())
	fmt.Printf(displayfmt, "  --label", "choose which label to display - syntax is labelname#alias here alias represents the column name to show in the output")
	fmt.Printf(displayfmt, "  --noinfo", "disable printing of cluster info")
	fmt.Printf(displayfmt, "  --pods", "show pod usage instead of node usage")
	os.Exit(1)
}

// PrintArgs is used for Printing an arguments
func PrintArgs(args utils.Inputs) {
	// print key value pairs
	t := reflect.TypeOf(args)
	v := reflect.ValueOf(args)

	if args.Debug {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)
			fmt.Printf("%s: %v\n", field.Name, value)
		}
	}
}

func checkinputs(args *utils.Inputs) {
	// Check if help is requested
	if args.Help {
		usage()
	}

	// Check if metrics is valid
	if !utils.IsValidMetric(args.Metrics) {
		utils.Logger.Error("Invalid metric: ", args.Metrics)
		usage()
	}

	// Check if sortby is valid
	if args.SortBy != "" && !utils.IsValidSort(args.SortBy) {
		utils.Logger.Error("Invalid sort: ", args.SortBy)
		usage()
	}

	// Check if filtercolor is valid
	if args.FilterColor != "" && !utils.IsValidColor(args.FilterColor) {
		utils.Logger.Error("Invalid color: ", args.FilterColor)
		usage()
	}

	// Check if all filters are on
	IsAllFiltersOn(args)

	// Set log level
	if args.Debug {
		utils.Logger.SetLevel(logrus.DebugLevel)
	}

	// Set label alias if not set
	if args.LabelToDisplay != "" {
		if !strings.Contains(args.LabelToDisplay, "#") {
			args.LabelAlias = strcase.ToCamel(args.LabelToDisplay)
		} else {
			args.LabelAlias = strings.Split(args.LabelToDisplay, "#")[1]
			args.LabelToDisplay = strings.Split(args.LabelToDisplay, "#")[0]
		}
	}
}

func IsAllFiltersOn(args *utils.Inputs) {
	if args.FilterLabel != "" && args.FilterNodes != "" && args.FilterColor != "" {
		utils.Logger.Error("Only one filter can be used at a time")
		os.Exit(2)
	}
}

func main() {
	var args utils.Inputs

	// Initialize logger
	utils.InitLogger()

	// Flags
	flag.StringVar(&args.Metrics, "metrics", "memory", "Metrics to display")
	flag.StringVar(&args.SortBy, "sortby", "", "Sort by field")
	flag.StringVar(&args.FilterNodes, "filternodes", "", "Filter nodes")
	flag.StringVar(&args.FilterColor, "filtercolor", "", "Filter by color")
	flag.StringVar(&args.FilterLabel, "filterlabel", "", "Filter by label")
	flag.StringVar(&args.LabelToDisplay, "label", "", "Label to display")
	flag.BoolVar(&args.ReverseFlag, "desc", false, "Reverse sort")
	flag.BoolVar(&args.Debug, "debug", false, "Debug mode")
	flag.BoolVar(&args.NoInfo, "noinfo", false, "No info")
	flag.BoolVar(&args.Pods, "pods", false, "Show pods")
	flag.BoolVar(&args.Help, "help", false, "Help")
	flag.Parse()

	// Check inputs
	checkinputs(&args)

	// Print args if debug is enabled
	PrintArgs(args)

	// Initialize the appropriate model based on the --pods flag
	var mdl tea.Model
	if args.Pods {
		mdl = podmodel.NewPodUsage(&args)
	} else {
		mdl = nodemodel.NewNodeUsage(&args)
	}

	// Run the program
	p := tea.NewProgram(mdl)
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
