# Hyperledger Fabric SLA Chaincode


## Preparations

To deploy the network, the kafka configuration is needed for the consumers. To create it:

1. Install `sudo apt install openjdk-8-jre-headless` or similar (`keytool` is needed).
2. Run `scripts/JKS2PEM.sh`.
   For example run:
   ```bash
    ./scripts/JKS2PEM.sh ./kafka-config/kafka.client.truststore.jks ./kafka-config/server.cer.pem
   ```
3. Copy all kafka configuration files to `config/kafka`.

## Deploy on Kubernetes

1. Run `bash ./fabric-k8s.sh deploy`

## Changes needed to run on cloud

Override the corresponding variables from `network-k8s.sh` with the proper ones.