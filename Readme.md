# KubeNodeUsage

KubeNodeUsage is a tool designed to provide insights into Kubernetes node usage. It offers various options for customization to help you analyze and filter node metrics effectively.

> This is a beta release - feel free to report bugs and feature requests


## Download

You can clone this project or download the suitable binary from releases directory

## Usage

```bash
KubeNodeUsage [options]
```

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

## Todo

* `FilterLabels` Filter by Label feature to be added
* Pod Usage stats to be added as a feature