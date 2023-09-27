package k8s

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Node struct {
	Name string
	capacity_disk int
	capacity_memory int
	capacity_cpu int
	usage_disk int
	usage_memory int
	usage_cpu float32
	usage_disk_percent float32
	usage_memory_percent float32
	usage_cpu_percent float32
}

var NodeStatsList []Node

func Nodes()(NodeStatsList []Node) {
	kubeconfig := flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
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
		panic(err.Error())
	}

	// To fetch kubectl get nodes information
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{}); 
	if err != nil {
		fmt.Println("Failed to Get Nodes")
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
			for _, node := range nodes.Items{
			if node.Name == nm.Name {
				fmt.Println("Node Name:", node.Name)
				nodestats.Name = node.Name

				
				// Ki - Kibibyte - 1024 bytes
				nodestats.capacity_disk, err = strconv.Atoi(strings.TrimSuffix(node.Status.Capacity.StorageEphemeral().String(), "Ki"))
				if err != nil {
					fmt.Println("Error converting Disk capacity")
				}
				fmt.Println("Capacity Disk:", nodestats.capacity_disk)
				
				// Ki - Kibibyte - 1024 bytes
				nodestats.capacity_memory, err = strconv.Atoi(strings.TrimSuffix(node.Status.Capacity.Memory().String(), "Ki"))
				if err != nil {
					fmt.Println("Error converting Memory capacity")
				}
				fmt.Println("Capacity Memory:", nodestats.capacity_memory)


				capacity_cpu, err := strconv.Atoi(node.Status.Capacity.Cpu().String())
				// Converting to millicore 1 CPU 1000 millicore
				nodestats.capacity_cpu = capacity_cpu * 1000
				if err != nil {
					fmt.Println("Error converting CPU capacity")
				}
				fmt.Println("Capacity CPU:", nodestats.capacity_cpu * 1000)


				// Disk usage is taken from Node Status Allocatable - not from Node Metrics
				// the result would be on bytes - need to convert to Ki
				fmt.Println(node.Status.Allocatable.StorageEphemeral().String())
				if diskusage, err := strconv.Atoi(node.Status.Allocatable.StorageEphemeral().String()); err == nil {
					nodestats.usage_disk = diskusage / 1024
					fmt.Println("Usage Disk:", nodestats.usage_disk)	
				} else {
					fmt.Println("Error converting Disk usage")
				}
				
				// output is in Ki - Kibibyte - 1024 bytes
				if memusage, err := strconv.Atoi(strings.TrimSuffix(nm.Usage.Memory().String(), "Ki")); err != nil {
					fmt.Println("Error converting Memory usage")
				} else {
					nodestats.usage_memory = memusage
					fmt.Println("Usage Memory:", nodestats.usage_memory)
				}
				
				
				
				cpu_in_nanocore, err := strconv.ParseFloat(strings.TrimSuffix(nm.Usage.Cpu().String(), "n"), 32); if err == nil {
					cpu_in_millicore := cpu_in_nanocore / 1000000
					nodestats.usage_cpu = float32(cpu_in_millicore)
					fmt.Println("Usage CPU:", nodestats.usage_cpu)
				} else {
					fmt.Println("Error converting CPU usage to millicore")
				}

				nodestats.usage_disk_percent = float32(nodestats.usage_disk) / float32(nodestats.capacity_disk) * 100
				// Since we are dividing the allocatable / capacity - the result would be the free space so we need to subtract it from 100 to get the usage
				nodestats.usage_disk_percent = 100 - nodestats.usage_disk_percent
				fmt.Println("Usage Disk Percent:", nodestats.usage_disk_percent)

				nodestats.usage_memory_percent = float32(nodestats.usage_memory) / float32(nodestats.capacity_memory) * 100
				fmt.Println("Usage Memory Percent:", nodestats.usage_memory_percent)


				nodestats.usage_cpu_percent = nodestats.usage_cpu / float32(nodestats.capacity_cpu) * 100
				fmt.Println("Usage CPU Percent:", nodestats.usage_cpu_percent)


				
				NodeStatsList = append(NodeStatsList, nodestats)
			}
			
			
		}
	}
	return NodeStatsList
	// fmt.Println(NodeStatsList)
}
