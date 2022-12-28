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

def usage(type):
    # getnodes first

    nodeslist=init()
    print("")
    print("# Disk Usage\n\n%-30s%-10s%-10s%s" % ("NodeName", "Free (GB)", "Max(GB)", "Usage(%)") ) if type == "disk" else ""
    print("# Memory Usage\n\n%-30s%-10s%-10s%s" % ("NodeName", "Free (GB)", "Max(GB)", "Usage(%)") ) if type == "mem" else ""
    print("# CPU Usage\n\n%-30s%-10s%-10s%s" % ("NodeName", "Free (m)", "Max (m)", "Usage(%)") ) if type == "cpu" else ""

    print("-"*80)
    
    for item in nodeslist['items']:
        node_name=item['metadata']['name']
        if type == "disk":
            allocatable=item['status']['allocatable']['ephemeral-storage']
            maximum=item['status']['capacity']['ephemeral-storage'].split('Ki')[0]
            max_inmb=round(int(maximum) / 1024 / 1024)
            allocatable_inmb=round_down(int(allocatable)/1024/1024/1024)
        elif type == "mem":
            allocatable=item['status']['allocatable']['memory'].split('Ki')[0]
            maximum=item['status']['capacity']['memory'].split('Ki')[0]
            max_inmb=round(int(maximum) / 1024 /1024 )
            allocatable_inmb=round_down(int(allocatable)/1024/1024)
        elif type == "cpu":
            allocatable=item['status']['allocatable']['cpu'].split('m')[0]
            maximum=int(item['status']['capacity']['cpu'])*1000
            max_inmb=round(int(maximum))
            allocatable_inmb=round_down(int(allocatable))

        usage=max_inmb - allocatable_inmb
        #usage_percent = (usage/max_inmb) * 100
        usage_percent = random.randint(1,100)
        
        old_stdout = sys.stdout
        sys.stdout = f
        usage=str(printbar(usage_percent))
        sys.stdout = old_stdout
        print("%-30s%-10d%-10d%s" % (node_name, allocatable_inmb, max_inmb, usage) )

def main():

    greet()    
    state=''
    input = sys.argv[1:]
    if not input:
        print('Collecting Disk Usage...')
        usage("disk")
    else:
        if len(input) > 1:
            print("Ignoring the second argument",input[1:])    
        elif input[0] == "--memory" or sys.argv[1] == "-m":
            print("Collecting the Memory Usage")
            usage("mem")
        elif input[0] == "--cpu" or sys.argv[1] == "-c":
            print("Collecting the CPU Usage")
            usage("cpu")
        elif input[0] == "--disk" or sys.argv[1] == "-d":
            print("Collecting the Disk Usage")
            usage("disk")
        elif input[0] == "--all" or sys.argv[1] == "-a":
            print("Collecting the Disk, Memory and CPU Usage")   
            usage("cpu")
            usage("mem") 
            usage("disk") 
        else:
            print("Invalid Option",input[0])
            print("Valid options are --memory | -m , --cpu | -c , --disk | -d")



    
if __name__ == '__main__':
    main()


