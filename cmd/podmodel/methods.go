package podmodel

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
	}
	return unit
}

func DebugView(m PodUsage, output *strings.Builder) {
	if m.Args.Debug {
		fmt.Fprint(output, " \nDebug mode enabled")
		fmt.Fprint(output, "\nArgs: ", m.Args)
		fmt.Fprint(output, "\nPods: ", m.Podstats)
	}
}

func RightMetric(m PodUsage, index int) float32 {
	switch m.Args.Metrics {
	case "memory":
		if m.Args.SortBy == "free" {
			if m.Podstats[index].Limit_memory > 0 {
				return float32(m.Podstats[index].Limit_memory - m.Podstats[index].Usage_memory)
			}
			return float32(m.Podstats[index].Capacity_memory - m.Podstats[index].Usage_memory)
		} else if m.Args.SortBy == "capacity" || m.Args.SortBy == "max" {
			return float32(m.Podstats[index].Capacity_memory)
		} else if m.Args.SortBy == "limit" {
			return float32(m.Podstats[index].Limit_memory)
		} else if m.Args.SortBy == "request" {
			return float32(m.Podstats[index].Request_memory)
		} else if m.Args.SortBy == "color" || m.Args.SortBy == "usage" {
			return m.Podstats[index].Usage_memory_percent
		}
	case "cpu":
		if m.Args.SortBy == "free" {
			if m.Podstats[index].Limit_cpu > 0 {
				return m.Podstats[index].Limit_cpu - m.Podstats[index].Usage_cpu
			}
			return float32(m.Podstats[index].Capacity_cpu/1000) - m.Podstats[index].Usage_cpu
		} else if m.Args.SortBy == "capacity" || m.Args.SortBy == "max" {
			return float32(m.Podstats[index].Capacity_cpu)
		} else if m.Args.SortBy == "limit" {
			return m.Podstats[index].Limit_cpu
		} else if m.Args.SortBy == "request" {
			return m.Podstats[index].Request_cpu
		} else if m.Args.SortBy == "color" || m.Args.SortBy == "usage" {
			return m.Podstats[index].Usage_cpu_percent
		}
	}
	// default return
	return m.Podstats[index].Usage_memory_percent
}

func SortByHandler(m PodUsage) {
	if m.Args.SortBy != "" && m.Args.SortBy != "name" && m.Args.SortBy != "pod" && m.Args.SortBy != "namespace" {
		if !m.Args.ReverseFlag {
			sort.Slice(m.Podstats, func(i, j int) bool {
				return RightMetric(m, i) < RightMetric(m, j)
			})
		} else {
			sort.Slice(m.Podstats, func(i, j int) bool {
				return RightMetric(m, i) > RightMetric(m, j)
			})
		}
	} else if m.Args.SortBy == "namespace" {
		if !m.Args.ReverseFlag {
			sort.Slice(m.Podstats, func(i, j int) bool {
				return m.Podstats[i].Namespace < m.Podstats[j].Namespace
			})
		} else {
			sort.Slice(m.Podstats, func(i, j int) bool {
				return m.Podstats[i].Namespace > m.Podstats[j].Namespace
			})
		}
	} else {
		if !m.Args.ReverseFlag {
			sort.Slice(m.Podstats, func(i, j int) bool {
				return m.Podstats[i].Name < m.Podstats[j].Name
			})
		} else {
			sort.Slice(m.Podstats, func(i, j int) bool {
				return m.Podstats[i].Name > m.Podstats[j].Name
			})
		}
	}
}

func ApplyFilters(m PodUsage) []k8s.Pod {
	if m.Args.FilterLabel != "" {
		return FilterForLabel(m)
	} else if m.Args.FilterNodes != "" {
		return FilterForNode(m)
	} else if m.Args.FilterColor != "" {
		return FilterForColor(m)
	} else {
		return m.Podstats
	}
}

func FilterForNode(m PodUsage) []k8s.Pod {
	var filteredPods []k8s.Pod
	FilterNodeInput := strings.Split(m.Args.FilterNodes, ",")

	for _, pod := range m.Podstats {
		for _, FilteredNode := range FilterNodeInput {
			if matched, _ := regexp.MatchString(FilteredNode, pod.NodeName); matched {
				filteredPods = append(filteredPods, pod)
				break
			}
			// Also filter by pod name if that's what user wanted
			if matched, _ := regexp.MatchString(FilteredNode, pod.Name); matched {
				filteredPods = append(filteredPods, pod)
				break
			}
		}
	}

	if len(filteredPods) > 0 {
		utils.Logger.Debug("Filter For Node results", filteredPods)
		m.Podstats = filteredPods
		return m.Podstats
	} else {
		utils.Logger.Errorf("No matching Pods found.. Exiting")
		os.Exit(2)
		return m.Podstats
	}
}

func FilterForLabel(m PodUsage) []k8s.Pod {
	var filteredPods []k8s.Pod

	FilterKey := strings.Split(m.Args.FilterLabel, "=")[0]
	FilterValue := strings.Split(m.Args.FilterLabel, "=")[1]

	if FilterKey == "" || FilterValue == "" {
		utils.Logger.Errorf("Filter Key or Value is empty.. Exiting")
		os.Exit(2)
	}

	for _, pod := range m.Podstats {
		if _, ok := pod.Labels[FilterKey]; ok {
			if pod.Labels[FilterKey] == FilterValue {
				filteredPods = append(filteredPods, pod)
			}
		}
	}

	if len(filteredPods) > 0 {
		utils.Logger.Debug("Filter For Label results", filteredPods)
		m.Podstats = filteredPods
		return m.Podstats
	} else {
		utils.Logger.Errorf("No matching Pods found.. Exiting")
		os.Exit(2)
		return m.Podstats
	}
}

func FilterForColor(m PodUsage) []k8s.Pod {
	utils.Logger.Debug("Filter for Color called")
	var filteredPods []k8s.Pod
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

	// Filter pods based on metric and threshold values
	for _, pod := range m.Podstats {
		var usagepercent float64
		switch m.Args.Metrics {
		case "memory":
			usagepercent = float64(pod.Usage_memory_percent) / 100.0
		case "cpu":
			usagepercent = float64(pod.Usage_cpu_percent) / 100.0
		default:
			if m.Args.Debug {
				fmt.Println("No Matching Metric", m.Args.Metrics)
			}
		}

		if (usagepercent*100) >= thresholdMin && (usagepercent*100) < thresholdMax {
			filteredPods = append(filteredPods, pod)
		}
	}
	if m.Args.Debug {
		fmt.Println("Filter For Color result:", filteredPods)
	}
	return filteredPods
}

func PrintDesign(output *strings.Builder, maxNameWidth int, maxNsWidth int) {
	output.WriteString(strings.Repeat("-", maxNameWidth+maxNsWidth+60) + "\n")
}

func headlinePrinter(m *PodUsage, output *strings.Builder, Pods *[]k8s.Pod, maxNameWidth *int, maxNsWidth *int) {
	unit := getUnit(m.Args.Metrics)
	usageHeading := "Usage(" + unit + ")"
	requestHeading := "Request(" + unit + ")"
	limitHeading := "Limit(" + unit + ")"

	// Adjust format based on label display
	if m.Args.LabelToDisplay != "" {
		m.Format = "%-" + strconv.Itoa(*maxNameWidth) + "s %-" + strconv.Itoa(*maxNsWidth) + "s %-20s %-10s %-10s %-10s %-12s %s\n"
		values := []interface{}{"Name", "Namespace", "Node", usageHeading, requestHeading, limitHeading, m.Args.LabelAlias, "Usage%"}
		fmt.Fprintf(output, m.Format, values...)
	} else {
		m.Format = "%-" + strconv.Itoa(*maxNameWidth) + "s %-" + strconv.Itoa(*maxNsWidth) + "s %-20s %-10s %-10s %-10s %s\n"
		values := []interface{}{"Name", "Namespace", "Node", usageHeading, requestHeading, limitHeading, "Usage%"}
		fmt.Fprintf(output, m.Format, values...)
	}
}

func MetricsHandler(m PodUsage, output *strings.Builder) {
	// Pods Filtering based on filters
	filteredPods := ApplyFilters(m)

	m.Podstats = filteredPods
	SortByHandler(m)

	// decide formatting and Maximum width
	maxNameWidth := 15
	maxNsWidth := 12
	for _, pod := range filteredPods {
		if maxNameWidth < len(pod.Name) {
			maxNameWidth = len(pod.Name)
		}
		if maxNsWidth < len(pod.Namespace) {
			maxNsWidth = len(pod.Namespace)
		}
	}

	// Allow for reasonable padding
	maxNameWidth += 2
	maxNsWidth += 2

	// Header and Version info
	fmt.Fprint(output, "\n# KubeNodeUsage - Pod View\n# Version: 3.0.2\n# https://github.com/AKSarav/Kube-Node-Usage\n\n")

	if !m.Args.NoInfo {
		fmt.Fprint(output, "\n# Context: ", m.ClusterInfo.Context, "\n# Version: ", m.ClusterInfo.Version, "\n# URL: ", m.ClusterInfo.URL, "\n\n")
	}

	fmt.Fprint(output, "# ", strcase.ToCamel(m.Args.Metrics), " Metrics for Pods\n\n")

	headlinePrinter(&m, output, &filteredPods, &maxNameWidth, &maxNsWidth)
	PrintDesign(output, maxNameWidth, maxNsWidth)

	if m.Args.Metrics == "memory" {
		for _, pod := range filteredPods {
			prog := GetBar(float64(pod.Usage_memory_percent) / 100.0)

			// Truncate node name if too long
			nodeName := pod.NodeName
			if len(nodeName) > 20 {
				nodeName = nodeName[:9] + "…"
			}

			if m.Args.LabelToDisplay != "" {
				values := []interface{}{
					pod.Name,
					pod.Namespace,
					nodeName,
					strconv.Itoa(pod.Usage_memory),
					strconv.Itoa(pod.Request_memory),
					strconv.Itoa(pod.Limit_memory),
					pod.LabelToDisplay,
					prog.ViewAs(float64(pod.Usage_memory_percent) / 100.0),
				}
				fmt.Fprintf(output, m.Format, values...)
			} else {
				values := []interface{}{
					pod.Name,
					pod.Namespace,
					nodeName,
					strconv.Itoa(pod.Usage_memory),
					strconv.Itoa(pod.Request_memory),
					strconv.Itoa(pod.Limit_memory),
					prog.ViewAs(float64(pod.Usage_memory_percent) / 100.0),
				}
				fmt.Fprintf(output, m.Format, values...)
			}
		}
	} else if m.Args.Metrics == "cpu" {
		for _, pod := range filteredPods {
			prog := GetBar(float64(pod.Usage_cpu_percent) / 100.0)

			// Truncate node name if too long
			nodeName := pod.NodeName
			if len(nodeName) > 20 {
				nodeName = nodeName[:9] + "…"
			}

			if m.Args.LabelToDisplay != "" {
				values := []interface{}{
					pod.Name,
					pod.Namespace,
					nodeName,
					fmt.Sprintf("%.2f", pod.Usage_cpu),
					fmt.Sprintf("%.2f", pod.Request_cpu),
					fmt.Sprintf("%.2f", pod.Limit_cpu),
					pod.LabelToDisplay,
					prog.ViewAs(float64(pod.Usage_cpu_percent) / 100.0),
				}
				fmt.Fprintf(output, m.Format, values...)
			} else {
				values := []interface{}{
					pod.Name,
					pod.Namespace,
					nodeName,
					fmt.Sprintf("%.2f", pod.Usage_cpu),
					fmt.Sprintf("%.2f", pod.Request_cpu),
					fmt.Sprintf("%.2f", pod.Limit_cpu),
					prog.ViewAs(float64(pod.Usage_cpu_percent) / 100.0),
				}
				fmt.Fprintf(output, m.Format, values...)
			}
		}
	} else if m.Args.Metrics == "disk" {
		for _, pod := range filteredPods {
			prog := GetBar(float64(pod.Usage_disk_percent) / 100.0)

			// Truncate node name if too long
			nodeName := pod.NodeName
			if len(nodeName) > 20 {
				nodeName = nodeName[:9] + "…"
			}

			if m.Args.LabelToDisplay != "" {
				values := []interface{}{
					pod.Name,
					pod.Namespace,
					nodeName,
					fmt.Sprintf("%.2f", pod.Usage_cpu),
					fmt.Sprintf("%.2f", pod.Request_cpu),
					fmt.Sprintf("%.2f", pod.Limit_cpu),
					pod.LabelToDisplay,
					prog.ViewAs(float64(pod.Usage_cpu_percent) / 100.0),
				}
				fmt.Fprintf(output, m.Format, values...)
			} else {
				values := []interface{}{
					pod.Name,
					pod.Namespace,
					nodeName,
					fmt.Sprintf("%.2f", pod.Usage_cpu),
					fmt.Sprintf("%.2f", pod.Request_cpu),
					fmt.Sprintf("%.2f", pod.Limit_cpu),
					prog.ViewAs(float64(pod.Usage_cpu_percent) / 100.0),
				}
				fmt.Fprintf(output, m.Format, values...)
			}
		}
	}

}
