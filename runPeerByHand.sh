#!/bin/bash

##########################################
## ADAPT THOSE
##########################################
export CHANNEL_NAME="sla"
export CC_NAME="slasc_bridge"
CC_VERSION=1

##########################################
## COPY THOSE
##########################################

# export CC_SRC_LANGUAGE="golang"
# export CC_LABEL="${CC_NAME}_${CC_VERSION}"

pushd ../test-network

export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

popd

# Now just run
# peer chaincode query -C ${CHANNEL_NAME} -n ${CC_NAME} -c '{"Args":["GetAllAssets"]}'
# peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "../test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C ${CHANNEL_NAME} -n ${CC_NAME} --peerAddresses localhost:7051 --tlsRootCertFiles "../test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"  -c '{"function":"InitLedger","Args":[]}'
