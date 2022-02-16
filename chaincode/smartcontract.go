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

// Returns the users balance.
func (s *SmartContract) UserBalance(ctx contractapi.TransactionContextInterface, identifier string) (int, error) {
	_, userData, err := s.ReadUser(ctx, identifier)
	if err != nil {
		return 0, fmt.Errorf("could not read user: %v", err)
	}

	var currentBalance int

	// Error handling not needed since Itoa() was used when setting the account balance,
	// guaranteeing it was an integer.
	currentBalance, _ = strconv.Atoi(string(userData))

	return currentBalance, nil
}

func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface,
	user string, id string, initialBalance int) error {

	if initialBalance < 0 {
		return fmt.Errorf("initial amount must be zero or positive")
	}

	nameExists, err := s.UserExists(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to get user info")
	}
	if nameExists {
		return fmt.Errorf("user already exists")
	}

	idExists, err := s.UserExists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user info")
	}
	if idExists {
		return fmt.Errorf("public key already exists")
	}

	userIndex, err := ctx.GetStub().CreateCompositeKey(index, []string{user, id})
	if err != nil {
		return fmt.Errorf("could not create composite key:  %v", err)
	}

	return ctx.GetStub().PutState(userIndex, []byte(strconv.Itoa(initialBalance)))
}

// Mint creates new tokens and adds them to minter's account balance
func (s *SmartContract) Mint(ctx contractapi.TransactionContextInterface, user string, amount int) (string, error) {
	if amount <= 0 {
		return "", fmt.Errorf("mint amount must be a positive integer")
	}
	currentBalance, err := s.UserBalance(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to read minter account %s from world state: %v", user, err)
	}

	updatedBalance := currentBalance + amount

	err = s.UpdateUserBalance(ctx, user, updatedBalance)
	if err != nil {
		return "", fmt.Errorf("could not update user balance: %v", err)
	}

	return fmt.Sprintf("New balance is: %d\n", updatedBalance), nil
}

func (s *SmartContract) TransferTokens(ctx contractapi.TransactionContextInterface,
	from string, to string, amount int) error {
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

	err = s.UpdateUserBalance(ctx, from, updatedFromBalance)
	if err != nil {
		return fmt.Errorf("could not update sender's balance: %v", err)
	}

	err = s.UpdateUserBalance(ctx, to, updatedToBalance)
	if err != nil {
		return fmt.Errorf("could not update receiver's balance: %v", err)
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

	exists, err = s.UserExists(ctx, sla.Details.Provider.ID)
	if err != nil {
		return fmt.Errorf("provider account %s could not be read: %v", sla.Details.Provider.ID, err)
	}
	if !exists {
		err = s.CreateUser(ctx, sla.Details.Provider.Name, sla.Details.Provider.ID, 500)
		if err != nil {
			return fmt.Errorf("could not create provider: %v", err)
		}
	}

	exists, err = s.UserExists(ctx, sla.Details.Client.ID)
	if err != nil {
		return fmt.Errorf("client account %s could not be read: %v", sla.Details.Client.ID, err)
	}
	if !exists {
		err = s.CreateUser(ctx, sla.Details.Client.Name, sla.Details.Client.ID, 500)
		if err != nil {
			return fmt.Errorf("could not create client: %v", err)
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

// ReadUser returns the User stored in the world state with given name or public key.
func (s *SmartContract) ReadUser(ctx contractapi.TransactionContextInterface, identifier string) (string, []byte, error) {
	UserJSONIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(index, []string{identifier})
	if err != nil {
		return "", []byte{}, fmt.Errorf("failed to read from world state: %v", err)
	}
	if !UserJSONIterator.HasNext() {
		return "", []byte{}, fmt.Errorf("user does not exist")
	}

	userKeyValue, err := UserJSONIterator.Next()
	if err != nil {
		return "", []byte{}, fmt.Errorf("failed to get user key value pair: %v", err)
	}
	return userKeyValue.Key, userKeyValue.Value, nil
}

func (s *SmartContract) UpdateUserBalance(ctx contractapi.TransactionContextInterface,
	identifier string, newBalance int) error {

	userKey, _, err := s.ReadUser(ctx, identifier)
	if err != nil {
		return fmt.Errorf("failed to read user %v", err)
	}
	return ctx.GetStub().PutState(userKey, []byte(strconv.Itoa(newBalance)))
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

// UserExists returns true when a User with given name or public key exists in world state
func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, identifier string) (bool, error) {
	UserJSONIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(index, []string{identifier})
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	defer UserJSONIterator.Close()

	return UserJSONIterator.HasNext(), nil
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

	err = s.TransferTokens(ctx, contract.SLA.Details.Provider.ID, contract.SLA.Details.Client.ID, contract.Value)
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
