import logging
import os
import yaml

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
            print(err)
            if err.status != 409:
                logging.error(
                    f"[SPLITSPLACER] Error creating SplitsPlacer: {err}")
                raise

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
                raise

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
                raise

    def collect_result(self):
        splitsplacer = self.get()

        if splitsplacer["status"]["state"] == constants.STATUS_ERROR:
            return {
                "state": constants.STATUS_ERROR
            }

        links_bandwidth = splitsplacer["status"]["remainingBandwidth"]
        creation_timestamp = splitsplacer["metadata"]["creationTimestamp"]

        pods = K8S.list_pods()

        initialization_time = {}
        for pod in pods:
            pod_logs = K8S.logs(pod).split("\n")

            timestamp = pod_logs[0]

            initialization_time[pod] = timestamp

        return {
            "links_bandwidth": links_bandwidth,
            "creation_timestamp": creation_timestamp,
            "initialization_time": initialization_time,
            "state": constants.STATUS_FINISHED
        }
