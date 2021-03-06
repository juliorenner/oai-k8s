import logging
import os
import yaml
import time

from datetime import datetime
from utils.k8s import K8S
import utils.constants as constants

from kubernetes import config, client
from kubernetes.client.rest import ApiException

from tenacity import retry, stop_after_delay, retry_if_exception_type, TryAgain, wait_fixed


class SplitsPlacer:

    template_file = ""
    splitsplacer = None

    def __init__(self, template_file: str):
        template_folder = f"{os.path.dirname(os.path.abspath(__file__))}/templates"
        self.template_file = f"{template_folder}/{template_file}"
        self.splitsplacer = next(yaml.safe_load_all(
            open(self.template_file, "r").read()))

    @retry(stop=stop_after_delay(120), retry=retry_if_exception_type(TryAgain),
           wait=wait_fixed(5), reraise=True)
    def create(self):
        """
            Creates the SplitsPlacer resource.
        """

        k8s_cli = K8S.get_custom_object_client()
        try:
            res = k8s_cli.create_namespaced_custom_object(group=constants.CRD_GROUP,
                                                          plural=constants.CRD_KIND_SPLITSPLACER,
                                                          version=constants.APIVERSION_V1BETA1,
                                                          namespace=constants.NAMESPACE_OAI,
                                                          body=self.splitsplacer)
            self.splitsplacer = res
        except ApiException as err:
            if err.status != 409:
                logging.error(
                    f"[SPLITSPLACER] Error creating SplitsPlacer: {err}")
                raise TryAgain

    @retry(stop=stop_after_delay(120), retry=retry_if_exception_type(TryAgain),
           wait=wait_fixed(5), reraise=True)
    def get(self):
        """
            Get SplitsPlacer.
        """
        splitsplacer = None
        splitsplacer_name = self.splitsplacer["metadata"]["name"]
        k8s_cli = K8S.get_custom_object_client()

        try:
            splitsplacer = k8s_cli.get_namespaced_custom_object(group=constants.CRD_GROUP,
                                                                version=constants.APIVERSION_V1BETA1,
                                                                namespace=constants.NAMESPACE_OAI,
                                                                plural=constants.CRD_KIND_SPLITSPLACER,
                                                                name=splitsplacer_name)
        except ApiException as err:
            if err.status != 404:
                logging.error(
                    f"[SPLITSPLACER] Error getting splitsplacer: {err}")
                raise TryAgain

        if splitsplacer is not None:
            self.splitsplacer = splitsplacer

        return splitsplacer

    @retry(stop=stop_after_delay(1200), retry=retry_if_exception_type(TryAgain),
           wait=wait_fixed(constants.WAIT_FIXED_INTERVAL), reraise=True)
    def wait_to_be_finished(self):
        """
            Wait for the splitsplacer to be in a defined states, and if not throws an error.
        """
        try:
            splitsplacer = self.get()
            if splitsplacer is None or ("status" not in splitsplacer) or ("state" not in splitsplacer["status"]):
                raise TryAgain
            status = splitsplacer["status"]["state"]

            if status == constants.STATUS_FINISHED or status == constants.STATUS_ERROR:
                return True
        except ApiException as err:
            logging.error(
                f"[SPLITSPLACER] Unexpected error waiting for splitsplacer to be finished")
            raise

    @retry(stop=stop_after_delay(120), retry=retry_if_exception_type(TryAgain),
           wait=wait_fixed(5), reraise=True)
    def delete(self):
        """
            Deletes the SplitsPlacer.
        """
        splitsplacer = self.get()
        if splitsplacer is None:
            return

        splitsplacer_name = splitsplacer["metadata"]["name"]
        k8s_cli = K8S.get_custom_object_client()
        try:
            k8s_cli.delete_namespaced_custom_object(group=constants.CRD_GROUP,
                                                    version=constants.APIVERSION_V1BETA1,
                                                    namespace=constants.NAMESPACE_OAI,
                                                    plural=constants.CRD_KIND_SPLITSPLACER,
                                                    name=splitsplacer_name,
                                                    body=client.V1DeleteOptions())
        except ApiException as err:
            if err.status != 404:
                logging.error(
                    f"[SPLITSPLACER] Error deleting splitsplacer: {err}")
                raise TryAgain

    def collect_result(self):
        splitsplacer = self.get()

        if splitsplacer["status"]["state"] == constants.STATUS_ERROR:
            return {
                "state": constants.STATUS_ERROR
            }

        links_bandwidth = splitsplacer["status"]["remainingBandwidth"]
        creation_timestamp = splitsplacer["metadata"]["creationTimestamp"]

        hops_count = {}
        for ru in splitsplacer["spec"]["rus"]:
            if "path" not in ru or len(ru["path"]) == 0:
                continue
            hops = len(ru["path"])-1
            hops_count[ru["splitName"]] = hops

        average_hops = sum(hops_count.values())/len(hops_count)

        while True:
            pods = K8S.list_pods()
            ready = True
            if len(pods.items) < splitsplacer["status"]["allocatedRUs"] * 3:
                ready = False
            else:
                for pod in pods.items:
                    if pod.status.phase != "Running":
                        logging.info(
                            f"pod {pod.metadata.name} in state {pod.status.phase}")
                        ready = False
                        break
            if ready:
                logging.info("all pods running")
                break
            logging.info("waiting pods to be ready")
            time.sleep(5)

        initialization_time = {}
        duration = []
        for pod in pods.items:
            logging.info(f"getting logs for pod {pod.metadata.name}")
            pod_logs = K8S.logs(pod.metadata.name).split("\n")

            timestamp = pod_logs[0]

            if "Starting replacer" in timestamp:
                timestamp = timestamp.split("\"")[1]
                timestamp = timestamp[:-1]

            initialization_time[pod.metadata.name] = timestamp
            init_time = datetime.strptime(timestamp, "%Y-%m-%dT%H:%M:%S")
            creation_time = datetime.strptime(
                creation_timestamp, "%Y-%m-%dT%H:%M:%SZ")
            difference = (init_time - creation_time)
            duration.append(difference.total_seconds())

        average_initialization_time = 0
        for v in duration:
            average_initialization_time += v
        average_initialization_time = average_initialization_time/len(duration)

        return {
            "placement": splitsplacer["spec"]["rus"],
            "links_bandwidth": links_bandwidth,
            "creation_timestamp": creation_timestamp,
            "initialization_time": initialization_time,
            "state": constants.STATUS_FINISHED,
            "average_initialization_time": average_initialization_time,
            "average_hops": average_hops,
            "hops_count": sum(hops_count.values()),
            "allocation_time": splitsplacer["status"]["allocationTime"],
            "allocated_rus": splitsplacer["status"]["allocatedRUs"],
            "allocated_percentage": (splitsplacer["status"]["allocatedRUs"]/len(splitsplacer["spec"]["rus"]))*100
        }
