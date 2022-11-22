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

  apply_nginx_ingress
  apply_cert_manager
  apply_storage

  sleep 2

  wait_for_cert_manager
  wait_for_nginx_ingress

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
  if [ "${CLUSTER_RUNTIME}" == "kind" ] || [ "${CLUSTER_RUNTIME}" == "k3s" ]; then
    kubectl apply -f kube/ingress-nginx-"${CLUSTER_RUNTIME}".yaml
  elif [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    microk8s enable ingress

    kubectl -n ingress patch daemonset.apps/nginx-ingress-microk8s-controller --patch\
  "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"nginx-ingress-microk8s\",\"ports\":[{\"name\":\"http\",\"hostPort\":${NGINX_HTTP_PORT},\"containerPort\":80},{\"name\":\"https\",\"hostPort\":${NGINX_HTTPS_PORT},\"containerPort\":443}]}]}}}}"
  else
    echo "Unknown CLUSTER_RUNTIME ${CLUSTER_RUNTIME}"
    exit 1
  fi

  pop_fn
}

function delete_nginx_ingress() {
  push_fn "Deleting ${CLUSTER_RUNTIME} ingress controller"

  if [ "${CLUSTER_RUNTIME}" == "kind" ] || [ "${CLUSTER_RUNTIME}" == "k3s" ]; then
    kubectl delete -f kube/ingress-nginx-"${CLUSTER_RUNTIME}".yaml
  elif [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    microk8s disable ingress
  else
    echo "Unknown CLUSTER_RUNTIME ${CLUSTER_RUNTIME}"
    exit 1
  fi
  pop_fn
}

function wait_for_nginx_ingress() {
  push_fn "Waiting for ingress controller"

  if [ "${CLUSTER_RUNTIME}" == "kind" ] || [ "${CLUSTER_RUNTIME}" == "k3s" ]; then
    kubectl wait --namespace ingress-nginx \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/component=controller \
      --timeout=2m
  elif [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    :
  else
    echo "Unknown CLUSTER_RUNTIME ${CLUSTER_RUNTIME}"
    exit 1
  fi

  pop_fn
}

function apply_cert_manager() {
  push_fn "Launching cert-manager"

  # Install cert-manager to manage TLS certificates
  if [ "${CLUSTER_RUNTIME}" == "kind" ] || [ "${CLUSTER_RUNTIME}" == "k3s" ]; then
    kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.yaml
  elif [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    microk8s enable cert-manager
  else
    echo "Unknown CLUSTER_RUNTIME ${CLUSTER_RUNTIME}"
    exit 1
  fi
  pop_fn
}

function delete_cert_manager() {
  push_fn "Deleting cert-manager"

  # Remove cert-manager
  if [ "${CLUSTER_RUNTIME}" == "kind" ] || [ "${CLUSTER_RUNTIME}" == "k3s" ]; then
    kubectl delete -f https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.yaml
  elif [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    microk8s disable cert-manager
  else
    echo "Unknown CLUSTER_RUNTIME ${CLUSTER_RUNTIME}"
    exit 1
  fi

  pop_fn
}

function wait_for_cert_manager() {
  push_fn "Waiting for cert-manager"

  if [ "${CLUSTER_RUNTIME}" == "kind" ] || [ "${CLUSTER_RUNTIME}" == "k3s" ]; then
    kubectl -n cert-manager rollout status deploy/cert-manager
    kubectl -n cert-manager rollout status deploy/cert-manager-cainjector
    kubectl -n cert-manager rollout status deploy/cert-manager-webhook
  elif [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    :
  else
    echo "Unknown CLUSTER_RUNTIME ${CLUSTER_RUNTIME}"
    exit 1
  fi
  pop_fn
}

function apply_storage() {
  push_fn "Launching storage"

  if [ "${CLUSTER_RUNTIME}" == "kind" ] || [ "${CLUSTER_RUNTIME}" == "k3s" ]; then
    :
  elif [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    cat <<EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
EOF
  else
    echo "Unknown CLUSTER_RUNTIME ${CLUSTER_RUNTIME}"
    exit 1
  fi

  pop_fn
}

function delete_storage() {
  push_fn "Launching storage"

  if [ "${CLUSTER_RUNTIME}" == "kind" ] || [ "${CLUSTER_RUNTIME}" == "k3s" ]; then
    :
  elif [ "${CLUSTER_RUNTIME}" == "microk8s" ]; then
    cat <<EOF | kubectl delete -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
EOF
  else
    echo "Unknown CLUSTER_RUNTIME ${CLUSTER_RUNTIME}"
    exit 1
  fi

  pop_fn
}

function cluster_clean() {
  delete_nginx_ingress
  delete_cert_manager
  delete_storage
}
