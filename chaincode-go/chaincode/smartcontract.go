package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Contract
type SmartContract struct {
	contractapi.Contract
}

// Contract describes basic details of what makes up a simple Contract
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

// InitLedger adds a base set of Contracts to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// Contracts := []SLA{
	// 	{ID: "Contract1", Customer: "blue", Metric: "Downtime", Provider: "Tomoko", Value: 300, Status: 1},
	// 	{ID: "Contract2", Customer: "red", Metric: "Downtime", Provider: "Brad", Value: 400, Status: 1},
	// 	{ID: "Contract3", Customer: "green", Metric: "Downtime", Provider: "Jin Soo", Value: 500, Status: 1},
	// 	{ID: "Contract4", Customer: "yellow", Metric: "Downtime", Provider: "Max", Value: 600, Status: 1},
	// 	{ID: "Contract5", Customer: "black", Metric: "Downtime", Provider: "Adriana", Value: 700, Status: 1},
	// 	{ID: "Contract6", Customer: "white", Metric: "Downtime", Provider: "Michel", Value: 800, Status: 1},
	// }

	// for _, Contract := range Contracts {
	// 	ContractJSON, err := json.Marshal(Contract)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = ctx.GetStub().PutState(Contract.ID, ContractJSON)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to put to world state. %v", err)
	// 	}
	// }

	return nil
}

// CreateContract issues a new Contract to the world state with given details.
func (s *SmartContract) CreateContract(ctx contractapi.TransactionContextInterface, id string, customer string, metric string, provider string, value int, status int) error {
	exists, err := s.ContractExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the Contract %s already exists", id)
	}

	Contract := SLA{
		ID:       id,
		Customer: customer,
		Metric:   metric,
		Provider: provider,
		Value:    value,
		Status:   status,
	}
	ContractJSON, err := json.Marshal(Contract)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, ContractJSON)
}

// ReadContract returns the Contract stored in the world state with given id.
func (s *SmartContract) ReadContract(ctx contractapi.TransactionContextInterface, id string) (*SLA, error) {
	ContractJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if ContractJSON == nil {
		return nil, fmt.Errorf("the Contract %s does not exist", id)
	}

	var Contract SLA
	err = json.Unmarshal(ContractJSON, &Contract)
	if err != nil {
		return nil, err
	}

	return &Contract, nil
}

// UpdateContract updates an existing Contract in the world state with provided parameters.
func (s *SmartContract) UpdateContract(ctx contractapi.TransactionContextInterface, id string, customer string, metric string, provider string, value int) error {
	exists, err := s.ContractExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the Contract %s does not exist", id)
	}

	// overwriting original Contract with new Contract
	Contract := SLA{
		ID:       id,
		Customer: customer,
		Metric:   metric,
		Provider: provider,
		Value:    value,
		Status:   status,
	}
	ContractJSON, err := json.Marshal(Contract)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, ContractJSON)
}

// DeleteContract deletes an given Contract from the world state.
func (s *SmartContract) DeleteContract(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.ContractExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the Contract %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// ContractExists returns true when Contract with given ID exists in world state
func (s *SmartContract) ContractExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	ContractJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return ContractJSON != nil, nil
}

// SLAViolated transfers the Contracts from the provider to the customer.
func (s *SmartContract) SLAViolated(ctx contractapi.TransactionContextInterface, id string, newStatus int) error {
	Contract, err := s.ReadContract(ctx, id)
	if err != nil {
		return err
	}

	Contract.Status = newStatus
	ContractJSON, err := json.Marshal(Contract)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, ContractJSON)
}

// GetAllContracts returns all Contracts found in world state
func (s *SmartContract) GetAllContracts(ctx contractapi.TransactionContextInterface) ([]*SLA, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all Contracts in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var Contracts []*SLA
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var Contract SLA
		err = json.Unmarshal(queryResponse.Value, &Contract)
		if err != nil {
			return nil, err
		}
		Contracts = append(Contracts, &Contract)
	}

	return Contracts, nil
}
