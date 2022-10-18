#!/bin/bash

set -o errexit

cd "$(dirname "$0")"

. scripts/utils.sh

context FABRIC_VERSION            2.4.4
context FABRIC_CA_VERSION         1.5.5

context CLUSTER_RUNTIME           kind                  # or k3s for Rancher
context CONTAINER_CLI             docker                # or nerdctl for containerd
context CONTAINER_NAMESPACE       ""                    # or "--namespace k8s.io" for containerd / nerdctl
context STORAGE_CLASS             standard
context KUSTOMIZE_BUILD           "kubectl kustomize"
context STAGE_DOCKER_IMAGES       true
context FABRIC_CONTAINER_REGISTRY hyperledger

context NAME                      test-network
context NS                        "${NAME}"
context CLUSTER_NAME              "$CLUSTER_RUNTIME"
context KUBE_DNS_DOMAIN           "${NS}".svc.cluster.local
context INGRESS_DOMAIN            localho.st
context COREDNS_DOMAIN_OVERRIDE   true
context LOG_FILE                  network.log
context DEBUG_FILE                network-debug.log
context LOG_ERROR_LINES           1
context USE_LOCAL_REGISTRY        true
context LOCAL_REGISTRY_NAME       kind-registry
context LOCAL_REGISTRY_PORT       5000
context LOCAL_REGISTRY_INTERFACE  127.0.0.1
context NGINX_HTTP_PORT           80
context NGINX_HTTPS_PORT          443

context CONSOLE_DOMAIN            "${INGRESS_DOMAIN}"
context CONSOLE_USERNAME          admin
context CONSOLE_PASSWORD          password

# TODO: use new cc logic from test-network
context CHANNEL_NAME              mychannel
context ORDERER_TIMEOUT           15s
context CHAINCODE_NAME            asset-transfer-basic
context CHAINCODE_IMAGE           ghcr.io/hyperledgendary/fabric-ccaas-asset-transfer-basic:latest
context CHAINCODE_LABEL           basic_1.0

context CA_IMAGE                  "${FABRIC_CONTAINER_REGISTRY}"/fabric-ca
context CA_IMAGE_LABEL            "${FABRIC_CA_VERSION}"
context PEER_IMAGE                "${FABRIC_CONTAINER_REGISTRY}"/fabric-peer
context PEER_IMAGE_LABEL          "${FABRIC_VERSION}"
context ORDERER_IMAGE             "${FABRIC_CONTAINER_REGISTRY}"/fabric-orderer
context ORDERER_IMAGE_LABEL       "${FABRIC_VERSION}"
context TOOLS_IMAGE               "${FABRIC_CONTAINER_REGISTRY}"/fabric-tools
context TOOLS_IMAGE_LABEL         "${FABRIC_VERSION}"
context OPERATOR_IMAGE            ghcr.io/hyperledger-labs/fabric-operator
context OPERATOR_IMAGE_LABEL      latest-amd64
context INIT_IMAGE                registry.access.redhat.com/ubi8/ubi-minimal
context INIT_IMAGE_LABEL          latest
context GRPCWEB_IMAGE             ghcr.io/hyperledger-labs/grpc-web
context GRPCWEB_IMAGE_LABEL       latest
context COUCHDB_IMAGE             couchdb
context COUCHDB_IMAGE_LABEL       3.2.2
context CONSOLE_IMAGE             ghcr.io/hyperledger-labs/fabric-console
context CONSOLE_IMAGE_LABEL       latest
context DEPLOYER_IMAGE            ghcr.io/ibm-blockchain/fabric-deployer
context DEPLOYER_IMAGE_LABEL      latest-amd64

export FABRIC_OPERATOR_IMAGE=${OPERATOR_IMAGE}:${OPERATOR_IMAGE_LABEL}
export FABRIC_CONSOLE_IMAGE=${CONSOLE_IMAGE}:${CONSOLE_IMAGE_LABEL}
export FABRIC_DEPLOYER_IMAGE=${DEPLOYER_IMAGE}:${DEPLOYER_IMAGE_LABEL}
export FABRIC_CA_IMAGE=${CA_IMAGE}:${CA_IMAGE_LABEL}
export FABRIC_PEER_IMAGE=${PEER_IMAGE}:${PEER_IMAGE_LABEL}
export FABRIC_ORDERER_IMAGE=${ORDERER_IMAGE}:${ORDERER_IMAGE_LABEL}
export FABRIC_TOOLS_IMAGE=${TOOLS_IMAGE}:${TOOLS_IMAGE_LABEL}

export TEMP_DIR=${PWD}/temp


function print_help() {
  log
  log "--- Fabric Information"
  log "Fabric Version     \t\t: ${FABRIC_VERSION}"
  log "Fabric CA Version    \t: ${FABRIC_CA_VERSION}"
  log "Container Registry   \t: ${FABRIC_CONTAINER_REGISTRY}"
  log "Network name       \t\t: ${NAME}"
  log "Channel name       \t\t: ${CHANNEL_NAME}"
  log
  log "--- Chaincode Information"
  log "Chaincode name      \t\t: ${CHAINCODE_NAME}"
  log "Chaincode image      \t: ${CHAINCODE_IMAGE}"
  log "Chaincode label      \t: ${CHAINCODE_LABEL}"
  log
  log "--- Cluster Information"
  log "Cluster runtime      \t: ${CLUSTER_RUNTIME}"
  log "Cluster name       \t\t: ${CLUSTER_NAME}"
  log "Cluster namespace    \t: ${NS}"
  log "Fabric Registry      \t: ${FABRIC_CONTAINER_REGISTRY}"
  log "Local Registry     \t\t: ${LOCAL_REGISTRY_NAME}"
  log "Local Registry port  \t: ${LOCAL_REGISTRY_PORT}"
  log "nginx http port      \t: ${NGINX_HTTP_PORT}"
  log "nginx https port     \t: ${NGINX_HTTPS_PORT}"
  log
  log "--- Script Information"
  log "Log file           \t\t: ${LOG_FILE}"
  log "Debug log file     \t\t: ${DEBUG_FILE}"
  log

  echo todo: help output, parse mode, flags, env, etc.
}

. scripts/prereqs.sh
. scripts/kind.sh
# . scripts/cluster.sh
# . scripts/console.sh
# . scripts/test_network.sh
# . scripts/channel.sh
# . scripts/chaincode.sh

# check for kind, kubectl, etc.
check_prereqs
exit 0
# Initialize the logging system - control output to 'network.log' and everything else to 'network-debug.log'
logging_init

## Parse mode
if [[ $# -lt 1 ]] ; then
  print_help
  exit 0
else
  MODE=$1
  shift
fi

if [ "${MODE}" == "kind" ]; then
  log "Initializing kind cluster \"${CLUSTER_NAME}\":"
  kind_init
  log "🏁 - Cluster is ready."

elif [ "${MODE}" == "unkind" ]; then
  log "Deleting kind cluster \"${CLUSTER_NAME}\":"
  kind_unkind
  log "🏁 - Cluster is gone."

elif [[ "${MODE}" == "cluster" || "${MODE}" == "k8s" || "${MODE}" == "kube" ]]; then
  cluster_command_group "$@"

elif [ "${MODE}" == "channel" ]; then
  channel_command_group "$@"

elif [[ "${MODE}" == "chaincode" || "${MODE}" == "cc" ]]; then
  chaincode_command_group "$@"

elif [ "${MODE}" == "up" ]; then
  log "Launching network \"${NAME}\":"
  network_up
  log "🏁 - Network is ready."

elif [ "${MODE}" == "down" ]; then
  log "Shutting down test network  \"${NAME}\":"
  network_down
  log "🏁 - Fabric network is down."

elif [ "${MODE}" == "operator" ]; then
  log "Launching Fabric operator"
  launch_operator
  log "🏁 - Operator is ready."

elif [ "${MODE}" == "console" ]; then
  log "Launching Fabric Operations Console"
  console_up
  log "🏁 - Console is ready"

else
  print_help
  exit 1
fi