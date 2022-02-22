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

type User struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	PubKey  string `json:"pubkey"`
	Balance string `json:"balance"`
}

// InitLedger is just a template for now.
// Used to test the connection and verify that applications can connect to the chaincode.
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	return nil
}

// Returns the users balance.
func (s *SmartContract) UserBalance(ctx contractapi.TransactionContextInterface, id string) (int, error) {
	user, err := s.ReadUser(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("could not read user: %v", err)
	}

	var currentBalance int

	currentBalance, err = strconv.Atoi(string(user.Balance))
	if err != nil {
		return 0, fmt.Errorf("could not convert balance: %v", err)
	}

	return currentBalance, nil
}

func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface,
	name, id, pubkey string, initialBalance int) error {

	if initialBalance < 0 {
		return fmt.Errorf("initial amount must be zero or positive")
	}

	exists, err := s.UserExists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user info")
	}
	if exists {
		return fmt.Errorf("user already exists")
	}
	// TODO: Add check if public key exists

	user := User{
		ID:      id,
		Name:    name,
		PubKey:  pubkey,
		Balance: strconv.Itoa(initialBalance),
	}
	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("unable to marshal json: %v", err)
	}
	return ctx.GetStub().PutState(fmt.Sprintf("user_%v", id), userBytes)
}

// Mint creates new tokens and adds them to minter's account balance
func (s *SmartContract) Mint(ctx contractapi.TransactionContextInterface, id string, amount int) (string, error) {
	if amount <= 0 {
		return "", fmt.Errorf("mint amount must be a positive integer")
	}
	currentBalance, err := s.UserBalance(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to read minter account %s from world state: %v", id, err)
	}

	updatedBalance := currentBalance + amount

	err = s.UpdateUserBalance(ctx, id, updatedBalance)
	if err != nil {
		return "", fmt.Errorf("could not update user balance: %v", err)
	}

	return fmt.Sprintf("New balance is: %d\n", updatedBalance), nil
}

func (s *SmartContract) TransferTokens(ctx contractapi.TransactionContextInterface,
	from, to string, amount int) error {
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
		err = s.CreateUser(ctx, sla.Details.Provider.Name, sla.Details.Provider.ID, "", 500)
		if err != nil {
			return fmt.Errorf("could not create provider: %v", err)
		}
	}

	exists, err = s.UserExists(ctx, sla.Details.Client.ID)
	if err != nil {
		return fmt.Errorf("client account %s could not be read: %v", sla.Details.Client.ID, err)
	}
	if !exists {
		err = s.CreateUser(ctx, sla.Details.Client.Name, sla.Details.Client.ID, "", 500)
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

	return ctx.GetStub().PutState(fmt.Sprintf("contract_%v", contract.SLA.ID), slaContractJSON)
}

// ReadContract returns the Contract stored in the world state with given id.
func (s *SmartContract) ReadContract(ctx contractapi.TransactionContextInterface, id string) (*sla_contract, error) {
	ContractJSON, err := ctx.GetStub().GetState(fmt.Sprintf("contract_%v", id))
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
func (s *SmartContract) ReadUser(ctx contractapi.TransactionContextInterface, id string) (User, error) {
	userBytes, err := ctx.GetStub().GetState(fmt.Sprintf("user_%v", id))
	if err != nil {
		return User{}, fmt.Errorf("user with id %v could not be read from world state: %v", id, err)
	}
	var user User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		return User{}, fmt.Errorf("failed to unmarshal file: %v", err)
	}
	return user, nil
}

func (s *SmartContract) QueryUsersByPublicKey(ctx contractapi.TransactionContextInterface,
	publicKey string) (*User, error) {
	queryString := fmt.Sprintf(`{"selector:{"user": "pubkey", "%s"}}`, publicKey)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	if !resultsIterator.HasNext() {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	queryResult, err := resultsIterator.Next()
	if err != nil {
		return nil, fmt.Errorf("taking result from iterator failed: %v", err)
	}

	var user User
	err = json.Unmarshal(queryResult.Value, &user)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshall user: %v", err)
	}

	return &user, nil

}

func (s *SmartContract) UpdateUserBalance(ctx contractapi.TransactionContextInterface,
	id string, newBalance int) error {

	user, err := s.ReadUser(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read user %v", err)
	}
	user.Balance = strconv.Itoa(newBalance)

	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshall user: %v", err)
	}
	return ctx.GetStub().PutState(fmt.Sprintf("user_%v", id), userBytes)
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
	ContractJSON, err := ctx.GetStub().GetState(fmt.Sprintf("contract_%v", id))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return ContractJSON != nil, nil
}

// UserExists returns true when a User with given name or public key exists in world state
func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	UserJSON, err := ctx.GetStub().GetState(fmt.Sprintf("user_%v", id))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return UserJSON != nil, nil
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

// TODO: Needs to be redone!
// GetAllContracts returns all Contracts found in world state
func (s *SmartContract) GetAllContracts(ctx contractapi.TransactionContextInterface) ([]*sla_contract, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all Contracts in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("contract_0", "")
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
		log.Panicf("Error creating slasc_bridge chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting slasc_bridge chaincode: %v", err)
	}
}
