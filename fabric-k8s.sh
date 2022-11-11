#!/bin/bash

export SLA_CHANNEL_NAME=sla
export VRU_CHANNEL_NAME=vru
export PARTS_CHANNEL_NAME=parts

export SLA_CHAINCODE_NAME=slasc-bridge
export VRU_CHAINCODE_NAME=vru-positions
export PARTS_CHAINCODE_NAME=parts

export SLA_CC_SRC_PATH="${PWD}/ccas_sla"
export VRU_CC_SRC_PATH="${PWD}/ccas_vru"
export PARTS_CC_SRC_PATH="${PWD}/ccas_parts"

export TEST_NETWORK_NETWORK_NAME=pledger-dlt
export TEST_NETWORK_LOCAL_REGISTRY_DOMAIN=localhost:8080
export TEST_NETWORK_LOCAL_REGISTRY_PORT=8080

function log_line() {
    echo -e "==============================================" >>network-debug.log
}

function log() {
    echo -e "$@"
}

function down() {
    ./network-k8s.sh unkind
}
function deploy() {
    printf '' >network-debug.log

    ./network-k8s.sh kind
    log_line

    ./network-k8s.sh cluster init
    log_line

    ./network-k8s.sh up
    log_line

    # Initialize the channel
    ./network-k8s.sh channel init
    log_line

    ./network-k8s.sh channel create "$SLA_CHANNEL_NAME" 1
    log_line

    ./network-k8s.sh channel create "$VRU_CHANNEL_NAME" 2
    log_line

    ./network-k8s.sh channel create "$PARTS_CHANNEL_NAME" 3
    log_line

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
    printf '' >network-debug.log
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
    printf '' >network-debug.log
    log "Building SLA client pod"
    export CHANNEL_NAME=${SLA_CHANNEL_NAME}
    export CHAINCODE_NAME=${SLA_CHAINCODE_NAME}

    cp config/kafka/consumer.properties application/sla_client/
    cp config/kafka/kafka.client.keystore.jks application/sla_client/
    cp config/kafka/kafka.client.truststore.jks application/sla_client/
    cp config/kafka/server.cer.pem application/sla_client/

    ./network-k8s.sh application create 1
    docker build -t ${TEST_NETWORK_LOCAL_REGISTRY_DOMAIN}/sla-client application/sla_client >>network-debug.log
    docker push ${TEST_NETWORK_LOCAL_REGISTRY_DOMAIN}/sla-client >>network-debug.log
    # Maybe todo: change the namespace here
    kubectl -n "${TEST_NETWORK_NETWORK_NAME}" delete -f kube/sla-client-deployment.yaml
    kubectl -n "${TEST_NETWORK_NETWORK_NAME}" apply -f kube/sla-client-deployment.yaml
    log "🏁 SLA client pod built"
}

function vru_client() {
    printf '' >network-debug.log
    log "Building VRU client pod"
    export CHANNEL_NAME=${VRU_CHANNEL_NAME}
    export CHAINCODE_NAME=${VRU_CHAINCODE_NAME}

    cp config/kafka/consumer.properties application/vru_client/
    cp config/kafka/kafka.client.keystore.jks application/vru_client/
    cp config/kafka/kafka.client.truststore.jks application/vru_client/
    cp config/kafka/server.cer.pem application/vru_client/

    ./network-k8s.sh application create 2
    docker build -t ${TEST_NETWORK_LOCAL_REGISTRY_DOMAIN}/vru-client application/vru_client >>network-debug.log
    docker push ${TEST_NETWORK_LOCAL_REGISTRY_DOMAIN}/vru-client >>network-debug.log
    # Maybe todo: change the namespace here
    kubectl -n "${TEST_NETWORK_NETWORK_NAME}" delete -f kube/vru-client-deployment.yaml
    kubectl -n "${TEST_NETWORK_NETWORK_NAME}" apply -f kube/vru-client-deployment.yaml
    log "🏁 VRU client pod built"
}

function parts_client() {
    printf '' >network-debug.log
    log "Building Parts client pod"
    export CHANNEL_NAME=${PARTS_CHANNEL_NAME}
    export CHAINCODE_NAME=${PARTS_CHAINCODE_NAME}

    cp config/kafka/consumer.properties application/parts_client/
    cp config/kafka/kafka.client.keystore.jks application/parts_client/
    cp config/kafka/kafka.client.truststore.jks application/parts_client/
    cp config/kafka/server.cer.pem application/parts_client/

    ./network-k8s.sh application create 3
    docker build -t ${TEST_NETWORK_LOCAL_REGISTRY_DOMAIN}/parts-client application/parts_client >>network-debug.log
    docker push ${TEST_NETWORK_LOCAL_REGISTRY_DOMAIN}/parts-client >>network-debug.log
    # Maybe todo: change the namespace here
    kubectl -n "${TEST_NETWORK_NETWORK_NAME}" delete -f kube/parts-client-deployment.yaml
    kubectl -n "${TEST_NETWORK_NETWORK_NAME}" apply -f kube/parts-client-deployment.yaml
    log "🏁 Parts client pod built"
}


function api() {
     printf '' >network-debug.log
    log "Building API pod"
    ./network-k8s.sh application api

}

function explorer() {
    printf '' >network-debug.log
    ./network-k8s.sh application explorer

}


## Parse mode
if [[ $# -lt 1 ]]; then
    log "Only valid mode is 'deploy'"
    exit 0
else
    MODE=$1
    shift
fi

if [ "${MODE}" == "deploy" ]; then
    down
    deploy
    init_application_config
    sla_client
    vru_client
    parts_client
    identity_management
    api
    explorer
elif [ "${MODE}" == "down" ]; then
    down
else
    log "Only valid modes are 'deploy' and 'down'"
    exit 1
fi
