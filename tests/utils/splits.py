import os
import yaml
import logging
import time

from datetime import datetime

from kubernetes import config, client, utils
from utils.k8s import K8S
from utils.bfs import count_hops

from kubernetes.client.rest import ApiException
from tenacity import retry, stop_after_delay, retry_if_exception_type, TryAgain, wait_fixed

import utils.constants as constants


class Splits:

    template_file = ""
    splits = []

    def __init__(self, template_file: str):
        template_folder = f"{os.path.dirname(os.path.abspath(__file__))}/templates"
        self.template_file = f"{template_folder}/{template_file}"
        files = yaml.safe_load_all(open(self.template_file, "r").read())

        try:
            while True:
                split = next(files)
                self.splits.append(split)
        except StopIteration:
            return

    @retry(stop=stop_after_delay(120), retry=retry_if_exception_type(TryAgain),
           wait=wait_fixed(5), reraise=True)
    def create(self):
        """
            Creates the Splits resource from yaml file.
        """

        k8s_cli = K8S.get_custom_object_client()

        created_crds = []

        for s in self.splits:
            try:
                created_crds.append(k8s_cli.create_namespaced_custom_object(group=constants.CRD_GROUP,
                                                                            plural=constants.CRD_KIND_SPLITS,
                                                                            version=constants.APIVERSION_V1BETA1,
                                                                            namespace=constants.NAMESPACE_OAI,
                                                                            body=s))
            except ApiException as err:
                if err.status != 409:
                    logging.error(
                        f"[SPLITS] Error creating Splits: {err}")
                    raise TryAgain

        if len(created_crds) > 0:
            self.splits = created_crds

        return created_crds

    @retry(stop=stop_after_delay(120), retry=retry_if_exception_type(TryAgain),
           wait=wait_fixed(5), reraise=True)
    def get(self):
        """
            Gets all the splits created.
        """
        splits = None
        k8s_cli = K8S.get_custom_object_client()
        try:
            splits = k8s_cli.list_namespaced_custom_object(group=constants.CRD_GROUP,
                                                           version=constants.APIVERSION_V1BETA1,
                                                           namespace=constants.NAMESPACE_OAI,
                                                           plural=constants.CRD_KIND_SPLITS)
        except ApiException as err:
            logging.error(f"[SPLITS] Error listing splits: {err}")
            raise TryAgain

        if splits["items"] is not None:
            self.splits = splits["items"]

        return splits["items"]

    @retry(stop=stop_after_delay(120), retry=retry_if_exception_type(TryAgain),
           wait=wait_fixed(5), reraise=True)
    def delete(self):
        """
            Deletes all the splits.
        """
        if self.splits is None:
            return

        k8s_cli = K8S.get_custom_object_client()
        for s in self.splits:
            try:
                name = s["metadata"]["name"]
                k8s_cli.delete_namespaced_custom_object(group=constants.CRD_GROUP,
                                                        version=constants.APIVERSION_V1BETA1,
                                                        namespace=constants.NAMESPACE_OAI,
                                                        plural=constants.CRD_KIND_SPLITS,
                                                        name=name,
                                                        body=client.V1DeleteOptions())
            except ApiException as err:
                if err.status != 404:
                    logging.error(
                        f"[SPLITS] Error deleting split: {err}")
                    raise TryAgain

    def wait_pods_to_be_running(self):
        time.sleep(5)
        while True:
            splits = self.get()
            pods = K8S.list_pods()
            ready = True
            if len(pods.items) < len(splits) * 3:
                ready = False
            else:
                for s in splits:
                    if ("status" not in s or ("cuNode" not in s["status"]) or
                        ("duNode" not in s["status"]) or ("ruNode" not in s["status"]) or
                            s["status"]["cuNode"] == "" or s["status"]["duNode"] == "" or s["status"]["ruNode"] == ""):
                        ready = False
                if ready:
                    for pod in pods.items:
                        if pod.status.phase != "Running":
                            logging.info(
                                f"pod {pod.metadata.name} in state {pod.status.phase}")
                            ready = False
                            break
            if ready:
                logging.info("all pods running")
                break
            time.sleep(5)

    def get_initialization_time(self):
        initialization_time = {}
        pods = K8S.list_pods()
        for pod in pods.items:
            logging.info(f"getting logs for pod {pod.metadata.name}")
            pod_logs = K8S.logs(pod.metadata.name).split("\n")

            timestamp = pod_logs[0]

            if "Starting replacer" in timestamp:
                timestamp = timestamp.split("\"")[1]
                timestamp = timestamp[:-1]

            initialization_time[pod.metadata.name] = timestamp

        return initialization_time

    def collect_result(self):
        self.wait_pods_to_be_running()

        initialization_time = self.get_initialization_time()

        splits = self.get()
        placement = {}
        creation_timestamp = {}
        duration = []
        for s in splits:
            creation_timestamp[s["metadata"]["name"]
                               ] = s["metadata"]["creationTimestamp"]
            if "status" not in s:
                n = s["metadata"]["name"]
                logging.info(f"split {n} does not have status")
                time.sleep(30)

            placement[s["metadata"]["name"]] = {
                "cu": s["status"]["cuNode"],
                "du": s["status"]["duNode"],
                "ru": s["status"]["ruNode"],
                "status": s["status"]["state"]
            }

            for v in initialization_time:
                split_name = v.split("-")[1]
                if s["metadata"]["name"] == split_name:
                    creation_time = datetime.strptime(
                        s["metadata"]["creationTimestamp"], "%Y-%m-%dT%H:%M:%SZ")
                    init_time = datetime.strptime(
                        initialization_time[v], "%Y-%m-%dT%H:%M:%S")
                    difference = (init_time - creation_time)
                    duration.append(difference.total_seconds())

        average_initialization_time = 0
        for v in duration:
            average_initialization_time += v
        average_initialization_time = average_initialization_time/len(duration)

        hops_count = count_hops(placement)
        average_hops = sum(hops_count.values())/len(hops_count)

        return {
            "creation_timestamp": creation_timestamp,
            "initialization_time": initialization_time,
            "placement": placement,
            "hops_count": sum(hops_count.values()),
            "average_hops": average_hops,
            "average_initialization_time": average_initialization_time,
        }
