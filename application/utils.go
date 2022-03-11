package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

const colorReset = "\033[0m"
const colorBlue = "\033[34m"
const colorRed = "\033[31m"
const colorGreen = "\033[32m"
const colorCyan = "\033[36m"

// setDiscoveryAsLocalhost sets the environmental variable DISCOVERY_AS_LOCALHOST
func setDiscoveryAsLocalhost(value bool) error {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", strconv.FormatBool(value))
	if err != nil {
		return fmt.Errorf("failed to set DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}
	return nil
}

// A wallet can hold multiple identities.
func populateWallet(wallet *gateway.Wallet, orgID int, userID int) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		fmt.Sprintf("org%d.example.com", orgID),
		"users",
		fmt.Sprintf("User%d@org%d.example.com", userID, orgID),
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity(fmt.Sprintf("Org%dMSP", orgID), string(cert), string(key))

	return wallet.Put("appUser", identity)
}
