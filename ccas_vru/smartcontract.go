package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Contract
type SmartContract struct {
	contractapi.Contract
}

type vru_st struct {
	Timestamp int64        `json:"timestamp"`
	Trams     []lib.Tram_s `json:"trams"`
	OBUs      []lib.OBU_s  `json:"obus"`
}

// InitLedger is just a template for now.
// Used to test the connection and verify that applications can connect to the chaincode.
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	return nil
}

func (s *SmartContract) CreateContract(ctx contractapi.TransactionContextInterface, contractJSON string) error {
	var vru lib.VRU
	var vruCC vru_st

	err := json.Unmarshal([]byte(contractJSON), &vru)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %v", err)
	}

	exists, err := s.ContractExists(ctx, strconv.FormatInt(vru.Timestamp, 10))
	if err != nil {
		return fmt.Errorf("failed to check contract existence: %v", err)
	}
	if exists {
		vruCC.Trams = append(vruCC.Trams, vru.Tram)
		vruCC.OBUs = append(vruCC.OBUs, vru.OBUs...)
	} else {
		vruCC.Timestamp = vru.Timestamp
		vruCC.Trams = make([]lib.Tram_s, 1)
		vruCC.Trams[0] = vru.Tram
		vruCC.OBUs = vru.OBUs
	}

	vruCCJson, err := json.Marshal(vruCC)
	if err != nil {
		return fmt.Errorf("could not marshal vru chaincode struct: %v", err)
	}

	return ctx.GetStub().PutState(fmt.Sprintf("%v", vru.Timestamp), []byte(vruCCJson))
}

// ContractExists returns true when Contract with given ID exists in world state
func (s *SmartContract) ContractExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	ContractJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return ContractJSON != nil, nil
}

func (s *SmartContract) getAssetByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]*vru_st, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*vru_st
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset vru_st
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (s *SmartContract) GetAssetRiskInRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) (lib.Risk, error) {
	assets, err := s.getAssetByRange(ctx, startKey, endKey)
	if err != nil {
		return lib.Risk{}, err
	}
	var risk = lib.Risk{}

	for _, asset := range assets {
		for _, OBU := range asset.OBUs {
			if OBU.Risk == "CRITICAL" {
				risk.Critical += 1
			}
			if OBU.Risk == "WARNING" {
				risk.Warning += 1
			}
			if OBU.Risk == "HIGHRISK" {
				risk.HighRisk += 1
			}
			if OBU.Risk == "LOWRISK" {
				risk.LowRisk += 1
			}
			if OBU.Risk == "NORISK" {
				risk.NoRisk += 1
			}
		}
	}
	return risk, nil
}
