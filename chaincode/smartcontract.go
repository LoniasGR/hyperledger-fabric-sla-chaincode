package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Contract
type SmartContract struct {
	contractapi.Contract
}

type sla_contract struct {
	lib.SLA
	Value      int `json:"Value"` // compensation amount
	Violations int `json:"Violations"`
}

// InitLedger is just a template for now.
// Used to test the connection and verify that applications can connect to the chaincode.
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	return nil
}

// Returns the users balance. Can be used to check for existence of user.
func (s *SmartContract) UserBalance(ctx contractapi.TransactionContextInterface, name string) (int, error) {
	currentBalanceBytes, err := ctx.GetStub().GetState(name)
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}

	var currentBalance int

	// If user current balance doesn't yet exist, we'll create it with a current balance of 0
	if currentBalanceBytes == nil {
		currentBalance = 0
	} else {
		// Error handling not needed since Itoa() was used when setting the account balance,
		// guaranteeing it was an integer.
		currentBalance, _ = strconv.Atoi(string(currentBalanceBytes))
	}

	return currentBalance, nil
}

// Mint creates new tokens and adds them to minter's account balance
func (s *SmartContract) Mint(ctx contractapi.TransactionContextInterface, user string, amount int) (string, error) {
	// clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	// if err != nil {
	// 	return "", fmt.Errorf("failed to get MSPID: %v", err)
	// }

	// Example of how to manage who can get tokens
	// if clientMSPID != "Org1MSP" {
	// 	return fmt.Errorf("client is not authorized to mint new tokens")
	// }

	// minter, err := ctx.GetClientIdentity().GetID()
	// if err != nil {
	// 	return "", fmt.Errorf("failed to get client ID: %v", err)
	// }
	if amount <= 0 {
		return "", fmt.Errorf("mint amount must be a positive integer")
	}
	currentBalance, err := s.UserBalance(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to read minter account %s from world state: %v", user, err)
	}

	updatedBalance := currentBalance + amount

	err = ctx.GetStub().PutState(user, []byte(strconv.Itoa(updatedBalance)))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("New balance is: %d\n", updatedBalance), nil
}

func (s *SmartContract) TransferTokens(ctx contractapi.TransactionContextInterface, from string, to string, amount int) error {
	if from == to {
		return fmt.Errorf("cannot transfer from and to the same account")
	}

	fromBalance, err := s.UserBalance(ctx, from)
	if err != nil {
		return fmt.Errorf("could not get balance of transferer during token transfer: %v", err)
	}
	if fromBalance < amount {
		return fmt.Errorf("transferer does not have enough tokens to complete transfer")
	}

	toBalance, err := s.UserBalance(ctx, to)
	if err != nil {
		return fmt.Errorf("could not get balance of transferee during token transfer: %v", err)
	}

	updatedFromBalance := fromBalance - amount
	updatedToBalance := toBalance + amount

	err = ctx.GetStub().PutState(from, []byte(strconv.Itoa(updatedFromBalance)))
	if err != nil {
		return fmt.Errorf("failed to update sender's balance: %v", err)
	}

	err = ctx.GetStub().PutState(to, []byte(strconv.Itoa(updatedToBalance)))
	if err != nil {
		return fmt.Errorf("failed to update receiver's balance: %v", err)
	}
	return nil
}

// CreateContract issues a new Contract to the world state with given details.
func (s *SmartContract) CreateContract(ctx contractapi.TransactionContextInterface, contractJSON string) error {
	var sla lib.SLA
	err := json.Unmarshal([]byte(contractJSON), &sla)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %v", err)
	}

	exists, err := s.ContractExists(ctx, sla.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the Contract %s already exists", sla.ID)
	}

	providerBalance, err := s.UserBalance(ctx, sla.Details.Provider.Name)
	if err != nil {
		return fmt.Errorf("failed to read provider account %s from world state: %v", sla.Details.Provider.ID, err)
	}

	if providerBalance == 0 {
		_, err = s.Mint(ctx, sla.Details.Provider.Name, rand.Intn(500)+500)
		if err != nil {
			return fmt.Errorf("failed to mint tokens for provider %s: %v", sla.Details.Provider.ID, err)
		}
	}

	contract := sla_contract{
		SLA:        sla,
		Value:      rand.Intn(20) + 10,
		Violations: 0,
	}
	slaContractJSON, err := json.Marshal(contract)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(contract.SLA.ID, slaContractJSON)
}

// ReadContract returns the Contract stored in the world state with given id.
func (s *SmartContract) ReadContract(ctx contractapi.TransactionContextInterface, id string) (*sla_contract, error) {
	ContractJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if ContractJSON == nil {
		return nil, fmt.Errorf("the Contract %s does not exist", id)
	}

	var contract sla_contract
	err = json.Unmarshal(ContractJSON, &contract)
	if err != nil {
		return nil, err
	}

	return &contract, nil
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

// SLAViolated changes the number of violations that have happened.
func (s *SmartContract) SLAViolated(ctx contractapi.TransactionContextInterface, id string) error {
	contract, err := s.ReadContract(ctx, id)
	if err != nil {
		return err
	}
	if contract.SLA.State == "stopped" {
		return fmt.Errorf("the contract %s is completed, no violations can happen", id)
	}

	contract.Violations += 1
	ContractJSON, err := json.Marshal(contract)
	if err != nil {
		return err
	}

	err = s.TransferTokens(ctx, contract.SLA.Details.Provider.Name, contract.SLA.Details.Client.Name, contract.Value)
	if err != nil {
		return fmt.Errorf("could not transfer tokens from violation: %v", err)
	}

	return ctx.GetStub().PutState(id, ContractJSON)
}

// GetAllContracts returns all Contracts found in world state
func (s *SmartContract) GetAllContracts(ctx contractapi.TransactionContextInterface) ([]*sla_contract, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all Contracts in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("a0", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var Contracts []*sla_contract
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var Contract sla_contract
		err = json.Unmarshal(queryResponse.Value, &Contract)
		if err != nil {
			return nil, err
		}
		Contracts = append(Contracts, &Contract)
	}

	return Contracts, nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}
