# OAI-K8S

## Overview

This github project contains an implementation of two K8S operators to manage the Deployment and Orchestration
of the RAN disagregated functions (CU, DU and RU). In the folder [operator](operator) you can find the developed code.

The operators are currently name `Split` and `SplitsPlacer` but will be renamed to `RANDeployer` and `RANPlacer` respectively.

The operators manage the OpenAirInterface software for now, but could use any ran implementation in the future, as the main
orchestration logic will still be the same, only the `RANDeployer` would need to be adapted to the requirements of the
new RAN software.

Currently, only the disagregation that considers the CU, DU and RU as separated pieces is implemented, but the idea is that
the `RANPlacer` can be extended without much effort to accept different dissagregations. The current implementation is only
a prototype and there is a lot of room for improvements based on the experience acquired implementing it.

Also, there are the folders [replacer](replacer) and [tests](tests). The folder [replacer](replacer) keeps a golang code that is
used in the OAI image initialization to get the configuration information from the `RANDeployer` and provide it to the OAI software.
Therefore, its binary is embedeed in the OAI images. The [tests](tests) folder contains python code that was used to automatically
execute the validations over the operators. The code generates csv files and can:

1. Get initialization metrics, comparing the operators performance with the K8S native behaviour
2. Get the percentage of allocated chains
3. Get the average number of hops resulted from the chains placement
4. Evaluate the usage of CPU and Memory according to the number of chains placed.

All the tests can be executed `N` times according to the argument `--number-of-executions`.

## OAI Images

The OAI images orchestrated by the operators are not in this repository, they can be found in the private github
repository [k8s-build](https://github.com/CROSSHAUL/OAI_Containerized/tree/feat/k8s-build). Access needs to be requested
to have access to it, at least for now.
