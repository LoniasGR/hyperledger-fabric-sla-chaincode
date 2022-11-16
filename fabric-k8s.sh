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

export TEST_NETWORK_NETWORK_NAME=pledger-dlt
export TEST_NETWORK_LOCAL_REGISTRY_DOMAIN=localhost:5000
export TEST_NETWORK_LOCAL_REGISTRY_PORT=5000

function log_line() {
    echo -e "==============================================" >>network-debug.log
}

function log() {
    echo -e "$@"
}

function unkind() {
    ./network-k8s.sh unkind
}

function kind() {
    ./network-k8s.sh kind
}

function init_cluster() {
    ./network-k8s.sh cluster init
}

function up() {
    ./network-k8s.sh up
    log_line
}

function set_channels() {
    ./network-k8s.sh channel init
    log_line

    ./network-k8s.sh channel create "$SLA_CHANNEL_NAME" 1
    log_line

    ./network-k8s.sh channel create "$VRU_CHANNEL_NAME" 2
    log_line

    ./network-k8s.sh channel create "$PARTS_CHANNEL_NAME" 3
    log_line

    ./network-k8s.sh channel create "$SLA2_CHANNEL_NAME" 4
    log_line
}

function deploy_chaincodes() {

    export CHANNEL_NAME=${SLA_CHANNEL_NAME}
    ./network-k8s.sh chaincode deploy 1 $SLA_CHAINCODE_NAME "$SLA_CC_SRC_PATH"
    log_line

    export CHANNEL_NAME=${VRU_CHANNEL_NAME}
    ./network-k8s.sh chaincode deploy 2 $VRU_CHAINCODE_NAME "$VRU_CC_SRC_PATH"
    log_line

    export CHANNEL_NAME=${PARTS_CHANNEL_NAME}
    ./network-k8s.sh chaincode deploy 3 $PARTS_CHAINCODE_NAME "$PARTS_CC_SRC_PATH"
}

function init_application_config() {
    ./network-k8s.sh application init
}

function identity_management() {
    log "Building identity management pod"
    docker build -t ${TEST_NETWORK_LOCAL_REGISTRY_DOMAIN}/identity-management application/identity_management >>network-debug.log
    docker push ${TEST_NETWORK_LOCAL_REGISTRY_DOMAIN}/identity-management >>network-debug.log
    # Maybe todo: change the namespace here
    kubectl -n "${TEST_NETWORK_NETWORK_NAME}" delete -f kube/identity-management-client.yaml
    kubectl -n "${TEST_NETWORK_NETWORK_NAME}" apply -f kube/identity-management-client.yaml
    log "🏁 Identity management pod built"
    log_line
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

function api() {
    log "Building API pod"
    ./network-k8s.sh application api

}

function explorer() {
    ./network-k8s.sh application explorer

}

function print_help() {
    log "Only valid modes are 'deploy' and 'down'"
}

## Parse mode
if [[ $# -lt 1 ]]; then
    log "USAGE:"
    log "    kind: Set up the the KIND cluster and the container registry"
    log "    cluster: Initialize the cluster"
    log "    up: Bring up all the peers, CAs and orderers of the network, as well as the channels"
    log "    deploy: Bring up the chaincodes and the clients."

    exit 0
else
    MODE=$1
    shift
fi

if [ "${MODE}" == "up" ]; then
    up
    set_channels
elif [ "${MODE}" == "deploy" ]; then
    deploy_chaincodes
    init_application_config
    sla_client
    vru_client
    parts_client
    identity_management
    api
    explorer
elif [ "${MODE}" == "kind" ]; then
    kind
elif [ "${MODE}" == "cluster" ]; then
    init_cluster
elif [ "${MODE}" == "unkind" ]; then
    unkind
else
    print_help
    exit 1
fi
