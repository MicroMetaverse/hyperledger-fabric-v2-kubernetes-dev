package main

import (
	"encoding/json"
	"errors"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
	"strconv"
)

// Define objectType names for prefix
const balancePrefix = "balance"
const nftPrefix = "nft"
const approvalPrefix = "approval"

// Define key names for options
const nameKey = "name"
const symbolKey = "symbol"

// ERC721Contract contract for handling writing and reading from the world state
type ERC721Contract struct {
	contractapi.Contract
}

type NFT struct {
	TokenId  int    `json:"tokenId"`
	Owner    string `json:"owner"`
	TokenURI string `json:"tokenURI"`
	Approved string `json:"approved"`
}

// TransferEvent event provides an organized struct for emitting events
type TransferEvent struct {
	From    string `json:"from"`
	To      string `json:"to"`
	TokenId int    `json:"tokenId"`
}

//ApprovalEvent
type ApprovalEvent struct {
	Owner    string `json:"owner"`
	Approved string `json:"approved"`
	TokenId  int    `json:"tokenId"`
}

type Approval struct {
	Owner    string `json:"owner"`
	Operator string `json:"operator"`
	Approved bool   `json:"approved"`
}

/*
 * BalanceOf counts all non-fungible tokens assigned to an owner
 *
 * @param {Context} ctx the transaction context
 * @param {String} owner An owner for whom to query the balance
 * @returns {Number} The number of non-fungible tokens owned by the owner, possibly zero
 */
func BalanceOf(ctx contractapi.TransactionContextInterface, owner string) int {
	// There is a key record for every non-fungible token in the format of balancePrefix.owner.tokenId.
	// BalanceOf() queries for and counts all records matching balancePrefix.owner.*
	var balance = 0
	keys := []string{owner}
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey(balancePrefix, keys)
	if err != nil {
		log.Printf("BalanceOf GetStateByPartialCompositeKey balancePrefix key: %s is err", owner)
		return balance
	}

	// Count the number of returned composite keys
	result, _ := iterator.Next()

	for result != nil {
		balance++
		result, _ = iterator.Next()
	}
	return balance
}

/**
 * OwnerOf finds the owner of a non-fungible token
 *
 * @param {Context} ctx the transaction context
 * @param {String} tokenId The identifier for a non-fungible token
 * @returns {String} Return the owner of the non-fungible token
 */
func (sc *ERC721Contract) OwnerOf(ctx contractapi.TransactionContextInterface, tokenId string) (string, error) {
	var strEmpty string = ""
	nft, err := _readNFT(ctx, tokenId)
	if err != nil {
		log.Printf("failed to _readNFT: %v", err)
		return strEmpty, err
	}

	owner := nft.Owner
	if len(owner) == 0 {
		return strEmpty, errors.New(`no owner is assigned to this token`)
	}

	return owner, nil
}

/**
 * TransferFrom transfers the ownership of a non-fungible token
 * from one owner to another owner
 *
 * @param {Context} ctx the transaction context
 * @param {String} from The current owner of the non-fungible token
 * @param {String} to The new owner
 * @param {String} tokenId the non-fungible token to transfer
 * @returns {Boolean} Return whether the transfer was successful or not
 */
func (sc *ERC721Contract) TransferFrom(ctx contractapi.TransactionContextInterface, from string, to string, tokenId string) bool {
	var err error
	var sender string
	sender, err = ctx.GetClientIdentity().GetID()
	if err != nil {
		return false
	}
	var nft NFT
	nft, err = _readNFT(ctx, tokenId)

	// Check if the sender is the current owner, an authorized operator,
	// or the approved client for this non-fungible token.
	owner := nft.Owner
	tokenApproval := nft.Approved
	operatorApproval := IsApprovedForAll(ctx, owner, sender)
	if owner != sender && tokenApproval != sender && !operatorApproval {
		log.Printf(`The sender is not allowed to transfer the non-fungible token`)
		return false
	}

	// Check if `from` is the current owner
	if owner != from {
		log.Printf(`The from is not the current owner.`)
		return false
	}

	// Clear the approved client for this non-fungible token
	nft.Approved = ""

	// Overwrite a non-fungible token to assign a new owner.
	nft.Owner = to
	tokenIds := []string{tokenId}
	var nftKey string
	nftKey, err = ctx.GetStub().CreateCompositeKey(nftPrefix, tokenIds)

	var nftBytes []byte
	nftBytes, err = json.Marshal(nft)
	if err != nil {
		log.Printf("failed to json.Marshal: %v", err)
		return false
	}
	err = ctx.GetStub().PutState(nftKey, nftBytes)
	if err != nil {
		log.Printf("failed to PutState: %v", err)
		return false
	}

	// Remove a composite key from the balance of the current owner
	fromTokenIds := []string{from, tokenId}
	var balanceKeyFrom string
	balanceKeyFrom, err = ctx.GetStub().CreateCompositeKey(balancePrefix, fromTokenIds)
	if err != nil {
		log.Printf("failed to CreateCompositeKey: %v", err)
		return false
	}
	err = ctx.GetStub().DelState(balanceKeyFrom)
	if err != nil {
		log.Printf("failed to DelState: %v", err)
		return false
	}

	// Save a composite key to count the balance of a new owner
	toTokenIds := []string{to, tokenId}
	var balanceKeyTo string
	balanceKeyTo, err = ctx.GetStub().CreateCompositeKey(balancePrefix, toTokenIds)
	if err != nil {
		log.Printf("failed to balanceKeyTo CreateCompositeKey: %v", err)
		return false
	}
	stateByte := []byte(string('\u0000'))
	err = ctx.GetStub().PutState(balanceKeyTo, stateByte)
	if err != nil {
		log.Printf("failed to balanceKeyTo PutState: %v", err)
		return false
	}

	// Emit the Transfer event
	tokenIdInt, err := strconv.Atoi(tokenId)
	if err != nil {
		log.Printf("failed to tokenIdInt: %v", err)
		return false
	}
	transferEvent := TransferEvent{from, to, tokenIdInt}
	var transferEventBytes []byte
	transferEventBytes, _ = json.Marshal(transferEvent)
	err = ctx.GetStub().SetEvent("Transfer", transferEventBytes)
	if err != nil {
		log.Printf("failed to Transfer SetEvent: %v", err)
		return false
	}

	return true
}

/**
 * Approve changes or reaffirms the approved client for a non-fungible token
 *
 * @param {Context} ctx the transaction context
 * @param {String} approved The new approved client
 * @param {String} tokenId the non-fungible token to approve
 * @returns {Boolean} Return whether the approval was successful or not
 */
func (sc *ERC721Contract) Approve(ctx contractapi.TransactionContextInterface, approved string, tokenId string) bool {
	sender, _ := ctx.GetClientIdentity().GetID()
	nft, _ := _readNFT(ctx, tokenId)

	// Check if the sender is the current owner of the non-fungible token
	// or an authorized operator of the current owner
	owner := nft.Owner
	operatorApproval := IsApprovedForAll(ctx, owner, sender)
	if owner != sender && !operatorApproval {
		log.Printf(`The sender is not the current owner nor an authorized operator`)
		return false
	}

	// Update the approved client of the non-fungible token
	nft.Approved = approved
	tokenIds := []string{tokenId}
	nftKey, _ := ctx.GetStub().CreateCompositeKey(nftPrefix, tokenIds)
	var err error
	var nftBytes []byte
	nftBytes, err = json.Marshal(nft)
	if err != nil {
		log.Printf("failed to json.Marshal: %v", err)
		return false
	}
	err = ctx.GetStub().PutState(nftKey, nftBytes)
	if err != nil {
		return false
	}

	// Emit the Approval event
	tokenIdInt, err := strconv.Atoi(tokenId)
	if err != nil {
		log.Printf("failed to tokenIdInt: %v", err)
		return false
	}
	approvalEvent := ApprovalEvent{owner, approved, tokenIdInt}
	var approvalEventBytes []byte
	approvalEventBytes, _ = json.Marshal(approvalEvent)
	err = ctx.GetStub().SetEvent("Approval", approvalEventBytes)
	if err != nil {
		log.Printf("failed to GetStub().SetEvent(\"Approval\", approvalEventBytes): %v", err)
		return false
	}

	return true
}

/**
 * SetApprovalForAll enables or disables approval for a third party ("operator")
 * to manage all of message sender's assets
 *
 * @param {Context} ctx the transaction context
 * @param {String} operator A client to add to the set of authorized operators
 * @param {Boolean} approved True if the operator is approved, false to revoke approval
 * @returns {Boolean} Return whether the approval was successful or not
 */
func (sc *ERC721Contract) SetApprovalForAll(ctx contractapi.TransactionContextInterface, operator string, approved bool) bool {
	sender, _ := ctx.GetClientIdentity().GetID()

	approval := Approval{sender, operator, approved}
	approvalAttrs := []string{sender, operator}
	approvalKey, _ := ctx.GetStub().CreateCompositeKey(approvalPrefix, approvalAttrs)
	var approvalBytes []byte
	approvalBytes, _ = json.Marshal(approval)

	var err error
	err = ctx.GetStub().PutState(approvalKey, approvalBytes)
	if err != nil {
		log.Printf("failed to PutState in SetApprovalForAll(...): %v", err)
		return false
	}

	// Emit the ApprovalForAll event
	approvalForAllEvent := Approval{
		sender, operator, approved,
	}
	var approvalForAllEventBytes []byte
	approvalForAllEventBytes, _ = json.Marshal(approvalForAllEvent)
	err = ctx.GetStub().SetEvent("ApprovalForAll", approvalForAllEventBytes)
	if err != nil {
		log.Printf("failed to SetEvent in SetApprovalForAll(...): %v", err)
		return false
	}

	return true
}

/**
 * GetApproved returns the approved client for a single non-fungible token
 *
 * @param {Context} ctx the transaction context
 * @param {String} tokenId the non-fungible token to find the approved client for
 * @returns {Object} Return the approved client for this non-fungible token, or null if there is none
 */
func (sc *ERC721Contract) GetApproved(ctx contractapi.TransactionContextInterface, tokenId string) string {
	nft, _ := _readNFT(ctx, tokenId)
	return nft.Approved
}

/**
 * IsApprovedForAll returns if a client is an authorized operator for another client
 *
 * @param {Context} ctx the transaction context
 * @param {String} owner The client that owns the non-fungible tokens
 * @param {String} operator The client that acts on behalf of the owner
 * @returns {Boolean} Return true if the operator is an approved operator for the owner, false otherwise
 */
func IsApprovedForAll(ctx contractapi.TransactionContextInterface, owner string, operator string) bool {
	var err error
	var approvalKey string
	tokenIds := []string{owner, operator}
	approvalKey, err = ctx.GetStub().CreateCompositeKey(approvalPrefix, tokenIds)
	if err != nil {
		log.Printf("failed to get client id: %v", err)
		return false
	}
	var approvalBytes []byte
	approvalBytes, err = ctx.GetStub().GetState(approvalKey)

	var approved bool
	if approvalBytes != nil && len(approvalBytes) > 0 {
		var approval map[string]bool
		err := json.Unmarshal(approvalBytes, &approval)
		if err != nil {
			log.Printf("failed to json.Unmarshal: %v", &approval)
			return false
		}
		approved = approval["approved"]
	} else {
		approved = false
	}

	return approved
}

// ============== ERC721 metadata extension ===============

/**
 * Name returns a descriptive name for a collection of non-fungible tokens in this contract
 *
 * @param {Context} ctx the transaction context
 * @returns {String} Returns the name of the token
 */
func (sc *ERC721Contract) Name(ctx contractapi.TransactionContextInterface) string {
	nameAsBytes, err := ctx.GetStub().GetState(nameKey)
	if err != nil {
		log.Printf("failed to GetState: %s", nameKey)
		return ""
	}
	return string(nameAsBytes)
}

/**
 * Symbol returns an abbreviated name for non-fungible tokens in this contract.
 *
 * @param {Context} ctx the transaction context
 * @returns {String} Returns the symbol of the token
 */
func (sc *ERC721Contract) Symbol(ctx contractapi.TransactionContextInterface) string {
	symbolAsBytes, _ := ctx.GetStub().GetState(symbolKey)
	return string(symbolAsBytes)
}

/**
 * TokenURI returns a distinct Uniform Resource Identifier (URI) for a given token.
 *
 * @param {Context} ctx the transaction context
 * @param {string} tokenId The identifier for a non-fungible token
 * @returns {String} Returns the URI of the token
 */
func (sc *ERC721Contract) TokenURI(ctx contractapi.TransactionContextInterface, tokenId string) string {
	nft, _ := _readNFT(ctx, tokenId)
	return nft.TokenURI
}

// ============== ERC721 enumeration extension ===============

/**
 * TotalSupply counts non-fungible tokens tracked by this contract.
 *
 * @param {Context} ctx the transaction context
 * @returns {Number} Returns a count of valid non-fungible tokens tracked by this contract,
 * where each one of them has an assigned and queryable owner.
 */
func (sc *ERC721Contract) TotalSupply(ctx contractapi.TransactionContextInterface) int {
	// There is a key record for every non-fungible token in the format of nftPrefix.tokenId.
	// TotalSupply() queries for and counts all records matching nftPrefix.*
	var keys []string
	iterator, _ := ctx.GetStub().GetStateByPartialCompositeKey(nftPrefix, keys)

	// Count the number of returned composite keys
	totalSupply := 0
	result, _ := iterator.Next()
	for result != nil {
		totalSupply++
		result, _ = iterator.Next()
	}
	return totalSupply
}

// ============== Extended Functions for this sample ===============

/**
 * Set optional information for a token.
 *
 * @param {Context} ctx the transaction context
 * @param {String} name The name of the token
 * @param {String} symbol The symbol of the token
 */
func (sc *ERC721Contract) SetOption(ctx contractapi.TransactionContextInterface, name string, symbol string) bool {

	// Check minter authorization - this sample assumes Org1 is the issuer with privilege to set the name and symbol
	clientMSPID, _ := ctx.GetClientIdentity().GetMSPID()
	if clientMSPID != "Org1MSP" {
		log.Printf(`client is not authorized to set the name and symbol of the token`)
		return false
	}
	var err error
	err = ctx.GetStub().PutState(nameKey, []byte(name))
	if err != nil {
		log.Printf("failed to PutState(nameKey, []byte(name)) in SetOption: %s", nameKey)
		return false
	}
	err = ctx.GetStub().PutState(symbolKey, []byte(symbol))
	if err != nil {
		log.Printf("failed to PutState(symbolKey, []byte(symbol)) in SetOption: %s", symbolKey)
		return false
	}
	return true
}

/**
 * Mint a new non-fungible token
 *
 * @param {Context} ctx the transaction context
 * @param {String} tokenId Unique ID of the non-fungible token to be minted
 * @param {String} tokenURI URI containing metadata of the minted non-fungible token
 * @returns {Object} Return the non-fungible token object
 */
func (sc *ERC721Contract) MintWithTokenURI(ctx contractapi.TransactionContextInterface, tokenId string, tokenURI string) (NFT, error) {

	// Check minter authorization - this sample assumes Org1 is the issuer with privilege to mint a new token
	var err error
	clientMSPID, _ := ctx.GetClientIdentity().GetMSPID()
	if clientMSPID != "Org1MSP" {
		err = errors.New(`client is not authorized to mint new tokens`)
		log.Printf("failed to GetMSPID in MintWithTokenURI: %v", err)
		return NFT{}, err
	}

	// Get ID of submitting client identity
	minter, _ := ctx.GetClientIdentity().GetID()

	// Check if the token to be minted does not exist
	exists := _nftExists(ctx, tokenId)
	if exists {
		err = errors.New(`the token ${tokenId} is already minted`)
		log.Printf("failed to GetID in MintWithTokenURI: %v", err)
		return NFT{}, err
	}

	// Add a non-fungible token
	//tokenIdInt := parseInt(tokenId)
	//if isNaN(tokenIdInt) {
	//	throw
	//	new
	//	Error(`The tokenId ${tokenId} is invalid. tokenId must be an integer`)
	//}
	var tokenIdInt int
	tokenIdInt, err = strconv.Atoi(tokenId)
	if err != nil {
		err = errors.New(`the tokenId ${tokenId} is invalid. tokenId must be an integer`)
		log.Printf("failed to tokenIdInt: %v", err)
		return NFT{}, err
	}
	nft := NFT{tokenIdInt, minter, tokenURI, ""}
	attrs := []string{tokenId}
	nftKey, _ := ctx.GetStub().CreateCompositeKey(nftPrefix, attrs)

	var nftBytes []byte
	nftBytes, _ = json.Marshal(nft)
	err = ctx.GetStub().PutState(nftKey, nftBytes)
	if err != nil {
		log.Printf("failed to ctx.GetStub().PutState(nftKey, nftBytes): %v", err)
		return NFT{}, err
	}

	// A composite key would be balancePrefix.owner.tokenId, which enables partial
	// composite key query to find and count all records matching balance.owner.*
	// An empty value would represent a delete, so we simply insert the null character.
	balanceAttrs := []string{minter, tokenId}
	balanceKey, _ := ctx.GetStub().CreateCompositeKey(balancePrefix, balanceAttrs)

	err = ctx.GetStub().PutState(balanceKey, []byte(string('\u0000')))
	if err != nil {
		log.Printf("PutState(balanceKey, []byte(string('\\u0000'))): %v", err)
		return NFT{}, err
	}

	// Emit the Transfer event
	transferEvent := TransferEvent{"0x0", minter, tokenIdInt}
	var transferEventBytes []byte
	transferEventBytes, _ = json.Marshal(transferEvent)
	err = ctx.GetStub().SetEvent("Transfer", transferEventBytes)
	if err != nil {
		return NFT{}, err
	}

	return nft, nil
}

/**
 * Burn a non-fungible token
 *
 * @param {Context} ctx the transaction context
 * @param {String} tokenId Unique ID of a non-fungible token
 * @returns {Boolean} Return whether the burn was successful or not
 */
func (sc *ERC721Contract) Burn(ctx contractapi.TransactionContextInterface, tokenId string) bool {
	owner, _ := ctx.GetClientIdentity().GetID()

	// Check if a caller is the owner of the non-fungible token
	nft, _ := _readNFT(ctx, tokenId)
	if nft.Owner != owner {
		log.Printf(`Non-fungible token ${tokenId} is not owned by ${owner}`)
		return false
	}

	// Delete the token
	attrs := []string{tokenId}
	nftKey, _ := ctx.GetStub().CreateCompositeKey(nftPrefix, attrs)

	var err error
	err = ctx.GetStub().DelState(nftKey)
	if err != nil {
		return false
	}

	// Remove a composite key from the balance of the owner
	balanceAttrs := []string{owner, tokenId}
	balanceKey, _ := ctx.GetStub().CreateCompositeKey(balancePrefix, balanceAttrs)

	err = ctx.GetStub().DelState(balanceKey)
	if err != nil {
		return false
	}

	// Emit the Transfer event
	//const tokenIdInt = parseInt(tokenId)
	var tokenIdInt int
	tokenIdInt, err = strconv.Atoi(tokenId)
	transferEvent := TransferEvent{owner, "0x0", tokenIdInt}

	var transferEventBytes []byte
	transferEventBytes, _ = json.Marshal(transferEvent)
	err = ctx.GetStub().SetEvent("Transfer", transferEventBytes)
	if err != nil {
		return false
	}

	return true
}

func _readNFT(ctx contractapi.TransactionContextInterface, tokenId string) (NFT, error) {
	var err error
	var nftKey string
	tokenIds := []string{tokenId}
	nftKey, err = ctx.GetStub().CreateCompositeKey(nftPrefix, tokenIds)
	if err != nil {
		log.Printf("failed to get client id: %v", err)
		return NFT{}, err
	}
	var nftBytes []byte
	nftBytes, err = ctx.GetStub().GetState(nftKey)
	if nftBytes == nil || len(nftBytes) == 0 {
		return NFT{}, errors.New(`the tokenId ${tokenId} is invalid. It does not exist`)
	}
	var nft NFT
	err = json.Unmarshal(nftBytes, &nft)
	return nft, nil
}

func _nftExists(ctx contractapi.TransactionContextInterface, tokenId string) bool {
	var err error
	var nftKey string
	tokenIds := []string{tokenId}
	nftKey, err = ctx.GetStub().CreateCompositeKey(nftPrefix, tokenIds)
	if err != nil {
		log.Printf("failed to get client id: %v", err)
		return false
	}
	var nftBytes []byte
	nftBytes, err = ctx.GetStub().GetState(nftKey)
	if nftBytes == nil || len(nftBytes) == 0 {
		log.Printf(`The tokenId ${tokenId} is invalid. It does not exist`)
		return false
	}
	return true
}

/**
 * ClientAccountBalance returns the balance of the requesting client's account.
 *
 * @param {Context} ctx the transaction context
 * @returns {Number} Returns the account balance
 */
func (sc *ERC721Contract) ClientAccountBalance(ctx contractapi.TransactionContextInterface) int {
	// Get ID of submitting client identity
	clientAccountID, _ := ctx.GetClientIdentity().GetID()
	return BalanceOf(ctx, clientAccountID)
}

// ClientAccountID returns the id of the requesting client's account.
// In this implementation, the client account ID is the clientId itself.
// Users can use this function to get their own account id, which they can then give to others as the payment address
func (sc *ERC721Contract) ClientAccountID(ctx contractapi.TransactionContextInterface) string {
	// Get ID of submitting client identity
	clientAccountID, _ := ctx.GetClientIdentity().GetID()
	return clientAccountID
}
