package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AKSarav/KubeNodeUsage/v3/utils"

	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Node struct {
	Name                 string
	Capacity_disk        int
	Capacity_memory      int
	Capacity_cpu         int
	Usage_disk           int
	Usage_memory         int
	Usage_cpu            float32
	Free_disk            int
	Free_memory          int
	Free_cpu             float32
	Usage_disk_percent   float32
	Usage_memory_percent float32
	Usage_cpu_percent    float32
	TotalPods            string
	LabelToDisplay       string
	Labels               map[string]string
	Uptime               string
	Status               string
}

type Cluster struct {
	Context string
	Version string
	URL     string
}

var NodeStatsList []Node
var K8sinfo Cluster

// KubeletStats represents the structure returned by /stats/summary
type KubeletStats struct {
	Node struct {
		NodeName         string `json:"nodeName"`
		SystemContainers []struct {
			Name      string `json:"name"`
			UsedBytes int64  `json:"usedBytes"`
		} `json:"systemContainers"`
		Runtime struct {
			ImageFs struct {
				UsedBytes     int64 `json:"usedBytes"`
				CapacityBytes int64 `json:"capacityBytes"`
			} `json:"imageFs"`
		} `json:"runtime"`
		Fs struct {
			UsedBytes     int64 `json:"usedBytes"`
			CapacityBytes int64 `json:"capacityBytes"`
		} `json:"fs"`
	} `json:"node"`
	Pods []struct {
		PodRef struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"podRef"`
		Containers []struct {
			Name   string `json:"name"`
			Rootfs struct {
				UsedBytes     int64 `json:"usedBytes"`
				CapacityBytes int64 `json:"capacityBytes"`
			} `json:"rootfs"`
			Logs struct {
				UsedBytes     int64 `json:"usedBytes"`
				CapacityBytes int64 `json:"capacityBytes"`
			} `json:"logs"`
		} `json:"containers"`
		EphemeralStorage struct {
			UsedBytes     int64 `json:"usedBytes"`
			CapacityBytes int64 `json:"capacityBytes"`
		} `json:"ephemeral-storage"`
		VolumeStats []struct {
			Name    string `json:"name"`
			FsStats struct {
				UsedBytes     int64 `json:"usedBytes"`
				CapacityBytes int64 `json:"capacityBytes"`
			} `json:"fs"`
		} `json:"volume-stats"`
	} `json:"pods"`
}

func ClusterInfo() Cluster {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	confvar := clientcmd.GetConfigFromFileOrDie(kubeconfig)

	K8sinfo := Cluster{}
	K8sinfo.Context = confvar.CurrentContext

	utils.InitLogger()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println("Not able to access .kube/config file from the Home Directory path: ", kubeconfig)
		os.Exit(2)
	}

	mc, err := metricsv.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	K8sinfo.URL = config.Host

	// Validate Version of Server
	if version, err := mc.ServerVersion(); err != nil {
		fmt.Println("\n# ERROR: Unable to Establish Connection to Kubernetes Cluster")
		fmt.Println("# Kubernetes Context:", K8sinfo.Context)
		fmt.Println("# Kubernetes URL:", K8sinfo.URL)
		fmt.Println("# Please check your kubernetes configuration and permissions\n")
		os.Exit(2)
	} else {
		K8sinfo.Version = version.String()
	}

	return K8sinfo

}

// This Go function takes in node statistics, node information, node metrics, a specific metric, and
// returns an array of nodes.
// responsible for collecting memory, cpu, and disk statistics for each node
func GetMetricsForNode(nodestats *Node, node *core.Node, nm *v1beta1.NodeMetrics, metric string, clientset *kubernetes.Clientset) []Node {

	NodeMetrics := []Node{}

	switch metric {
	case "memory":
		// Ki - Kibibyte - 1024 bytes
		memcapcity, err := strconv.Atoi(strings.TrimSuffix(node.Status.Capacity.Memory().String(), "Ki"))
		if err != nil {
			fmt.Println("Error converting Memory capacity")
		} else {
			nodestats.Capacity_memory = memcapcity
		}

		// output is in Ki - Kibibyte - 1024 bytes
		if memusage, err := strconv.Atoi(strings.TrimSuffix(nm.Usage.Memory().String(), "Ki")); err != nil {
			// Try with Mi - Mebibyte - 1024 * 1024 bytes
			if memusage, err := strconv.Atoi(strings.TrimSuffix(nm.Usage.Memory().String(), "Mi")); err != nil {
				fmt.Println("Error converting Memory usage", err)
			} else {
				nodestats.Usage_memory = memusage
				nodestats.Free_memory = nodestats.Capacity_memory - nodestats.Usage_memory
			}
		} else {
			nodestats.Usage_memory = memusage
			nodestats.Free_memory = nodestats.Capacity_memory - nodestats.Usage_memory
		}

		nodestats.Usage_memory_percent = float32(nodestats.Usage_memory) / float32(nodestats.Capacity_memory) * 100

		NodeMetrics = append(NodeMetrics, *nodestats)

	case "cpu":
		Capacity_cpu, err := strconv.Atoi(node.Status.Capacity.Cpu().String())
		// Converting to millicore 1 CPU 1000 millicore
		nodestats.Capacity_cpu = Capacity_cpu * 1000
		if err != nil {
			fmt.Println("Error converting CPU capacity")
		}
		// fmt.Println("Capacity CPU:", nodestats.Capacity_cpu * 1000)

		cpu_in_nanocore, err := strconv.ParseFloat(strings.TrimSuffix(nm.Usage.Cpu().String(), "n"), 32)
		if err == nil {
			cpu_in_millicore := cpu_in_nanocore / 1000000
			nodestats.Usage_cpu = float32(cpu_in_millicore)
			nodestats.Free_cpu = float32(nodestats.Capacity_cpu) - nodestats.Usage_cpu
		} else {
			// fmt.Println("Error converting CPU usage to millicore")
		}

		nodestats.Usage_cpu_percent = nodestats.Usage_cpu / float32(nodestats.Capacity_cpu) * 100
		// fmt.Println("Usage CPU Percent:", nodestats.Usage_cpu_percent)

		NodeMetrics = append(NodeMetrics, *nodestats)

	case "disk":
		var pods *core.PodList // Declare pods at the start of the case

		// Get disk capacity from ephemeral-storage
		if capacity, ok := node.Status.Capacity["ephemeral-storage"]; ok {
			capacityValue := capacity.Value()
			nodestats.Capacity_disk = int(capacityValue)
		} else {
			fmt.Println("No ephemeral-storage capacity found")
			nodestats.Capacity_disk = -1
		}

		// Try to get filesystem stats from kubelet API
		if stats, err := getKubeletStats(clientset, node); err == nil {
			// Use the filesystem stats from kubelet
			nodestats.Usage_disk = int(stats.Node.Fs.UsedBytes)
			if stats.Node.Fs.CapacityBytes > 0 {
				// If kubelet reports capacity, use that instead
				nodestats.Capacity_disk = int(stats.Node.Fs.CapacityBytes)
			}
		} else {
			// Fall back to metrics API
			if fsStats := nm.Usage.StorageEphemeral(); fsStats != nil {
				usageValue := fsStats.Value()
				if usageValue > 0 {
					nodestats.Usage_disk = int(usageValue)
				}
			}

			// If still no usage data, try to estimate from pods
			if nodestats.Usage_disk == 0 {
				var err error
				pods, err = clientset.CoreV1().Pods("").List(context.TODO(), v1.ListOptions{
					FieldSelector: fmt.Sprintf("spec.nodeName=%s", node.Name),
				})
				if err == nil {
					var podStorageUsage int64 = 0
					for _, pod := range pods.Items {
						// Get actual storage usage from pod status if available
						for _, containerStatus := range pod.Status.ContainerStatuses {
							if containerStatus.RestartCount > 0 {
								podStorageUsage += 10 * 1024 * 1024 // 10MB per restart
							}
						}
						// Add volume sizes if available
						for _, volume := range pod.Spec.Volumes {
							if volume.EmptyDir != nil && volume.EmptyDir.SizeLimit != nil {
								sizeLimit := volume.EmptyDir.SizeLimit.Value()
								podStorageUsage += sizeLimit / 2
							}
						}
					}
					nodestats.Usage_disk = int(podStorageUsage)
				}
			}
		}

		if nodestats.Capacity_disk > 0 {
			nodestats.Free_disk = nodestats.Capacity_disk - nodestats.Usage_disk
			nodestats.Usage_disk_percent = float32(nodestats.Usage_disk) / float32(nodestats.Capacity_disk) * 100
		} else {
			fmt.Println("Invalid disk capacity")
			nodestats.Usage_disk = -1
			nodestats.Free_disk = -1
			nodestats.Usage_disk_percent = 0
		}

		NodeMetrics = append(NodeMetrics, *nodestats)
	}
	return NodeMetrics
}

// getKubeletStats retrieves disk usage statistics from the kubelet's /stats/summary endpoint
func getKubeletStats(clientset *kubernetes.Clientset, node *core.Node) (*KubeletStats, error) {
	// Get the node's internal IP
	var nodeIP string
	for _, addr := range node.Status.Addresses {
		if addr.Type == core.NodeInternalIP {
			nodeIP = addr.Address
			break
		}
	}
	if nodeIP == "" {
		return nil, fmt.Errorf("could not find internal IP for node %s", node.Name)
	}

	// Create a proxy request to the kubelet
	request := clientset.CoreV1().RESTClient().Get().
		Resource("nodes").
		Name(fmt.Sprintf("%s:10250", node.Name)). // kubelet port
		SubResource("proxy").
		Suffix("stats/summary")

	// Get raw bytes
	raw, err := request.DoRaw(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get kubelet stats: %v", err)
	}

	// Unmarshal into our struct
	result := &KubeletStats{}
	if err := json.Unmarshal(raw, result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal kubelet stats: %v", err)
	}

	return result, nil
}

// getPodStats retrieves disk usage statistics for a specific pod from the kubelet's /stats/summary endpoint
func getPodStats(clientset *kubernetes.Clientset, node *core.Node, podName string, podNamespace string) (int64, error) {
	stats, err := getKubeletStats(clientset, node)
	if err != nil {
		return 0, fmt.Errorf("failed to get kubelet stats: %v", err)
	}

	// Find the pod in the stats
	for _, pod := range stats.Pods {
		if pod.PodRef.Name == podName && pod.PodRef.Namespace == podNamespace {
			var totalUsage int64 = 0

			// Add ephemeral storage usage
			totalUsage += pod.EphemeralStorage.UsedBytes

			// Add container storage usage
			for _, container := range pod.Containers {
				containerUsage := container.Rootfs.UsedBytes + container.Logs.UsedBytes
				totalUsage += containerUsage
			}

			// Add volume storage usage
			for _, volume := range pod.VolumeStats {
				volumeUsage := volume.FsStats.UsedBytes
				totalUsage += volumeUsage
			}

			// Add filesystem usage from the node's stats for this pod
			// We'll estimate this as 0.1% of the node's total filesystem usage
			if stats.Node.Fs.UsedBytes > 0 {
				fsUsage := stats.Node.Fs.UsedBytes / 1000 // 0.1% of node's filesystem usage
				totalUsage += fsUsage
			}

			// Add a base amount for pod overhead (metadata, etc.)
			baseOverhead := int64(1 * 1024 * 1024) // 1MB base overhead
			totalUsage += baseOverhead

			return totalUsage, nil
		}
	}

	return 0, fmt.Errorf("pod %s/%s not found in kubelet stats", podNamespace, podName)
}

func Nodes(inputs *utils.Inputs) (NodeStatsList []Node) {
	metric := inputs.Metrics

	utils.InitLogger()
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

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

	// To fetch kubectl top nodes metrics
	nodeMetrics, err := mc.MetricsV1beta1().NodeMetricses().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Println("Unable to Get NodeMetrics. Is Metrics Server running ?")
		os.Exit(2)
	}

	// To fetch kubectl get nodes information
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Println("Failed to Get Nodes")
		panic(err.Error())
	}

	// To fetch kubectl get pods information
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Println("Failed to Get Pods")
		panic(err.Error())
	}

	// output comes in this format
	//{{270301052 -9} {<nil>} 270301052n DecimalSI}
	//is reminiscent of how Kubernetes represents resource quantities,
	//where the first value is the quantity in base units (e.g. bytes for memory, cores for CPU),
	//the second value is the quantity in scaled units (e.g. megabytes for memory, millicores for CPU),
	//and the third value is the suffix for the scaled units (e.g. Mi for megabytes, m for millicores).
	//The fourth value is the format to use when printing the quantity as a string.
	//The DecimalSI format is the most human-readable format, and is the default format used by the Kubernetes CLI.
	//The BinarySI format is the most machine-readable format, and is the default format used by the Kubernetes API.
	//The ExponentSI format is the most compact human-readable format, and is used by the Kubernetes API for fields that are not human-facing.

	nodestats := Node{}

	// Parsing Every Node and collecting information
	for _, nm := range nodeMetrics.Items {
		for _, node := range nodes.Items {
			if node.Name == nm.Name {
				// fmt.Println("Node Name:", node.Name)
				nodestats.Name = node.Name

				// Capture the Ready status of the node
				nodestats.Status = "NotReady" // Default to NotReady
				for _, condition := range node.Status.Conditions {
					if condition.Type == core.NodeReady {
						if condition.Status == core.ConditionTrue {
							nodestats.Status = "Ready"
						}
						break
					}
				}

				// capture Uptime
				uptimeDuration := time.Since(node.CreationTimestamp.Time)

				// Convert the time duration to a simple display format
				// Use days if more than 24 hours, hours if more than 60 minutes, otherwise minutes
				if uptimeDuration.Hours() > 24 {
					days := int(uptimeDuration.Hours() / 24)
					nodestats.Uptime = fmt.Sprintf("%dd", days)
				} else if uptimeDuration.Hours() >= 1 {
					hours := int(uptimeDuration.Hours())
					nodestats.Uptime = fmt.Sprintf("%dh", hours)
				} else {
					minutes := int(uptimeDuration.Minutes())
					nodestats.Uptime = fmt.Sprintf("%dm", minutes)
				}

				// Counting Total Pods in the Node
				var totalpods int
				for _, pod := range pods.Items {
					if pod.Spec.NodeName == node.Name {
						totalpods++
					}
				}
				nodestats.TotalPods = strconv.Itoa(totalpods)

				// Display Label if provided - Logic
				if inputs.LabelToDisplay != "" {
					// check if the label exists in the node - if not, set output to "Not Found"
					if _, ok := node.Labels[inputs.LabelToDisplay]; !ok {
						nodestats.LabelToDisplay = "NA"
					} else {
						nodestats.LabelToDisplay = node.Labels[inputs.LabelToDisplay]
					}
				}

				// Collect all the labels and store in a map
				nodestats.Labels = node.Labels

				NodeStatsList = append(NodeStatsList, GetMetricsForNode(&nodestats, &node, &nm, metric, clientset)[0])

			}

		}
	}

	utils.Logger.Debug(NodeStatsList)
	return NodeStatsList

}
