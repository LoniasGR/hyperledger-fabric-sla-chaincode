package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Contract
type SmartContract struct {
	contractapi.Contract
}

type sla_contract struct {
	lib.SLA
	Modifier   int `json:"Modifier"` // compensation amount
	Violations int `json:"Violations"`
}

type User struct {
	DocType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
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

	user, err := s.QueryUsersByPublicKey(ctx, pubkey)
	if err != nil {
		return fmt.Errorf("querying for public key failed: %v", err)
	}
	if (user != User{}) {
		return fmt.Errorf("public key already exists")
	}

	user = User{
		DocType: "user",
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
		return fmt.Errorf("provider does not exist")
	}

	exists, err = s.UserExists(ctx, sla.Details.Client.ID)
	if err != nil {
		return fmt.Errorf("client account %s could not be read: %v", sla.Details.Client.ID, err)
	}
	if !exists {
		return fmt.Errorf("client does not exist")
	}

	contract := sla_contract{
		SLA:        sla,
		Modifier:   10,
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
	publicKey string) (User, error) {
	publicKey = strings.ReplaceAll(publicKey, "\n", "")
	queryString := fmt.Sprintf(`{"selector":{"docType":"user","pubkey":"%s"}}`, publicKey)
	fmt.Println(queryString)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return User{}, fmt.Errorf("query failed: %v", err)
	}
	defer resultsIterator.Close()

	if !resultsIterator.HasNext() {
		return User{}, nil
	}

	queryResult, err := resultsIterator.Next()
	if err != nil {
		return User{}, fmt.Errorf("taking result from iterator failed: %v", err)
	}

	var user User
	err = json.Unmarshal(queryResult.Value, &user)
	if err != nil {
		return User{}, fmt.Errorf("could not unmarshall user: %v", err)
	}

	return user, nil

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
func (s *SmartContract) SLAViolated(ctx contractapi.TransactionContextInterface, violationJSON string) error {
	var violation lib.Violation

	err := json.Unmarshal([]byte(violationJSON), &violation)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %v", err)
	}

	contract, err := s.ReadContract(ctx, violation.SLAID)
	if err != nil {
		return err
	}
	if contract.SLA.State == "stopped" {
		return fmt.Errorf("the contract %s is completed, no violations can happen", contract.ID)
	}

	contract.Violations += 1
	ContractJSON, err := json.Marshal(contract)
	if err != nil {
		return err
	}

	compensationAmount := contract.Modifier * violation.Importance

	err = s.TransferTokens(ctx, contract.SLA.Details.Provider.ID, contract.SLA.Details.Client.ID, compensationAmount)
	if err != nil {
		return fmt.Errorf("could not transfer tokens from violation: %v", err)
	}

	return ctx.GetStub().PutState(contract.ID, ContractJSON)
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
