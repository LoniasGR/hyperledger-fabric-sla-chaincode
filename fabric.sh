#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0
#
# Run as
# ./fabric {MODE}
#
#####################################

trap "restoreConfigtx" EXIT

. scripts/utils.sh

# Get docker sock path from environment variable
SOCK="${DOCKER_HOST:-/var/run/docker.sock}"
DOCKER_SOCK="${SOCK##unix://}"

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1
starttime=$(date +%s)
CC_SRC_LANGUAGE="go"
SLA_CC_SRC_PATH="${PWD}/chaincode_sla"
VRU_CC_SRC_PATH="${PWD}/chaincode_vru"
PARTS_CC_SRC_PATH="${PWD}/chaincode_parts"

SLA_CHANNEL_NAME=sla
VRU_CHANNEL_NAME=vru
PARTS_CHANNEL_NAME=parts

SLA_CHAINCODE_NAME=slasc_bridge
VRU_CHAINCODE_NAME=vru_positions
PARTS_CHAINCODE_NAME=parts

function startLogSpout() {
  ./monitordocker.sh
}

function cleanExplorer() {
  if [ -d "../explorer-local" ]; then
    println "Cleaning up explorer containers"
    pushd ../explorer-local || fatalln "Could not find directory ../explorer-local"
    docker-compose down -v
    popd || fataln "Could not exit directory"
  fi
}

function startExplorer() {
  if [ -d "../explorer-local" ]; then
    pushd ../explorer-local || fatalln "Could not find directory ../explorer-local"
    bash ./restart-explorer.sh
    popd || fataln "Could not exit directory"
  else
    println "Explorer does not exist! Skipping!"
  fi
}

function networkDown() {
  cleanExplorer

  # launch network; create channel and join peer to channel
  pushd ../test-network || fatalln "Could not find directory ../test-network"
  docker stop logspout || infoln "Logspout not running"
  ./network.sh down
  popd || fataln "Could not exit directory"
}

function deployChaincodes() {
  pushd ../test-network || fatalln "Could not find directory ../test-network"
  ./network.sh deployCC -c ${SLA_CHANNEL_NAME} -ccn ${SLA_CHAINCODE_NAME} -ccl ${CC_SRC_LANGUAGE} -ccp "${SLA_CC_SRC_PATH}"
  ./network.sh deployCC -c ${VRU_CHANNEL_NAME} -ccn ${VRU_CHAINCODE_NAME} -ccl ${CC_SRC_LANGUAGE} -ccp "${VRU_CC_SRC_PATH}"
  ./network.sh deployCC -c ${PARTS_CHANNEL_NAME} -ccn ${PARTS_CHAINCODE_NAME} -ccl ${CC_SRC_LANGUAGE} -ccp "${PARTS_CC_SRC_PATH}"
  popd || fataln "Could not exit directory"
}

function development() {

  # Shut down the network and then restart it
  networkDown

  pushd ../test-network || fatalln "Could not find directory ../test-network"
  ./network.sh up createChannel -c ${SLA_CHANNEL_NAME} -ca -s couchdb
  ./network.sh up createChannel -c ${VRU_CHANNEL_NAME} -ca -s couchdb
  ./network.sh up createChannel -c ${PARTS_CHANNEL_NAME} -ca -s couchdb
  popd || fataln "Could not exit directory"

  deployChaincodes

  startExplorer

  println "Total setup execution time : $(($(date +%s) - starttime)) secs ..."
  println
  println "Next, use the application to interact with the deployed contract."
  println
  println "Start by changing into the \"application\" directory:"
  println "cd application"
  println
  println "Then, install dependencies and run the test using:"
  println "bash runclient.sh"
  exit 0
}

function createChannel() {
  pushd ../test-network || fatalln "Could not find directory ../test-network"
  export PATH=${PWD}/../bin:$PATH
  export FABRIC_CFG_PATH=${PWD}/configtx
  export ORDERER_CA=${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  export ORDERER_ADMIN_TLS_SIGN_CERT=${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt
  export ORDERER_ADMIN_TLS_PRIVATE_KEY=${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key

  configtxgen -profile "$2" -outputBlock ./channel-artifacts/"$1".block -channelID "$1"

  osnadmin channel join --channelID "$1" --config-block ./channel-artifacts/"$1".block -o localhost:7053 --ca-file "$ORDERER_CA" --client-cert "$ORDERER_ADMIN_TLS_SIGN_CERT" --client-key "$ORDERER_ADMIN_TLS_PRIVATE_KEY"

  # List channels
  osnadmin channel list -o localhost:7053 --ca-file "$ORDERER_CA" --client-cert "$ORDERER_ADMIN_TLS_SIGN_CERT" --client-key "$ORDERER_ADMIN_TLS_PRIVATE_KEY"

  popd || fataln "Could not exit directory"
}

function addOrgToChannel() {
  pushd ../test-network || fatalln "Could not find directory ../test-network"
  # Join Org1 peer on SLA
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="Org${1}MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org${1}.example.com/peers/peer0.org${1}.example.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org${1}.example.com/users/Admin@org${1}.example.com/msp
  export CORE_PEER_ADDRESS=localhost:${2}

  export FABRIC_CFG_PATH=$PWD/../config/

  peer channel join -b ./channel-artifacts/"${3}".block
  popd || fataln "Could not exit directory"
}

function addOrg3() {
  pushd ../test-network || fatalln "Could not find directory ../test-network"
  # Create the third org
  cd addOrg3 || fatalln "Could not enter addOrg3 directory"

  export PATH=${PWD}/../../bin:${PWD}:$PATH
  export FABRIC_CFG_PATH=${PWD}

  if [[ $(fabric-ca-client version) -ne 0 ]]; then
    echo "ERROR! fabric-ca-client binary not found.."
    echo
    echo "Follow the instructions in the Fabric docs to install the Fabric Binaries:"
    echo "https://hyperledger-fabric.readthedocs.io/en/latest/install.html"
    exit 1
  fi

  infoln "Generating certificates using Fabric CA"
  docker-compose -f compose/compose-ca-org3.yaml -f compose/docker/docker-compose-ca-org3.yaml up -d 2>&1

  # shellcheck source=../test-network/addOrg3/fabric-ca/registerEnroll.sh
  . fabric-ca/registerEnroll.sh

  sleep 10

  infoln "Creating Org3 Identities"
  createOrg3

  infoln "Generating CCP files for Org3"
  ./ccp-generate.sh

  ## Generate Org3 definitions
  ../../bin/configtxgen -printOrg Org3MSP >../organizations/peerOrganizations/org3.example.com/org3.json

  DOCKER_SOCK="${DOCKER_SOCK}" docker-compose -f compose/compose-org3.yaml -f compose/compose-couch-org3.yaml -f compose/docker/docker-compose-couch-org3.yaml \
    -f compose/docker/docker-compose-org3.yaml up -d
  popd || fataln "Could not exit directory"
}

function deployChaincode() {
  pushd ../test-network || fatalln "Could not find directory ../test-network"
  CHANNEL_NAME=${1}
  CC_NAME=${2}
  CC_SRC_PATH=${3}
  PEER_NUMBER=${4}
  CC_VERSION=${5:-"1.0"}
  CC_SEQUENCE=${6:-"1"}
  CC_END_POLICY=""
  CC_COLL_CONFIG=""
  CC_INIT_FCN="InitLedger"
  DELAY="3"
  MAX_RETRY="5"
  VERBOSE="false"
  CC_RUNTIME_LANGUAGE=golang

  println "executing with the following"
  println "- CHANNEL_NAME: ${C_GREEN}${CHANNEL_NAME}${C_RESET}"
  println "- CC_NAME: ${C_GREEN}${CC_NAME}${C_RESET}"
  println "- CC_SRC_PATH: ${C_GREEN}${CC_SRC_PATH}${C_RESET}"
  println "- CC_SRC_LANGUAGE: ${C_GREEN}${CC_SRC_LANGUAGE}${C_RESET}"
  println "- CC_VERSION: ${C_GREEN}${CC_VERSION}${C_RESET}"
  println "- CC_SEQUENCE: ${C_GREEN}${CC_SEQUENCE}${C_RESET}"
  println "- CC_END_POLICY: ${C_GREEN}${CC_END_POLICY}${C_RESET}"
  println "- CC_COLL_CONFIG: ${C_GREEN}${CC_COLL_CONFIG}${C_RESET}"
  println "- CC_INIT_FCN: ${C_GREEN}${CC_INIT_FCN}${C_RESET}"
  println "- DELAY: ${C_GREEN}${DELAY}${C_RESET}"
  println "- MAX_RETRY: ${C_GREEN}${MAX_RETRY}${C_RESET}"
  println "- VERBOSE: ${C_GREEN}${VERBOSE}${C_RESET}"

  FABRIC_CFG_PATH=$PWD/../config/

  infoln "Vendoring Go dependencies at $CC_SRC_PATH"
  pushd "$CC_SRC_PATH" ||  fatalln "${CC_SRC_PATH} not found"
  GO111MODULE=on go mod vendor
  popd || fatalln "Could not leave directory"
  successln "Finished vendoring Go dependencies"

  # import utils

  # shellcheck source=../test-network/scripts/envVar.sh
  . scripts/envVar.sh

    # shellcheck source=../test-network/scripts/ccutils.sh
  . scripts/ccutils.sh

  packageChaincode

  ## Install chaincode on peer
  infoln "Installing chaincode on peer"
  installChaincode "${PEER_NUMBER}"

  ## query whether the chaincode is installed
  queryInstalled "${PEER_NUMBER}"

  ## approve the definition for org
  approveForMyOrg "${PEER_NUMBER}"

  ## Check commit readiness
  peer lifecycle chaincode checkcommitreadiness --channelID "${CHANNEL_NAME}" --name "${CC_NAME}" --version "${CC_VERSION}" --sequence "${CC_SEQUENCE}" --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" --output json

  commitChaincodeDefinition "${PEER_NUMBER}"
  peer lifecycle chaincode querycommitted --channelID "${CHANNEL_NAME}" --name "${CC_NAME}" --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"
}

packageChaincode() {
  set -x
  peer lifecycle chaincode package "${CC_NAME}".tar.gz --path "${CC_SRC_PATH}" --lang "${CC_RUNTIME_LANGUAGE}" --label "${CC_NAME}_${CC_VERSION}" >&log.txt
  res=$?
  PACKAGE_ID=$(peer lifecycle chaincode calculatepackageid "${CC_NAME}".tar.gz)
  { set +x; } 2>/dev/null
  cat log.txt
  verifyResult $res "Chaincode packaging has failed"
  successln "Chaincode is packaged"
}

function testing() {
  infoln "Shutting down old network"

  networkDown
  infoln "Running test network"

  # Start the two orgs network
  pushd ../test-network || fatalln "Could not find directory ../test-network"
  ./network.sh up -ca -s couchdb
  popd || fatalln "Could not leave directory ../test-network"

  # Start logspout for logging
  infoln "Starting logspout"
  startLogSpout

  # Add the third organisation
  infoln "Adding Organisation 3"
  addOrg3

  # Move our own configtx.yaml in place
  infoln "Replacing configtx"
  pushd ../test-network || fatalln "Could not find directory ../test-network"
  mv ./configtx/configtx.yaml ./configtx/configtx-backup.yaml
  popd || fataln "Could not exit directory"
  cp ./configtx/configtx.yaml ../test-network/configtx

  infoln "Creating channels"
  createChannel ${SLA_CHANNEL_NAME} Org1ApplicationGenesis
  createChannel ${VRU_CHANNEL_NAME} Org2ApplicationGenesis
  createChannel ${PARTS_CHANNEL_NAME} Org3ApplicationGenesis

  sleep 5

  infoln "Adding orgs to channels"
  addOrgToChannel 1 7051 ${SLA_CHANNEL_NAME}
  addOrgToChannel 2 9051 ${VRU_CHANNEL_NAME}
  addOrgToChannel 3 11051 ${PARTS_CHANNEL_NAME}

  infoln "Deploing chaincode to each channel"
    deployChaincode ${SLA_CHANNEL_NAME} ${SLA_CHAINCODE_NAME} "${SLA_CC_SRC_PATH}" 1
    infoln "SLA Chaincode deployed"

  deployChaincode ${VRU_CHANNEL_NAME} ${VRU_CHAINCODE_NAME} "${VRU_CC_SRC_PATH}" 2
  infoln "VRU Chaincode deployed"

  deployChaincode ${PARTS_CHANNEL_NAME} ${PARTS_CHAINCODE_NAME} "${PARTS_CC_SRC_PATH}" 3
  infoln "Parts Chaincode deployed"

  restoreConfigtx
}

function restoreConfigtx() {
  infoln "Returing original configtx"
  pushd ../test-network || fatalln "Could not find directory ../test-network"
  mv ./configtx/configtx-backup.yaml ./configtx/configtx.yaml
  popd || fataln "Could not exit directory"
  exit 0
}

function printHelp() {
  println "USAGE:"
  println
  println "bash fabric development  Starts the development version, for testing the chaincodes"
  println "bash fabric testing      Starts the expiramental 3 Orgs with separate channels configuration"
  println "bash fabric down         Shuts down the whole network"
}

## Parse mode
if [[ $# -lt 1 ]]; then
  printHelp
  exit 0
else
  MODE=$1
  shift
fi

if [ "$MODE" == "development" ]; then
  infoln "Running development build"
  development
elif [ "$MODE" == "testing" ]; then
  infoln "Running testing build"
  testing
elif [ "$MODE" == "down" ]; then
  infoln "Shutting down the network"
  networkDown
else
  printHelp
  exit 1
fi
