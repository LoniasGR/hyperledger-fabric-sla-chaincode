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
	RefundValue     int     `json:"RefundValue"` // compensation amount
	TotalViolations []int   `json:"TotalViolations"`
	DailyValue      float64 `json:"DailyValue"`
	DailyViolations []int   `json:"DailyViolations"`
}

type User struct {
	DocType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
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
func (s *SmartContract) UserBalance(ctx contractapi.TransactionContextInterface, id string) (float64, error) {
	user, err := s.ReadUser(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("could not read user: %v", err)
	}

	var currentBalance float64

	currentBalance, err = strconv.ParseFloat(string(user.Balance), 64)
	if err != nil {
		return 0, fmt.Errorf("could not convert balance: %v", err)
	}

	return currentBalance, nil
}

func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface,
	name, pubkey string, initialBalance int) error {

	if initialBalance < 0 {
		return fmt.Errorf("initial amount must be zero or positive")
	}

	exists, err := s.UserExists(ctx, name)
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
		Name:    name,
		PubKey:  pubkey,
		Balance: strconv.Itoa(initialBalance),
	}
	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("unable to marshal json: %v", err)
	}
	return ctx.GetStub().PutState(fmt.Sprintf("user_%v", name), userBytes)
}

// Mint creates new tokens and adds them to minter's account balance
func (s *SmartContract) Mint(ctx contractapi.TransactionContextInterface, id string, amount float64) (string, error) {
	if amount <= 0 {
		return "", fmt.Errorf("mint amount must be a positive integer")
	}
	currentBalance, err := s.UserBalance(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to read minter account %s from world state: %v", id, err)
	}

	updatedBalance := currentBalance + amount

	err = s.updateUserBalance(ctx, id, updatedBalance)
	if err != nil {
		return "", fmt.Errorf("could not update user balance: %v", err)
	}

	return fmt.Sprintf("New balance is: %f\n", updatedBalance), nil
}

func (s *SmartContract) transferTokens(ctx contractapi.TransactionContextInterface,
	from, to string, amount float64) error {
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

	err = s.updateUserBalance(ctx, from, updatedFromBalance)
	if err != nil {
		return fmt.Errorf("could not update sender's balance: %v", err)
	}

	err = s.updateUserBalance(ctx, to, updatedToBalance)
	if err != nil {
		return fmt.Errorf("could not update receiver's balance: %v", err)
	}
	return nil
}

// CreateOrUpdateContract issues a new Contract to the world state with given details.
func (s *SmartContract) CreateOrUpdateContract(ctx contractapi.TransactionContextInterface, contractJSON string) error {
	var sla lib.SLA
	err := json.Unmarshal([]byte(contractJSON), &sla)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %v", err)
	}

	exists, err := s.UserExists(ctx, sla.Details.Provider.Name)
	if err != nil {
		return fmt.Errorf("provider account %s could not be read: %v", sla.Details.Provider.ID, err)
	}
	if !exists {
		return fmt.Errorf("provider does not exist")
	}

	exists, err = s.UserExists(ctx, sla.Details.Client.Name)
	if err != nil {
		return fmt.Errorf("client account %s could not be read: %v", sla.Details.Client.ID, err)
	}
	if !exists {
		return fmt.Errorf("client does not exist")
	}

	exists, err = s.ContractExists(ctx, sla.ID)
	if err != nil {
		return err
	}

	value := rand.Intn(20) + 10
	totalViolations := make([]int, 1)
	dailyViolations := make([]int, 1)
	dailyValue := 0.0

	if exists {
		contract, err := s.ReadContract(ctx, sla.ID)
		if err != nil {
			return err
		}
		value = contract.RefundValue
		totalViolations = contract.TotalViolations
		dailyViolations = contract.DailyViolations
		dailyValue = contract.DailyValue
	}

	contract := sla_contract{
		SLA:             sla,
		RefundValue:     value,
		TotalViolations: totalViolations,
		DailyViolations: dailyViolations,
		DailyValue:      dailyValue,
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
		return User{}, fmt.Errorf("could not unmarshal user: %v", err)
	}

	return user, nil

}

func (s *SmartContract) updateUserBalance(ctx contractapi.TransactionContextInterface,
	id string, newBalance float64) error {

	user, err := s.ReadUser(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read user %v", err)
	}
	user.Balance = fmt.Sprintf("%f", newBalance)

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

	return ctx.GetStub().DelState(fmt.Sprintf("contract_%v", id))
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
func (s *SmartContract) SLAViolated(ctx contractapi.TransactionContextInterface, violation string) error {
	var vio lib.Violation
	err := json.Unmarshal([]byte(violation), &vio)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %v", err)
	}

	contract, err := s.ReadContract(ctx, vio.SLAID)
	if err != nil {
		return err
	}
	if contract.SLA.State == "stopped" {
		return fmt.Errorf("the contract %s is completed, no violations can happen", vio.SLAID)
	}

	switch vio.GuaranteeID {
	case "40":
		// This should happen only the first time the SLA is violated, but it's the time
		// we actually have information about the violation itself.
		if len(contract.DailyViolations) < 3 {
			contract.DailyViolations = make([]int, 3)
			contract.TotalViolations = make([]int, 3)
		}
		switch vio.ImportanceName {
		case "Warning":
			contract.DailyValue += (1 - 0.985) * float64(contract.RefundValue)
			contract.DailyViolations[0] += 1
		case "Serious":
			contract.DailyValue += (1 - 0.965) * float64(contract.RefundValue)
			contract.DailyViolations[1] += 1
		case "Catastrophic":
			contract.DailyValue += (1 - 0.945) * float64(contract.RefundValue)
			contract.DailyViolations[2] += 1
		}
	// If we don't know the type of guarantee
	default:
		contract.DailyViolations[0] += 1
	}
	ContractJSON, err := json.Marshal(contract)
	if err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("could not transfer tokens from violation: %v", err)
	}

	return ctx.GetStub().PutState(fmt.Sprintf("contract_%v", vio.SLAID), ContractJSON)
}

func (s *SmartContract) RefundSLA(ctx contractapi.TransactionContextInterface, id string) error {
	contract, err := s.ReadContract(ctx, id)
	if err != nil {
		return err
	}

	if contract.SLA.State == "stopped" {
		return fmt.Errorf("the contract %s is completed, no violations can happen", id)
	}

	err = s.transferTokens(ctx, contract.SLA.Details.Provider.Name, contract.SLA.Details.Client.Name, contract.DailyValue)
	if err != nil {
		return err
	}

	for i := 0; i < len(contract.DailyViolations); i++ {
		contract.TotalViolations[i] += contract.DailyViolations[i]
		contract.DailyViolations[i] = 0
	}
	contract.DailyValue = 0.0

	ContractJSON, err := json.Marshal(contract)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(fmt.Sprintf("contract_%v", id), ContractJSON)
}

func (s *SmartContract) RefundAllSLAs(ctx contractapi.TransactionContextInterface) error {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all kv pairs in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		if strings.HasPrefix(queryResponse.Key, "user_") {
			continue
		}

		var contract sla_contract
		err = json.Unmarshal(queryResponse.Value, &contract)
		if err != nil {
			return err
		}

		err = s.RefundSLA(ctx, contract.ID)
		if err != nil {
			return err
		}
	}
	return nil
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
