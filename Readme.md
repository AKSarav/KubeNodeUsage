
# KubeNodeUsage

![Alt text](KubeNodeUsage.png)

KubeNodeUsage is a tool designed to provide insights into Kubernetes node usage. It offers various options for customization to help you analyze and filter node metrics effectively.

KubeNodeUsage use your local KubeConfig file to connect to the cluster and use Kubernetes API directly using Kubernetes GO SDK

It fetch the Node Metrics from Kubernetes API and apply filters and aggreations to display it in a nice human readable Graphical format as Bar Charts.

It has lot of capabilities to filter the output based on

* NodeName
* Usage
* Free/Availability of Disk,CPU, Memory
* Max/Capacity of Disk,CPU,Memory
* Color - We use Green, Red and Orange to represent the usage 
  * Green - Below 30% Usage 
  * Orange - Between 30% to 70% Usage
  * Red - Above 70% Usage

Here is a quick demo recorded with Live Kubernetes Cluster

![Alt text](KubeNodeUsage-demo.gif)


&nbsp;

## Kubernetes Supported Versions / Clusters

As KubeNodeUsage use the Kubernetes Go SDK and directly connects to the API - It supports all the Kubernetes cluster which supports `.kube/config` file based authentication 

**Clusters**

I have tested it with the following K8s clusters

* EKS - Elastic Kubernetes Service from AWS
* Azure Kubernetes Service
* GKE - Google Kubernetes Engine
* Minikube
* Kind Cluster

**Versions**

I have tested KubeNodeUsage starting from **1.19 - 1.28** ( the latest stable version as of Dec2023)

&nbsp;

## Download

You can clone this project or download the suitable binary from releases directory

&nbsp;

## Usage

```bash
KubeNodeUsage [options]
```

&nbsp;
## Options

- help: Display help information.

- metrics: Choose which metric to display. Valid options include:

    - memory
    - disk
    - cpu

- filternodes: Filter nodes based on node name using a regular expression. (Note: Only one filter can be used at a time, and the input should be enclosed in quotes.)

- filtercolor: Filter nodes based on color categories. Valid options include:

    - red
    - green
    - orange

- desc: Enable reverse sort order.

- debug: Enable debug mode. ( Prints more logging for debug)


- sortby: Sort the output by a specific metric. Valid options include:

    - name (Sort by node name alphabetically)
    - node (Sort by node name alphabetically, same as 'name')

    - free (Sort by available resources)

    - usage (Sort by resource usage)
    - color (Sort by color category, same as usage)

    - capacity (Sort by resource capacity)
    - max (Sort by maximum resource value, same as 'capacity')

&nbsp;
## Examples:

```bash
# Display help information
KubeNodeUsage --help

# Display node usage with default settings (memory is the default metric)
KubeNodeUsage

# Display node usage sorted by node name
KubeNodeUsage --sortby=name

# Display node usage sorted by free resources in descending order
KubeNodeUsage --sortby=free --desc

# Display node usage sorted by usage in ascending order
KubeNodeUsage --sortby=usage

# Display node usage sorted by capacity in descending order
KubeNodeUsage --sortby=capacity --desc

# Filter nodes with a name starting with "web"
KubeNodeUsage --filternodes="web.*"

# Filter nodes with color category "green"
KubeNodeUsage --filtercolor=green

# Display memory usage for all nodes
KubeNodeUsage --metrics=memory

# Display disk usage for nodes with a name containing "data"
KubeNodeUsage --metrics=disk --filternodes=".*data.*"

# Show CPU usage for nodes with color category "red" in descending order
KubeNodeUsage --metrics=cpu --filtercolor=red --desc

# Display node usage sorted by maximum resource value in ascending order
KubeNodeUsage --sortby=max

# Display node usage sorted by capacity, show memory usage, and filter nodes with a name starting with "prod"
KubeNodeUsage --sortby=capacity --metrics=memory --filternodes="prod.*"

# Show CPU usage for nodes with color category "orange" and filter nodes with a name containing "IP range"
KubeNodeUsage --metrics=cpu --filtercolor=orange --filternodes=".*172-31.*"

# Display node usage sorted by name, filter nodes with a name starting with "app", and enable debug mode
KubeNodeUsage --sortby=name --filternodes="app.*" --debug

```
&nbsp;
## Todo

* `FilterLabels` Filter by Label feature to be added
* Pod Usage stats to be added as a feature

&nbsp;

#### Contributions are welcome

Feel free to send your Pull requests and Issues to make this better.

&nbsp;

>Please share and Leave a **Github Star** if you like KubeNodeUsage - It would motivate me