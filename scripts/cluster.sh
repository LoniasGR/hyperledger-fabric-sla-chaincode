#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

# cluster "group" commands.  Like "main" for the fabric-cli "cluster" sub-command
function cluster_command_group() {

  # Default COMMAND is 'init' if not specified
  if [ "$#" -eq 0 ]; then
    COMMAND="init"

  else
    COMMAND=$1
    shift
  fi

  if [ "${COMMAND}" == "init" ]; then
    log "Initializing K8s cluster"
    cluster_init
    log "🏁 - Cluster is ready"

  elif [ "${COMMAND}" == "clean" ]; then
    log "Cleaning k8s cluster"
    cluster_clean
    log "🏁 - Cluster is cleaned"

  else
    print_help
    exit 1
  fi
}

function cluster_init() {
  init_namespace

  if [ "$SELF_SIGNED_REGISTRY" == 1 ]; then
    add_registry_key
  fi
  apply_dns
  apply_nginx_ingress
  apply_cert_manager
  apply_metrics_server
  if [ "$NO_VOLUMES" -eq 0 ]; then
    apply_storage
  fi

  sleep 2

  # wait_for_dns
  wait_for_cert_manager
  wait_for_nginx_ingress
  wait_for_metrics_server
}

function cluster_clean() {
  delete_dns
  delete_nginx_ingress
  delete_cert_manager
  if [ "$NO_VOLUMES" -eq 0 ]; then
    delete_storage
  fi
  delete_metrics_server
  delete_namespace

  if [ "$SELF_SIGNED_REGISTRY" == 1 ]; then
    delete_registry_key_pod
  fi
}

function add_registry_key() {
  # More info:
  # http://hypernephelist.com/2021/03/23/kubernetes-containerd-certificate.html
  # https://github.com/containerd/containerd/issues/6485
  #
  push_fn "Adding registry key"
  kubectl -n "$NS" create configmap trusted-ca --from-file=config/docker/ca.crt || true
  envsubst <kube/add-private-registry-tls-cert-on-nodes.yaml | kubectl -n ${NS} apply -f -
  pop_fn
}

function delete_registry_key_pod() {
  push_fn "Deleting registry key pod"
  kubectl -n "$NS" delete configmap trusted-ca || true
  envsubst <kube/add-private-registry-tls-cert-on-nodes.yaml | kubectl -n ${NS} delete -f - || true
  pop_fn
}

function apply_dns() {
  echo ${CLUSTER_RUNTIME}
  push_fn "Launching DNS"
  if [ "${CLUSTER_RUNTIME}" == "kind" ]; then
    :
  elif [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    microk8s enable dns
  else
    envsubst <kube/coredns-deployment.yaml | kubectl apply -f -
  fi
  pop_fn
}

function delete_dns() {
  push_fn "Deleting DNS"
  if [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    microk8s disable dns
  elif [ "${CLUSTER_RUNTIME}" == "kind" ]; then
    :
  else
    envsubst <kube/coredns-deployment.yaml | kubectl delete -f - || :
  fi
  pop_fn
}

function wait_for_dns() {
  push_fn "Waiting for dns"

  kubectl -n ${NS} rollout status deploy/coredns

  pop_fn
}

function apply_nginx() {
  apply_nginx_ingress
  wait_for_nginx_ingress
}

function apply_nginx_ingress() {
  push_fn "Launching ${CLUSTER_RUNTIME} ingress controller"

  # 1.1.2 static ingress with modifications to enable ssl-passthrough
  # k3s : 'cloud'
  # kind : 'kind'
  # kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.2/deploy/static/provider/cloud/deploy.yaml
  if [ "${CLUSTER_RUNTIME}" == "kind" ]; then
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
  else
    envsubst <kube/ingress-deployment.yaml | kubectl apply -f -
  fi

  pop_fn
}

function delete_nginx_ingress() {
  push_fn "Deleting ${CLUSTER_RUNTIME} ingress controller"

  if [ "${CLUSTER_RUNTIME}" == "kind" ]; then
    kubectl delete -f kube/ingress-nginx-kind.yaml || :
  else
    envsubst <kube/ingress-deployment.yaml | kubectl delete -f - || :
  fi
  pop_fn
}

function wait_for_nginx_ingress() {
  push_fn "Waiting for ingress controller"

  kubectl wait --namespace ${NS} \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=2m

  pop_fn
}

function apply_cert_manager() {
  push_fn "Launching cert-manager"

  # Install cert-manager to manage TLS certificates
  envsubst <kube/cert-manager-deployment.yaml | kubectl apply -f -
  pop_fn
}

function delete_cert_manager() {
  push_fn "Deleting cert-manager"

  # Remove cert-manager
  envsubst <kube/cert-manager-deployment.yaml | kubectl delete -f - || :
  pop_fn
}

function wait_for_cert_manager() {
  push_fn "Waiting for cert-manager"

  kubectl -n ${NS} rollout status deploy/cert-manager
  kubectl -n ${NS} rollout status deploy/cert-manager-cainjector
  kubectl -n ${NS} rollout status deploy/cert-manager-webhook
  pop_fn
}

function apply_storage() {
  push_fn "Launching storage"

  if [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    cat <<EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
EOF
  fi
  pop_fn
}

function delete_storage() {
  push_fn "Deleting storage"

  if [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    cat <<EOF | kubectl delete -f - || true
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
EOF
  fi

  pop_fn
}

function apply_metrics_server() {
  push_fn "Launching Metrics Server"
  envsubst <kube/metrics-server-deployment.yaml | kubectl apply -f -
  pop_fn
}

function delete_metrics_server() {
  push_fn "Deleting Metrics Server"
  envsubst <kube/metrics-server-deployment.yaml | kubectl delete -f - || :
  pop_fn

}

function wait_for_metrics_server() {
  kubectl -n ${NS} rollout status deploy/metrics-server
}
