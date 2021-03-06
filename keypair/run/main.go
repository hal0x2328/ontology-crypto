package main

/*
 * Utility to convert OWallet key format to WIF - by hal0x2328
 *
 * Instructions:
 * 1: Edit the keyjson and pwd fields below with the information from your OWallet key export .dat file
 * 2: Open a terminal and run the following commands in the ontology-crypto/keypair/run folder:
 *    go run main.go
 */

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ontio/ontology-crypto/ec"
	"github.com/ontio/ontology-crypto/keypair"
)

var pwd = []byte("123456")

var keyjson = `{
      "address": "ASZ92mGiwVM6GixyRFU9WPxXAeYXRP6CoF",
      "algorithm": "ECDSA",
      "parameters": {
        "curve": "P-256"
      },
      "key": "4BA6AgUAIudxLtN2Lu7F4EwNpac5qjcS994MbjiGsF0/zVmTKcuDd6OEABBU/Yhi",
      "enc-alg": "aes-256-gcm",
      "salt": "LxRs+5bjsQkRT2HofV/cQw=="
}`

func main() {
	var pro keypair.ProtectedKey
	json.Unmarshal([]byte(keyjson), &pro)
	d, err := ONTDecrypt(&pro, pwd)
	if err != nil {
		fmt.Println(err)
		return
	}
	wif, err := keypair.Key2WIF(d)
	if err != nil {
		fmt.Println("Error converting Key to WIF", err)
		return
	}
	fmt.Println("WIF version of key is:", string(wif))
}

func ONTDecrypt(prot *keypair.ProtectedKey, pass []byte) (keypair.PrivateKey, error) {
	pri, err := keypair.DecryptPrivateKey(prot, pass)
	if err != nil {
		return nil, err
	}

	v, ok := pri.(*ec.PrivateKey)
	if !ok {
		return nil, errors.New("decryption error: wrong key type")
	}
	if v.Algorithm != ec.ECDSA {
		return nil, errors.New("decryption error: wrong algorithm")
	}
	return v, nil
}
