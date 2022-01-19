package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

// KeyValueContract contract for handling writing and reading from the world state
type KeyValueContract struct {
	contractapi.Contract
}

const TokenDataLen = 100

type RecordData struct {
	Key         string                 `json:"key"`
	TokenData   map[string]interface{} `json:"tokenData,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
}

// Create adds a new key with value to the world state
func (sc *KeyValueContract) Create(ctx contractapi.TransactionContextInterface, key string, value string) error {
	return putDataByKey(ctx, key, value)
}

func putDataByKey(ctx contractapi.TransactionContextInterface, key string, value string) error {
	existing, err := ctx.GetStub().GetState(key)
	if err != nil {
		return errors.New("unable to interact with world state")
	}
	if existing != nil {
		return fmt.Errorf("cannot create world state pair with key %s. Already exists", key)
	}

	var recordData RecordData
	err = json.Unmarshal([]byte(value), &recordData)
	if err != nil {
		log.Printf("failed to json.Unmarshal([]byte(value), &recordData) in Create: %v", err)
		return err
	}

	if len(recordData.TokenData) > TokenDataLen {
		err = errors.New("len(recordData.TokenData) > (TokenDataLen=100),len is must <= 100")
		log.Print(err)
		return err
	}

	var recordDataBytes []byte
	recordDataBytes, _ = json.Marshal(recordData)
	err = ctx.GetStub().PutState(key, recordDataBytes)

	if err != nil {
		return errors.New("unable to interact with world state")
	}

	return nil
}

// Update changes the value with key in the world state
func (sc *KeyValueContract) Update(ctx contractapi.TransactionContextInterface, key string, value string) error {
	return putDataByKey(ctx, key, value)
}

// Read returns the value at key in the world state
func (sc *KeyValueContract) Read(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	existing, err := ctx.GetStub().GetState(key)

	if err != nil {
		return "", errors.New("unable to interact with world state")
	}

	if existing == nil {
		return "", fmt.Errorf("cannot read world state pair with key %s. Does not exist", key)
	}

	return string(existing), nil
}
