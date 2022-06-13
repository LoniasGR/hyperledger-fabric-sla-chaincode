#!/bin/bash
set -e

COMMAND="go run . -f ../../kafka-config/producer.properties.dev"

pushd ./sla_producer
eval "$COMMAND"
popd
pushd ./vru_producer
eval "$COMMAND"
popd
pushd ./parts_producer
eval "$COMMAND"
popd
