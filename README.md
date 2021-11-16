# Hyperledger Fabric SLA Chaincode

* [Basic example that uses SDK to query and execute transaction](https://github.com/hyperledger/fabric-sdk-go/blob/main/test/integration/e2e/end_to_end.go)
  
## TODO

1) `client.go` integrate Kafka consumer
2) `client.go` Wallet management (provider wallet, customer wallet)
3) `smartcontract.go` function SLAViolation Compensation scheme
4) Violation Kafka topic: `sla_violation`
5) SLA Smart Contract Bridge (Kafka topic:)