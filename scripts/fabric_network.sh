#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

function launch_orderers() {
  push_fn "Launching orderers"
  local org0_ca_pod
  # Had some issues when developing with pushing an image, but getting a previous version
  # so creating random tags to avoid that
  if [ $RANDOM_TAG == 1 ]; then
    local TAG=$(random_chars)
  else
    local TAG="latest"
  fi
  if [ "$NO_VOLUMES" -eq 1 ]; then
    # Find the org-ca pod.
    org0_ca_pod=$(pod_from_name org0-ca "$NS")

    # Get the hyperledger (MSP) data from the pod
    kubectl -n "${NS}" cp "$org0_ca_pod":/var/hyperledger ./tmp
    rm -rf ./tmp/fabric/config
    rm -rf ./tmp/organizations

    # Build the order container with that information
    mkdir -p "${TEMP_DIR}/docker"
    envsubst <docker/Dockerfile.orderer >"${TEMP_DIR}/docker/Dockerfile.orderer"
    docker build -f "${TEMP_DIR}/docker/Dockerfile.orderer" -t "${CONTAINER_REGISTRY_ADDRESS}/orderer:${TAG}" ./tmp
    docker push "${CONTAINER_REGISTRY_ADDRESS}/orderer:${TAG}"
    export FABRIC_ORDERER_IMAGE="${CONTAINER_REGISTRY_ADDRESS}/orderer:${TAG}"

    apply_template kube/org0/org0-orderer1-no-volume.yaml "$ORG0_NS"
    apply_template kube/org0/org0-orderer2-no-volume.yaml "$ORG0_NS"
    apply_template kube/org0/org0-orderer3-no-volume.yaml "$ORG0_NS"

    rm -rf ./tmp
  else
    apply_template kube/org0/org0-orderer1.yaml "$ORG0_NS"
    apply_template kube/org0/org0-orderer2.yaml "$ORG0_NS"
    apply_template kube/org0/org0-orderer3.yaml "$ORG0_NS"
  fi

  sleep 2

  kubectl -n "$ORG0_NS" rollout status deploy/org0-orderer1
  kubectl -n "$ORG0_NS" rollout status deploy/org0-orderer2
  kubectl -n "$ORG0_NS" rollout status deploy/org0-orderer3

  pop_fn
}

function launch_peers() {
  push_fn "Launching peers"
  # Had some issues when developing with pushing an image, but getting a previous version
  # so creating random tags to avoid that
  if [ $RANDOM_TAG == 1 ]; then
    local TAG=$(random_chars)
  else
    local TAG="latest"
  fi

  # TODO: Add org here
  for org in org1 org2 org3 org4; do
    local namespace
    namespace=$(get_namespace $org)

    if [ "$NO_VOLUMES" -eq 1 ]; then
      # Find the org-ca pod.
      local org_ca_pod
      org_ca_pod=$(pod_from_name $org-ca "$NS")

      # Get the hyperledger (MSP) data from the pod
      kubectl -n "${NS}" cp "$org_ca_pod":/var/hyperledger ./tmp
      rm -rf ./tmp/fabric/config
      rm -rf ./tmp/organizations

      # Build the order container with that information
      mkdir -p "${TEMP_DIR}/docker"
      envsubst <docker/Dockerfile.peer >"${TEMP_DIR}/docker/Dockerfile.peer"
      docker build -f "${TEMP_DIR}/docker/Dockerfile.peer" -t "${CONTAINER_REGISTRY_ADDRESS}/peer-${org}:${TAG}" ./tmp
      docker push "${CONTAINER_REGISTRY_ADDRESS}/peer-${org}:${TAG}"
      export FABRIC_PEER_IMAGE="${CONTAINER_REGISTRY_ADDRESS}/peer-${org}:${TAG}"

      apply_template kube/$org/$org-peer1-no-volume.yaml "$namespace"
      apply_template kube/$org/$org-peer2-no-volume.yaml "$namespace"

      rm -rf ./tmp
    else
      apply_template kube/$org/$org-peer1.yaml "$namespace"
      apply_template kube/$org/$org-peer2.yaml "$namespace"
    fi
  done

  # TODO: Add org here
  kubectl -n "$ORG1_NS" rollout status deploy/org1-peer1
  kubectl -n "$ORG1_NS" rollout status deploy/org1-peer2
  kubectl -n "$ORG2_NS" rollout status deploy/org2-peer1
  kubectl -n "$ORG2_NS" rollout status deploy/org2-peer2
  kubectl -n "$ORG3_NS" rollout status deploy/org3-peer1
  kubectl -n "$ORG3_NS" rollout status deploy/org3-peer2
  kubectl -n "$ORG4_NS" rollout status deploy/org4-peer1
  kubectl -n "$ORG4_NS" rollout status deploy/org4-peer2

  pop_fn
}

# Each network node needs a registration, enrollment, and MSP config.yaml
function create_node_local_MSP() {
  local node_type=$1
  local org=$2
  local node=$3
  local csr_hosts=$4
  local ns=$5
  local id_name=${org}-${node}
  local id_secret=${node_type}pw
  local ca_name=${org}-ca

  # Register the node admin
  rc=0
  fabric-ca-client register \
    --id.name "${id_name}" \
    --id.secret "${id_secret}" \
    --id.type "${node_type}" \
    --url https://"${ca_name}.${DOMAIN}:${NGINX_HTTPS_PORT}" \
    --tls.certfiles "$TEMP_DIR/cas/${ca_name}/tlsca-cert.pem" \
    --mspdir "$TEMP_DIR/enrollments/${org}/users/${RCAADMIN_USER}/msp" ||
    rc=$? # trap error code from registration without exiting the network driver script"

  if [ $rc -eq 1 ]; then
    echo "CA admin was (probably) previously registered - continuing"
  fi

  # Enroll the node admin user from within k8s.  This will leave the certificates available on a volume share in the
  # cluster for access by the nodes when launching in a container.
  cat <<EOF | kubectl -n "${ns}" exec deploy/"${ca_name}" -i -- /bin/sh

  set -x
  export FABRIC_CA_CLIENT_HOME=/var/hyperledger/fabric-ca-client
  export FABRIC_CA_CLIENT_TLS_CERTFILES=/var/hyperledger/fabric/config/tls/ca.crt

  fabric-ca-client enroll \
    --url https://${id_name}:${id_secret}@${ca_name}:8443 \
    --csr.hosts ${csr_hosts} \
    --mspdir /var/hyperledger/fabric/organizations/${node_type}Organizations/${org}.example.com/${node_type}s/${id_name}.${org}.example.com/msp

  mv /var/hyperledger/fabric/organizations/${node_type}Organizations/${org}.example.com/${node_type}s/${id_name}.${org}.example.com/msp/cacerts/${org}-ca* \
    /var/hyperledger/fabric/organizations/${node_type}Organizations/${org}.example.com/${node_type}s/${id_name}.${org}.example.com/msp/cacerts/${org}-ca.pem \

  # Create local MSP config.yaml
  echo "NodeOUs:
    Enable: true
    ClientOUIdentifier:
      Certificate: cacerts/${org}-ca.pem
      OrganizationalUnitIdentifier: client
    PeerOUIdentifier:
      Certificate: cacerts/${org}-ca.pem
      OrganizationalUnitIdentifier: peer
    AdminOUIdentifier:
      Certificate: cacerts/${org}-ca.pem
      OrganizationalUnitIdentifier: admin
    OrdererOUIdentifier:
      Certificate: cacerts/${org}-ca.pem
      OrganizationalUnitIdentifier: orderer" > /var/hyperledger/fabric/organizations/${node_type}Organizations/${org}.example.com/${node_type}s/${id_name}.${org}.example.com/msp/config.yaml
EOF
}

function create_orderer_local_MSP() {
  local org=$1
  local orderer=$2
  local csr_hosts=${org}-${orderer}

  create_node_local_MSP orderer "$org" "$orderer" "$csr_hosts" "$ORG0_NS"
}

function create_peer_local_MSP() {
  local org=$1
  local peer=$2
  local ns=$3
  local csr_hosts=localhost,${org}-${peer},${org}-peer-gateway-svc

  create_node_local_MSP peer "$org" "$peer" "$csr_hosts" "${ns}"
}

function create_local_MSP() {
  push_fn "Creating local node MSP"

  create_orderer_local_MSP org0 orderer1
  create_orderer_local_MSP org0 orderer2
  create_orderer_local_MSP org0 orderer3

  create_peer_local_MSP org1 peer1 "$ORG1_NS"
  create_peer_local_MSP org1 peer2 "$ORG1_NS"

  create_peer_local_MSP org2 peer1 "$ORG2_NS"
  create_peer_local_MSP org2 peer2 "$ORG2_NS"

  # TODO: Add org here
  create_peer_local_MSP org3 peer1 "$ORG3_NS"
  create_peer_local_MSP org3 peer2 "$ORG3_NS"

  create_peer_local_MSP org4 peer1 "$ORG4_NS"
  create_peer_local_MSP org4 peer2 "$ORG4_NS"
  pop_fn
}

function cas_up() {

  # Kube config
  if [ "$NO_VOLUMES" -eq 0 ]; then
    init_storage_volumes
  fi
  load_org_config

  # Service account permissions for the k8s builder
  if [ "${CHAINCODE_BUILDER}" == "k8s" ]; then
    apply_k8s_builder_roles
    apply_k8s_builders
  fi

  # Network TLS CAs
  init_tls_cert_issuers

  # Network ECert CAs
  launch_ECert_CAs
  enroll_bootstrap_ECert_CA_users

  # Test Network
  create_local_MSP
}

function orderers_and_peers_up() {
  # launch_orderers
  launch_peers
}

function stop_services() {
  push_fn "Stopping Fabric services"
  # TODO: Add org here
  for ns in $ORG0_NS $ORG1_NS $ORG2_NS $ORG3_NS; do
    kubectl -n "$ns" delete ingress --all
    kubectl -n "$ns" delete deployment --all
    kubectl -n "$ns" delete pod --all
    kubectl -n "$ns" delete service --all
    kubectl -n "$ns" delete configmap --all
    kubectl -n "$ns" delete cert --all
    kubectl -n "$ns" delete issuer --all
    kubectl -n "$ns" delete secret --all
  done

  pop_fn
}

function scrub_org_volumes() {
  push_fn "Scrubbing Fabric volumes"
  if [ "${CLUSTER_RUNTIME}" == "kind" ] || [ "${CLUSTER_RUNTIME}" == "k3s" ]; then
    # TODO: Add org here
    for org in org0 org1 org2 org3; do
      # clean job to make this function can be rerun
      local namespace_variable=${org^^}_NS
      kubectl -n ${!namespace_variable} delete jobs --all

      # scrub all pv contents
      kubectl -n ${!namespace_variable} create -f kube/${org}/${org}-job-scrub-fabric-volumes.yaml
      kubectl -n ${!namespace_variable} wait --for=condition=complete --timeout=60s job/job-scrub-fabric-volumes
      kubectl -n ${!namespace_variable} delete jobs --all
    done
  else
    # TODO: Add org here
    export STORAGE_CLASS="local-path"
    envsubst <kube/pvc-fabric-org0.yaml | kubectl -n "$ORG0_NS" delete -f - || true
    envsubst <kube/pvc-fabric-org1.yaml | kubectl -n "$ORG1_NS" delete -f - || true
    envsubst <kube/pvc-fabric-org2.yaml | kubectl -n "$ORG2_NS" delete -f - || true
    envsubst <kube/pvc-fabric-org3.yaml | kubectl -n "$ORG3_NS" delete -f - || true
    envsubst <kube/pvc-fabric-org4.yaml | kubectl -n "$ORG4_NS" delete -f - || true

    envsubst <kube/pv-fabric-org0.yaml | kubectl -n "$ORG0_NS" delete -f - || true
    envsubst <kube/pv-fabric-org1.yaml | kubectl -n "$ORG1_NS" delete -f - || true
    envsubst <kube/pv-fabric-org2.yaml | kubectl -n "$ORG2_NS" delete -f - || true
    envsubst <kube/pv-fabric-org3.yaml | kubectl -n "$ORG3_NS" delete -f - || true
    envsubst <kube/pv-fabric-org4.yaml | kubectl -n "$ORG4_NS" delete -f - || true
  fi

  pop_fn
}

function network_down() {

  set +e
  # TODO: Add org here
  for ns in $ORG0_NS $ORG1_NS $ORG2_NS $ORG3_NS; do
    kubectl get namespace "$ns" >/dev/null
    if [[ $? -ne 0 ]]; then
      echo "No namespace $ns found - nothing to do."
      return
    fi
  done
  set -e

  stop_services
  scrub_org_volumes

  # delete_namespace

  rm -rf "$PWD"/build
}
