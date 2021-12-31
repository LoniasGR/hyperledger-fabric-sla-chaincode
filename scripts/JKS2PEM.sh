#!/bin/sh


keytool -importkeystore -srckeystore ../kafka-config/kafka.client.truststore.jks -storepass -destkeystore server.p12 -deststoretype PKCS12
openssl pkcs12 -in server.p12 -nokeys -out server.cer.pem
mv ./server.cer.pem ../kafka-config/
