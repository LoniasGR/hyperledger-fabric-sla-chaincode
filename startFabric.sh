#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error
set -e

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1
starttime=$(date +%s)
CC_SRC_LANGUAGE="go"
CC_SRC_PATH="${PWD}/chaincode"


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
./network.sh up createChannel -c sla -ca -s couchdb
./network.sh deployCC -c sla -ccn sla_contract -ccl ${CC_SRC_LANGUAGE} -ccp ${CC_SRC_PATH}
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