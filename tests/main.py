import argparse
import time
import os
import logging
import csv

from datetime import datetime

import utils.constants as constants

from utils.k8s import K8S
from utils.splitsplacer import SplitsPlacer
from utils.splits import Splits

logging.basicConfig(level=logging.INFO)


def TestSplitsPlacer(exec_number: int, topology_name: str, resources_validation: bool=False):

    for n in range(exec_number):
        splitsplacer = SplitsPlacer(topology_name)

        try:
            logging.info("creating splitsplacer")
            splitsplacer.create()
            logging.info("waiting to be finished")
            splitsplacer.wait_to_be_finished()

            if resources_validation:
                resources_validation and time.sleep(60)
            else:
                logging.info("collecting results")
                result = splitsplacer.collect_result()

                logging.info("outputing csv")
                output_csv(result, topology_name, True, n)

                logging.info("outputing results")
                output_result(result, topology_name, n)
        finally:
            logging.info("deleting splitsplacer")
            splitsplacer.delete()
            logging.info("waiting for clean up to finish")
            wait_cleanup_finished()


def TestSplits(exec_number: int, template_file: str, resources_validation: bool=False):

    for n in range(exec_number):
        splits = Splits(template_file)

        try:
            logging.info("creating splits")
            splits.create()

            if resources_validation:
                time.sleep(60)
            else:
                logging.info("collecting results")
                result = splits.collect_result()

                logging.info("outputing csv")
                output_csv(result, template_file, False, n)

                logging.info("outputing results")
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

    if "average_initialization_time" in result:
        logs_file.write("average initialization time: {}\n".format(
            result["average_initialization_time"]))

    if "average_hops" in result:
        logs_file.write("average_hops: {}\n".format(
            result["average_hops"]))

    if "hops_count" in result:
        logs_file.write("hops_count: {}\n".format(
            result["hops_count"]))

    if "allocated_rus" in result:
        logs_file.write("allocated rus: {}\n".format(
            result["allocated_rus"]))

    if "allocated_percentage" in result:
        logs_file.write("allocated rus percentage: {}\n".format(
            result["allocated_percentage"]))

    if "allocation_time" in result:
        logs_file.write("splits placer allocation_time: {}\n".format(
            result["allocation_time"]))

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


def output_csv(result: object, file_name: str, splitsPlacer: bool, exec_number: int):
    output_filename = file_name.split(".")[0]
    output_file = "{}/results/{}.csv".format(os.getcwd(),
                                             output_filename)
    with open(output_file, "a") as csv_file:
        csv_writer = csv.writer(csv_file, delimiter=';',
                                quotechar='|', quoting=csv.QUOTE_MINIMAL)
        output_line = []
        output_line.append(exec_number)
        if splitsPlacer:
            output_line.append(result["state"])

        output_line.append(result["average_initialization_time"])
        output_line.append(result["average_hops"])
        output_line.append(result["hops_count"])
        if splitsPlacer:
            output_line.append(result["allocated_rus"])
            output_line.append(result["allocated_percentage"])
            output_line.append(result["allocation_time"])
        csv_writer.writerow(output_line)


def wait_cleanup_finished():
    pods = K8S.list_pods()

    while len(pods.items) > 0:
        time.sleep(5)

        pods = K8S.list_pods()


def output_start_end_times(prefix: str):
    o_file = open("{}/results/{}.txt".format(os.getcwd(), "times"), "a")
    now = datetime.now().strftime("%d/%m/%Y %H:%M:%S\n")
    o_file.write(f"{prefix}: {now}")

    o_file.close()


def main():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument('--number-of-executions', type=int, default=10)
    parser.add_argument('--resources-validation', type=bool, default=False)

    args = parser.parse_args()

    output_start_end_times("start")

    if args.resources_validation:
        output_start_end_times("resources validation")

        for i in range(1, 7):
            TestSplits(args.number_of_executions,
                    f"resources-{i}.yaml", args.resources_validation)        

        output_start_end_times("resources validation")
    else:
        for i in range(1, 7):
            TestSplitsPlacer(args.number_of_executions,
                            f"bw-max-delay-min-{i*3}.yaml")
        for i in range(1, 7):
            TestSplitsPlacer(args.number_of_executions,
                            f"bw-random-link-delay-{i*3}.yaml")

        # for i in range(1, 7):
        #     TestSplitsPlacer(args.number_of_executions,
        #                     f"bw-min-link-delay-{i*3}.yaml")

        # for i in range(1, 7):
        #     TestSplits(args.number_of_executions,
        #             f"scheduler-{i*3}.yaml")

    output_start_end_times("end")


main()
