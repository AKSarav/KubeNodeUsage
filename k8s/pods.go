package k8s

import (
	"context"
	"fmt"
	"kubenodeusage/utils"
	"os"
	"path/filepath"

	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Pod struct {
	Name                 string
	Namespace            string
	NodeName             string
	Capacity_memory      int
	Capacity_cpu         int
	Usage_memory         int
	Usage_cpu            float32
	Request_memory       int
	Request_cpu          float32
	Limit_memory         int
	Limit_cpu            float32
	Usage_memory_percent float32
	Usage_cpu_percent    float32
	Usage_disk           int64   // Total disk usage in bytes
	Usage_disk_percent   float32 // Disk usage percentage
	Status               string
	LabelToDisplay       string
	Labels               map[string]string
}

var PodStatsList []Pod

func Pods(inputs *utils.Inputs) (PodStatsList []Pod) {
	metric := inputs.Metrics

	utils.InitLogger()
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println("Not able to access .kube/config file from the Home Directory path: ", kubeconfig)
		os.Exit(2)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	mc, err := metricsv.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// To fetch kubectl top pods metrics
	podMetrics, err := mc.MetricsV1beta1().PodMetricses("").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Println("Unable to Get PodMetrics. Is Metrics Server running?")
		os.Exit(2)
	}

	// To fetch kubectl get pods information
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Println("Failed to Get Pods")
		panic(err.Error())
	}

	// To fetch node information for capacity context
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Println("Failed to Get Nodes")
		panic(err.Error())
	}

	// Create a map of node names to node objects
	nodeMap := make(map[string]*core.Node)
	for i := range nodes.Items {
		nodeMap[nodes.Items[i].Name] = &nodes.Items[i]
	}

	// Parsing Every Pod and collecting information
	for _, pm := range podMetrics.Items {
		for _, pod := range pods.Items {
			if pod.Name == pm.Name && pod.Namespace == pm.Namespace {
				podstats := Pod{}
				podstats.Name = pod.Name
				podstats.Namespace = pod.Namespace
				podstats.NodeName = pod.Spec.NodeName
				podstats.Status = string(pod.Status.Phase)

				// Get pod container metrics
				switch metric {
				case "memory":
					// Calculate total memory usage and capacity
					var totalMemUsage int64 = 0
					var totalMemRequest int64 = 0
					var totalMemLimit int64 = 0

					for _, container := range pm.Containers {
						memUsage, _ := container.Usage.Memory().AsInt64()
						totalMemUsage += memUsage
					}

					for _, container := range pod.Spec.Containers {
						if container.Resources.Requests.Memory() != nil {
							memReq, _ := container.Resources.Requests.Memory().AsInt64()
							totalMemRequest += memReq
						}

						if container.Resources.Limits.Memory() != nil {
							memLimit, _ := container.Resources.Limits.Memory().AsInt64()
							totalMemLimit += memLimit
						}
					}

					// Convert to MB (from bytes)
					podstats.Usage_memory = int(totalMemUsage / (1024 * 1024))
					podstats.Request_memory = int(totalMemRequest / (1024 * 1024))
					podstats.Limit_memory = int(totalMemLimit / (1024 * 1024))

					// If pod is on a node, get node capacity for context
					if node, exists := nodeMap[pod.Spec.NodeName]; exists {
						nodeCap, _ := node.Status.Capacity.Memory().AsInt64()
						podstats.Capacity_memory = int(nodeCap / (1024 * 1024))
					}

					// Calculate percentage based on limit or node capacity
					if podstats.Limit_memory > 0 {
						podstats.Usage_memory_percent = float32(podstats.Usage_memory) / float32(podstats.Limit_memory) * 100
					} else if podstats.Capacity_memory > 0 {
						podstats.Usage_memory_percent = float32(podstats.Usage_memory) / float32(podstats.Capacity_memory) * 100
					}

				case "cpu":
					// Calculate total CPU usage and capacity
					var totalCpuUsage float32 = 0
					var totalCpuRequest float32 = 0
					var totalCpuLimit float32 = 0

					for _, container := range pm.Containers {
						cpuUsage := float32(container.Usage.Cpu().MilliValue())
						totalCpuUsage += cpuUsage
					}

					for _, container := range pod.Spec.Containers {
						if container.Resources.Requests.Cpu() != nil {
							cpuReq := float32(container.Resources.Requests.Cpu().MilliValue())
							totalCpuRequest += cpuReq
						}

						if container.Resources.Limits.Cpu() != nil {
							cpuLimit := float32(container.Resources.Limits.Cpu().MilliValue())
							totalCpuLimit += cpuLimit
						}
					}

					// Convert to cores (from millicores)
					podstats.Usage_cpu = totalCpuUsage / 1000
					podstats.Request_cpu = totalCpuRequest / 1000
					podstats.Limit_cpu = totalCpuLimit / 1000

					// If pod is on a node, get node capacity for context
					if node, exists := nodeMap[pod.Spec.NodeName]; exists {
						nodeCap := node.Status.Capacity.Cpu().MilliValue()
						podstats.Capacity_cpu = int(nodeCap)
					}

					// Calculate percentage based on limit or node capacity
					if podstats.Limit_cpu > 0 {
						podstats.Usage_cpu_percent = (podstats.Usage_cpu / podstats.Limit_cpu) * 100
					} else if podstats.Capacity_cpu > 0 {
						podstats.Usage_cpu_percent = (podstats.Usage_cpu / float32(podstats.Capacity_cpu/1000)) * 100
					}

				case "disk":
					// Calculate total disk usage
					var totalDiskUsage int64 = 0

					for _, container := range pm.Containers {
						// Get filesystem stats if available
						if stats := container.Usage.StorageEphemeral(); stats != nil {
							diskUsage, _ := stats.AsInt64()
							totalDiskUsage += diskUsage
						}
					}

					// Convert to MB and store
					podstats.Usage_disk = totalDiskUsage

					// If pod is on a node, try to get node storage capacity for context
					if node, exists := nodeMap[pod.Spec.NodeName]; exists {
						if storage, ok := node.Status.Capacity["ephemeral-storage"]; ok {
							storageCapacity, _ := storage.AsInt64()
							if storageCapacity > 0 {
								podstats.Usage_disk_percent = float32(totalDiskUsage) / float32(storageCapacity) * 100
							}
						}
					}
				}

				// Display Label if provided
				if inputs.LabelToDisplay != "" {
					if _, ok := pod.Labels[inputs.LabelToDisplay]; !ok {
						podstats.LabelToDisplay = "Not Found"
					} else {
						podstats.LabelToDisplay = pod.Labels[inputs.LabelToDisplay]
					}
				}

				// Collect all labels
				podstats.Labels = pod.Labels

				PodStatsList = append(PodStatsList, podstats)
			}
		}
	}

	utils.Logger.Debug(PodStatsList)
	return PodStatsList
}
