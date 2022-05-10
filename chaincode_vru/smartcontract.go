package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

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

	exists, err := s.ContractExists(ctx, strconv.FormatInt(vru.Timestamp, 10))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the Contract %v already exists", vru.Timestamp)
	}

	return ctx.GetStub().PutState(fmt.Sprintf("%v", vru.Timestamp), []byte(contractJSON))
}

// ContractExists returns true when Contract with given ID exists in world state
func (s *SmartContract) ContractExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	ContractJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return ContractJSON != nil, nil
}

func (s *SmartContract) GetAssetByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]*lib.VRU, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*lib.VRU
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset lib.VRU
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (s *SmartContract) GetTotalAssetsInRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) (int, error) {
	assets, err := s.GetAssetByRange(ctx, startKey, endKey)
	if err != nil {
		return 0, err
	}

	return len(assets), nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		log.Panicf("Error creating vru_positions chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting vru_positions chaincode: %v", err)
	}
}
