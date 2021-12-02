package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	nftContract := new(NFTContract)

	cc, err := contractapi.NewChaincode(nftContract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
