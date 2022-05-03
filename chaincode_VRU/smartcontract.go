package main

import (
	"encoding/json"
	"fmt"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Contract
type SmartContract struct {
	contractapi.Contract
}

// InitLedger is just a template for now.
// Used to test the connection and verify that applications can connect to the chaincode.
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	return nil
}

func (s *SmartContract) CreateContract(ctx contractapi.TransactionContextInterface, contractJSON string) error {
	var vru lib.VRU
	err := json.Unmarshal([]byte(contractJSON), &vru)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %v", err)
	}

	exists, err := s.ContractExists(ctx, vru.Timestamp)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the Contract %s already exists", vru.Timestamp)
	}

	return ctx.GetStub().PutState(fmt.Sprintf("contract_%v", vru.Timestamp), contractJSON)
}

// ContractExists returns true when Contract with given ID exists in world state
func (s *SmartContract) ContractExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	ContractJSON, err := ctx.GetStub().GetState(fmt.Sprintf("contract_%v", id))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return ContractJSON != nil, nil
}
