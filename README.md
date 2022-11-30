# Hyperledger Fabric SLA Chaincode

An as-flexible-as-it can-be deployment system for the DLT infrastructure on Kubernetes.

Can probably be used in any kubernetes runtime with minimal configuration. For now, the only expectations
are:

* A kubernetes cluster running somewhere locally or on the cloud.
* A Container Networking Interface (CNI), with or without support for NetworkPolicies. (This should be bundled with your
kubernetes runtime, unless you go for bare metal)
* A KV database (any will do, most K8s installations come with etcd installed).
* A DNS and service discovery service (K8s usually comes bundled with CoreDNS out of the box, or easy very easy to enable).

The rest (Ingress, Storage, Certificates management) will be installed in the cluster creation process.

## Preparations

To deploy the network, the kafka configuration is needed for the consumers. To create it:

1. Install `openjdk-8-jre-headless` or similar (`keytool` is needed) and `openssl`.
2. Run `scripts/JKS2PEM.sh`.
   For example run:
   ```bash
    ./scripts/JKS2PEM.sh ./kafka-config/kafka.client.truststore.jks ./kafka-config/server.cer.pem
   ```
3. Copy all kafka configuration files to `config/kafka`.
4. Copy `docker_credentials.json.example` to `docker_credentials.json` and change the credentials so that you can
   push on the registry of your choice.

## Deploy on Kubernetes

RUNTIME marks your K8s runtime.

Important enviromental variables:

```
NO_VOLUMES (0 or 1): Controlls wether there are volumes used or emptyDir
SLA_CHANNEL_NAME: Name of SLA channel
VRU_CHANNEL_NAME: Name of VRU channel
PARTS_CHANNEL_NAME: Name of parts channel
SLA2_CHANNEL_NAME: Name of SLA 2.0 channel
SLA_CHAINCODE_NAME: SLA chaincode name
VRU_CHAINCODE_NAME: VRU chaincode name
PARTS_CHAINCODE_NAME: Parts chaincode name
SLA_CC_SRC_PATH: SLA chaincode path
VRU_CC_SRC_PATH: VRU chaincode path
PARTS_CC_SRC_PATH: Parts chaincode path
PLEDGER_NETWORK_CONTAINER_REGISTRY_HOSTNAME: Container registry hostname
PLEDGER_NETWORK_CONTAINER_REGISTRY_PORT: Container registry port
```

`fabric-k8s.sh` arguments:
* `--no-volumes`: Disable volume mounting and uses emptyDirs. (EXPERIMENTAL: DOES NOT WORK)
* `--skip-sla1`: Disable the creation of SLAv1 channel, chaincode and client. (EXPERIMENTAL: MIGHT BE BUGGY)
* `--skip-sla2`: Disable the creation of SLAv2 channel and client. (EXPERIMENTAL: MIGHT BE BUGGY)

1. Run `./fabric-k8s.sh RUNTIME build [--tag TAG] [--no-push]`
   This will build with a specific optional tag and push all the container images.
2. Run `./fabric-k8s.sh RUNTIME init`.
   Creates the KIND cluster, sets up ingress and cert-manager
3. Run `./fabric-k8s.sh RUNTIME up`.
   Brings up the CAs, orderers and peers.
4. Run `./fabric-k8s.sh RUNTIME channels`.
   Brings up the channels.
5. Login to the container registry by running `./fabric-k8s.sh login`.
   This needs to happen now, because namespace to have been created.
6. Run `./fabric-k8s.sh RUNTIME chaincodes`.
   Brings up the channels.
7. Run `./fabric-k8s.sh RUNTIME applications`.
   Deploys chaincodes and clients.

## Shut down network

Run `./fabric-k8s.sh RUNTIME down`

## Remove the KIND cluster

Run `./fabric-k8s.sh unkind`


## Changes needed to run on cloud

Override the corresponding variables from `network-k8s.sh` with the proper ones.
