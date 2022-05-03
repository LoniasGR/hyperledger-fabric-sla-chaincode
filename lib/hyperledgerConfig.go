package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

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
