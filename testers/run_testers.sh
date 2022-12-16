#!/bin/bash

C_RESET='\033[0m'
C_RED='\033[0;31m'
C_GREEN='\033[0;32m'
C_BLUE='\033[0;34m'
C_YELLOW='\033[1;33m'

# println echos string
function println() {
    echo -e "$1"
}

# errorln echos i red color
function errorln() {
    println "${C_RED}${1}${C_RESET}"
}

# successln echos in green color
function successln() {
    println "${C_GREEN}${1}${C_RESET}"
}

# infoln echos in blue color
function infoln() {
    println "${C_BLUE}${1}${C_RESET}"
}

# warnln echos in yellow color
function warnln() {
    println "${C_YELLOW}${1}${C_RESET}"
}

function getFiles() {
    ENV=$1
    if [ "$ENV" == "prod" ]; then
        cp ../../config/kafka/producer.properties ${PWD}
        cp ../../config/kafka/server.cer.pem ${PWD}
        cp ../../config/kafka/kafka.client.keystore.jks ${PWD}
        cp ../../config/kafka/kafka.client.truststore.jks ${PWD}
    elif [ "$ENV" == "dev" ]; then
        cp ../config/kafka/producer.properties.dev ${PWD}/producer.properties
    else
        errorln "Unknown environment provided"
        printHelp
        exit 1
    fi
}

function removeFiles() {
    rm producer.properties
    rm server.cer.pem || true
    rm kafka.client.keystore.jks || true
    rm kafka.client.truststore.jks || true
}

function runTesters() {
    ENV=$1
    TESTER=$2
    local COMMAND="go run . -f ./producer.properties"

    if [ "$TESTER" = "sla" ]; then
        infoln "Running SLA producer"
        pushd ./sla_producer || println "Did not find sla tester folder"
        getFiles "$ENV"
        eval "$COMMAND"
        removeFiles
        popd || true
    elif [ "$TESTER" = "vru" ]; then
        infoln "Running VRU producer"
        pushd ./vru_producer || println "Did not find sla tester folder"
        getFiles "$ENV"
        eval "$COMMAND"
        removeFiles
        popd || true
    elif [ "$TESTER" = "parts" ]; then
        infoln "Running Parts producer"
        pushd ./parts_producer || println "Did not find sla tester folder"
        getFiles "$ENV"
        eval "$COMMAND"
        removeFiles
        popd || true
    else
        errorln "Tester ${TESTER} does not exist."
        exit 1
    fi

}

function printHelp() {
    local file=$1
    println "USAGE:"
    println
    println "$file {env} {testers}"
    println "ENVIRONMENT:"
    println "    dev: Development environment"
    println "    prod: Production environment"
    println "TESTERS:"
    println "    all: Run all testers [sla, vru, parts]"
    println "    sla: Run sla tester"
    println "    vru: Run vru tester"
    println "    parts: Run parts tester"
}

if [ $# -lt 2 ] || [ $# -gt 4 ]; then
    printHelp "$0"
    exit 1
fi

ENV=$1
shift

if [ "$1" = "all" ]; then
    runTesters "$ENV" "sla"
    runTesters "$ENV" "vru"
    runTesters "$ENV" "parts"
else
    for var in "$@"; do
        runTesters "$ENV" "$var"
    done
fi
