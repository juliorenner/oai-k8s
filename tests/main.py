import argparse
import time
import os
import logging

from datetime import datetime

import utils.constants as constants

from utils.k8s import K8S
from utils.splitsplacer import SplitsPlacer

logging.basicConfig(level = logging.INFO)

def TestSplitsPlacer(exec_number: int, topology_name: str):

    for n in range(exec_number):
        splitsplacer = SplitsPlacer(topology_name)

        try:
            logging.info("creating splitsplacer")
            splitsplacer.create()
            logging.info("waiting to be finished")
            splitsplacer.wait_to_be_finished()

            logging.info("collecting results")
            result = splitsplacer.collect_result()
            logging.info("outputing results")
            output_result(result, topology_name, n)
        finally:
            logging.info("deleting splitsplacer")
            splitsplacer.delete()
            logging.info("waiting for clean up to finish")
            wait_cleanup_finished()


def output_result(result: object, file_name: str, exec_number: int):
    logs_file = open("{}/results/{}.txt".format(os.getcwd(), file_name.split(".")[0]), "a")
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


def output_start_end_times():
    logs_file = open("{}/results/{}.txt".format(os.getcwd(), "times"), "a")
    now = datetime.now().strftime("%d/%m/%Y %H:%M:%S")
    logs_files.write(f"now: {now}")

    logs_file.close()

def main():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument('--number-of-executions', type=int,default=30)
    
    args = parser.parse_args()
    
    output_start_end_times()

    TestSplitsPlacer(args.number_of_executions, "bw-max-delay-min.yaml")
    TestSplitsPlacer(args.number_of_executions, "bw-min-link-delay.yaml")
    TestSplitsPlacer(args.number_of_executions, "bw-random-link-delay.yaml")

    output_start_end_times()

main()