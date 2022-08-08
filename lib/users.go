package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

type UserKeys struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

type UserRequest struct {
	Username     string `json:"username"`
	Organization int    `json:"org"`
}

func UserExistsOrCreate(contract *gateway.Contract, name, id string, org int) (bool, string, error) {
	result, err := contract.EvaluateTransaction("UserExists", id)
	if err != nil {
		err = fmt.Errorf(string(ColorRed)+"failed to submit transaction: %s\n"+string(ColorReset), err)
		return false, "", err
	}
	result_bool, err := strconv.ParseBool(string(result))
	if err != nil {
		err = fmt.Errorf(string(ColorRed)+"failed to parse boolean: %s\n"+string(ColorReset), err)
		return false, "", err
	}
	if !result_bool {
		postBody, err := json.Marshal(UserRequest{
			Username:     name,
			Organization: org,
		})
		if err != nil {
			err = fmt.Errorf(string(ColorRed)+"failed to marshall post request: %s\n"+string(ColorReset), err)
			return false, "", err
		}
		responseBody := bytes.NewBuffer(postBody)
		resp, err := http.Post(createUserUrl, "application/json", responseBody)
		if err != nil {
			err = fmt.Errorf(string(ColorRed)+"failed to send post request: %s\n"+string(ColorReset), err)
			return false, "", err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf(string(ColorRed)+"failed to get response body: %s\n"+string(ColorReset), err)
			return false, "", err
		}

		// We use interface{} since we don't know the contents of the JSON beforehand
		var responseBodyJSON map[string]interface{}
		err = json.Unmarshal(body, &responseBodyJSON)
		if err != nil {
			err = fmt.Errorf(string(ColorRed)+"failed to unmarshal response body: %s\n"+string(ColorReset), err)
			return false, "", err
		}
		if responseBodyJSON["success"] == true {
			// get the data of the internal JSON
			data, ok := responseBodyJSON["data"].(map[string]interface{})
			if !ok {
				err = fmt.Errorf(string(ColorRed) + "failed to convert interface to struct\n" + string(ColorReset))
				return false, "", err
			}
			// convert interface{} to string
			privateKey := fmt.Sprintf("%v", data["privateKey"])
			publicKey := fmt.Sprintf("%v", data["publicKey"])

			publicKeyStripped := splitCertificate(publicKey)
			// privateKeyStripped := splitCertificate(privateKey)

			err = saveCertificates(name, privateKey, publicKey)
			if err != nil {
				err = fmt.Errorf(string(ColorRed)+"failed to save certificates: %s"+string(ColorReset), err)
				return false, "", err
			}
			return false, strings.ReplaceAll(publicKeyStripped, "\n", ""), nil
		} else if responseBodyJSON["success"] == false && responseBodyJSON["error"] == "User already exists" {
			return true, "", nil
		} else {
			return false, "", fmt.Errorf(string(ColorRed)+"response failure: %v"+string(ColorReset), responseBodyJSON["error"])
		}
	}
	return true, "", nil
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
	filename := fmt.Sprintf(keysFolder+"%v.keys", name)
	err := os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("failed to write keys: %v", err)
	}
	return nil
}
