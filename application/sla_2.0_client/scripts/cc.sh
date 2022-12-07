#!/bin/sh

package_cc() {
  _cc_name=$1
  _cc_label=$2
  _cc_archive=$3

  _cc_folder=$(dirname "$_cc_archive")
  # archive_name=$(basename "$cc_archive")

  mkdir -p "${_cc_folder}"

  # Allow the user to override the service URL for the endpoint.  This allows, for instance,
  # local debugging at the 'host.docker.internal' DNS alias.
  _cc_address="{{.peername}}-ccaas-${_cc_name}:8999"

  cat <<EOF >"${_cc_folder}"/connection.json
{
  "address": "${_cc_address}",
  "dial_timeout": "10s",
  "tls_required": false
}
EOF

  cat <<EOF >"${_cc_folder}"/metadata.json
{
  "type": "ccaas",
  "label": "${_cc_label}"
}
EOF

  tar -C "${_cc_folder}" -zcf "${_cc_folder}"/code.tar.gz connection.json
  tar -C "${_cc_folder}" -zcf "${_cc_archive}" code.tar.gz metadata.json

  rm "${_cc_folder}"/code.tar.gz
}

set_chaincode_id() {
  _cc_package=$1

  cc_sha256=$(shasum -a 256 "${_cc_package}" | tr -s ' ' | cut -d ' ' -f 1)
  cc_label=$(tar zxfO "${_cc_package}" metadata.json | jq -r '.label')

  echo "${cc_label}:${cc_sha256}"
}

export_context() {
  orgNr=$1
  peerNr=$2
  tlsCertPath=$3

  export FABRIC_CFG_PATH="${PWD}/config"
  export CORE_PEER_ADDRESS="org${orgNr}-peer${peerNr}:8051"
  export CORE_PEER_MSPCONFIGPATH="${PWD}/msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="${tlsCertPath}"
}

install_cc() {
  orgNr=$1
  peerNr=$2
  tlsCertPath=$3
  cc_package=$4

  export_context "${orgNr}" "${peerNr}" "${tlsCertPath}"
  peer lifecycle chaincode install "${cc_package}"
}

approve_cc() {
  orgNr=$1
  peerNr=1
  cc_name=$2
  cc_id=$3
  tlsCertPath=$4

  export_context "${orgNr}" "${peerNr}" "${tlsCertPath}"

  peer lifecycle \
    chaincode approveformyorg \
    --channelID "${CHANNEL_NAME}" \
    --name "${cc_name}" \
    --package-id "${cc_id}" \
    --version 1 \
    --sequence 1 \
    --orderer "org0-orderer1:8050" \
    --connTimeout "${ORDERER_TIMEOUT}" \
    --tls --cafile "${PWD}/orderer-cert.pem"
}

# commit the named chaincode for an org
commit_cc() {
  orgNr=$1
  peerNr=1
  cc_name=$2
  tlsCertPath=$3

  export_context "${orgNr}" "${peerNr}" "${tlsCertPath}"

  peer lifecycle chaincode commit \
    --channelID "${CHANNEL_NAME}" \
    --name "${cc_name}" \
    --version 1 \
    --sequence 1 \
    --orderer "org0-orderer1:8050" \
    --connTimeout "${ORDERER_TIMEOUT}" \
    --tls --cafile "${PWD}/orderer-cert.pem"

}

query_cc_installed() {
  orgNr=$1
  tlsCertPath=$2
  export_context "${orgNr}" 1 "${tlsCertPath}"

  peer lifecycle chaincode queryinstalled \
    --peerAddresses ${CORE_PEER_ADDRESS} \
    --tlsRootCertFiles ${CORE_PEER_TLS_ROOTCERT_FILE}

  export_context "${orgNr}" 2 "${tlsCertPath}"

  peer lifecycle chaincode queryinstalled \
    --peerAddresses ${CORE_PEER_ADDRESS} \
    --tlsRootCertFiles ${CORE_PEER_TLS_ROOTCERT_FILE}

}

query_cc_installed_one_peer() {
  orgNr=$1
  tlsCertPath=$2
  peerNr=$3
  export_context "${orgNr}" "${peerNr}" "${tlsCertPath}"

  peer lifecycle chaincode queryinstalled \
    --peerAddresses ${CORE_PEER_ADDRESS} \
    --tlsRootCertFiles ${CORE_PEER_TLS_ROOTCERT_FILE}
}

query_cc_metadata() {
  cc_name=$1
  orgNr=$2
  tlsCertPath=$3

  args='{"Args":["org.hyperledger.fabric:GetMetadata"]}'

  echo ''
  echo "Org${orgNr}-Peer1:"
  export_context "${orgNr}" 1 "${tlsCertPath}"
  peer chaincode query -n $cc_name -C $CHANNEL_NAME -c $args

  echo ''
  echo "Org${orgNr}-Peer2:"
  export_context "${orgNr}" 2 "${tlsCertPath}"
  peer chaincode query -n $cc_name -C $CHANNEL_NAME -c $args
}

check_commit_readiness() {

  orgNr=$1
  peerNr=1
  cc_name=$2
  tlsCertPath=$3

  export_context "${orgNr}" "${peerNr}" "${tlsCertPath}"

  peer lifecycle chaincode checkcommitreadiness \
    --channelID "${CHANNEL_NAME}" \
    --name "${cc_name}" \
    --version 1 \
    --sequence 1 \
    --orderer "org0-orderer1:8050" \
    --connTimeout "${ORDERER_TIMEOUT}" \
    --tls --cafile "${PWD}/orderer-cert.pem" \
    --output json
}
## Parse mode
if [ $# -lt 1 ]; then
  exit 1
else
  MODE=$1
  shift
fi

if [ "${MODE}" = "package" ]; then
  package_cc "$@"
elif [ "${MODE}" = "id" ]; then
  set_chaincode_id "$@"
elif [ "${MODE}" = "install" ]; then
  install_cc "$@"
elif [ "${MODE}" = "approve" ]; then
  approve_cc "$@"
elif [ "${MODE}" = "commit" ]; then
  commit_cc "$@"
elif [ "${MODE}" = "checkcommitreadiness" ]; then
  check_commit_readiness "$@"
elif [ "${MODE}" = "query" ]; then
  ## Parse mode
  if [ $# -lt 1 ]; then
    exit 1
  else
    SUBMODE=$1
    shift
  fi

  if [ "${SUBMODE}" = "installed" ]; then
    query_cc_installed "$@"
  elif [ "${SUBMODE}" = "installed_one" ]; then
    query_cc_installed_one_peer "$@"
  elif [ "${SUBMODE}" = "metadata" ]; then
    query_cc_metadata "$@"
  fi
else
  exit 1
fi
