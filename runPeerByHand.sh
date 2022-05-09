set -e


##########################################
## ADAPT THOSE
##########################################
CHANNEL_NAME="sla"
CC_NAME="slasc_bridge"

##########################################
## COPY THOSE
##########################################

CC_SRC_LANGUAGE="golang"
CC_LABEL=`echo ${CC_NAME}_${CC_VERSION}`

pushd ../test-network

export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

popd

# Now just run
# peer chaincode query -C ${CHANNEL_NAME} -n ${CC_NAME} -c '{"Args":["GetAllAssets"]}'
