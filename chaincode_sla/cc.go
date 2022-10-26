package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	assetChaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		log.Panicf("Error creating slasc_bridge chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting slasc_bridge chaincode: %v", err)
	}
}
