#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

function kind_create() {
  push_fn "Creating cluster \"${CLUSTER_NAME}\""

  # prevent the next kind cluster from using the previous Fabric network's enrollments.
  rm -rf "$PWD"/build

  # todo: always delete?  Maybe return no-op if the cluster already exists?
  kind delete cluster --name "$CLUSTER_NAME"

  local ingress_http_port=${NGINX_HTTP_PORT}
  local ingress_https_port=${NGINX_HTTPS_PORT}

  # the 'ipvs'proxy mode permits better HA abilities

  cat <<EOF | kind create cluster --name "$CLUSTER_NAME" --config=-
---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
      - containerPort: 80
        hostPort: ${ingress_http_port}
        protocol: TCP
      - containerPort: 443
        hostPort: ${ingress_https_port}
        protocol: TCP
    extraMounts:
    - hostPath: ${HOME}/hyperledger
      containerPath: /var/hyperledger

#networking:
#  kubeProxyMode: "ipvs"

containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.configs."${CONTAINER_REGISTRY_HOSTNAME}".tls]
      insecure_skip_verify = true
EOF

  # workaround for https://github.com/hyperledger/fabric-samples/issues/550 - pods can not resolve external DNS
  for node in $(kind get nodes); do
    docker exec "$node" sysctl net.ipv4.conf.all.route_localnet=1
  done

  pop_fn
}

function kind_delete() {
  push_fn "Deleting KIND cluster ${CLUSTER_NAME}"

  kind delete cluster --name "$CLUSTER_NAME"

  pop_fn
}

function kind_init() {
  kind_create
}

function kind_unkind() {
  kind_delete
}
