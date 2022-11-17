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
4. Copy `docker_credentials.json.example` to `docker_credentials.json` and change the credentials so that you can
   push on the registry of your choice.

## Deploy on Kubernetes (KIND)

1. Run `./fabric-k8s.sh kind`.
   This creates the KIND cluster and the registry
2. Run `./fabric-k8s.sh cluster`.
   Sets up ingress and other important containers
3. Run `./fabric-k8s.sh up`.
   Brings up the CAs, orderers and peers.
4. Run `./fabric-k8s.sh channels`.
   Brings up the channels.
5. Login to the container registry by running `./fabric-k8s.sh login`.
   This needs to happen now, because namespace to have been created.
6. Run `./fabric-k8s.sh chaincodes`.
   Brings up the channels.
7. Run `./fabric-k8s.sh applications`.
   Deploys chaincodes and clients.

## Shut down network

Run `./fabric-k8s.sh down`

## Remove the cluster

Run `./fabric-k8s.sh unkind`


## Changes needed to run on cloud

Override the corresponding variables from `network-k8s.sh` with the proper ones.