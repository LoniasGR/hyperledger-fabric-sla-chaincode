#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

function app_one_line_pem {
  echo "`awk 'NF {sub(/\\n/, ""); printf "%s\\\\\\\n",$0;}' $1`"
}

function app_json_ccp {
  local ORG=$1
  local PP=$(app_one_line_pem "$2")
  local CP=$(app_one_line_pem "$3")

  sed -e "s/\${ORG}/$ORG/" \
    -e "s#\${PEERPEM}#$PP#" \
    -e "s#\${CAPEM}#$CP#" \
    -e "s/\${NS}/$NS/" \
    scripts/ccp-template.json
}

function app_id {
  local MSP=$1
  local CERT=$(app_one_line_pem "$2")
  local PK=$(app_one_line_pem "$3")

  sed -e "s#\${CERTIFICATE}#$CERT#" \
    -e "s#\${PRIVATE_KEY}#$PK#" \
    -e "s#\${MSPID}#$MSP#" \
    scripts/appuser.id.template
}

function construct_global_configmap() {
  push_fn "Creating persistent volumes and mounts for applications"

  # Both KIND and k3s use the Rancher local-path provider.  In KIND, this is installed
  # as the 'standard' storage class, and in Rancher as the 'local-path' storage class.
  if [ "${CLUSTER_RUNTIME}" == "kind" ]; then
    export STORAGE_CLASS="standard"

  elif [ "${CLUSTER_RUNTIME}" == "k3s" ]; then
    export STORAGE_CLASS="local-path"
  else
    echo "Unknown CLUSTER_RUNTIME ${CLUSTER_RUNTIME}"
    exit 1
  fi

  envsubst <kube/pv-fabric-applications.yaml | kubectl -n "$NS" create -f - || true
  envsubst <kube/pvc-fabric-applications.yaml | kubectl -n "$NS" create -f - || true

  pop_fn

  push_fn "Constructing application connection profiles"

  ENROLLMENT_DIR=${TEMP_DIR}/enrollments
  CHANNEL_MSP_DIR=${TEMP_DIR}/channel-msp

  mkdir -p build/application/wallet
  mkdir -p build/application/gateways

  local peer_pem=$CHANNEL_MSP_DIR/peerOrganizations/org1/msp/tlscacerts/tlsca-signcert.pem
  local ca_pem=$CHANNEL_MSP_DIR/peerOrganizations/org1/msp/cacerts/ca-signcert.pem

  echo "$(app_json_ccp 1 $peer_pem $ca_pem)" >build/application/gateways/org1_ccp.json

  peer_pem=$CHANNEL_MSP_DIR/peerOrganizations/org2/msp/tlscacerts/tlsca-signcert.pem
  ca_pem=$CHANNEL_MSP_DIR/peerOrganizations/org2/msp/cacerts/ca-signcert.pem

  echo "$(app_json_ccp 2 $peer_pem $ca_pem)" >build/application/gateways/org2_ccp.json

  peer_pem=$CHANNEL_MSP_DIR/peerOrganizations/org3/msp/tlscacerts/tlsca-signcert.pem
  ca_pem=$CHANNEL_MSP_DIR/peerOrganizations/org3/msp/cacerts/ca-signcert.pem

  echo "$(app_json_ccp 3 $peer_pem $ca_pem)" >build/application/gateways/org3_ccp.json

  peer_pem=$CHANNEL_MSP_DIR/peerOrganizations/org4/msp/tlscacerts/tlsca-signcert.pem
  ca_pem=$CHANNEL_MSP_DIR/peerOrganizations/org4/msp/cacerts/ca-signcert.pem

  echo "$(app_json_ccp 4 $peer_pem $ca_pem)" >build/application/gateways/org4_ccp.json

  pop_fn

  push_fn "Getting Application Identities"

  local cert=$ENROLLMENT_DIR/org1/users/org1admin/msp/signcerts/cert.pem
  local pk=$ENROLLMENT_DIR/org1/users/org1admin/msp/keystore/key.pem

  echo "$(app_id Org1MSP $cert $pk)" >build/application/wallet/appuser_org1.id

  local cert=$ENROLLMENT_DIR/org2/users/org2admin/msp/signcerts/cert.pem
  local pk=$ENROLLMENT_DIR/org2/users/org2admin/msp/keystore/key.pem

  echo "$(app_id Org2MSP $cert $pk)" >build/application/wallet/appuser_org2.id

  local cert=$ENROLLMENT_DIR/org3/users/org3admin/msp/signcerts/cert.pem
  local pk=$ENROLLMENT_DIR/org3/users/org3admin/msp/keystore/key.pem

  echo "$(app_id Org3MSP $cert $pk)" >build/application/wallet/appuser_org3.id

  local cert=$ENROLLMENT_DIR/org4/users/org4admin/msp/signcerts/cert.pem
  local pk=$ENROLLMENT_DIR/org4/users/org4admin/msp/keystore/key.pem

  echo "$(app_id Org4MSP $cert $pk)" >build/application/wallet/appuser_org4.id

  pop_fn

  push_fn "Creating ConfigMap \"app-fabric-org1-tls-v1-map\" with TLS certificates for the application"
  kubectl -n $NS delete configmap app-fabric-org1-tls-v1-map || true
  kubectl -n $NS create configmap app-fabric-org1-tls-v1-map --from-file=$CHANNEL_MSP_DIR/peerOrganizations/org1/msp/tlscacerts
  pop_fn

  push_fn "Creating ConfigMap \"app-fabric-org2-tls-v1-map\" with TLS certificates for the application"
  kubectl -n $NS delete configmap app-fabric-org2-tls-v1-map || true
  kubectl -n $NS create configmap app-fabric-org2-tls-v1-map --from-file=$CHANNEL_MSP_DIR/peerOrganizations/org2/msp/tlscacerts
  pop_fn

  push_fn "Creating ConfigMap \"app-fabric-org3-tls-v1-map\" with TLS certificates for the application"
  kubectl -n $NS delete configmap app-fabric-org3-tls-v1-map || true
  kubectl -n $NS create configmap app-fabric-org3-tls-v1-map --from-file=$CHANNEL_MSP_DIR/peerOrganizations/org3/msp/tlscacerts
  pop_fn

  push_fn "Creating ConfigMap \"app-fabric-org4-tls-v1-map\" with TLS certificates for the application"
  kubectl -n $NS delete configmap app-fabric-org4-tls-v1-map || true
  kubectl -n $NS create configmap app-fabric-org4-tls-v1-map --from-file=$CHANNEL_MSP_DIR/peerOrganizations/org4/msp/tlscacerts
  pop_fn

  push_fn "Creating ConfigMap \"app-fabric-ids-v1-map\" with identities for the application"
  kubectl -n $NS delete configmap app-fabric-ids-v1-map || true
  kubectl -n $NS create configmap app-fabric-ids-v1-map --from-file=${TEMP_DIR}/application/wallet
  pop_fn

  push_fn "Creating ConfigMap \"app-fabric-ccp-v1-map\" with ConnectionProfile for the application"
  kubectl -n $NS delete configmap app-fabric-ccp-v1-map || true
  kubectl -n $NS create configmap app-fabric-ccp-v1-map --from-file=${TEMP_DIR}/application/gateways
  pop_fn
}

function construct_application_configmap() {
  local org=$1

  push_fn "Creating ConfigMap \"app-fabric-org${org}-v1-map\" with Organization ${org} information for the application"

  cat <<EOF >"build/app-fabric-org${org}-v1-map.yaml"
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-fabric-org${org}-v1-map
data:
  fabric_channel: ${CHANNEL_NAME}
  fabric_contract: ${CHAINCODE_NAME}
  fabric_wallet_dir: /fabric/application/wallet
  fabric_gateway_hostport: org${org}-peer-gateway-svc:8051
  fabric_gateway_sslHostOverride: org${org}-peer-gateway-svc
  fabric_user: appuser_org${org}
  fabric_gateway_tlsCertPath: /fabric/tlscacerts/tlsca-signcert.pem
  identity_endpoint: http://identity-management:8000
  data_folder: /fabric/data
  consumer_group: org${org}-consumer-group
EOF

  kubectl -n $NS apply -f "build/app-fabric-org${org}-v1-map.yaml"

  pop_fn
}

function construct_api_configmap() {
  push_fn "Creating ConfigMap \"app-fabric-api-v1-map\""

  cat <<EOF >"build/app-fabric-api-v1-map.yaml"
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-fabric-api-v1-map
data:
  fabric_sla_channel: ${SLA_CHANNEL_NAME}
  fabric_sla_contract: ${SLA_CHAINCODE_NAME}
  fabric_vru_channel: ${VRU_CHANNEL_NAME}
  fabric_vru_contract: ${VRU_CHAINCODE_NAME}
  fabric_parts_channel: ${PARTS_CHANNEL_NAME}
  fabric_parts_contract: ${PARTS_CHAINCODE_NAME}
  fabric_sla2_channel: ${SLA2_CHANNEL_NAME}
  fabric_wallet_dir: /fabric/application/wallet
  fabric_org1_gateway_hostport: org1-peer-gateway-svc:8051
  fabric_org2_gateway_hostport: org2-peer-gateway-svc:8051
  fabric_org3_gateway_hostport: org3-peer-gateway-svc:8051
  fabric_org4_gateway_hostport: org4-peer-gateway-svc:8051
  fabric_org1_gateway_sslHostOverride: org1-peer-gateway-svc
  fabric_org2_gateway_sslHostOverride: org2-peer-gateway-svc
  fabric_org3_gateway_sslHostOverride: org3-peer-gateway-svc
  fabric_org4_gateway_sslHostOverride: org4-peer-gateway-svc

  identity_endpoint: http://identity-management:8000
EOF

  kubectl -n "$NS" apply -f "build/app-fabric-api-v1-map.yaml"
  pop_fn
}

function deploy_api() {
  construct_api_configmap
  docker build -t "${CONTAINER_REGISTRY_ADDRESS}/api" application/api
  docker push "${CONTAINER_REGISTRY_ADDRESS}/api"
  envsubst <kube/api-deployment.yaml | kubectl -n "$NS" apply -f -
}

function explorer_config() {
  push_fn "Create Config Maps and secrets"
  export orgNr=4
  export EXPLORER_CHANNEL_NAME=sla2.0
  export EXPLORER_ORG_MSP=Org${orgNr}MSP
  export EXPLORER_ORG_PEER_GATEWAY=org${orgNr}-peer-gateway-svc
  export EXPLORER_CA_CERT_PATH=/fabric/tlscacerts/tlsca-signcert.pem
  export EXPLORER_ORG_GATEWAY_PORT=org${orgNr}-peer-gateway-svc:8051
  export EXPLORER_ADMIN_CERT=/fabric/keys/cert.pem
  export EXPLORER_ADMIN_PK=/fabric/keys/key.pem

  local PK="${TEMP_DIR}/enrollments/org${orgNr}/users/org${orgNr}admin/msp/keystore/key.pem"
  local CERT="${TEMP_DIR}/enrollments/org${orgNr}/users/org${orgNr}admin/msp/signcerts/cert.pem"

  kubectl -n $NS delete configmap app-fabric-explorer-pk-v1 || true
  kubectl -n $NS create configmap app-fabric-explorer-pk-v1 --from-file=$PK

  kubectl -n $NS delete configmap app-fabric-explorer-cert-v1 || true
  kubectl -n $NS create configmap app-fabric-explorer-cert-v1 --from-file=$CERT

  mkdir -p "${TEMP_DIR}/explorer-local"
  envsubst <./explorer-local/connection-profile/test-network.json >${TEMP_DIR}/explorer-local/network.json

  kubectl -n $NS delete configmap app-fabric-explorer-network-v1 || true
  kubectl -n $NS create configmap app-fabric-explorer-network-v1 --from-file=${TEMP_DIR}/explorer-local/network.json

  envsubst <./explorer-local/config.json >${TEMP_DIR}/explorer-local/config.json
  kubectl -n $NS delete configmap app-fabric-explorer-config-v1 || true
  kubectl -n $NS create configmap app-fabric-explorer-config-v1 --from-file=${TEMP_DIR}/explorer-local/config.json

  cat <<EOF >build/app-fabric-explorer-v1-map.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-fabric-explorer-v1-map
data:
  DATABASE_HOST: explorerdb
  DATABASE_DATABASE: fabricexplorer
  DATABASE_PORT: "8432"
  DATABASE_USERNAME: hppoc
  DATABASE_PASSWORD: password
  LOG_LEVEL_APP: debug
  LOG_LEVEL_DB: info
  LOG_LEVEL_CONSOLE: debug
  LOG_CONSOLE_STDOUT: "true"
  DISCOVERY_AS_LOCALHOST: "false"
  PORT: "8000"
EOF

  kubectl -n $NS apply -f "build/app-fabric-explorer-v1-map.yaml"

  pop_fn

  push_fn "Run deployment"

  envsubst <kube/explorer-db-deployment.yaml | kubectl -n $NS delete -f - || true
  envsubst <kube/explorer-db-deployment.yaml | kubectl -n $NS apply -f -

  sleep 10s

  envsubst <kube/explorer-deployment.yaml | kubectl -n $NS delete -f - || true
  envsubst <kube/explorer-deployment.yaml | kubectl -n $NS apply -f -

  pop_fn
}

function deploy_sla_client() {
  push_fn "Setting up files"
  export CHANNEL_NAME=${SLA_CHANNEL_NAME}
  export CHAINCODE_NAME=${SLA_CHAINCODE_NAME}

  cp config/kafka/consumer.properties application/sla_client/
  cp config/kafka/kafka.client.keystore.jks application/sla_client/
  cp config/kafka/kafka.client.truststore.jks application/sla_client/
  cp config/kafka/server.cer.pem application/sla_client/

  pop_fn

  construct_application_configmap 1
  push_fn "Creating and deploying container"

  docker build -t "${CONTAINER_REGISTRY_ADDRESS}/sla-client" application/sla_client
  docker push "${CONTAINER_REGISTRY_ADDRESS}/sla-client"

  envsubst <kube/sla-client-deployment.yaml | kubectl -n "${NS}" delete -f - || true
  envsubst <kube/sla-client-deployment.yaml | kubectl -n "${NS}" apply -f -
  pop_fn

  rm application/sla_client/consumer.properties
  rm application/sla_client/kafka.client.keystore.jks
  rm application/sla_client/kafka.client.truststore.jks
  rm application/sla_client/server.cer.pem
}

function deploy_vru_client() {
  push_fn "Setting up files"
  export CHANNEL_NAME=${VRU_CHANNEL_NAME}
  export CHAINCODE_NAME=${VRU_CHAINCODE_NAME}

  cp config/kafka/consumer.properties application/vru_client/
  cp config/kafka/kafka.client.keystore.jks application/vru_client/
  cp config/kafka/kafka.client.truststore.jks application/vru_client/
  cp config/kafka/server.cer.pem application/vru_client/

  pop_fn
  construct_application_configmap 2

  push_fn "Creating and deploying container"

  docker build -t "${CONTAINER_REGISTRY_ADDRESS}/vru-client" application/vru_client
  docker push "${CONTAINER_REGISTRY_ADDRESS}/vru-client"

  envsubst <kube/vru-client-deployment.yaml | kubectl -n "${NS}" delete -f - || true
  envsubst <kube/vru-client-deployment.yaml | kubectl -n "${NS}" apply -f -
  pop_fn

  rm application/vru_client/consumer.properties
  rm application/vru_client/kafka.client.keystore.jks
  rm application/vru_client/kafka.client.truststore.jks
  rm application/vru_client/server.cer.pem
}

function deploy_parts_client() {
  push_fn "Setting up files"
  export CHANNEL_NAME=${PARTS_CHANNEL_NAME}
  export CHAINCODE_NAME=${PARTS_CHAINCODE_NAME}

  cp config/kafka/consumer.properties application/parts_client/
  cp config/kafka/kafka.client.keystore.jks application/parts_client/
  cp config/kafka/kafka.client.truststore.jks application/parts_client/
  cp config/kafka/server.cer.pem application/parts_client/

  pop_fn

  construct_application_configmap 3

  push_fn "Creating and deploying container"

  docker build -t "${CONTAINER_REGISTRY_ADDRESS}/parts-client" application/parts_client
  docker push "${CONTAINER_REGISTRY_ADDRESS}/parts-client"

  envsubst <kube/parts-client-deployment.yaml | kubectl -n "${NS}" delete -f - || true
  envsubst <kube/parts-client-deployment.yaml | kubectl -n "${NS}" apply -f -
  pop_fn

  rm application/parts_client/consumer.properties
  rm application/parts_client/kafka.client.keystore.jks
  rm application/parts_client/kafka.client.truststore.jks
  rm application/parts_client/server.cer.pem

}

function deploy_sla_2_client() {
  push_fn "Setting up files"
  export CHANNEL_NAME=${SLA2_CHANNEL_NAME}
  export CHAINCODE_NAME=sla-version-2

  cp config/kafka/consumer.properties application/sla_2.0_client/
  cp config/kafka/kafka.client.keystore.jks application/sla_2.0_client/
  cp config/kafka/kafka.client.truststore.jks application/sla_2.0_client/
  cp config/kafka/server.cer.pem application/sla_2.0_client/

  cp -r config/org4/ application/sla_2.0_client/config

  cp -r ${TEMP_DIR}/enrollments/org4/users/org4admin/msp application/sla_2.0_client/msp

  cp ${TEMP_DIR}/channel-msp/ordererOrganizations/org0/orderers/org0-orderer1/tls/signcerts/tls-cert.pem application/sla_2.0_client/orderer-cert.pem

  pop_fn

  construct_application_configmap 4
  push_fn "Creating and deploying container"

  docker build -t "${CONTAINER_REGISTRY_ADDRESS}/sla2-client" application/sla_2.0_client
  docker push "${CONTAINER_REGISTRY_ADDRESS}/sla2-client"

  envsubst <kube/deploy-permissions-rbac.yaml | kubectl -n "${NS}" delete -f - || true
  envsubst <kube/deploy-permissions-rbac.yaml | kubectl -n "${NS}" apply -f -

  envsubst <kube/sla2-client-deployment.yaml | kubectl -n "${NS}" delete -f - || true
  envsubst <kube/sla2-client-deployment.yaml | kubectl -n "${NS}" apply -f -


  pop_fn

  # Cleanup

  rm -r application/sla_2.0_client/config
  rm -r application/sla_2.0_client/msp

  rm application/sla_2.0_client/consumer.properties
  rm application/sla_2.0_client/kafka.client.keystore.jks
  rm application/sla_2.0_client/kafka.client.truststore.jks
  rm application/sla_2.0_client/server.cer.pem
  rm application/sla_2.0_client/orderer-cert.pem
}


function identity_management() {
    push_fn "Building identity management pod"
    docker build -t ${CONTAINER_REGISTRY_ADDRESS}/identity-management application/identity_management
    docker push ${CONTAINER_REGISTRY_ADDRESS}/identity-management
    # Maybe todo: change the namespace here
    envsubst <kube/identity-management-client.yaml | kubectl -n "${NS}" delete -f - || true
    envsubst <kube/identity-management-client.yaml | kubectl -n "${NS}" apply -f -
    pop_fn "Identity management pod built"

}

function application_command_group() {
  # set -x

  COMMAND=$1
  shift

  if [ "${COMMAND}" == "init" ]; then
    log "Initializing global config"
    construct_global_configmap
    log "🏁 - Config is initialized."

  elif [ "${COMMAND}" == "create" ]; then
    if [ $# -ne 1 ]; then
      log "Usage: create ORG_NR"
      exit 1
    fi
    log "Creating config for organisation \"$1\":"
    construct_application_configmap "$1"
    log "🏁 - Config is ready."
  elif [ "${COMMAND}" == "sla" ]; then
    log "Deploying sla client:"
    deploy_sla_client
    log "🏁 - Client deployed"
  elif [ "${COMMAND}" == "vru" ]; then
    log "Deploying vru client:"
    deploy_vru_client
    log "🏁 - Client deployed"
  elif [ "${COMMAND}" == "parts" ]; then
    log "Deploying parts client:"
    deploy_parts_client
    log "🏁 - Client deployed"
  elif [ "${COMMAND}" == "sla2" ]; then
    log "Deploying sla 2 client:"
    deploy_sla_2_client
    log "🏁 - Client deployed"
  elif [ "${COMMAND}" == "api" ]; then
    log "Deploying api: "
    deploy_api
    log "🏁 - API is ready."
    elif [ "${COMMAND}" == "identity_management" ]; then
    log "Deploying identity management API: "
    identity_management
    log "🏁 - Identity Management is ready."
  elif [ "${COMMAND}" == "explorer" ]; then
    log "Deploying explorer"
    explorer_config
    log "🏁 explorer deployed"
  else
    log "Uknown command"
    exit 1
  fi
}
