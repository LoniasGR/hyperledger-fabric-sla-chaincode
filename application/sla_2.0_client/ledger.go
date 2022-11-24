package main

import (
	"log"
	"strconv"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func InitLedger(contract *client.Contract) error {

	log.Println(green("--> Submit Transaction: InitLedger, function the connection with the ledger"))

	_, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		return err
	}
	return nil
}

func CreateOrUpdateContract(contract *client.Contract, data string) error {
	log.Println(green(`--> Submit Transaction: CreateOrUpdateContract, creates new contract or updates existing one with SLA`))

	_, err := contract.SubmitTransaction("CreateOrUpdateContract", data)
	if err != nil {
		return err
	}
	return nil
}

func SLAViolated(contract *client.Contract, slaJSON string) (string, error) {
	log.Println(green("--> Submit Transaction: SLAViolated, updates contracts details with ID, newStatus"))
	result, err := contract.SubmitTransaction("SLAViolated", slaJSON)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func CreateUser(contract *client.Contract, name, publicKey string, balance int) error {
	log.Println(green("--> Submit Transaction: CreateUser, creates new user with name, ID, publickey and an initial balance"))
	_, err := contract.SubmitTransaction("CreateUser", name, publicKey, strconv.Itoa(balance))
	if err != nil {
		return err
	}
	return nil
}

func UserExists(contract *client.Contract, name string) (bool, error) {
	log.Println(green("--> Evaluate Transaction: UserExists, checks if a user exists on chaincode"))
	result, err := contract.EvaluateTransaction("UserExists", name)
	if err != nil {
		return false, err
	}
	result_bool, err := strconv.ParseBool(string(result))
	if err != nil {
		return false, err
	}
	return result_bool, nil
}
