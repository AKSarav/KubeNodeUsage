package k8s

import (
	"context"
	"fmt"
	"kubenodeusage/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	TotalPods             string
	LabelToDisplay 		 string
	LabelAlias		 string
}

type Cluster struct{
	Context string
	Version string
	URL		string
}

var NodeStatsList []Node
var K8sinfo Cluster

func ClusterInfo()(Cluster){
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	confvar := clientcmd.GetConfigFromFileOrDie(kubeconfig);
	
	K8sinfo := Cluster{}
	K8sinfo.Context = confvar.CurrentContext

	utils.InitLogger()
	

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println("Not able to access .kube/config file from the Home Directory path: ",kubeconfig)
		os.Exit(2)
	}
	
	mc, err := metricsv.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	K8sinfo.URL = config.Host

	// Validate Version of Server
	if version, err := mc.ServerVersion(); err != nil{
		fmt.Println("\n# ERROR: Unable to Establish Connection to Kubernetes Cluster")
		fmt.Println("# Kubernetes Context:",K8sinfo.Context)
		fmt.Println("# Kubernetes URL:",K8sinfo.URL)
		fmt.Println("# Please check your kubernetes configuration and permissions\n")
		os.Exit(2)
	} else{
		K8sinfo.Version = version.String()
	}

	

	return K8sinfo
	
}

func GetMetricsForNode(nodestats *Node, node *core.Node, nm *v1beta1.NodeMetrics, metric string)([]Node) {

	NodeMetrics := []Node{}

	switch metric {
	case "memory":
		// Ki - Kibibyte - 1024 bytes
		memcapcity, err := strconv.Atoi(strings.TrimSuffix(node.Status.Capacity.Memory().String(), "Ki"))
		if err != nil {
			fmt.Println("Error converting Memory capacity")
		} else{
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

		// Ki - Kibibyte - 1024 bytes
		diskcapacity, err := strconv.Atoi(strings.TrimSuffix(node.Status.Capacity.StorageEphemeral().String(), "Ki"))
		if err != nil {
			fmt.Println("Error converting Disk capacity")
		} else{
			nodestats.Capacity_disk = diskcapacity
		}
		// fmt.Println("Capacity Disk:", nodestats.Capacity_disk)

		// Disk usage is taken from Node Status Allocatable - not from Node Metrics
		// the result would be on bytes - need to convert to Ki
		// fmt.Println(node.Status.Allocatable.StorageEphemeral().String())
		if diskfree, err := strconv.Atoi(node.Status.Allocatable.StorageEphemeral().String()); err == nil {
			nodestats.Free_disk = diskfree / 1024
			nodestats.Usage_disk = nodestats.Capacity_disk - nodestats.Free_disk
		} else {
			fmt.Println("Error converting Disk usage")
		}

		nodestats.Usage_disk_percent = float32(nodestats.Usage_disk) / float32(nodestats.Capacity_disk) * 100

		NodeMetrics = append(NodeMetrics, *nodestats)
	}
	return NodeMetrics
}


func Nodes(inputs *utils.Inputs) (NodeStatsList []Node) {

	metric := inputs.Metrics

	utils.InitLogger()
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println("Not able to access .kube/config file from the Home Directory path: ",kubeconfig)
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

	for _, nm := range nodeMetrics.Items {
		for _, node := range nodes.Items {
			if node.Name == nm.Name {
				// fmt.Println("Node Name:", node.Name)
				nodestats.Name = node.Name

				// Counting Total Pods in the Node
				var totalpods int
				for _, pod := range pods.Items {
					if pod.Spec.NodeName == node.Name {
						totalpods++
					}
				}
				nodestats.TotalPods = strconv.Itoa(totalpods)
				
				lblToDisplay := inputs.LabelToDisplay
				if lblToDisplay != "" {
					fmt.Println("Label to Display:", node.Labels[lblToDisplay])
				}

				NodeStatsList = append(NodeStatsList, GetMetricsForNode(&nodestats, &node, &nm, metric)[0])
				
			}

		}
	}

	utils.Logger.Debug(NodeStatsList)
	return NodeStatsList

}
