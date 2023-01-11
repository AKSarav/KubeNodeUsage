# Kube Node Usage

![Alt text](KubeNodeUsage-cpu-sort-desc.png)


**Kubernetes Node Usage** or **Kube-Node-Usage** is a CLI tool to get the Memory, CPU and Disk Usage of Kubernetes Nodes

it is designed on python and relies on the `kubectl` installed in your local.

No Authentication data are directly handled.

You can think of `Kube-Node-Usage`  as a wrapper over `kubectl get nodes` command

Kube-Node-Usage simply execute the `kubectl get nodes` command and parse the output and present it to you with a nice formatting and Usage Bar



### Prerequisites

1) `Kubectl` must be installed and configured 
2) `Python3` must be installed and used to run the kube-node-usage
3) `pip` package manager is required to install the necassary python packages
4) Must have required Kubernetes Cluster accessr. As we have mentioned. kube-node-usage simply run the `kubectl get nodes` command and parse the output and present it to you.
   

### Release Notes of V1.0.2

1. In earlier release `v1.0.1` we supported the short hand arguments like `-d` , `-m` , `-c` to represent the disk, memory and cpu respectively. It is now removed for much cleaner approach. Only full forms are supported `--cpu, --memory, --disk`

2. Release `v1.0.2` is powered with unix style `getopts` comparing to the if - else style on the `v1.0.1`

3. Two startup arguments/options are added with this release
   *  `--sort=` to sort by the output fields in ascending order by default
   *  `--reverse` to enable reverse sorting, in descending order

<br/>

### How to Set up / Install Kube-Node-Usage

1. Clone the repository

```
git clone https://github.com/AKSarav/Kube-Node-Usage.git
```

2. Install the necassary packages with the following PIP command.

We presume you have pip and python3 installed


```
pip install -r requirements.txt
```

3. Execute the command to list the kubernetes nodes with their Usage Information

<br/>
<br/>

For some reason, If you do not wish to install the required python packages into the entire system

You can create your virtual environment (virtualenv) and install the packages

Here are the commands for the same

```
# python -m venv venv
# source venv/bin/activate
# pip install -r requirements.txt
```
Once you have used the `kube-node-usage` you can execute the `deactivate` command

```
# deactivate
```

### How to use Kube-Node-Usage

Here are the list of commands Kube-Node-Usage supports and how it can be used.

> **Note:** 
> By default the results are sorted in ascending order by the Node Name. You can control the sort behaviour with `--sort` and `--reverse` args
 

<br/>
##### List the Node with Disk Usage

To list the Kubernetes nodes with Disk Usage execute the following command

```
# python kube-node-usage.py --disk
```

To List the Nodes with CPU Usage with `sort` 

You can sort the output based on each displayed field

1. Node Name 
2. Free
3. Max
4. Usage

```
# python kube-node-usage.py --disk --sort=max 

# python kube-node-usage.py --disk --sort=free 

# python kube-node-usage.py --disk --sort=node 

# python kube-node-usage.py --disk --sort=usage
```

<br/>

<br/>

##### List the Node with CPU Usage

To list the Kubernetes nodes with CPU Usage execute the following command

```
# python kube-node-usage.py --cpu
```

To List the Nodes with CPU Usage with `sort` 

You can sort the output based on each displayed field

1. Node Name 
2. Free
3. Max
4. Usage

```
# python kube-node-usage.py --cpu --sort=max 

# python kube-node-usage.py --cpu --sort=free 

# python kube-node-usage.py --cpu --sort=node 

# python kube-node-usage.py --cpu --sort=usage
```

<br/>

##### List the Node with Memory Usage

To list the Kubernetes nodes with Memory Usage execute the following command

```
# python kube-node-usage.py --memory
```

To List the Nodes with Memory Usage with `sort` 

You can sort the output based on each displayed field

1. Node Name 
2. Free
3. Max
4. Usage

```
# python kube-node-usage.py --memory --sort=max 

# python kube-node-usage.py --memory --sort=free 

# python kube-node-usage.py --memory --sort=node 

# python kube-node-usage.py --memory --sort=usage
```

<br/>

##### List the Node with All - CPU, MEMORY, DISK Usage

To list the Kubernetes nodes with All ( CPU, Memory, Disk) Usage execute the following command

```
# python kube-node-usage.py --all
```

To List the Nodes with  All ( CPU, Memory, Disk) Usage with `sort` 

You can sort the output based on each displayed field

1. Node Name 
2. Free
3. Max
4. Usage

```
# python kube-node-usage.py --all --sort=max 

# python kube-node-usage.py --all --sort=free 

# python kube-node-usage.py --all --sort=node 

# python kube-node-usage.py --all --sort=usage
```

</br>

##### DESCENDING (or) REVERSE sorting 

By Default, `sort` option do the sort in  `ASCENDING` order

If you want to do the `sort` in `DESCENDING` order

<br>

**All Option Commands with Sort and Reverse**

```

# python kube-node-usage.py --all --sort=max --reverse

# python kube-node-usage.py --all --sort=free --reverse

# python kube-node-usage.py --all --sort=node --reverse

# python kube-node-usage.py --all --sort=usage --reverse
```

<br>

**Disk related Commands with Memory and Reverse**

```

# python kube-node-usage.py --memory --sort=max --reverse

# python kube-node-usage.py --memory --sort=free --reverse

# python kube-node-usage.py --memory --sort=node --reverse

# python kube-node-usage.py --memory --sort=usage --reverse
```
<br>

**CPU related Commands with Sort and Reverse**


```
# python kube-node-usage.py --cpu --sort=max --reverse

# python kube-node-usage.py --cpu --sort=free --reverse

# python kube-node-usage.py --cpu --sort=node --reverse

# python kube-node-usage.py --cpu --sort=usage --reverse
```
<br>

**Disk related Commands with Sort and Reverse**

```

# python kube-node-usage.py --disk --sort=max --reverse

# python kube-node-usage.py --disk --sort=free --reverse

# python kube-node-usage.py --disk --sort=node --reverse

# python kube-node-usage.py --disk --sort=usage --reverse

```

### Screenshots 

</br>

> **Note**  
> All the data shown here are created with Random Usage data. The Free, Max and the Usage% may not add up

</br>
**Kubernetes Nodes CPU Usage - Release 1.0.2**
This is a screenshot taken from Release 1.0.2 

Sort By Usage ( ASC )
![Alt text](KubeNodeUsage-cpu-sort-asc.png)

Sort by Usage ( DESC)
![Alt text](KubeNodeUsage-cpu-sort-desc.png)

**Kubernetes Nodes Memory Usage - Release 1.0.2**
This is a screenshot taken from Release 1.0.2 


Sort by Usage ( DESC)
![Alt text](KubeNodeUsage-memory-sort-desc.png)

Sort by Usage ( ASC)
![Alt text](KubeNodeUsage-memory-sort-asc.png)

**Kubernetes Nodes Disk Usage - Release 1.0.2**
This is a screenshot taken from Release 1.0.2


Sort by Usage ( DESC)
![Alt text](KubeNodeUsage-disk-sort-desc.png)

Sort by Usage ( ASC)
![Alt text](KubeNodeUsage-disk-sort-asc.png)

</br>

### Pull requests and Issues are welcome

Feel free to send your Pull requests to make this tool better.

If you happen to see any issues. please create an issue and I will have it checked.


</br>

### If you like this tool. please let me know by clicking on the Github Stars 


### How to reach me

Linked in : https://www.linkedin.com/in/saravakdevopsjunction/
Website: https://middlewareinventory.com





