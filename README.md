# Hyperledger Fabric SLA Chaincode

## Setup

* Install fabric-samples

```bash
mkdir -p $HOME/go/src/github.com/<your_github_userid>
cd $HOME/go/src/github.com/<your_github_userid>
curl -sSL https://bit.ly/2ysbOFE | bash -s
```

* Install Kafka on the home folder

```bash
cd $HOME
wget https://dlcdn.apache.org/kafka/3.0.0/kafka_2.13-3.0.0.tgz
tar -xzf kafka_2.13-3.0.0.tgz
```

* Clone project into fabric-samples folder

```bash
cd $HOME/go/src/github.com/<your_github_userid>/fabric-samples 
git clone https://github.com/LoniasGR/hyperledger-fabric-sla-chaincode.git
```

* Run Kafka and Zookeeper on one terminal

```bash
bash startKafka.sh
```

* Open another terminal and start Fabric Network

```bash
bash startFabric.sh
```

* Run fabric application

```bash
cd ../../application
bash runclient.st
```

* Run kafka producer on a third terminal

```bash
cd testers/producer
go run sample-producer.go
```

## TODO

- [x] `client.go` integrate Kafka consumer
- [x] Violation Kafka topic: `sla_violation`
- [x] SLA Smart Contract Bridge (Kafka topic:`sla`)
- [x] Keep track how many times an SLA is violated.
- [x] Have the work so far to be demonstrable - alpha version.
- [ ] ERC20 fabric-sample
- [ ] `client.go` Wallet management (provider wallet, customer wallet)
- [ ] `smartcontract.go` function SLAViolation Compensation scheme
