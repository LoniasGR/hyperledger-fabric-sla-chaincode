#!/bin/bash

function create_server_cer() {
    local openssl_major
    local legacy

    openssl_major=$(openssl version | cut -d ' ' -f2 | cut -d '.' -f1)
    echo -e "Using openssl version ${openssl_major}"

    # Check version of openssl
    if [ "${openssl_major}" -eq 3 ]; then
        legacy="-legacy"
    else
        legacy=""
    fi

    keytool -importkeystore -srckeystore "$1" \
        -destkeystore truststore.p12 \
        -srcstoretype jks \
        -deststoretype pkcs12

    openssl pkcs12 -in truststore.p12 -out "$2" "${legacy}"

    rm truststore.p12
}

function check_tools() {

    if ! which keytool >>/dev/null; then
        echo -e "Keytool not found. Please install openjdk-8-jre-headless"
    fi

    if ! which openssl >>/dev/null; then
        echo -e "Openssl not found. Please install it."
    fi
}

function print_help() {
    echo -e
    echo -e "Usage:"
    echo -e "     $0 INFILE OUTFILE"
    echo -e
}

if [ $# -ne 2 ]; then
    print_help
else
    check_tools
    create_server_cer "$@"
fi
