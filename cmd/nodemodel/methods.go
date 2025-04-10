package nodemodel

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/AKSarav/KubeNodeUsage/v3/k8s"
	"github.com/AKSarav/KubeNodeUsage/v3/utils"

	"github.com/iancoleman/strcase"
)

func getUnit(metricType string) string {
	unit := ""
	switch metricType {
	case "memory":
		unit = "MB"
	case "cpu":
		unit = "Cores"
	case "disk":
		unit = "GB"
	}
	return unit
}

func DebugView(m NodeUsage, output *strings.Builder) {
	if m.Args.Debug {
		fmt.Fprint(output, " \nDebug mode enabled")
		fmt.Fprint(output, "\nArgs: ", m.Args)
		fmt.Fprint(output, "\nNodes: ", m.Nodestats)
	}
}

func RightMetric(m NodeUsage, index int) float32 {

	switch m.Args.Metrics {
	case "memory":
		if m.Args.SortBy == "free" {
			return float32(m.Nodestats[index].Free_memory)
		} else if m.Args.SortBy == "capacity" || m.Args.SortBy == "max" {
			return float32(m.Nodestats[index].Capacity_memory)
		} else if m.Args.SortBy == "color" || m.Args.SortBy == "usage" {
			return m.Nodestats[index].Usage_memory_percent
		}
	case "cpu":
		if m.Args.SortBy == "free" {
			return float32(m.Nodestats[index].Free_cpu)
		} else if m.Args.SortBy == "capacity" || m.Args.SortBy == "max" {
			return float32(m.Nodestats[index].Capacity_cpu)
		} else if m.Args.SortBy == "color" || m.Args.SortBy == "usage" {
			return m.Nodestats[index].Usage_cpu_percent
		}
	case "disk":
		if m.Args.SortBy == "free" {
			return float32(m.Nodestats[index].Free_disk)
		} else if m.Args.SortBy == "capacity" || m.Args.SortBy == "max" {
			return float32(m.Nodestats[index].Capacity_disk)
		} else if m.Args.SortBy == "color" || m.Args.SortBy == "usage" {
			return m.Nodestats[index].Usage_disk_percent
		}
	}
	// default return
	return m.Nodestats[index].Usage_memory_percent
}

func SortByHandler(m NodeUsage) {

	if m.Args.SortBy != "" && m.Args.SortBy != "name" && m.Args.SortBy != "node" {
		if !m.Args.ReverseFlag {
			sort.Slice(m.Nodestats, func(i, j int) bool {
				return RightMetric(m, i) < RightMetric(m, j)
			})
		} else {
			sort.Slice(m.Nodestats, func(i, j int) bool {
				return RightMetric(m, i) > RightMetric(m, j)
			})
		}
	} else if m.Args.SortBy != "name" || m.Args.SortBy != "node" {
		if !m.Args.ReverseFlag {
			sort.Slice(m.Nodestats, func(i, j int) bool {
				return m.Nodestats[i].Name < m.Nodestats[j].Name
			})
		} else {
			sort.Slice(m.Nodestats, func(i, j int) bool {
				return m.Nodestats[i].Name > m.Nodestats[j].Name
			})
		}
	}

}
func ApplyFilters(m NodeUsage) []k8s.Node {
	if m.Args.FilterLabel != "" {
		return FilterForLabel(m)
	} else if m.Args.FilterNodes != "" {
		return FilterForNode(m)
	} else if m.Args.FilterColor != "" {
		return FilterForColor(m)
	} else {
		return FilterForColor(m)
	}
}

func FilterForNode(m NodeUsage) []k8s.Node {
	var filteredNodes []k8s.Node
	FilterNodeInput := strings.Split(m.Args.FilterNodes, ",")

	// Creating a new map to store the values of NodeStats list
	// Choosing Map over Nested Array for comparision is best for TimeComplexity
	//NodesMap := make(map[string]k8s.Node)

	for _, node := range m.Nodestats {
		// NodesMap[node.Name] = node
		for _, FilteredNode := range FilterNodeInput {
			if matched, _ := regexp.MatchString(FilteredNode, node.Name); matched {
				filteredNodes = append(filteredNodes, node)
			}
		}
	}

	if len(filteredNodes) > 0 {
		utils.Logger.Debug("Filter For Node results", filteredNodes)
		m.Nodestats = filteredNodes
		return m.Nodestats
	} else {
		utils.Logger.Errorf("No matching Nodes found.. Exiting")
		os.Exit(2)
		return m.Nodestats
	}

}

func FilterForLabel(m NodeUsage) []k8s.Node {
	var filteredNodes []k8s.Node

	FilterKey := strings.Split(m.Args.FilterLabel, "=")[0]
	FilterValue := strings.Split(m.Args.FilterLabel, "=")[1]

	if FilterKey == "" || FilterValue == "" {
		utils.Logger.Errorf("Filter Key or Value is empty.. Exiting")
		os.Exit(2)
	}

	for _, node := range m.Nodestats {
		if _, ok := node.Labels[FilterKey]; ok {
			if node.Labels[FilterKey] == FilterValue {
				filteredNodes = append(filteredNodes, node)
			}
		}
	}

	if len(filteredNodes) > 0 {
		utils.Logger.Debug("Filter For Label results", filteredNodes)
		m.Nodestats = filteredNodes
		return m.Nodestats
	} else {
		utils.Logger.Errorf("No matching Nodes found.. Exiting")
		os.Exit(2)
		return m.Nodestats
	}
}

func FilterForColor(m NodeUsage) []k8s.Node {
	utils.Logger.Debug("Filter for Color called")
	var filteredNodes []k8s.Node
	var thresholdMin, thresholdMax float64

	// Define the color threshold values
	switch m.Args.FilterColor {
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

	// Filter nodes based on metric and threshold values
	for _, node := range m.Nodestats {
		var usagepercent float64
		switch m.Args.Metrics {
		case "memory":
			usagepercent = float64(node.Usage_memory_percent) / 100.0
		case "cpu":
			usagepercent = float64(node.Usage_cpu_percent) / 100.0
		case "disk":
			usagepercent = float64(node.Usage_disk_percent) / 100.0
		default:
			if m.Args.Debug {
				fmt.Println("No Matching Metric", m.Args.Metrics)
			}
		}

		if (usagepercent*100) >= thresholdMin && (usagepercent*100) < thresholdMax {
			filteredNodes = append(filteredNodes, node)
		}
	}
	if m.Args.Debug {
		fmt.Println("Filter For Color result:", filteredNodes)
	}
	return filteredNodes

}

func headlinePrinter(m *NodeUsage, output *strings.Builder, Nodes *[]k8s.Node, maxNameWidth *int) {
	unit := getUnit(m.Args.Metrics)
	freeHeading := "Free(" + unit + ")"
	maxHeading := "Max(" + unit + ")"
	uptimeHeading := "Uptime"
	statusHeading := "Status"

	if m.Args.LabelToDisplay != "" {
		m.Format = "%-" + strconv.Itoa(*maxNameWidth) + "s %-10s %-10s %-5s %-8s %-10s %-15s %-30s\n"
		fmt.Fprintf(output, m.Format, "Name", freeHeading, maxHeading, "Pods", uptimeHeading, statusHeading, m.Args.LabelAlias, "Usage%")
	} else {
		m.Format = "%-" + strconv.Itoa(*maxNameWidth) + "s %-10s %-10s %-5s %-8s %-10s %-30s\n"
		fmt.Fprintf(output, m.Format, "Name", freeHeading, maxHeading, "Pods", uptimeHeading, statusHeading, "Usage%")
	}
}

func PrintDesign(output *strings.Builder, maxNameWidth int) {
	if maxNameWidth < 30 {
		maxNameWidth = 30
	}
	lines := strings.Repeat("-", maxNameWidth+83)
	fmt.Fprint(output, lines)
	fmt.Fprint(output, "\n")
}

func MetricsHandler(m NodeUsage, output *strings.Builder) {

	// Nodes Filtering based on filters
	filteredNodes := ApplyFilters(m)

	m.Nodestats = filteredNodes
	SortByHandler(m)

	// decide formatting and Maximum width
	maxNameWidth := 30
	for _, node := range filteredNodes {
		if maxNameWidth < len(node.Name) {
			maxNameWidth = len(node.Name)
		}
	}
	// Header and Version info
	fmt.Fprintf(output, "\n# KubeNodeUsage\n# Version: %s\n# https://github.com/AKSarav/KubeNodeUsage\n\n", utils.Version)

	if !m.Args.NoInfo {
		fmt.Fprint(output, "\n# Context: ", m.ClusterInfo.Context, "\n# Version: ", m.ClusterInfo.Version, "\n# URL: ", m.ClusterInfo.URL, "\n\n")
	}

	fmt.Fprint(output, "# ", strcase.ToCamel(m.Args.Metrics), " Metrics\n\n")
	headlinePrinter(&m, output, &filteredNodes, &maxNameWidth)
	PrintDesign(output, maxNameWidth)

	if m.Args.Metrics == "memory" {
		for _, node := range filteredNodes {
			prog := GetBar(float64(node.Usage_memory_percent) / 100.0)
			if m.Args.LabelToDisplay != "" {
				fmt.Fprintf(output, m.Format,
					node.Name,
					strconv.Itoa(node.Free_memory/1024),
					strconv.Itoa(node.Capacity_memory/1024),
					node.TotalPods,
					node.Uptime,
					node.Status,
					node.LabelToDisplay,
					prog.ViewAs(float64(node.Usage_memory_percent)/100.0))
			} else {
				fmt.Fprintf(output, m.Format,
					node.Name,
					strconv.Itoa(node.Free_memory/1024),
					strconv.Itoa(node.Capacity_memory/1024),
					node.TotalPods,
					node.Uptime,
					node.Status,
					prog.ViewAs(float64(node.Usage_memory_percent)/100.0))
			}
		}
	} else if m.Args.Metrics == "cpu" {
		for _, node := range filteredNodes {
			prog := GetBar(float64(node.Usage_cpu_percent) / 100.0)
			if m.Args.LabelToDisplay != "" {
				fmt.Fprintf(output, m.Format,
					node.Name,
					strconv.Itoa(int(node.Free_cpu)),
					strconv.Itoa(node.Capacity_cpu),
					node.TotalPods,
					node.Uptime,
					node.Status,
					node.LabelToDisplay,
					prog.ViewAs(float64(node.Usage_cpu_percent)/100.0))
			} else {
				fmt.Fprintf(output, m.Format,
					node.Name,
					strconv.Itoa(int(node.Free_cpu)),
					strconv.Itoa(node.Capacity_cpu),
					node.TotalPods,
					node.Uptime,
					node.Status,
					prog.ViewAs(float64(node.Usage_cpu_percent)/100.0))
			}
		}
	} else if m.Args.Metrics == "disk" {
		for _, node := range filteredNodes {
			prog := GetBar(float64(node.Usage_disk_percent) / 100.0)
			// Convert bytes to GB (1 GB = 1024^3 bytes)
			gbDivisor := 1024 * 1024 * 1024
			if m.Args.LabelToDisplay != "" {
				fmt.Fprintf(output, m.Format,
					node.Name,
					fmt.Sprintf("%.1f", float64(node.Free_disk)/float64(gbDivisor)),
					fmt.Sprintf("%.1f", float64(node.Capacity_disk)/float64(gbDivisor)),
					node.TotalPods,
					node.Uptime,
					node.Status,
					node.LabelToDisplay,
					prog.ViewAs(float64(node.Usage_disk_percent)/100.0))
			} else {
				fmt.Fprintf(output, m.Format,
					node.Name,
					fmt.Sprintf("%.1f", float64(node.Free_disk)/float64(gbDivisor)),
					fmt.Sprintf("%.1f", float64(node.Capacity_disk)/float64(gbDivisor)),
					node.TotalPods,
					node.Uptime,
					node.Status,
					prog.ViewAs(float64(node.Usage_disk_percent)/100.0))
			}
		}
	}
}
