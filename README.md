
# KubeNodeUsage

![GitHub release](https://img.shields.io/github/v/release/AKSarav/Kube-Node-Usage?style=for-the-badge)![GitHub all releases](https://img.shields.io/github/downloads/AKSarav/Kube-Node-Usage/total?style=for-the-badge)![GitHub stars](https://img.shields.io/github/stars/AKSarav/Kube-Node-Usage?style=for-the-badge)

![Go-Version](https://img.shields.io/github/go-mod/go-version/AKSarav/Kube-Node-Usage)![GitHub](https://img.shields.io/github/license/AKSarav/Kube-Node-Usage)![GitHub issues](https://img.shields.io/github/issues/AKSarav/Kube-Node-Usage)![GitHub pull requests](https://img.shields.io/github/issues-pr/AKSarav/Kube-Node-Usage)


**KubeNodeUsage** is a Terminal App designed to provide insights into Kubernetes node and pod usage. It offers both interactive exploration and command-line filtering options to help you analyze your cluster effectively right from your terminal .

Available to download from `Brew` and `Goinstall`

You no longer have to wait for that `prometheus` and `grafana` - Just check the usage visually right from terminal.

![Alt text](assets/KubeNodeUsage3.0.4.gif)

KubeNodeUsage use your local KubeConfig file to connect to the cluster and use Kubernetes API directly using Kubernetes GO SDK

It fetches the Node and Pod Metrics from Kubernetes API and apply filters and aggreations to display it in a nice human readable Graphical format as Bar Charts.



## 🌟  Newly added Features ![v3.0.4](https://img.shields.io/badge/v3.0.4-8A2BE2)
- **Smart Search**: Press `S` to instantly filter and highlight matching entries
  - Real-time filtering as you type
  - Headers remain visible for context
  - Match count display
  - Press ESC to exit search mode
- **Horizontal Scrolling**: Use `←` and `→` arrows to view wide content
  - Smooth scrolling for large tables
  - Preserves column alignment
- **New Pod Usage**:
  - Now you can see Pod usage in KubeNodeUsage
- **Extra fields in NodeUsage**
  - Thanks to the Horizontal scrolling - we can show more fields like `Uptime` and `Status`
- **More accurate diskusage calculation**
  - Bringing you the accurate diskusage calculation for POD and Node using /stats/summary endpoint in Kubelet



## 🚀 Existing Features 

- Filter output by Node, Color, Label
- Display any specific Node Label as a Custom field
- Metrics are dynamically updated every 2 seconds, so you get the real time data always
- Choose to not display the ClusterInfo for additional security
- No data collected everything is local
- Sorting with Ascending and Descending

### Options for Filtering and Sorting
Perfect for scripting and automation:
- **Node/Pod Filtering**: 
  - `--filternodes`: Filter by node name using regex
  - `--filtercolor`: Filter by usage levels (green/orange/red)
  - `--filterlabel`: Filter by label key-value pairs
- **Sorting Options**:
  - `--sortby`: Sort by various metrics (name, free, usage, etc.)
  - `--desc`: Reverse sort order
- **Display Options**:
  - `--noinfo`: To not display the Cluster information ( can sometimes be confidential ) - Default is to display
  - `--label`: To display a specific lable with custom FieldName



&nbsp;

> Refer the Options and Screenshots section for more details

&nbsp;
## Screenshots 📸

:sparkle: Screenshot of Latest Version 3.0.1 with Cluster Info 

![Alt text](assets/KubeNodeUsageV3.0.1.png)

:sparkle: Screenshot of Version 3.0.2 with Label Column and Total Pods Count with `-noinfo` flag ( to turn off Cluster Info) 

![Alt text](assets/KubeNodeUsage-V3.0.2.png)

:sparkle: Screenshot of Multiple RegEx patterns with FilterNodes

![Alt text](assets/filternodes-regex.png)

&nbsp;

:sparkle: Screenshot of Label Filtering with FilterLabel

![Alt text](assets/KubeNodeUsage-FilterLabel.png)

:sparkle: Screenshot of Search and PodUsage
![Alt text](assets/KubeNodeUsagev3.0.4.png)

:sparkle: More powerful DiskUsage for Pods and Nodes powered by Kubelet /stats endpoint

![Alt text](assets/KubeNodeUsage-Disk-Pods.png)

:sparkle: Pods Memory by Sort - All the filters and sorts are available for Pods too

![Alt text](assets/KubeNodeUsageMemorySort.png)

## Kubernetes Supported Versions / Clusters :white_check_mark:

As KubeNodeUsage use the Kubernetes Go SDK and directly connects to the API - It supports all the Kubernetes cluster which supports `.kube/config` file based authentication 

**Clusters**

I have tested it with the following K8s clusters

* EKS - Elastic Kubernetes Service from AWS
* Azure Kubernetes Service
* GKE - Google Kubernetes Engine
* Minikube
* Kind Cluster
* Rancher Kubernetes Engine

**Versions**

I have tested KubeNodeUsage starting from **1.19 - 1.29** ( the latest stable version as of 30th June)

&nbsp; 

## How it Works :bulb:

KubeNodeUsage is written in `GoLang` and uses `client-go` and `kubernetes` SDK libraries

When you start KubeNodeUsage - It try to read the `$HOME/.kube/config` file in your HOME directory

KubeNodeUsage connects to the Default cluster set by the `CurrentContext` on the ./kube/config file

You can manually edit this file and update it but the recommended way to update current-context is to use `kubectl config use-context`

KubeNodeUsage does not use `kubectl` directly and it relies on the `.kube/config` file and the authentication method defined in there

KubeNodeUsage works the same way - kubectl works based on the configuration information found on .kube/config file

If the `kubectl`commands are not working - its highly likely KubeNodeUsage would fail too - In this case you have to check your Kube config file

&nbsp; 

#### Is it secure ? How about Data Privacy :lock:
KubeNodeUsage do not collect any data of your cluster or usage. 
&nbsp;

## How to Download :arrow_double_down:

You can clone this project and run it as shown below

#### Brew Install 🍺

```
brew tap AKSarav/kubenodeusage
brew install kubenodeusage
```

> Note: If you are using Brew install - You have to use all lowercase letters for the command `kubenodeusage` instead of `KubeNodeUsage` 

#### With GO install command  🚀

```
go install github.com/AKSarav/KubeNodeUsage/v3@v3.0.3

```

#### Clone and Run 👨‍💻

```
git clone https://github.com/AKSarav/KubeNodeUsage.git
cd Kube-Node-Usage
go run main.go
```

#### Download the Binaries from the Release Page 📦

Goto the https://github.com/AKSarav/KubeNodeUsage/releases and download the suitable binary for your OS and use it

```
unzip KubeNodeUsage-windows-386.exe.zip
./KubeNodeUsage-windows-386
```

&nbsp;

## How to use :book:

KubeNodeUsage is a command line utility with lot of startup options/flags 


```bash
KubeNodeUsage [options]
```

You can find the detail information on all available options here

&nbsp;
**Available Options** :memo:

- `help`: Display help information.

- `noinfo` : Disable the Cluster Info display. ( New feature in V3.0.2)

- `metrics`: Choose which metric to display. Valid options include:

    - memory
    - disk
    - cpu

- `filternodes`: Filter nodes based on node name using a regular expression. (Note: Only one filter can be used at a time, and the input should be enclosed in quotes.)

- `filtercolor`: Filter nodes based on color categories. Valid options include:

    - `red` 
    - `green`
    - `orange`

- `filterlabel`: Filter nodes based on the label key-value pair. ( New feature in V3.0.2) Syntax is `--filterlabel=<label-key>=<label-value>`

- `debug`: Enable debug mode. ( Prints more logging for debug)

- `sortby`: Sort the output by a specific metric. Valid options include:

    - `name` (Sort by node name alphabetically)
    - `node` (Sort by node name alphabetically, same as 'name')

    - `free` (Sort by available resources)

    - `usage` (Sort by resource usage)
    - `color` (Sort by color category, same as usage)

    - `capacity` (Sort by resource capacity)
    - `max` (Sort by maximum resource value, same as 'capacity')
-  `desc`: Enable reverse sort order.
-  `label`: Display the Label information as a new column in the output. ( New feature in V3.0.2) Syntax is `--label=<label-key>#<columnname>`
  

&nbsp;
## Examples 📝

```bash
# Display help information
KubeNodeUsage --help

# Display node usage with default settings (memory is the default metric)
KubeNodeUsage

# To disable the Cluster Info display use --noinfo flag with any other options
KubeNodeUsage --noinfo
KubeNodeUsage --noinfo --metrics cpu
KubeNodeUsage --noinfo --metrics cpu --sortby usage --desc

# Display node usage sorted by node name
KubeNodeUsage --sortby name

# Display node usage sorted by free resources in descending order
KubeNodeUsage --sortby free --desc

# Display node usage sorted by usage in ascending order
KubeNodeUsage --sortby usage

# Display node usage sorted by capacity in descending order
KubeNodeUsage --sortby capacity --desc

# Filter nodes with a name starting with "web" - supports regular expression
KubeNodeUsage --filternodes "web.*"

# Filter nodes with color category "green"
KubeNodeUsage --filtercolor green

# Display memory usage for all nodes
KubeNodeUsage --metrics memory

# Display disk usage for nodes with a name containing "data"
KubeNodeUsage --metrics disk --filternodes ".*data.*"

# Show CPU usage for nodes with color category "red" in descending order
KubeNodeUsage --metrics cpu --filtercolor red --desc

# Display node usage sorted by maximum resource value in ascending order
KubeNodeUsage --sortby max

# Display node usage sorted by capacity, show memory usage, and filter nodes with a name starting with "prod"
KubeNodeUsage --sortby capacity --metrics memory --filternodes "prod.*"

# Show CPU usage for nodes with a name containing "IP range" - here we are using multiple regex patterns as comma separated values
KubeNodeUsage --metrics cpu --filternodes ".*172-31.*",".*172-32.*"

# Display node usage sorted by name, filter nodes with a name starting with "app", and enable debug mode
KubeNodeUsage --sortby name --filternodes "app.*" --debug

# Display Label as a new column in the output. use the `--label` the syntax is "--label=<label-name>:<alias/columnname>" ( New feature in V3.0.2)
KubeNodeUsage --label eks.amazonaws.com/capacityType#capacity 
KubeNodeUsage --label beta.kubernetes.io/instance-type#InstanceType 

# Filter Nodes based on Label Key Value pair ( New feature in V3.0.2)
KubeNodeUsage --filterlabel eks.amazonaws.com/capacityType=OnDemand
KubeNodeUsage --filterlabel beta.kubernetes.io/instance-type=t3.medium
KubeNodeUsage --filterlabel topology.kubernetes.io/zone=us-east-1a


```

&nbsp;

### Contributions / Feature Requests are welcome :handshake:

Feel free to send your Pull requests and Issues to make this better.

&nbsp;

>Please share and Leave a **Github Star** :star: :star: :star: :star:  if you like KubeNodeUsage - It would motivate me
