#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0
#
# Run as
# ./startFabric sla slasc_bridge
# ./startFabric vru vru_positions
#
#####################################

# Exit on first error
set -e

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1
starttime=$(date +%s)
CC_SRC_LANGUAGE="go"
SLA_CC_SRC_PATH="${PWD}/chaincode_sla"
VRU_CC_SRC_PATH="${PWD}/chaincode_vru"

SLA_CHANNEL_NAME=sla
VRU_CHANNEL_NAME=vru

SLA_CHAINCODE_NAME=slasc_bridge
VRU_CHAINCODE_NAME=vru_positions

if [ -d "../explorer-local" ]
then
    echo "Cleaning up explorer containers"
    pushd ../explorer-local
    docker-compose down -v
    popd
fi

# launch network; create channel and join peer to channel
pushd ../test-network
./network.sh down
./network.sh up createChannel -c ${SLA_CHANNEL_NAME} -ca -s couchdb
./network.sh up createChannel -c ${VRU_CHANNEL_NAME} -ca -s couchdb

./network.sh deployCC -c ${SLA_CHANNEL_NAME} -ccn ${SLA_CHAINCODE_NAME} -ccl ${CC_SRC_LANGUAGE} -ccp ${SLA_CC_SRC_PATH}
./network.sh deployCC -c ${VRU_CHANNEL_NAME} -ccn ${VRU_CHAINCODE_NAME} -ccl ${CC_SRC_LANGUAGE} -ccp ${VRU_CC_SRC_PATH}

popd

if [ -d "../explorer-local" ]
then
    pushd ../explorer-local
    bash ./restart-explorer.sh
    popd
else
    echo "Explorer does not exist! Skipping!"
fi

cat <<EOF

Total setup execution time : $(($(date +%s) - starttime)) secs ...

Next, use the application to interact with the deployed contract.

Start by changing into the "application" directory:
  cd application

Then, install dependencies and run the test using:
  bash runclient.sh
EOF
