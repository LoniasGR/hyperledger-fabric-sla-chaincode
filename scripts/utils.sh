#!/bin/bash

C_RESET='\033[0m'
C_RED='\033[0;31m'
C_GREEN='\033[0;32m'
C_BLUE='\033[0;34m'
C_YELLOW='\033[1;33m'

# println echos string
function println() {
  echo -e "$1"
}

# errorln echos i red color
function errorln() {
  println "${C_RED}${1}${C_RESET}"
}

# successln echos in green color
function successln() {
  println "${C_GREEN}${1}${C_RESET}"
}

# infoln echos in blue color
function infoln() {
  println "${C_BLUE}${1}${C_RESET}"
}

# warnln echos in yellow color
function warnln() {
  println "${C_YELLOW}${1}${C_RESET}"
}

# fatalln echos in red color and exits with fail status
function fatalln() {
  errorln "$1"
  exit 1
}

# Set an environment variable based on an optional override (TEST_NETWORK_${name})
# from the calling shell.  If the override is not available, assign the parameter
# to a default value.
function context() {
  local variable=$1
  local default_value=$2
  local override_name=TEST_NETWORK_${variable}

  export "${variable}"="${!override_name:-${default_value}}"
}

function logging_init() {
  # Reset the output and debug log files
  printf '' | tee "${LOG_FILE}" "${DEBUG_FILE}"

  # Write all output to the control flow log to STDOUT
  tail -f "${LOG_FILE}" &

  # Call the exit handler when we exit.
  trap "exit_fn" EXIT

  # Send stdout and stderr from child programs to the debug log file
  exec 1>>"${DEBUG_FILE}" 2>>"${DEBUG_FILE}"

  # There can be a race between the tail starting and the next log statement
  sleep 0.5
}

function exit_fn() {
  rc=$?
  set +x

  # Write an error icon to the current logging statement.
  if [ "0" -ne $rc ]; then
    pop_fn $rc
  fi

  # always remove the log trailer when the process exits.
  pkill -P $$
}

function push_fn() {
  #echo -ne "   - entering ${FUNCNAME[1]} with arguments $@"

  echo -ne "   - $* ..." >> "${LOG_FILE}"
}

function log() {
  echo -e "$@" >> "${LOG_FILE}"
}

function pop_fn() {
#  echo exiting ${FUNCNAME[1]}

  if [ $# -eq 0 ]; then
    echo -ne "\r✅"  >> "${LOG_FILE}"
    echo "" >> "${LOG_FILE}"
    return
  fi

  local res=$1
  if [ "$res" -eq 0 ]; then
    echo -ne "\r✅\n"  >> "${LOG_FILE}"

  elif [ "$res" -eq 1 ]; then
    echo -ne "\r⚠️\n" >> "${LOG_FILE}"

  elif [ "$res" -eq 2 ]; then
    echo -ne "\r☠️\n" >> "${LOG_FILE}"

  elif [ "$res" -eq 127 ]; then
    echo -ne "\r☠️\n" >> "${LOG_FILE}"

  else
    echo -ne "\r\n" >> "${LOG_FILE}"
  fi

  if [ "$res" -ne 0 ]; then
    tail -"${LOG_ERROR_LINES}" network-debug.log >> "${LOG_FILE}"
  fi

  #echo "" >> "${LOG_FILE}"
}

function wait_for_deployment() {
  local name=$1
  push_fn "Waiting for deployment $name"

  kubectl -n "$NS" rollout status deploy "$name"

  pop_fn
}

function absolute_path() {
  local relative_path=$1
  local abspath

  abspath="$( cd "${relative_path}" && pwd )"

  echo "$abspath"
}

function apply_kustomization() {
  $KUSTOMIZE_BUILD "$1" | envsubst | kubectl -n "$NS" apply -f -
}

function undo_kustomization() {
  $KUSTOMIZE_BUILD "$1" | envsubst | kubectl -n "$NS" delete --ignore-not-found=true -f -
}

function create_image_pull_secret() {
  local secret=$1
  local registry=$2
  local username=$3
  local password=$4

  push_fn "Creating $secret for access to $registry"

  kubectl -n "$NS" delete secret "$secret" --ignore-not-found

  # todo: can this be moved to a kustomization overlay?
  kubectl -n "$NS" \
    create secret docker-registry \
    "$secret" \
    --docker-server="$registry" \
    --docker-username="$username" \
    --docker-password="$password"

  pop_fn
}

function export_peer_context() {
  local orgnum=$1
  local peernum=$2
  local org=org${orgnum}
  local peer=peer${peernum}

#  export FABRIC_LOGGING_SPEC=DEBUG

  export FABRIC_CFG_PATH=${PWD}/config
  export CORE_PEER_ADDRESS=${NS}-${org}-${peer}-peer.${INGRESS_DOMAIN}:443
  export CORE_PEER_LOCALMSPID=Org${orgnum}MSP
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_MSPCONFIGPATH=${TEMP_DIR}/enrollments/${org}/users/${org}admin/msp
  export CORE_PEER_TLS_ROOTCERT_FILE=${TEMP_DIR}/channel-msp/peerOrganizations/${org}/msp/tlscacerts/tlsca-signcert.pem

#  export | egrep "CORE_|FABRIC_"
}

export -f errorln
export -f successln
export -f infoln
export -f warnln
export -f context