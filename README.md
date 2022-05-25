# Hyperledger Fabric SLA Chaincode

## Setup

- Install fabric-samples

```bash
mkdir -p $HOME/go/src/github.com/<your_github_userid>
cd $HOME/go/src/github.com/<your_github_userid>
curl -sSL https://bit.ly/2ysbOFE | bash -s
```

- Clone project into fabric-samples folder

```bash
cd $HOME/go/src/github.com/<your_github_userid>/fabric-samples
git clone https://github.com/LoniasGR/hyperledger-fabric-sla-chaincode.git
```

- Spin up the Kafka container

```bash
cd $HOME/go/src/github.com/<your_github_userid>/fabric-samples/hyperledger-fabric-sla-chaincode/docker
docker-compose up -d
```

- Start Fabric Network

```bash
cd ..
bash startFabric.sh
```

- Run fabric application

```bash
cd ../../application
bash runclient.sh
```

- Run kafka producer on another terminal

```bash
cd testers/producer
go run sample-producer.go  -f ../../kafka-config/producer.properties.dev
```

## CouchDB

Default credentials:
URL: http://localhost:5984/\_utils/
Username: admin
Password: adminpw

## Fabric Explorer
Port: 8080
Username: exploreradmin
Password: exploreradminpw
## TODO

- [x] "provider": { "id": "my_id", "name": "Pledger Platform1" }, "client": { "id": "c02", "name": "A client" }.
- [x] Rename `mychannel` to `SLA`.
- [x] Rename `sla` chaincode to `sla_contracts` and `sla_violation` to `sla_violation`.
- [x] When a violation happens there will be a transfer of tokens (ERC-20 style).
- [x] `smartcontract.go` function SLAViolation Compensation scheme.
- [x] Slides ~30 Hyperledger - Deadline: Start of January.
- [x] Slides ~30 Etherium - Deadline: Start of January.
- [x] `client.go` Wallet management (provider wallet, customer wallet).
- [x] App that when given a user certificates connects to Hyperledger.
- [ ] Connect wallets, violation function and ERC-20
- [x] Chrome Extension - fix formatting
- [x] Use case 3 (see JSON) - create. channel name: "parts", topic: "uc3-dtl"
- [x] Use case 2: Return number of JSONs in time-range.
- [ ] Get number of products based time range based on quality (total, quality 1, quality 0). [10/05]
- [ ] Extension buttons to pick Use Case. []
- [ ] Different users/OUs to different channels. []
- [ ] Containerize the client.
- [ ] Move to Kubernetes.
- [x] Running a Status node (whisper protocol): https://status.im/technical/run_status_node.html
- [ ] Clients connect to our own status node. Check out how the client works. [10/05]
- [x] Use case 2: We got the data - think how to do it.
- [x] Check if UC2 client works w/ Partners. [05/05]
- [ ] Range queries for use case 2/3: new chaincode w/ name: vru/? - store & retrieve for a time range - use fabcar example
- [ ] Trusted Execution Environment module (new git branch).
