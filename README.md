# Hyperledger Fabric SLA Chaincode

## Setup

* Clone project into fabric-samples folder
* Install Kafka

```bash
wget https://dlcdn.apache.org/kafka/3.0.0/kafka_2.13-3.0.0.tgz
tar -xzf kafka_2.13-3.0.0.tgz
```

* Run Kafka and Zookeeper on two different terminals

```bash
cd kafka_2.13-3.0.0
bin/zookeeper-server-start.sh config/zookeeper.properties
```

```bash
cd kafka_2.13-3.0.0
bin/kafka-server-start.sh config/server.properties
```

## TODO

- [x] `client.go` integrate Kafka consumer
- [] `client.go` Wallet management (provider wallet, customer wallet)
- [] `smartcontract.go` function SLAViolation Compensation scheme
- [] Violation Kafka topic: `sla_violation`
- [] SLA Smart Contract Bridge (Kafka topic:)