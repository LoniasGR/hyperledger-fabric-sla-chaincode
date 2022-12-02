#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
set -o errexit

# todo: better handling for input parameters.  Argbash?
# todo: skip storage volume init if deploying to a remote / cloud cluster (ICP IKS ROKS etc...)
# todo: for logging, set up a stack and allow multi-line status output codes
# todo: user:pass auth for tls and ecert bootstrap admins.  here and in the server-config.yaml
# todo: refactor chaincode install to support other chaincode routines
# todo: allow the user to specify the chaincode name (hardcoded as 'basic') both in install and invoke/query
# todo: track down a nasty bug whereby the CA service endpoints (kube services) will occasionally reject TCP connections after network down/up.  This is patched by introducing a 10s sleep after the deployments are up...

# todo: allow relative paths for input arguments.
cd "$(dirname "$0")"

# Set an environment variable based on an optional override (PLEDGER_NETWORK_${name})
# from the calling shell.  If the override is not available, assign the parameter
# to a default value.
function context() {
  local name=$1
  local default_value=$2
  local override_name=PLEDGER_NETWORK_${name}

  export "${name}"="${!override_name:-${default_value}}"
}

context FABRIC_VERSION           2.4
context FABRIC_CA_VERSION        1.5

context CLUSTER_RUNTIME           kind   # or k3s for Rancher
context CONTAINER_NAMESPACE       "" # or "--namespace k8s.io" for containerd / nerdctl

context FABRIC_CONTAINER_REGISTRY   hyperledger
context FABRIC_PEER_IMAGE           "${FABRIC_CONTAINER_REGISTRY}"/fabric-peer:"${FABRIC_VERSION}"
context NETWORK_NAME                pledger-dlt
context CLUSTER_NAME                kind
context KUBE_NAMESPACE              "${NETWORK_NAME}"
context NS                          "${KUBE_NAMESPACE}"
context ORG0_NS                     "${NS}"
context ORG1_NS                     "${NS}"
context ORG2_NS                     "${NS}"
context ORG3_NS                     "${NS}"
context ORG4_NS                     "${NS}"
context DOMAIN                      localho.st
context ORDERER_TIMEOUT             10s # see https://github.com/hyperledger/fabric/issues/3372
context TEMP_DIR                    "${PWD}"/build
context CHAINCODE_BUILDER           ccaas # see https://github.com/hyperledgendary/fabric-builder-k8s/blob/main/docs/TES_NETWORK_K8S.md
context K8S_CHAINCODE_BUILDER_IMAGE ghcr.io/hyperledger-labs/k8s-fabric-peer
context K8S_CHAINCODE_BUILDER_VERSION v0.7.2

context LOG_FILE network.log
context DEBUG_FILE network-debug.log
context LOG_ERROR_LINES 2
context CONTAINER_REGISTRY_HOSTNAME 147.102.19.6
context CONTAINER_REGISTRY_ADDRESS "${CONTAINER_REGISTRY_HOSTNAME}/pledger"
context NGINX_HTTP_PORT 8080
context NGINX_HTTPS_PORT 8443

context RCAADMIN_USER rcaadmin
context RCAADMIN_PASS rcaadminpw
context NO_VOLUMES 0

function print_help() {
  set +x

  log
  log "--- Fabric Information"
  log "Fabric Version     \t\t: ${FABRIC_VERSION}"
  log "Fabric CA Version  \t\t: ${FABRIC_CA_VERSION}"
  log "Network name       \t\t: ${NETWORK_NAME}"
  log "Ingress domain     \t\t: ${DOMAIN}"
  log
  log "--- Cluster Information"
  log "Cluster runtime    \t\t: ${CLUSTER_RUNTIME}"
  log "Cluster name       \t\t: ${CLUSTER_NAME}"
  log "Cluster namespace  \t\t: ${NS}"
  log "Fabric Registry    \t\t: ${FABRIC_CONTAINER_REGISTRY}"
  log "Container Registry \t\t: ${CONTAINER_REGISTRY_ADDRESS}"
  log "nginx http port    \t\t: ${NGINX_HTTP_PORT}"
  log "nginx https port   \t\t: ${NGINX_HTTPS_PORT}"
  log
  log "--- Script Information"
  log "Log file           \t\t: ${LOG_FILE}"
  log "Debug log file     \t\t: ${DEBUG_FILE}"
  log

  echo todo: help output, parse mode, flags, env, etc.
}

. scripts/utils-k8s.sh
. scripts/prereqs.sh
. scripts/kind.sh
. scripts/cluster.sh
. scripts/fabric_config.sh
. scripts/fabric_CAs.sh
. scripts/fabric_network.sh
. scripts/channel.sh
. scripts/chaincode.sh
. scripts/application_connection.sh
. scripts/containers.sh

# check for kind, kubectl, etc.
check_prereqs

# Initialize the logging system - control output to 'network.log' and everything else to 'network-debug.log'
logging_init

## Parse mode
if [[ $# -lt 1 ]]; then
  print_help
  exit 0
else
  MODE=$1
  shift
fi

if [ "${MODE}" == "docker" ]; then
  log "Running docker command"
  docker_command_group "$@"
  log "🏁 - All images are built."
elif [ "${MODE}" == "kind" ]; then
  log "Creating KIND cluster \"${CLUSTER_NAME}\":"
  print_help
  kind_init
  log "🏁 - KIND cluster is ready"

elif [ "${MODE}" == "unkind" ]; then
  log "Deleting KIND cluster \"${CLUSTER_NAME}\":"
  kind_unkind
  log "🏁 - KIND Cluster is gone."

elif [[ "${MODE}" == "cluster" || "${MODE}" == "k8s" || "${MODE}" == "kube" ]]; then
  cluster_command_group "$@"

elif [ "${MODE}" == "up" ]; then
  log "Launching network \"${NETWORK_NAME}\":"
  network_up
  log "🏁 - Network is ready."

elif [ "${MODE}" == "down" ]; then
  log "Shutting down test network  \"${NETWORK_NAME}\":"
  network_down
  log "🏁 - Fabric network is down."

elif [ "${MODE}" == "channel" ]; then
  channel_command_group "$@"

elif [[ "${MODE}" == "chaincode" || "${MODE}" == "cc" ]]; then
  chaincode_command_group "$@"

elif [ "${MODE}" == "anchor" ]; then
  update_anchor_peers "$@"

elif [ "${MODE}" == "application" ]; then
  log "Getting application connection information:"
  application_command_group "$@"
  log "🏁 - Application connection information ready."

else
  print_help
  exit 1
fi
