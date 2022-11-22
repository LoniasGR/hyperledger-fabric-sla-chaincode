package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	result, err := contract.EvaluateTransaction("UserExists", name)
	if err != nil {
		return false, "", err
	}
	result_bool, err := strconv.ParseBool(string(result))
	if err != nil {
		return false, "", err
	}
	if result_bool {
		return true, "", nil
	}

	postBody, err := json.Marshal(UserRequest{
		Username:     name,
		Organization: org,
	})
	if err != nil {
		return false, "", fmt.Errorf("%w", err)
	}
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post((conf.IdentityEndpoint + "/create"), "application/json", responseBody)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}

	// We use interface{} since we don't know the contents of the JSON beforehand
	var responseBodyJSON map[string]interface{}
	err = json.Unmarshal(body, &responseBodyJSON)
	if err != nil {
		return false, "", err
	}
	if responseBodyJSON["success"] == false {
		if responseBodyJSON["error"] == "User already exists" {
			return false, "", fmt.Errorf("user does not exist on ledger, but exists on user service")
		}
		return false, "", fmt.Errorf("response failure: %v", responseBodyJSON["error"])
	}
	// get the data of the internal JSON
	data, ok := responseBodyJSON["data"].(map[string]interface{})
	if !ok {
		return false, "", fmt.Errorf("failed to convert interface to struct")
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
