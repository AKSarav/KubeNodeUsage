package main

import (
	"flag"
	"fmt"
	"kubenodeusage/cmd/nodemodel"
	"kubenodeusage/k8s"
	"kubenodeusage/utils"
	"os"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"

	tea "github.com/charmbracelet/bubbletea"
)

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
	fmt.Printf(displayfmt, "  --sortby", utils.PrintValidSorts())
	fmt.Printf(displayfmt, "  --filternodes", "filter based on node name")
	fmt.Printf(displayfmt, "  --filtercolor", "filter based on color category <30 Green, >30 <70 Orange, >70 Red")
	fmt.Printf(displayfmt, "  --filterlabel", "filter based on labels input should be key value pair in labelkey=labelvalue format")
	fmt.Printf(displayfmt, "  --desc", "to enable reverse sort")
	fmt.Printf(displayfmt, "  --debug", "enable debug mode")
	fmt.Printf(displayfmt, "  --metrics", utils.PrintValidMetrics())
	fmt.Printf(displayfmt, "  --label", "choose which label to display - syntax is labelname#alias here alias represents the column name to show in the output")
	fmt.Printf(displayfmt, "  --noinfo", "disable printing of cluster info")
	os.Exit(1)
}

// PrintArgs is used for Printing an arguments/*
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

	IsAllFiltersOn(args)	

	if args.FilterColor != "" {
		if !utils.IsValidColor(args.FilterColor) {
			fmt.Println("Not a valid color please choose one of the following colors", utils.PrintValidColors())
			os.Exit(2)
		}
	}

	if args.Metrics != "" {
		if !utils.IsValidMetric(args.Metrics) {
			fmt.Println("Not a valid Metric please choose one of", utils.PrintValidMetrics())
			os.Exit(2)
		}
	}


	if args.SortBy != "" {
		if !utils.IsValidSort(args.SortBy) {
			fmt.Println("Not a valid Sort by option please choose one of", utils.PrintValidSorts())
			os.Exit(2)
		}
	}

}



func IsAllFiltersOn(args *utils.Inputs) {

	var tempList []string
	tempList = append(tempList, args.FilterLabel, args.FilterNodes, args.FilterColor)
	filtersIntegrityValue := 0
	for _, filter := range tempList {
		if filter != ""{
			filtersIntegrityValue++
		}
	}
	if filtersIntegrityValue > 1 {
		fmt.Println("Only one filter can be used at a time")
		os.Exit(2)
	}
}

/*
Function: main
Description: main function
*/
func main() {

	// clearScreen()
	// parse command line arguments
	var (
		helpFlag     bool
		reverseFlag  bool
		debug        bool
		sortby       string
		filternodes  string
		filtercolor  string
		filterlabel string
		metrics      string
		label string
		lblAlias string
		noinfo bool
		pods bool
	)

	flag.BoolVar(&helpFlag, "help", false, "to display help")
	flag.BoolVar(&reverseFlag, "desc", false, "to display sort in descending order")
	flag.BoolVar(&debug, "debug", false, "enable debug mode")
	flag.StringVar(&sortby, "sortby", "name", "sort by name, free, capacity, usage")
	flag.StringVar(&filternodes, "filternodes", "", "filter nodes based on name")
	flag.StringVar(&filtercolor, "filtercolor", "", "filter nodes based on color")
	flag.StringVar(&filterlabel, "filterlabel", "", "filter nodes based on labels")
	flag.StringVar(&metrics, "metrics", "memory", "choose which metrics to display (memory, cpu, disk)")
	flag.StringVar(&label, "label", "", "choose which label to display")
	flag.BoolVar(&noinfo, "noinfo", false, "disable printing of cluster info")
	flag.BoolVar(&pods, "pods", false, "enable pod details")

	flag.Parse()

	if helpFlag {
		usage()
	}

	if debug {
		utils.InitLogger()
		utils.Logger.SetLevel(logrus.DebugLevel)
	}


	if (label != ""){
		if strings.Contains(label, "#") {
			lblAlias = strings.Split(label, "#")[1]
			label = strings.Split(label, "#")[0]
		} else {
			lblAlias = "label"
			label = strings.Split(label, "#")[0]
		}
	}

	args := utils.Inputs{
		HelpFlag:     helpFlag,
		ReverseFlag:  reverseFlag,
		Debug:        debug,
		SortBy:       sortby,
		FilterNodes:  filternodes,
		FilterColor:  filtercolor,
		FilterLabel: filterlabel,
		Metrics:      metrics,
		LabelToDisplay: label,
		LabelAlias: lblAlias,
		NoInfo: noinfo,
		Pods: pods,
	}

	checkinputs(&args) // sending the args using Address of Operator

	if debug {
		PrintArgs(args)
	}

	if pods {
		fmt.Println("Yet to implement")
		// KubePodUsage - For Pods
		// args.Pods = true
		// mdl := podusage{}
		// mdl.args = &args
		// mdl.clusterinfo = k8s.ClusterInfo()
		// mdl.podstats = k8s.Pods(&args)
		// if _, err := tea.NewProgram(mdl).Run(); err != nil {
		// 	fmt.Println("Oh no!", err)
		// 	os.Exit(1)
		// }
	} else{
		// KubeNodeUsage - For Nodes
		mdl := nodemodel.NodeUsage{}
		mdl.Args = &args
		mdl.ClusterInfo = k8s.ClusterInfo()
		mdl.Nodestats = k8s.Nodes(&args)
		
		if _, err := tea.NewProgram(mdl).Run(); err != nil {
			fmt.Println("Oh no!", err)
			os.Exit(1)
		}
	}
}
