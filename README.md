# Hyperledger Fabric SLA Chaincode

## Setup

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

* [Basic example that uses SDK to query and execute transaction](https://github.com/hyperledger/fabric-sdk-go/blob/main/test/integration/e2e/end_to_end.go)
  
## TODO

1) `client.go` integrate Kafka consumer
2) `client.go` Wallet management (provider wallet, customer wallet)
3) `smartcontract.go` function SLAViolation Compensation scheme
4) Violation Kafka topic: `sla_violation`
5) SLA Smart Contract Bridge (Kafka topic:)