#!/bin/bash

export SLA_CHANNEL_NAME=sla
export VRU_CHANNEL_NAME=vru
export PARTS_CHANNEL_NAME=parts
export SLA2_CHANNEL_NAME=sla2.0

export SLA_CHAINCODE_NAME=slasc-bridge
export VRU_CHAINCODE_NAME=vru-positions
export PARTS_CHAINCODE_NAME=parts

export SLA_CC_SRC_PATH="${PWD}/ccas_sla"
export VRU_CC_SRC_PATH="${PWD}/ccas_vru"
export PARTS_CC_SRC_PATH="${PWD}/ccas_parts"

export TEST_NETWORK_LOCAL_REGISTRY_HOSTNAME=147.102.19.6
export TEST_NETWORK_LOCAL_REGISTRY_PORT=443

function login() {
    ./network-k8s.sh docker
}

function init() {
    if [ "${RUNTIME}" == "kind" ]; then
        ./network-k8s.sh kind
    fi
    ./network-k8s.sh cluster init
}

function destroy() {
    ./network-k8s.sh cluster clean
    if [ "${RUNTIME}" == "kind" ]; then
        ./network-k8s.sh unkind
    fi
}

function up() {
    ./network-k8s.sh up
}

function down() {
    ./network-k8s.sh down
}

function set_channels() {
    ./network-k8s.sh channel init

    ./network-k8s.sh channel create "$SLA_CHANNEL_NAME" 1

    ./network-k8s.sh channel create "$VRU_CHANNEL_NAME" 2

    ./network-k8s.sh channel create "$PARTS_CHANNEL_NAME" 3

    ./network-k8s.sh channel create "$SLA2_CHANNEL_NAME" 4
}

function deploy_chaincodes() {

    export CHANNEL_NAME=${SLA_CHANNEL_NAME}
    ./network-k8s.sh chaincode deploy 1 $SLA_CHAINCODE_NAME "$SLA_CC_SRC_PATH"

    export CHANNEL_NAME=${VRU_CHANNEL_NAME}
    ./network-k8s.sh chaincode deploy 2 $VRU_CHAINCODE_NAME "$VRU_CC_SRC_PATH"

    export CHANNEL_NAME=${PARTS_CHANNEL_NAME}
    ./network-k8s.sh chaincode deploy 3 $PARTS_CHAINCODE_NAME "$PARTS_CC_SRC_PATH"
}

function init_application_config() {
    ./network-k8s.sh application init
}

function identity_management() {
    ./network-k8s.sh application identity_management
}

function sla_client() {
    ./network-k8s.sh application sla
}

function vru_client() {
    ./network-k8s.sh application vru
}

function parts_client() {
    ./network-k8s.sh application parts
}

function sla2_client() {
    ./network-k8s.sh application sla2
}

function api() {
    ./network-k8s.sh application api
}

function explorer() {
    ./network-k8s.sh application explorer
}

function applications() {
    # init_application_config
    # sla_client
    # vru_client
    parts_client
    # sla2_client
    # identity_management
    # api
    # explorer
}

function print_help() {
    echo "USAGE:"
    echo "$0 RUNTIME COMMAND"
    echo ""
    echo "RUNTIME:"
    echo "    kind: Kubernetes-in-Docker cluster"
    echo "    microk8s: Microk8s cluster"
    echo ""
    echo "COMMAND:"
    echo "    init: Set up the the cluster, the ingress and cert-manager"
    echo "    destroy: Bring down the cluster"
    echo "    up: Bring up all the peers, CAs and orderers of the network, as well as the channels"
    echo "    deploy: Bring up the chaincodes and the clients."
}

## Parse mode
if [[ $# -lt 2 ]]; then
    print_help "$@"
    exit 0
else
    RUNTIME=$1
    MODE=$2
    shift 2
fi

if [ "${RUNTIME}" == "kind" ]; then
    kubectl config use-context kind-kind
    export TEST_NETWORK_CLUSTER_RUNTIME=kind
    export TEST_NETWORK_CLUSTER_NAME=kind
    export TEST_NETWORK_NGINX_HTTP_PORT=8080
    export TEST_NETWORK_NGINX_HTTPS_PORT=8443
elif [ "${RUNTIME}" == "microk8s" ]; then
    kubectl config use-context microk8s
    export TEST_NETWORK_CLUSTER_RUNTIME=microk8s
    export TEST_NETWORK_CLUSTER_NAME=microk8s
    export TEST_NETWORK_NGINX_HTTP_PORT=9080
    export TEST_NETWORK_NGINX_HTTPS_PORT=9443
else
    print_help
    exit 1
fi

if [ "${MODE}" == "init" ]; then
    init
elif [ "${MODE}" == "destroy" ]; then
    destroy
elif [ "${MODE}" == "up" ]; then
    up
elif [ "${MODE}" == "channels" ]; then
    set_channels
elif [ "${MODE}" == "login" ]; then
    login
elif [ "${MODE}" == "chaincodes" ]; then
    deploy_chaincodes
elif [ "${MODE}" == "applications" ]; then
    applications
elif [ "${MODE}" == "down" ]; then
    down
elif [ "${MODE}" == "full" ]; then
    down
    destroy
    init
    up
    set_channels
    login
    deploy_chaincodes
    applications
else
    print_help
    exit 1
fi
