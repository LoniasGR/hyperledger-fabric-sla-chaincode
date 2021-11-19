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

# launch network; create channel and join peer to channel
pushd ../test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca
./network.sh deployCC -ccn SLA -ccl ${CC_SRC_LANGUAGE} -ccp ${CC_SRC_PATH}
popd

cat <<EOF

Total setup execution time : $(($(date +%s) - starttime)) secs ...

Next, use the FabCar applications to interact with the deployed FabCar contract.
The FabCar applications are available in multiple programming languages.
Follow the instructions for the programming language of your choice:

Go:

  Start by changing into the "go" directory:
    cd go

  Then, install dependencies and run the test using:
    go run client.go

  The test will invoke the sample client app which perform the following:
    - Import user credentials into the wallet (if they don't already exist there)
    - Submit a transaction to create a new car
    - Evaluate a transaction (query) to return details of this car
    - Submit a transaction to change the owner of this car
    - Evaluate a transaction (query) to return the updated details of this car

EOF