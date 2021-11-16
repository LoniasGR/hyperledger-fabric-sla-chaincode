/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/chaincode-go/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	assetChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}
