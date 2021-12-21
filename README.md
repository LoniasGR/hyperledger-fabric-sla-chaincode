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
bash runclient.sh
```

* Run kafka producer on a third terminal

```bash
cd testers/producer
go run sample-producer.go
```

## TODO

- [ ] "provider": { "id": "my_id", "name": "Pledger Platform1" }, "client": { "id": "c02", "name": "A client" },
- [ ] Rename mychannel to SLA
- [ ] Rename `sla` chaincode to `sla_contracts` and `sla_violation` to `sla_violation`
- [ ] When a violation happens there will be a transfer of tokens (ERC-20 style)
- [ ] `smartcontract.go` function SLAViolation Compensation scheme
- [ ] `client.go` Wallet management (provider wallet, customer wallet)
- [ ] App that when given a user certificates connects to Hyperledger
- [ ] Connect wallets, violation function and ERC-20
- [ ] Move to Kubernetes
- [ ] Slides ~30 Etherium - Deadline: Start of January
- [ ] Slides ~30 Hyperledger - Deadline: Start of January
- [ ] Intergrate whisper protocols in Fabric
