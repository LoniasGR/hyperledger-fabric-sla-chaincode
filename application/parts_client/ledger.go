package main

import (
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
)

func createContract(contract *client.Contract, value string) error {
	log.Println(lib.Green(`--> Submit Transaction:
    CreateContract, creates new parts entry with ID, Timestamp
    and all Document details`))
	_, err := contract.SubmitTransaction("CreateContract", value)
	if err != nil {
		return err
	}
	return nil
}

func initLedger(contract *client.Contract) error {

	log.Println(lib.Green("--> Submit Transaction: InitLedger, function the connection with the ledger"))

	_, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		return err
	}
	return nil
}
