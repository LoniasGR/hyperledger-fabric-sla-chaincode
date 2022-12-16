package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type UserKeys struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

type UserRequest struct {
	Username     string `json:"username"`
	Organization int    `json:"org"`
}

func UserExistsOrCreate(contract *client.Contract, name string, balance, org int, conf lib.Config) (bool, string, error) {
	exists, cert, err := userExists(contract, name, org, conf)
	if err != nil {
		return false, "", err
	}

	// The user exists both on cc and wallet
	if exists && cert == "" {
		return true, "", nil
	}

	// The user exists and cert is not empty, so we add the user on the cc
	if exists {
		publicKeyStripped := splitCertificate(cert)
		publicKeyOneLine := strings.ReplaceAll(publicKeyStripped, "\n", "")
		err = CreateUser(contract, name, publicKeyOneLine, balance)
		if err != nil {
			return false, "", err
		}
		return false, publicKeyOneLine, nil
	}

	// Now we know that the user does not exist

	postBody, err := json.Marshal(UserRequest{
		Username:     name,
		Organization: org,
	})
	if err != nil {
		return false, "", fmt.Errorf("%w", err)
	}

	data, err := processRequest(conf.IdentityEndpoint+"/create", postBody)
	if err != nil {
		return false, "", err
	}

	// convert interface{} to string
	privateKey := fmt.Sprintf("%v", data["privateKey"])
	publicKey := fmt.Sprintf("%v", data["publicKey"])

	publicKeyStripped := splitCertificate(publicKey)
	// privateKeyStripped := splitCertificate(privateKey)

	err = saveCertificates(name, privateKey, publicKey, conf)
	if err != nil {
		return false, "", fmt.Errorf("failed to save certificates: %s", err)
	}

	publicKeyOneLine := strings.ReplaceAll(publicKeyStripped, "\n", "")
	err = CreateUser(contract, name, publicKeyOneLine, balance)
	if err != nil {
		return false, "", err
	}
	return false, publicKeyOneLine, nil
}

// Remove header and footer from key
func splitCertificate(certificate string) string {
	certificateSplit := strings.Split(certificate, "-----")
	return strings.Trim(certificateSplit[2], "\n")
}

// Save the certificates of the user
func saveCertificates(name, privateKey, publicKey string, conf lib.Config) error {
	data := fmt.Sprintf("%v\n%v",
		privateKey, publicKey)
	filename := name + ".keys"
	path := filepath.Join(conf.DataFolder, "/keys/", filename)
	err := os.WriteFile(path, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("failed to write keys: %w", err)
	}
	return nil
}

func userExists(contract *client.Contract, name string, org int, conf lib.Config) (bool, string, error) {
	exists, err := UserExists(contract, name)
	if err != nil {
		return false, "", err
	}
	if exists {
		return true, "", nil
	}

	// Here we know that the user does not exists on chaincode.
	// We need to check if it exists on the wallet.

	postBody, err := json.Marshal(UserRequest{
		Username:     name,
		Organization: org,
	})
	if err != nil {
		return false, "", err
	}

	data, err := processRequest(conf.IdentityEndpoint+"/exists", postBody)
	if err != nil {
		return false, "", err
	}
	found := data["exists"].(bool)
	if !found {
		return false, "", nil
	}

	// The user is found on wallet, but not on chaincode, so they need to be added.
	cert := data["cert"].(string)
	return true, cert, nil

}

func processRequest(endpoint string, postBody []byte) (map[string]interface{}, error) {
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(endpoint, "application/json", responseBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// We use interface{} since we don't know the contents of the JSON beforehand
	var responseBodyJSON map[string]interface{}
	err = json.Unmarshal(body, &responseBodyJSON)
	if err != nil {
		return nil, err
	}
	if responseBodyJSON["success"] == false {
		switch responseBodyJSON["error"] {
		case "User already exists":
			return nil, fmt.Errorf("user does not exist on ledger, but exists on user service")
		case "Malformed request":
			return nil, fmt.Errorf("the request sent to identity service was malformed")
		default:
			return nil, fmt.Errorf("response failure: %v", responseBodyJSON["error"])
		}
	}
	// get the data of the internal JSON
	data, ok := responseBodyJSON["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert interface to struct")
	}
	return data, nil
}
