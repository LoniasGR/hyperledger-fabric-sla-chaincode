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

function runDispatcher() {
    JSON=$(realpath "${1}")
    CHANNEL=${2}
    ENV=${3}

    if [ "$ENV" = "prod" ]; then
        FILE="../../kafka-config/producer.properties"
    elif [ "$ENV" = "dev" ]; then
            FILE="../../kafka-config/producer.properties.dev"
    else
        errorln "Unknown environment provided"
        printHelp
        exit 1
    fi
    COMMAND="go run ."
    eval "$COMMAND -f $FILE -json $JSON -type $CHANNEL"
}

function printHelp() {
    println "USAGE:"
    println
    println "./run_dispatcher.sh {json} {channel} {env}"
    println "    json: The path to the json file"
    println "    channel: The channel to which it has to be submitted to"
    println "    env: prod or dev"
}

if [ $# -ne 3 ]; then
    printHelp
    exit 0
else
    runDispatcher "${1}" "${2}" "${3}"
fi
