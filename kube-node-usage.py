import getopt, sys
import json
import subprocess
from tqdm import tqdm
import math
import os
import sys
# Threading will be enabled in future
#import threading
import random

# For Redirecting the output to /dev/null
f = open(os.devnull, 'w')


def greet():
    print('''
# Kube-Node-Usage
# Release 1.0.2
# https://github.com/AKSarav/Kube-Node-Usage
''')

def round_down(n, decimals=0):
    multiplier = 10 ** decimals
    return math.floor(n * multiplier) / multiplier

def init():
    try:
        nodescmd=subprocess.run(["kubectl", "get" ,"nodes", "-o", "json"], capture_output=True,timeout=10)

        if nodescmd.stdout:
            nodeslist=json.loads(nodescmd.stdout.decode('utf-8'))
            return nodeslist
        else:
            print(nodescmd.stderr.decode('utf-8'))
            nodeslist=json.loads(nodescmd.stderr.decode('utf-8'))
            
        
    except Exception as e:
        print("Failed to execute the Kubectl get nodes command")
        exit(5)



def printbar(diskusage_percent):
        with tqdm(total=100, bar_format="{l_bar}{bar}", position=0, leave=True, file=sys.stdout, ncols=60) as diskbar:
        
            # For Debug
            #diskusage_percent=random.randint(1,100)
            if diskusage_percent < 18:
                diskbar.colour = 'green'
            if diskusage_percent > 18 and diskusage_percent < 70 :
                diskbar.colour = 'yellow'
            if diskusage_percent > 70:
                diskbar.colour = 'red'

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
    

def usage():
    print("-"*50)
    print("# KubeNodeUsage - Usage instructions")
    print("-"*50)
    print("\n* Valid Usage types are --memory | -m , --cpu | -c , --disk | -d, --all | -a")
    print("* Valid Sort by values are --sort=free | --sort=max | --sort=usage | --sort=node ")
    print("\n# Examples:")
    print("-"*10)
    print("* To display the Memory Usage with default Sorting\n# python kube-node-usage.py --memory \n")
    print("* To Display the Memory Usage sort by Usage Percentage\n# python kube-node-usage.py --memory --sort=usage \n")
    print("* To Display the CPU Usage sort by the free/allocatable cpu\n# python kube-node-usage.py --cpu --sort=free \n")
    print("* To Display the Disk Usage sort by the Max Disk\n# python kube-node-usage.py --disk --sort=max \n")
    print("* To Apply the reverse/desc sort with the existing command add --reverse\n# python kube-node-usage.py --disk --sort=max --reverse \n")
    print("-"*50)
    exit(6)


def action(type, sortkey, isreverse):
    # getnodes first
    # print("Starting with Args ",type,sortkey,isreverse)
    nodeslist=init()
    print("")
    
    node_name_len=0
    outputbuffer=[]
    
    for item in nodeslist['items']:
        node_name=item['metadata']['name']
        if type == "disk" or type == "d":
            allocatable=item['status']['allocatable']['ephemeral-storage']
            maximum=item['status']['capacity']['ephemeral-storage'].split('Ki')[0]
            max_inmb=round(int(maximum) / 1024 / 1024)
            allocatable_inmb=round_down(int(allocatable)/1024/1024/1024)
        elif type == "memory":
            allocatable=item['status']['allocatable']['memory'].split('Ki')[0]
            maximum=item['status']['capacity']['memory'].split('Ki')[0]
            max_inmb=round(int(maximum) / 1024 /1024 )
            allocatable_inmb=round_down(int(allocatable)/1024/1024)
        elif type == "cpu":
            allocatable=item['status']['allocatable']['cpu'].split('m')[0]
            maximum=int(item['status']['capacity']['cpu'])*1000
            max_inmb=round(int(maximum))
            allocatable_inmb=int(allocatable)

        if len(node_name)>node_name_len:
            node_name_len = len(node_name)

        usage=max_inmb - allocatable_inmb
        usage_percent = (usage/max_inmb) * 100
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
    print(col_fmt.format(*["NodeName", "Free(GB)", "Max(GB)", "Usage(%)"])) if type == "disk" else ""
    print(col_fmt.format(*["NodeName", "Free(GB)", "Max(GB)", "Usage(%)"])) if type == "memory" else ""
    print(col_fmt.format(*["NodeName", "Free(m)", "Max(m)", "Usage(%)"])) if type == "cpu" else ""
    print("-"*(node_name_len + 70))    
    
    for line in outputbuffer:
        print(col_fmt.format(*[line['node_name'], line['allocatable_inmb'], line['max_inmb'], line['usage']]) )

def main():
    try:
        opts, args = getopt.getopt(sys.argv[1:], ":h", ["help", "sort=", "reverse", "memory", "cpu", "disk", "all"])
    except getopt.GetoptError as err:
        # print help information and exit:
        print(err)  # will print something like "option -a not recognized"
        usage()

    # Defaults
    usagetype="disk"
    sortby="node_name"
    isreverse=False #ASC

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
            print("Fetching all the Usage")
            usagetype="all"
        elif o in ("--reverse"):
            # print("Reverse Sorting")
            isreverse=True
        else:
            print("Not a valid option",o)   
            exit(5)
    greet()
    if usagetype != "all":
        action(usagetype, sortby, isreverse)
    else:
        action("disk", sortby, isreverse)
        action("cpu", sortby, isreverse)
        action("memory", sortby, isreverse)
         


if __name__ == "__main__":
    main()
