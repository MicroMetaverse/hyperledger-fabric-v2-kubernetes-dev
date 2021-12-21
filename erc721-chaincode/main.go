package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	erc721Contract := new(ERC721Contract)

	cc, err := contractapi.NewChaincode(erc721Contract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
