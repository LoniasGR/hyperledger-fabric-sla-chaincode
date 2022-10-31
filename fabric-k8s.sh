#!/bin/bash


SLA_CHANNEL_NAME=sla
VRU_CHANNEL_NAME=vru
PARTS_CHANNEL_NAME=parts

SLA_CHAINCODE_NAME=slasc-bridge
VRU_CHAINCODE_NAME=vru-positions
PARTS_CHAINCODE_NAME=parts

SLA_CC_SRC_PATH="${PWD}/ccas_sla"
VRU_CC_SRC_PATH="${PWD}/ccas_vru"
PARTS_CC_SRC_PATH="${PWD}/ccas_parts"

function log_line() {
    echo -e "==============================================" >> network-debug.log
}

## Parse mode
if [[ $# -lt 1 ]]; then
    print_help
    exit 0
else
    MODE=$1
    shift
fi

if [ "${MODE}" == "deploy" ]; then
    ./network-k8s.sh unkind
    printf ''  > network-debug.log

    ./network-k8s.sh kind

    log_line

    ./network-k8s.sh cluster init

    log_line

    ./network-k8s.sh up

    log_line

    # Initialize the channel
    ./network-k8s.sh channel init
+
    log_line

    ./network-k8s.sh channel create "$SLA_CHANNEL_NAME" 1

    log_line

    ./network-k8s.sh channel create "$VRU_CHANNEL_NAME" 2

    log_line

    ./network-k8s.sh channel create "$PARTS_CHANNEL_NAME" 3

    log_line

    export CHANNEL_NAME=${SLA_CHANNEL_NAME}
    ./network-k8s.sh chaincode deploy 1 $SLA_CHAINCODE_NAME "$SLA_CC_SRC_PATH"

      export CHANNEL_NAME=${VRU_CHANNEL_NAME}
    ./network-k8s.sh chaincode deploy 2 $VRU_CHAINCODE_NAME "$VRU_CC_SRC_PATH"


      export CHANNEL_NAME=${PARTS_CHANNEL_NAME}
    ./network-k8s.sh chaincode deploy 3 $PARTS_CHAINCODE_NAME "$PARTS_CC_SRC_PATH"

    export CHANNEL_NAME=${SLA_CHANNEL_NAME}
    export CHAINCODE_NAME=${SLA_CHAINCODE_NAME}
    ./network-k8s.sh application

else
    print_help
    exit 1
fi
