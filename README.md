# Hyperledger Fabric SLA Chaincode


## Preparations

To deploy the network, the kafka configuration is needed for the consumers. To create it:

1. Install `openjdk-8-jre-headless` or similar (`keytool` is needed) and `openssl`.
2. Run `scripts/JKS2PEM.sh`.
   For example run:
   ```bash
    ./scripts/JKS2PEM.sh ./kafka-config/kafka.client.truststore.jks ./kafka-config/server.cer.pem
   ```
3. Copy all kafka configuration files to `config/kafka`.

## Deploy on Kubernetes (KIND)

1. Run `./fabric-k8s.sh kind`.
   This creates the KIND cluster and the registry
2. Run `./fabric-k8s.sh cluster`.
   Sets up ingress and other important containers
3. Run `./fabric-k8s.sh up`.
   Brings up the CAs, orderers, peers and channels.
3. Run `./fabric-k8s.sh deploy`.
   Deploys chaincodes and clients.

## Teardown

Run `bash ./fabric-k8s.sh unkind`


## Changes needed to run on cloud

Override the corresponding variables from `network-k8s.sh` with the proper ones.