package main

import (
	"os"
	"path/filepath"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
)

// Just a small hack so that things are more modular if we ever want to change to
// some other color implementation
var red = lib.Red
var green = lib.Green

func createKeysFolder(conf lib.Config) error {
	path := filepath.Join(conf.DataFolder, "/keys")
	err := os.MkdirAll(path, os.ModeDir)
	if err != nil {
		return err
	}
	return nil
}
