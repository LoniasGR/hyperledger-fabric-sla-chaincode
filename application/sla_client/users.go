package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

func UserExistsOrCreate(contract *client.Contract, name string, balance, org int) (bool, string, error) {
	result, err := contract.EvaluateTransaction("UserExists", name)
	if err != nil {
		err = fmt.Errorf(string(lib.ColorRed)+"failed to submit transaction: %s\n"+string(lib.ColorReset), err)
		return false, "", err
	}
	result_bool, err := strconv.ParseBool(string(result))
	if err != nil {
		err = fmt.Errorf(string(lib.ColorRed)+"failed to parse boolean: %s\n"+string(lib.ColorReset), err)
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
		err = fmt.Errorf(string(lib.ColorRed)+"failed to marshall post request: %s\n"+string(lib.ColorReset), err)
		return false, "", err
	}
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(lib.CreateUserUrl, "application/json", responseBody)
	if err != nil {
		err = fmt.Errorf(string(lib.ColorRed)+"failed to send post request: %s\n"+string(lib.ColorReset), err)
		return false, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf(string(lib.ColorRed)+"failed to get response body: %s\n"+string(lib.ColorReset), err)
		return false, "", err
	}

	// We use interface{} since we don't know the contents of the JSON beforehand
	var responseBodyJSON map[string]interface{}
	err = json.Unmarshal(body, &responseBodyJSON)
	if err != nil {
		err = fmt.Errorf(string(lib.ColorRed)+"failed to unmarshal response body: %s\n"+string(lib.ColorReset), err)
		return false, "", err
	}
	if responseBodyJSON["success"] == false {
		if responseBodyJSON["error"] == "User already exists" {
			return false, "", fmt.Errorf("user does not exist on ledger, but exists on user service")
		}
		return false, "", fmt.Errorf(string(lib.ColorRed)+"response failure: %v"+string(lib.ColorReset), responseBodyJSON["error"])
	}
	// get the data of the internal JSON
	data, ok := responseBodyJSON["data"].(map[string]interface{})
	if !ok {
		err = fmt.Errorf(string(lib.ColorRed) + "failed to convert interface to struct\n" + string(lib.ColorReset))
		return false, "", err
	}
	// convert interface{} to string
	privateKey := fmt.Sprintf("%v", data["privateKey"])
	publicKey := fmt.Sprintf("%v", data["publicKey"])

	publicKeyStripped := splitCertificate(publicKey)
	// privateKeyStripped := splitCertificate(privateKey)

	err = saveCertificates(name, privateKey, publicKey)
	if err != nil {
		err = fmt.Errorf(string(lib.ColorRed)+"failed to save certificates: %s"+string(lib.ColorReset), err)
		return false, "", err
	}

	publicKeyOneLine := strings.ReplaceAll(publicKeyStripped, "\n", "")
	log.Println(string(lib.ColorGreen), `--> Submit Transaction:
					CreateUser, creates new user with name, ID, publickey and an initial balance`, string(lib.ColorReset))
	_, err = contract.SubmitTransaction("CreateUser", name, publicKeyOneLine, strconv.Itoa(balance))
	if err != nil {
		return false, "", fmt.Errorf(string(lib.ColorRed)+"failed to submit transaction: %s\n"+string(lib.ColorReset), err)
	}

	return false, publicKeyOneLine, nil
}

// Remove header and footer from key
func splitCertificate(certificate string) string {
	certificateSplit := strings.Split(certificate, "-----")
	return strings.Trim(certificateSplit[2], "\n")
}

// Save the certificates of the user
func saveCertificates(name, privateKey, publicKey string) error {
	data := fmt.Sprintf("%v\n%v",
		privateKey, publicKey)
	filename := fmt.Sprintf(lib.KeysFolder+"%v.keys", name)
	err := os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("failed to write keys: %v", err)
	}
	return nil
}
