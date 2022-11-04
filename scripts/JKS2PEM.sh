#!/bin/bash

function create_server_cer() {

    keytool -importkeystore -srckeystore "$1" \
   -destkeystore truststore.p12 \
   -srcstoretype jks \
   -deststoretype pkcs12

   openssl pkcs12 -in truststore.p12 -out "$2" -legacy

   rm truststore.p12
}

function check_tools () {

    if ! which keytool >> /dev/null; then
        echo -e "Keytool not found. Please install openjdk-8-jre-headless"
    fi

    if ! which openssl >> /dev/null; then
        echo -e "Openssl not found. Please install it."
    fi
}

function print_help() {
    echo -e "Usage: .JKS2PEM.sh infile outfile"
}

if [ $# -ne 2 ]; then
    print_help
else
    check_tools
    create_server_cer "$@"
fi

