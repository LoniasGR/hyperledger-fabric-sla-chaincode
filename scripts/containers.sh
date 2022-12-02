#!/bin/bash

function build_containers() {
    TAG="${1:-pledger}"
    PUSH="${2:-0}"

    log "Building chaincode images"
    build_chaincode_image "$SLA_CC_SRC_PATH" "$TAG/$SLA_CHAINCODE_NAME:latest"

    build_chaincode_image "$VRU_CC_SRC_PATH" "$TAG/$VRU_CHAINCODE_NAME:latest"

    build_chaincode_image "$PARTS_CC_SRC_PATH" "$TAG/$PARTS_CHAINCODE_NAME:latest"

    log "Building client images"
    push_fn "Building $TAG/sla-client image"
    cp config/kafka/consumer.properties application/sla_client/
    cp config/kafka/kafka.client.keystore.jks application/sla_client/
    cp config/kafka/kafka.client.truststore.jks application/sla_client/
    cp config/kafka/server.cer.pem application/sla_client/
    docker build -t "$TAG/sla-client:latest" application/sla_client
    rm application/sla_client/consumer.properties
    rm application/sla_client/kafka.client.keystore.jks
    rm application/sla_client/kafka.client.truststore.jks
    rm application/sla_client/server.cer.pem
    pop_fn

    push_fn "Building $TAG/vru-client image"
    cp config/kafka/consumer.properties application/vru_client/
    cp config/kafka/kafka.client.keystore.jks application/vru_client/
    cp config/kafka/kafka.client.truststore.jks application/vru_client/
    cp config/kafka/server.cer.pem application/vru_client/
    docker build -t "$TAG/vru-client:latest" application/vru_client
    rm application/vru_client/consumer.properties
    rm application/vru_client/kafka.client.keystore.jks
    rm application/vru_client/kafka.client.truststore.jks
    rm application/vru_client/server.cer.pem
    pop_fn

    push_fn "Building$TAG/parts-client image"
    cp config/kafka/consumer.properties application/parts_client/
    cp config/kafka/kafka.client.keystore.jks application/parts_client/
    cp config/kafka/kafka.client.truststore.jks application/parts_client/
    cp config/kafka/server.cer.pem application/parts_client/
    docker build -t "$TAG/parts-client:latest" application/parts_client
    rm application/parts_client/consumer.properties
    rm application/parts_client/kafka.client.keystore.jks
    rm application/parts_client/kafka.client.truststore.jks
    rm application/parts_client/server.cer.pem
    pop_fn

    # push_fn "Building ${TAG}/sla2-client image"

    # cp config/kafka/consumer.properties application/sla_2.0_client/
    # cp config/kafka/kafka.client.keystore.jks application/sla_2.0_client/
    # cp config/kafka/kafka.client.truststore.jks application/sla_2.0_client/
    # cp config/kafka/server.cer.pem application/sla_2.0_client/

    # docker build -t "${TAG}/sla2-client:latest" application/sla_2.0_client
    # pop_fn

    push_fn "Building ${TAG}/identity-management client image"
    docker build -t "${TAG}/identity-management:latest" application/identity_management
    pop_fn

    push_fn "Building ${TAG}/api client image"
    docker build -t "${TAG}/api:latest" application/api
    pop_fn

    if [ $PUSH -ne 0 ]; then
        docker push "$TAG/$SLA_CHAINCODE_NAME:latest"
        docker push "$TAG/$VRU_CHAINCODE_NAME:latest"
        docker push "$TAG/sla-client:latest"
        docker push "$TAG/vru-client:latest"
        docker push "$TAG/parts-client:latest"
        # docker push "$TAG/sla2-client:latest"
        docker push "$TAG/api:latest"
        docker push "$TAG/identity-management:latest"

    fi

}

function docker_login() {
    set +x
    push_fn "Creating creating docker login credentials"
    local cred_file=${PWD}/config/docker/.docker_credentials.json
    # We use xargs to remove quotations from the stirngs
    jq .password <"${cred_file}" | xargs | docker login --username "$(jq .username <"${cred_file}" | xargs)" --password-stdin "${CONTAINER_REGISTRY_HOSTNAME}"
    kubectl create secret generic docker-secret -n ${NS} \
        --from-file=.dockerconfigjson=${HOME}/.docker/config.json \
        --type=kubernetes.io/dockerconfigjson
    pop_fn
}

function docker_command_group() {
    COMMAND=$1
    shift

    if [ "${COMMAND}" == "login" ]; then
        docker_login
    elif [ "${COMMAND}" == "build" ]; then
        build_containers "$@"
    else
        log "Uknown command"
        exit 1
    fi
}
