import argparse
import time
import os

import utils.constants as constants

from utils.k8s import K8S
from utils.splitsplacer import SplitsPlacer

def TestSplitsPlacer(exec_number):

    for n in range(exec_number):
        topology_name = "bw-max-delay-min.yaml"
        splitsplacer = SplitsPlacer(topology_name)

        splitsplacer.create()
        splitsplacer.wait_to_be_finished()

        result = splitsplacer.collect_result()
        output_result(result, topology_name, n)

        splitsplacer.delete()
        wait_cleanup_finished()


def output_result(result: object, file_name: str, exec_number: int):
    logs_file = open(f"{os.getcwd()}/tests/results/{file_name}.txt", "a")
    logs_file.write(f"Execution {exec_number}:\n")
    logs_file.write("State: {}\n".format(result["state"]))
    if result["state"] == constants.STATUS_ERROR:
        logs_file.close()
        return

    logs_file.write("links bandwidth: \n{}\n".format(result["links_bandwidth"]))
    logs_file.write("creation timestamp: {}\n".format(result["creation_timestamp"]))
    logs_file.write("initialization timestamp: {}\n".format(result["initialization_time"]))
    logs_file.close()

def wait_cleanup_finished():
    pods = K8S.list_pods()

    while len(pods.items) > 0:
        time.sleep(5)

        pods = K8S.list_pods()


def main():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument('--number-of-executions', type=int,default=5)
    
    args = parser.parse_args()

    TestSplitsPlacer(args.number_of_executions)

main()