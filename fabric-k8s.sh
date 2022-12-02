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

export PLEDGER_NETWORK_CONTAINER_REGISTRY_PORT=443
export PLEDGER_NETWORK_CONTAINER_REGISTRY_HOSTNAME=147.102.19.6
export PLEDGER_NETWORK_CONTAINER_REGISTRY_ADDRESS=$PLEDGER_NETWORK_CONTAINER_REGISTRY_HOSTNAME/pledger

export PLEDGER_NETWORK_NO_VOLUMES=0
export SKIP_SLA1=0
export SKIP_SLA2=0
export SELF_SIGNED_REGISTRY=0

export TAG=${PLEDGER_NETWORK_CONTAINER_REGISTRY_ADDRESS}
export PUSH=1

export HOST_PATH=${HOME}

function login() {
    ./network-k8s.sh docker login
}

function build() {
    ./network-k8s.sh docker build "$TAG" "$PUSH"
}

function init() {
    if [ "${RUNTIME}" == "kind" ]; then
        ./network-k8s.sh kind
    fi

    if [ "${RUNTIME}" == "minikube" ]; then
        if [ "$(docker ps | grep -c minikube)" == 0 ]; then
            minikube start
        fi
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

    if [ $SKIP_SLA1 -eq 1 ]; then
        ./network-k8s.sh channel create "$SLA_CHANNEL_NAME" 1
    fi
    ./network-k8s.sh channel create "$VRU_CHANNEL_NAME" 2

    ./network-k8s.sh channel create "$PARTS_CHANNEL_NAME" 3

    if [ $SKIP_SLA2 -eq 1 ]; then
        ./network-k8s.sh channel create "$SLA2_CHANNEL_NAME" 4
    fi
}

function deploy_chaincodes() {

    if [ $SKIP_SLA1 -eq 1 ]; then
        export CHANNEL_NAME=${SLA_CHANNEL_NAME}
        ./network-k8s.sh chaincode deploy 1 $SLA_CHAINCODE_NAME "$SLA_CC_SRC_PATH"
    fi

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
    init_application_config
    if [ $SKIP_SLA1 -eq 1 ]; then
        sla_client
    fi
    vru_client
    parts_client
    if [ $SKIP_SLA2 -eq 1 ]; then
        sla2_client
    fi
    identity_management
    api
    explorer
}

function print_help() {
    echo "USAGE:"
    echo "$0 RUNTIME COMMAND [ARGUMENTS]"
    echo ""
    echo "TESTED RUNTIMES (mileage may vary):"
    echo "    kind: Kubernetes-in-Docker cluster"
    echo "    microk8s: Microk8s cluster"
    echo "    minikube"
    echo ""
    echo "COMMAND:"
    echo "    build: Build all docker images"
    echo "    deploy: Bring up the chaincodes and the clients"
    echo "    destroy: Bring down the cluster"
    echo "    down: Bring down all the peers, CAs and all the other members of the channels"
    echo "    init: Set up the the cluster, the ingress and cert-manager"
    echo "    login: Login to the container registry. See README for more info"
    echo "    up: Bring up all the peers, CAs and orderers of the network, as well as the channels"
    echo ""
    echo "ARGUMENTS:"
    echo "    --skip-sla-1: Skip the deployment of the SLAv1 channel, chaincode and client"
    echo "    --skip-sla-2: Skip the deployment of the SLAv2 channel and client"
    echo "    --no-volumes: Do not create volumes for storage"
    echo "    --no-push: Used only with *build*. Do not push images to registry."
    echo "    --tag"
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

if [[ $# -ge 1 ]]; then
    while [ $# -gt 0 ]; do
        FLAG=$1
        case $FLAG in
        --no-volumes)
            export PLEDGER_NETWORK_NO_VOLUMES=1
            shift
            ;;
        --self-signed-registry)
            export SELF_SIGNED_REGISTRY=1
            shift
            ;;
        --skip-sla-1)
            export SKIP_SLA1=1
            shift
            ;;
        --skip-sla-2)
            export SKIP_SLA2=1
            shift
            ;;
        --tag)
            export TAG=$2
            shift 2
            ;;
        --no-push)
            export PUSH=0
            shift
            ;;
        *)
            print_help
            exit 0
            ;;
        esac

    done
fi

if [ "${RUNTIME}" == "kind" ]; then
    kubectl config use-context kind-kind
    export PLEDGER_NETWORK_CLUSTER_RUNTIME=kind
    export PLEDGER_NETWORK_CLUSTER_NAME=kind
    export PLEDGER_NETWORK_NGINX_HTTP_PORT=8080
    export PLEDGER_NETWORK_NGINX_HTTPS_PORT=8443
elif [ "${RUNTIME}" == "microk8s" ]; then
    kubectl config use-context microk8s
    export PLEDGER_NETWORK_CLUSTER_RUNTIME=microk8s
    export PLEDGER_NETWORK_CLUSTER_NAME=microk8s
    export PLEDGER_NETWORK_NGINX_HTTP_PORT=8080
    export PLEDGER_NETWORK_NGINX_HTTPS_PORT=8443
elif [ "${RUNTIME}" == "minikube" ]; then
    kubectl config use-context minikube
    export PLEDGER_NETWORK_CLUSTER_RUNTIME=minikube
    export PLEDGER_NETWORK_CLUSTER_NAME=minikube
    export PLEDGER_NETWORK_NGINX_HTTP_PORT=8080
    export PLEDGER_NETWORK_NGINX_HTTPS_PORT=8443
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
elif [ "${MODE}" == "build" ]; then
    build
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
