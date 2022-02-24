# Hyperledger Fabric SLA Chaincode

## Setup

* Install fabric-samples

```bash
mkdir -p $HOME/go/src/github.com/<your_github_userid>
cd $HOME/go/src/github.com/<your_github_userid>
curl -sSL https://bit.ly/2ysbOFE | bash -s
```

* Clone project into fabric-samples folder

```bash
cd $HOME/go/src/github.com/<your_github_userid>/fabric-samples
git clone https://github.com/LoniasGR/hyperledger-fabric-sla-chaincode.git
```

* Spin up the Kafka container

```bash
cd $HOME/go/src/github.com/<your_github_userid>/fabric-samples/hyperledger-fabric-sla-chaincode/docker
docker-compose up -d
```


* Start Fabric Network

```bash
cd ..
bash startFabric.sh
```

* Run fabric application

```bash
cd ../../application
bash runclient.sh
```

* Run kafka producer on another terminal

```bash
cd testers/producer
go run sample-producer.go
```

## TODO

- [x] "provider": { "id": "my_id", "name": "Pledger Platform1" }, "client": { "id": "c02", "name": "A client" },
- [x] Rename `mychannel` to `SLA`
- [x] Rename `sla` chaincode to `sla_contracts` and `sla_violation` to `sla_violation`
- [x] When a violation happens there will be a transfer of tokens (ERC-20 style)
- [x] `smartcontract.go` function SLAViolation Compensation scheme
- [x] Slides ~30 Hyperledger - Deadline: Start of January
- [x] Slides ~30 Etherium - Deadline: Start of January
- [ ] `client.go` Wallet management (provider wallet, customer wallet)
- [ ] App that when given a user certificates connects to Hyperledger
- [ ] Connect wallets, violation function and ERC-20
- [ ] Move to Kubernetes
- [ ] Containerize the client
- [ ] Intergrate whisper protocols in Fabric
