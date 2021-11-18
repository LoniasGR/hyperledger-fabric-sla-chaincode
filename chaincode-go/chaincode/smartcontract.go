package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
//Insert struct field in alphabetic order => to achieve determinism accross languages
// golang keeps the order when marshal to json but doesn't order automatically
type SLA struct {
	Customer string `json:"Customer"`
	ID       string `json:"String"`
	Metric   string `json:"Metric"`
	Provider string `json:"Provider"`
	Status   int    `json:"Status"`
	Value    int    `json:"Value"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// assets := []SLA{
	// 	{ID: "asset1", Customer: "blue", Metric: "Downtime", Provider: "Tomoko", Value: 300, Status: 1},
	// 	{ID: "asset2", Customer: "red", Metric: "Downtime", Provider: "Brad", Value: 400, Status: 1},
	// 	{ID: "asset3", Customer: "green", Metric: "Downtime", Provider: "Jin Soo", Value: 500, Status: 1},
	// 	{ID: "asset4", Customer: "yellow", Metric: "Downtime", Provider: "Max", Value: 600, Status: 1},
	// 	{ID: "asset5", Customer: "black", Metric: "Downtime", Provider: "Adriana", Value: 700, Status: 1},
	// 	{ID: "asset6", Customer: "white", Metric: "Downtime", Provider: "Michel", Value: 800, Status: 1},
	// }

	// for _, asset := range assets {
	// 	assetJSON, err := json.Marshal(asset)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = ctx.GetStub().PutState(asset.ID, assetJSON)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to put to world state. %v", err)
	// 	}
	// }

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, customer string, metric string, provider string, value int, status int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := SLA{
		ID:       id,
		Customer: customer,
		Metric:   metric,
		Provider: provider,
		Value:    value,
		Status:   status,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*SLA, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset SLA
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, customer string, metric string, provider string, value int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	asset := SLA{
		ID:       id,
		Customer: customer,
		Metric:   metric,
		Provider: provider,
		Value:    value,
		Status:   status,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// SLAViolated transfers the assets from the provider to the customer.
func (s *SmartContract) SLAViolated(ctx contractapi.TransactionContextInterface, id string, newStatus int) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	asset.Status = newStatus
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*SLA, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*SLA
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset SLA
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
