#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

function launch_ECert_CAs() {
  push_fn "Launching Fabric CAs"

  # TODO: Add org here
  for org in org0 org1 org2 org3 org4; do
    get_namespace $org
    if [ "$NO_VOLUMES" -eq 1 ]; then
      export VOLUME_CLAIM="emptyDir: {}"
    else
      VOLUME_CLAIM=$(
        cat <<EOF
persistentVolumeClaim:
            claimName: fabric-$org
EOF
      )
      export VOLUME_CLAIM
    fi
    apply_template kube/$org/$org-ca.yaml "$CURR_NS"

  done

  kubectl -n "$ORG0_NS" rollout status deploy/org0-ca
  kubectl -n "$ORG1_NS" rollout status deploy/org1-ca
  kubectl -n "$ORG2_NS" rollout status deploy/org2-ca
  kubectl -n "$ORG3_NS" rollout status deploy/org3-ca
  kubectl -n "$ORG4_NS" rollout status deploy/org4-ca

  # todo: this papers over a nasty bug whereby the CAs are ready, but sporadically refuse connections after a down / up
  sleep 5

  pop_fn
}

# experimental: create TLS CA issuers using cert-manager for each org.
function init_tls_cert_issuers() {
  push_fn "Initializing TLS certificate Issuers"

  # Create a self-signing certificate issuer / root TLS certificate for the blockchain.
  # TODO : Bring-Your-Own-Key - allow the network bootstrap to read an optional ECDSA key pair for the TLS trust root CA.
  # TODO: Add org here
  kubectl -n "$ORG0_NS" apply -f kube/root-tls-cert-issuer.yaml
  kubectl -n "$ORG0_NS" wait --timeout=30s --for=condition=Ready issuer/root-tls-cert-issuer
  kubectl -n "$ORG1_NS" apply -f kube/root-tls-cert-issuer.yaml
  kubectl -n "$ORG1_NS" wait --timeout=30s --for=condition=Ready issuer/root-tls-cert-issuer
  kubectl -n "$ORG2_NS" apply -f kube/root-tls-cert-issuer.yaml
  kubectl -n "$ORG2_NS" wait --timeout=30s --for=condition=Ready issuer/root-tls-cert-issuer
  kubectl -n "$ORG3_NS" apply -f kube/root-tls-cert-issuer.yaml
  kubectl -n "$ORG3_NS" wait --timeout=30s --for=condition=Ready issuer/root-tls-cert-issuer
  kubectl -n "$ORG4_NS" apply -f kube/root-tls-cert-issuer.yaml
  kubectl -n "$ORG4_NS" wait --timeout=30s --for=condition=Ready issuer/root-tls-cert-issuer

  # Use the self-signing issuer to generate three Issuers, one for each org.
  kubectl -n "$ORG0_NS" apply -f kube/org0/org0-tls-cert-issuer.yaml
  kubectl -n "$ORG1_NS" apply -f kube/org1/org1-tls-cert-issuer.yaml
  kubectl -n "$ORG2_NS" apply -f kube/org2/org2-tls-cert-issuer.yaml
  kubectl -n "$ORG3_NS" apply -f kube/org3/org3-tls-cert-issuer.yaml
  kubectl -n "$ORG4_NS" apply -f kube/org4/org4-tls-cert-issuer.yaml

  kubectl -n "$ORG0_NS" wait --timeout=30s --for=condition=Ready issuer/org0-tls-cert-issuer
  kubectl -n "$ORG1_NS" wait --timeout=30s --for=condition=Ready issuer/org1-tls-cert-issuer
  kubectl -n "$ORG2_NS" wait --timeout=30s --for=condition=Ready issuer/org2-tls-cert-issuer
  kubectl -n "$ORG3_NS" wait --timeout=30s --for=condition=Ready issuer/org3-tls-cert-issuer
  kubectl -n "$ORG4_NS" wait --timeout=30s --for=condition=Ready issuer/org4-tls-cert-issuer

  pop_fn
}

function enroll_bootstrap_ECert_CA_user() {
  local org=$1
  local ns=$2

  # Determine the CA information and TLS certificate
  CA_NAME=${org}-ca
  CA_DIR=${TEMP_DIR}/cas/${CA_NAME}
  mkdir -p "${CA_DIR}"

  # Read the CA's TLS certificate from the cert-manager CA secret
  echo "retrieving ${CA_NAME} TLS root cert"
  kubectl -n "$ns" get secret "${CA_NAME}"-tls-cert -o json |
    jq -r .data.\"ca.crt\" |
    base64 -d \
      >"${CA_DIR}"/tlsca-cert.pem

  # Enroll the root CA user
  # TODO: Added port here
  fabric-ca-client enroll \
    --url https://"${RCAADMIN_USER}:${RCAADMIN_PASS}@${CA_NAME}.${DOMAIN}:${NGINX_HTTPS_PORT}" \
    --tls.certfiles "$TEMP_DIR/cas/${CA_NAME}/tlsca-cert.pem" \
    --mspdir "$TEMP_DIR/enrollments/${org}/users/${RCAADMIN_USER}/msp"
}

function enroll_bootstrap_ECert_CA_users() {
  push_fn "Enrolling bootstrap ECert CA users"
  # TODO: Add org here

  enroll_bootstrap_ECert_CA_user org0 "$ORG0_NS"
  enroll_bootstrap_ECert_CA_user org1 "$ORG1_NS"
  enroll_bootstrap_ECert_CA_user org2 "$ORG2_NS"
  enroll_bootstrap_ECert_CA_user org3 "$ORG3_NS"
  enroll_bootstrap_ECert_CA_user org4 "$ORG4_NS"

  pop_fn
}
