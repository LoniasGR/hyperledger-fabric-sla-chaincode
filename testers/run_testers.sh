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


function runTesters() {
    TESTER=$1
    COMMAND="go run . -f ../../kafka-config/producer.properties.dev"

    if [ "$TESTER" = "sla" ]; then
        infoln "Running SLA producer"
        pushd ./sla_producer
        eval "$COMMAND"
        popd
    elif [ "$TESTER" = "vru" ]; then
        infoln "Running VRU producer"
        pushd ./vru_producer
        eval "$COMMAND"
        popd
    elif [ "$TESTER" = "parts" ]; then
        infoln "Running Parts producer"
        pushd ./parts_producer
        eval "$COMMAND"
        popd
    else
        errorln "Tester ${TESTER} does not exist."
        exit 1
    fi
}

function printHelp() {
    println "USAGE:"
    println
    println "./run_testers.sh {testers}"
    # println "ENVIRONMENT:"
    # println "    dev (default): Development environment"
    # println "    prod: "
    println "TESTERS:"
    println "    all: Run all testers [sla, vru, parts]"
    println "    sla: Run sla tester"
    println "    vru: Run vru tester"
    println "    parts: Run parts tester"
}

if [ $# -lt 1 ] || [ $# -gt 3 ]; then
    printHelp
    exit 0
elif [ "${1}" = "all" ]; then
    runTesters "sla"
    runTesters "vru"
    runTesters "parts"
else
    for var in "$@"; do
        runTesters "$var"
    done
fi
