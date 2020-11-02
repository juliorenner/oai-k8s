import argparse
import time
import os
import logging

from datetime import datetime

import utils.constants as constants

from utils.k8s import K8S
from utils.splitsplacer import SplitsPlacer
from utils.splits import Splits

logging.basicConfig(level=logging.INFO)


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


def TestSplits(exec_number: int):

    template_file = "scheduler.yaml"
    for n in range(exec_number):
        splits = Splits(template_file)

        try:
            logging.info("creating splits")
            splits.create()

            logging.info("collecting results")
            result = splits.collect_result()

            output_result(result, template_file, n)
        finally:
            logging.info("deleting splits")
            splits.delete()
            logging.info("waiting for clean up to finish")
            wait_cleanup_finished()


def output_result(result: object, file_name: str, exec_number: int):
    logs_file = open("{}/results/{}.txt".format(os.getcwd(),
                                                file_name.split(".")[0]), "a")
    logs_file.write(f"Execution {exec_number}:\n")

    if "state" in result:
        logs_file.write("State: {}\n".format(result["state"]))
        if result["state"] == constants.STATUS_ERROR:
            logs_file.close()
            return

    if "links_bandwidth" in result:
        logs_file.write("links bandwidth: \n{}\n".format(
            result["links_bandwidth"]))

    if "creation_timestamp" in result:
        logs_file.write("creation timestamp: {}\n".format(
            result["creation_timestamp"]))

    if "initialization_time" in result:
        logs_file.write("initialization timestamp: {}\n".format(
            result["initialization_time"]))

    if "placement" in result:
        logs_file.write("placement: {}\n".format(
            result["placement"]))

    logs_file.close()


def wait_cleanup_finished():
    pods = K8S.list_pods()

    while len(pods.items) > 0:
        time.sleep(5)

        pods = K8S.list_pods()


def output_start_end_times():
    o_file = open("{}/results/{}.txt".format(os.getcwd(), "times"), "a")
    now = datetime.now().strftime("%d/%m/%Y %H:%M:%S\n")
    o_file.write(f"now: {now}")

    o_file.close()


def main():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument('--number-of-executions', type=int, default=30)

    args = parser.parse_args()

    output_start_end_times()

    TestSplits(args.number_of_executions)

    # TestSplitsPlacer(args.number_of_executions, "bw-max-delay-min.yaml")
    # TestSplitsPlacer(args.number_of_executions, "bw-min-link-delay.yaml")
    # TestSplitsPlacer(args.number_of_executions, "bw-random-link-delay.yaml")

    output_start_end_times()


main()
