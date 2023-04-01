import getopt, sys
import json
import subprocess
from tqdm import tqdm
import math
import os
import sys
import time
# Threading will be enabled in future
#import threading
import random

# For Redirecting the output to /dev/null
f = open(os.devnull, 'w')


def greet():
    print('''
# Kube-Node-Usage
# Release 2.0.0
# https://github.com/AKSarav/Kube-Node-Usage
''')

def round_down(n, decimals=2):
    multiplier = 10 ** decimals
    return math.floor(n * multiplier) / multiplier

def init():
    try:
        nodescmd=subprocess.run(["kubectl", "get" ,"nodes", "-o", "json"], capture_output=True,timeout=10)

        if nodescmd.stdout:
            nodeslist=json.loads(nodescmd.stdout.decode('utf-8'))
            return nodeslist
        else:
            nodeslist=json.loads(nodescmd.stderr.decode('utf-8'))
            
        
    except Exception as e:
        print("Failed to execute the Kubectl get nodes command")
        exit(5)


def inittop():
    try:
        topcmd=subprocess.run(["kubectl", "top" ,"nodes"], capture_output=True,timeout=10)
        if topcmd.stdout:
            return topcmd.stdout.decode('utf-8')
        else:
            print(topcmd.stderr.decode('utf-8'))
            return False
    except Exception as e:
        print("Failed to execute the Kubectl top nodes command")
        exit(5)

def printbar(diskusage_percent):
        with tqdm(total=100, bar_format="{l_bar}{bar}", position=0, leave=True, file=sys.stdout, ncols=60) as diskbar:
        
            # For Debug
            #diskusage_percent=random.randint(1,100)
            if diskusage_percent <= 30:
                diskbar.colour = 'green'
            if diskusage_percent > 30 and diskusage_percent < 70 :
                diskbar.colour = 'yellow'
            if diskusage_percent >= 70:
                diskbar.colour = 'red'

            # to avoid multiple prints
            old_stdout = sys.stdout
            sys.stdout = f
            
            diskbar.refresh()
            diskbar.update(diskusage_percent)
            
            sys.stdout = old_stdout
            return diskbar
        # diskbar.close()

def sortbyusage(e):
    return e['usage_percent']

def sortbynode(e):
    return e['node_name']

def sortbyallocatable(e):
    return e['allocatable_inmb']

def sortbymax(e):
    return e['max_inmb']
    
def printargs(usagetype, sortby, isreverse, filternodes, filtercolors):
    print("""
Arguments Passed
-----------------
UsageType: {}
SortBy: {}
IsReverse: {}
FilterNodes: {}
FilterColors: {}""".format(usagetype, sortby, isreverse, filternodes, filtercolors))    

def usage():
    # Create Usage Instructions with latest features

    print("-"*50)
    print("# KubeNodeUsage - Usage instructions")
    print("-"*50)
    print("\n~ Valid Usage types are \n  --memory , --cpu , --disk, --all\n")
    print("~ Valid Sort by values are \n  --sort=free | --sort=max | --sort=usage | --sort=node\n")
    print("~ Valid Filter by values are ( multiple colors should be comma seperated ) \n  --filtercolors=red | --filtercolors=yellow | --filtercolors=green | --filtercolors=red,yellow,green\n")
    print("  red: Usage > 70% | yellow: Usage > 30% and < 70% | green: Usage < 30%\n")    
    print("~ Valid Filter by nodes are ( multiple node names should be comma seperated ) \n  --filternodes=node1 | --filternodes=node1,node2,node3\n")
    
    print("\n# Examples:")
    print("-"*10)
    print("> To display the Memory Usage with default Sorting\n\t# python kube-node-usage.py --memory \n")
    print("> To Display the Memory Usage sort by Usage Percentage\n\t# python kube-node-usage.py --memory --sort=usage \n")
    print("> To Display the CPU Usage sort by the free/allocatable cpu\n\t# python kube-node-usage.py --cpu --sort=free \n")
    print("> To Display the Disk Usage sort by the Max Disk\n\t# python kube-node-usage.py --disk --sort=max \n")
    print("> To Apply the reverse/desc sort with the existing command add --reverse\n\t# python kube-node-usage.py --disk --sort=max --reverse \n")
    print("> To Filter the nodes based on the usage color\n\t# python kube-node-usage.py --disk --filtercolors=red \n")
    print("> To Filter the nodes based on the color and sort by the usage\n\t# python kube-node-usage.py --disk --filtercolors=red --sort=usage \n")
    print("> To Filter the nodes based on the color and sort by the usage and apply the reverse sort\n\t# python kube-node-usage.py --disk --filtercolors=red --sort=usage --reverse \n")
    print("> To Filter multiple colors \n\t# python kube-node-usage.py --cpu --filtercolors=red,yellow \n")
    print("> To filter by nodenames \n\t# python kube-node-usage.py --cpu --filternodes=<nodename1>,<nodename2> \n")
    print("> To enable the debug mode use --debug with any of the previous command\n")
    print("-"*50)
    exit(6)

    # Create a valid commands and permutations filternodes and filtercolors should not be used together
    # python kube-node-usage.py (--memory|--cpu|--disk) (--sort=free|max|usage|node) (--reverse) (--debug) (--filtercolors=red,yellow,green|red|yellow|green) (--filternodes=node1,node2) 






def action(type, sortkey, isreverse, filternodes=[], filtercolors=[]):
    # getnodes first
    # print("Starting with Args ",type,sortkey,isreverse)
    nodeslist=init()
    nodetoplist=inittop()
    
    nodetopusage=[]
    nodeusagemap=[]
    
    for line in nodetoplist.splitlines():
        nodetopusage+=[line.split()]
    
    
    for i in range(1, len(nodetopusage)):
        usagemap={}
        for j in range(0,len(nodetopusage[i])):
            usagemap[nodetopusage[0][j]] = nodetopusage[i][j]    
        nodeusagemap.append(usagemap)
    
    
    node_name_len=0
    outputbuffer=[]
    
    for item in nodeslist['items']:
        node_name=item['metadata']['name']
        if type == "disk" or type == "d":
            allocatable=item['status']['allocatable']['ephemeral-storage']
            maximum=item['status']['capacity']['ephemeral-storage'].split('Ki')[0]
            max_inmb=round(int(maximum) / 1024 / 1024)
            allocatable_inmb=round_down(int(allocatable)/1024/1024/1024)
            usage=max_inmb - allocatable_inmb
            usage_percent = (usage/max_inmb) * 100

        elif type == "memory":
            maximum=item['status']['capacity']['memory'].split('Ki')[0]
            max_inmb=round(int(maximum) / 1024 / 1024)
            usage = int(list(filter(lambda x:x["NAME"] == node_name, nodeusagemap))[0]['MEMORY(bytes)'].split('Mi')[0])
            usage_percent = int(list(filter(lambda x:x["NAME"] == node_name, nodeusagemap))[0]['MEMORY%'].split('%')[0])
            usageinmb=int(usage)/1000
            allocatable_inmb=round_down(max_inmb-usageinmb)

        elif type == "cpu":
            maximum=int(item['status']['capacity']['cpu'])*1000
            max_inmb=round(int(maximum))
            usage = int(list(filter(lambda x:x["NAME"] == node_name, nodeusagemap))[0]['CPU(cores)'].split('m')[0])
            usageinmb=int(usage)
            allocatable_inmb=round_down(max_inmb-usageinmb)
            usage_percent = int(list(filter(lambda x:x["NAME"] == node_name, nodeusagemap))[0]['CPU%'].split('%')[0])

        if len(node_name)>node_name_len:
            node_name_len = len(node_name)

        
        #usage_percent = random.randint(1,100)
        
        old_stdout = sys.stdout
        sys.stdout = f
        usage=str(printbar(usage_percent))
        sys.stdout = old_stdout
        
        tmp={
            "node_name":node_name,
            "allocatable_inmb":allocatable_inmb,
            "max_inmb":max_inmb,
            "usage_percent": usage_percent,
            "usage": usage
            }
        outputbuffer.append(tmp)


    if sortkey == "usage":
        outputbuffer.sort(key=sortbyusage,reverse=isreverse)
    elif sortkey == "free":
        outputbuffer.sort(key=sortbyallocatable,reverse=isreverse)
    elif sortkey == "max":
        outputbuffer.sort(key=sortbymax,reverse=isreverse)
    elif sortkey == "node":
        outputbuffer.sort(key=sortbynode,reverse=isreverse)

    col_fmt="{:<"+str(node_name_len)+"}"+"\t{:<12}" * 3
    print("\r\n# Disk Usage\n\n"+col_fmt.format(*["NodeName", "Free(GB)", "Max(GB)", "Usage(%)"])) if type == "disk" else ""
    print("\r\n# Memory Usage\n\n"+col_fmt.format(*["NodeName", "Free(GB)", "Max(GB)", "Usage(%)"])) if type == "memory" else ""
    print("\r\n# CPU Usage\n\n"+col_fmt.format(*["NodeName", "Free(m)", "Max(m)", "Usage(%)"])) if type == "cpu" else ""
    print("-"*(node_name_len + 70))    
    
    for line in outputbuffer:  
        # case where only filtercolors is provided but not filternodes
        if len(filternodes) > 0 and len(filtercolors) == 0:
            if line['node_name'] in filternodes:
                print(col_fmt.format(*[line['node_name'], line['allocatable_inmb'], line['max_inmb'], line['usage']]),flush=True)                

        # case where only filternodes is provided but not filtercolors
        elif len(filternodes) == 0 and len(filtercolors) > 0:
            if 'red' in filtercolors:
                if line['usage_percent'] > 70:
                    print(col_fmt.format(*[line['node_name'], line['allocatable_inmb'], line['max_inmb'], line['usage']]),flush=True)
            if 'yellow' in filtercolors:
                if line['usage_percent'] > 30 and line['usage_percent'] < 70:
                    print(col_fmt.format(*[line['node_name'], line['allocatable_inmb'], line['max_inmb'], line['usage']]),flush=True)
            if 'green' in filtercolors:
                if line['usage_percent'] < 30:
                    print(col_fmt.format(*[line['node_name'], line['allocatable_inmb'], line['max_inmb'], line['usage']]),flush=True)

        # case where neither filternodes nor filtercolors are provided                    
        else:
            print(col_fmt.format(*[line['node_name'], line['allocatable_inmb'], line['max_inmb'], line['usage']]),flush=True)
def isdigit(s):
    try:
        int(s)
        return True
    except ValueError:
        return False

def main():
    try:
        opts, args = getopt.getopt(sys.argv[1:], ":h", ["help", "sort=", "reverse", "memory", "cpu", "disk", "all","interval=","filternodes=","filtercolors=","debug"])
    except getopt.GetoptError as err:
        # print help information and exit:
        print(err)  # will print something like "option -a not recognized"
        usage()

    # Defaults
    usagetype="disk"
    sortby="node_name"
    interval=0
    isreverse=False #ASC
    filtercolors=[]
    filternodes=[]
    debug=False

    for o, a in opts:
        if o in ("-h", "--help"):
            print("Printing Help Instructions and Exitting..")
            usage()
            exit(5)
        elif o in ("--cpu"):
            # print("Collecting CPU Usage")
            usagetype="cpu"          
        elif o in ("--memory"):
            # print("Collecting Memory Usage")
            usagetype="memory"
        elif o in ("--disk"):
            # print("Collecting disk Usage")
            usagetype="disk"
        elif o in ("--sort"):
            # print("Sorting")
            sortby=a
            if a not in ("max", "free", "usage", "node"):
                print("Invalid Argument for Sort. It should be one of max | free | usage | node")
                usage()
        elif o in ("--all"):
            print("--- Fetching all the Usage")
            usagetype="all"
        elif o in ("--reverse"):
            # print("Reverse Sorting")
            isreverse=True
        elif o in ("--interval"):
            if not a.isdigit():
                print("Invalid Argument for Interval. It should be a number")
                usage()
            # interval should be between 10 to 180 seconds
            if int(a) < 10 or int(a) > 180:
                print("\nERROR: Invalid Argument for Interval. It should be between 10 to 180 seconds\n")
                time.sleep(2)
                usage()
            print("--- Enabling Continious monitoring for every",a,"seconds")
            interval=a
        elif o in ("--filternodes"):
            if "," in a:
                filternodes=a.split(",")
            else:
                filternodes=a.split(" ")

        elif o in ("--filtercolors"):
            if "," in a:
                filtercolors=a.split(",")
            else:
                filtercolors=a.split(" ")

            # validate only red, green, yellow is passed
            for color in filtercolors:
                if color not in ("red", "green", "yellow"):
                    print("Invalid Argument for Filter Colors. It should be one of \n - red (above 70%) \n - green (below 30%) \n - yellow (between 30% and 70%))")
                    usage()
        elif o in ("--debug"):
            debug=True
        else:
            print("Not a valid option",o)   
            exit(5)
    
    os.system('clear')
    greet()
    time.sleep(2)  
    
    # print messagee if filtercolors and filternodes are enabled at the same time
    if len(filtercolors) > 0 and len(filternodes) > 0:
        print("--- ERROR: Filtering by NodeNames and Usage Colors are mutually exclusive. Please use either one of them")
        exit(5)
    if interval != 0:
        while True:          
            if usagetype != "all":
                printargs(usagetype, sortby, isreverse, filternodes, filtercolors) if debug else ""
                action(usagetype, sortby, isreverse, filternodes, filtercolors)
            else:
                printargs(usagetype, sortby, isreverse, filternodes, filtercolors) if debug else ""
                action("disk", sortby, isreverse, filternodes, filtercolors)
                action("cpu", sortby, isreverse, filternodes, filtercolors)
                action("memory", sortby, isreverse, filternodes, filtercolors)
            # sleep for 30 seconds        
            time.sleep(5)
            os.system('clear')
            
    else:
        if usagetype != "all":
            printargs(usagetype, sortby, isreverse, filternodes, filtercolors) if debug else ""
            action(usagetype, sortby, isreverse, filternodes, filtercolors)
        else:
            printargs(usagetype, sortby, isreverse, filternodes, filtercolors) if debug else ""
            action("disk", sortby, isreverse, filternodes, filtercolors)
            action("cpu", sortby, isreverse, filternodes, filtercolors)
            action("memory", sortby, isreverse, filternodes, filtercolors)     

    print("")
if __name__ == "__main__":
    main()
