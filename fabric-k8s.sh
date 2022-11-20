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



function log_line() {
    echo -e "==============================================" >>network-debug.log
}

function log() {
    echo -e "$@"
}

function login() {
    ./network-k8s.sh docker
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

function down() {
    ./network-k8s.sh down
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
    log "Building API pod"
    ./network-k8s.sh application api

}

function explorer() {
    ./network-k8s.sh application explorer

}

function applications() {
    init_application_config
    sla_client
    vru_client
    parts_client
    sla2_client
    identity_management
    api
    explorer
}

function print_help() {
    log "USAGE:"
    log "    kind: Set up the the KIND cluster and the container registry"
    log "    cluster: Initialize the cluster"
    log "    up: Bring up all the peers, CAs and orderers of the network, as well as the channels"
    log "    deploy: Bring up the chaincodes and the clients."
}

## Parse mode
if [[ $# -lt 1 ]]; then
    print_help
    exit 0
else
    MODE=$1
    shift
fi

if [ "${MODE}" == "up" ]; then
    up
elif [ "${MODE}" == "channels" ]; then
    set_channels
elif [ "${MODE}" == "login" ]; then
    login
elif [ "${MODE}" == "chaincodes" ]; then
    deploy_chaincodes
elif [ "${MODE}" == "applications" ]; then
    applications
elif [ "${MODE}" == "kind" ]; then
    kind
elif [ "${MODE}" == "cluster" ]; then
    init_cluster
elif [ "${MODE}" == "down" ]; then
    down
elif [ "${MODE}" == "unkind" ]; then
    unkind
elif [ "${MODE}" == "everything" ]; then
    down
    unkind
    kind
    init_cluster
    up
    set_channels
    login
    deploy_chaincodes
    applications
else
    print_help
    exit 1
fi
