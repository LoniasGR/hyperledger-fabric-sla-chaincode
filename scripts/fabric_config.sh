#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

function init_namespace() {
  local namespaces
  # TODO: Add org here
  namespaces=$(echo "$ORG0_NS $ORG1_NS $ORG2_NS $ORG3_NS $ORG4_NS" | xargs -n1 | sort -u)
  for ns in $namespaces; do
    push_fn "Creating namespace \"$ns\""
    kubectl create namespace "$ns" || true
    pop_fn
  done
}

function delete_namespace() {
  local namespaces
  # TODO: Add org here
  namespaces=$(echo "$ORG0_NS $ORG1_NS $ORG2_NS $ORG3_NS $ORG4_NS" | xargs -n1 | sort -u)
  for ns in $namespaces; do
    push_fn "Deleting namespace \"$ns\""
    kubectl delete namespace "$ns" || true
    pop_fn
  done
}

function init_storage_volumes() {
  push_fn "Provisioning volume storage"

  # Both KIND and k3s use the Rancher local-path provider.  In KIND, this is installed
  # as the 'standard' storage class, and in Rancher as the 'local-path' storage class.
  if [ "${CLUSTER_RUNTIME}" == "kind" ]; then
    export STORAGE_CLASS="standard"

  else
    export STORAGE_CLASS="local-path"
  fi

  # TODO: Add org here
  envsubst <kube/pv-fabric-org0.yaml | kubectl -n "$ORG0_NS" create -f - || true
  envsubst <kube/pv-fabric-org1.yaml | kubectl -n "$ORG1_NS" create -f - || true
  envsubst <kube/pv-fabric-org2.yaml | kubectl -n "$ORG2_NS" create -f - || true
  envsubst <kube/pv-fabric-org3.yaml | kubectl -n "$ORG3_NS" create -f - || true
  envsubst <kube/pv-fabric-org4.yaml | kubectl -n "$ORG4_NS" create -f - || true

  # TODO: Add org here
  envsubst <kube/pvc-fabric-org0.yaml | kubectl -n "$ORG0_NS" create -f - || true
  envsubst <kube/pvc-fabric-org1.yaml | kubectl -n "$ORG1_NS" create -f - || true
  envsubst <kube/pvc-fabric-org2.yaml | kubectl -n "$ORG2_NS" create -f - || true
  envsubst <kube/pvc-fabric-org3.yaml | kubectl -n "$ORG3_NS" create -f - || true
  envsubst <kube/pvc-fabric-org4.yaml | kubectl -n "$ORG4_NS" create -f - || true

  pop_fn
}

function load_org_config() {

  # TODO: Add org here
  kubectl -n "$ORG0_NS" delete configmap org0-ca-config || true
  kubectl -n "$ORG1_NS" delete configmap org1-ca-config || true
  kubectl -n "$ORG2_NS" delete configmap org2-ca-config || true
  kubectl -n "$ORG3_NS" delete configmap org3-ca-config || true
  kubectl -n "$ORG4_NS" delete configmap org4-ca-config || true

  # TODO: Add org here
  kubectl -n "$ORG0_NS" delete configmap org0-orderer-config || true
  kubectl -n "$ORG1_NS" delete configmap org1-peer-config || true
  kubectl -n "$ORG2_NS" delete configmap org2-peer-config || true
  kubectl -n "$ORG3_NS" delete configmap org3-peer-config || true
  kubectl -n "$ORG4_NS" delete configmap org4-peer-config || true

  push_fn "Creating fabric config maps"
  # TODO: Add org here
  mkdir -p ${TEMP_DIR}/config/org0/ca ${TEMP_DIR}/config/org0/orderer
  mkdir -p ${TEMP_DIR}/config/org1/ca
  mkdir -p ${TEMP_DIR}/config/org2/ca
  mkdir -p ${TEMP_DIR}/config/org3/ca
  mkdir -p ${TEMP_DIR}/config/org4/ca

  envsubst <config/org0/fabric-ca-server-config.yaml >${TEMP_DIR}/config/org0/ca/fabric-ca-server-config.yaml
  envsubst <config/org1/fabric-ca-server-config.yaml >${TEMP_DIR}/config/org1/ca/fabric-ca-server-config.yaml
  envsubst <config/org2/fabric-ca-server-config.yaml >${TEMP_DIR}/config/org2/ca/fabric-ca-server-config.yaml
  envsubst <config/org3/fabric-ca-server-config.yaml >${TEMP_DIR}/config/org3/ca/fabric-ca-server-config.yaml
  envsubst <config/org4/fabric-ca-server-config.yaml >${TEMP_DIR}/config/org4/ca/fabric-ca-server-config.yaml

  kubectl -n "$ORG0_NS" create configmap org0-ca-config --from-file ${TEMP_DIR}/config/org0/ca/fabric-ca-server-config.yaml
  kubectl -n "$ORG1_NS" create configmap org1-ca-config --from-file ${TEMP_DIR}/config/org1/ca/fabric-ca-server-config.yaml
  kubectl -n "$ORG2_NS" create configmap org2-ca-config --from-file ${TEMP_DIR}/config/org2/ca/fabric-ca-server-config.yaml
  kubectl -n "$ORG3_NS" create configmap org3-ca-config --from-file ${TEMP_DIR}/config/org3/ca/fabric-ca-server-config.yaml
  kubectl -n "$ORG4_NS" create configmap org4-ca-config --from-file ${TEMP_DIR}/config/org4/ca/fabric-ca-server-config.yaml

  cp config/org0/configtx-template.yaml config/org0/orderer.yaml ${TEMP_DIR}/config/org0/orderer

  kubectl -n "$ORG0_NS" create configmap org0-orderer-config --from-file=${TEMP_DIR}/config/org0/orderer
  kubectl -n "$ORG1_NS" create configmap org1-peer-config --from-file=config/org1/core.yaml
  kubectl -n "$ORG2_NS" create configmap org2-peer-config --from-file=config/org2/core.yaml
  kubectl -n "$ORG3_NS" create configmap org3-peer-config --from-file=config/org3/core.yaml
  kubectl -n "$ORG4_NS" create configmap org4-peer-config --from-file=config/org4/core.yaml

  pop_fn
}

function apply_k8s_builder_roles() {
  push_fn "Applying k8s chaincode builder roles"

  apply_template kube/fabric-builder-role.yaml "$ORG1_NS"
  apply_template kube/fabric-builder-rolebinding.yaml "$ORG1_NS"

  pop_fn
}

function apply_k8s_builders() {
  push_fn "Installing k8s chaincode builders"

  apply_template kube/org1/org1-install-k8s-builder.yaml "$ORG1_NS"
  apply_template kube/org2/org2-install-k8s-builder.yaml "$ORG1_NS"

  pop_fn
}
