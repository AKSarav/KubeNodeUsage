package cmd

import (
	"encoding/json"
	"fmt"
	"kubenodeusage/k8s"
	"os"
	"os/exec"
)

// Declaration of variables
var (
	nodes map[string]interface{}
	nodestop map[string]interface{}
)

// Function to execute kubectl top nodes and return the result as interface
func TopNodes() {
	cmd := exec.Command("kubectl","top","nodes")
	if out, err := cmd.CombinedOutput() ; err != nil{
		fmt.Println("Something went wrong",err)
	} else {
		fmt.Println(string(out))
	}

	/*
	Sample Output to parse
	NAME                           CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
	ip-10-0-106-148.ec2.internal   145m         3%     1901Mi          27%
	ip-10-0-106-6.ec2.internal     171m         8%     1850Mi          55%
	*/
	
	// nodes_top_result := strings.Fields(out)

}

// Function name has to be capitalized to be exported
func GetNodes()(interface{}){
	// Run kubectl get nodes and get the output and store it in a dictionary
	cmd := exec.Command("kubectl", "get", "nodes", "-o", "json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running kubectl get nodes")
		os.Exit(1)
	}

	// Unmarshal the JSON output of kubectl get nodes into an interface
	// Converting the `out` string to `byte array` and saving 
	json.Unmarshal([]byte(out), &nodes)

	return nodes
		
}

func NodeStats(){
	
	// // Doing a type assertion and saving the output to getnodesoutput
	// if nodeslist, ok := GetNodes().(map[string]interface{}); !ok {
	// 	fmt.Println("Output is not of map with String keys and interface values")
	// } else{
	// 	fmt.Println(nodeslist)
	// }
	fmt.Println("Calling GetNodes & Top Nodes")
	k8s.Nodes()

}
