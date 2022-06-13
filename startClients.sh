#!/bin/bash

set -e

if [ "$1" = "dev" ]; then
    CONSUMER_FILE="consumer.properties.dev"
else
    CONSUMER_FILE="consumer.properties"
fi

if [ "$1" = "down" ]; then
    for FOLDER in "sla_client" "vru_client" "parts_client"; do
        rm ./application/"$FOLDER"/$CONSUMER_FILE
        rm ./application/"$FOLDER"/connection-org1.yaml
        rm -rf ./application/"$FOLDER"/msp
    done
else
    for FOLDER in "sla_client" "vru_client" "parts_client"; do
        cp ./kafka-config/$CONSUMER_FILE ./application/"$FOLDER"
        cp ./organizations/peerOrganizations/org1.example.com/connection-org1.yaml ./application/"$FOLDER"
        cp -r ./organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp ./application/"$FOLDER"/msp
    done
fi
