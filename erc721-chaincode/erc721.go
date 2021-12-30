package main

import (
	"encoding/json"
	"errors"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
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
}

/*
 * BalanceOf counts all non-fungible tokens assigned to an owner
 *
 * @param {Context} ctx the transaction context
 * @param {String} owner An owner for whom to query the balance
 * @returns {Number} The number of non-fungible tokens owned by the owner, possibly zero
 */
func (sc *ERC721Contract) BalanceOf(ctx contractapi.TransactionContextInterface, owner string) int {
	// There is a key record for every non-fungible token in the format of balancePrefix.owner.tokenId.
	// BalanceOf() queries for and counts all records matching balancePrefix.owner.*
	keys := []string{owner}
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey(balancePrefix, keys)
	if err != nil {
		log.Printf("BalanceOf GetStateByPartialCompositeKey balancePrefix key: %s is err", owner)
	}

	// Count the number of returned composite keys
	var balance = 0
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
func (sc *ERC721Contract) OwnerOf(ctx contractapi.TransactionContextInterface, tokenId string) string {
	const nft = await
	this._readNFT(ctx, tokenId)
	const owner = nft.owner
	if !owner {
		throw
		new
		Error('No owner is assigned to this token')
	}

	return owner
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
	const sender = ctx.clientIdentity.getID()

	const nft = await
	this._readNFT(ctx, tokenId)

	// Check if the sender is the current owner, an authorized operator,
	// or the approved client for this non-fungible token.
	const owner = nft.owner
	const tokenApproval = nft.approved
	const operatorApproval = await
	this.IsApprovedForAll(ctx, owner, sender)
	if (owner != = sender && tokenApproval != = sender && !operatorApproval) {
		throw
		new
		Error('The sender is not allowed to transfer the non-fungible token')
	}

	// Check if `from` is the current owner
	if (owner != = from) {
		throw
		new
		Error('The from is not the current owner.')
	}

	// Clear the approved client for this non-fungible token
	nft.approved = ''

	// Overwrite a non-fungible token to assign a new owner.
	nft.owner = to
	const nftKey = ctx.stub.createCompositeKey(nftPrefix,[tokenId])
	await
	ctx.stub.putState(nftKey, Buffer.from(JSON.stringify(nft)))

	// Remove a composite key from the balance of the current owner
	const balanceKeyFrom = ctx.stub.createCompositeKey(balancePrefix,[from, tokenId])
	await
	ctx.stub.deleteState(balanceKeyFrom)

	// Save a composite key to count the balance of a new owner
	const balanceKeyTo = ctx.stub.createCompositeKey(balancePrefix,[to, tokenId])
	await
	ctx.stub.putState(balanceKeyTo, Buffer.from('\u0000'))

	// Emit the Transfer event
	const tokenIdInt = parseInt(tokenId)
	const transferEvent =
	{
	from:
		from, to: to, tokenId: tokenIdInt
	}
	ctx.stub.setEvent('Transfer', Buffer.from(JSON.stringify(transferEvent)))

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
func (sc *ERC721Contract) Approve(ctx, approved, tokenId) {
	const sender = ctx.clientIdentity.getID()

	const nft = await
	this._readNFT(ctx, tokenId)

	// Check if the sender is the current owner of the non-fungible token
	// or an authorized operator of the current owner
	const owner = nft.owner
	const operatorApproval = await
	this.IsApprovedForAll(ctx, owner, sender)
	if (owner != = sender && !operatorApproval) {
		throw
		new
		Error('The sender is not the current owner nor an authorized operator')
	}

	// Update the approved client of the non-fungible token
	nft.approved = approved
	const nftKey = ctx.stub.createCompositeKey(nftPrefix,[tokenId])
	await
	ctx.stub.putState(nftKey, Buffer.from(JSON.stringify(nft)))

	// Emit the Approval event
	const tokenIdInt = parseInt(tokenId)
	const approvalEvent =
	{
	owner:
		owner, approved: approved, tokenId: tokenIdInt
	}
	ctx.stub.setEvent('Approval', Buffer.from(JSON.stringify(approvalEvent)))

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
func (sc *ERC721Contract) SetApprovalForAll(ctx, operator, approved) {
	const sender = ctx.clientIdentity.getID()

	const approval =
	{
	owner:
		sender, operator: operator, approved: approved
	}
	const approvalKey = ctx.stub.createCompositeKey(approvalPrefix,[sender, operator])
	await
	ctx.stub.putState(approvalKey, Buffer.from(JSON.stringify(approval)))

	// Emit the ApprovalForAll event
	const approvalForAllEvent =
	{
	owner:
		sender, operator: operator, approved: approved
	}
	ctx.stub.setEvent('ApprovalForAll', Buffer.from(JSON.stringify(approvalForAllEvent)))

	return true
}

/**
 * GetApproved returns the approved client for a single non-fungible token
 *
 * @param {Context} ctx the transaction context
 * @param {String} tokenId the non-fungible token to find the approved client for
 * @returns {Object} Return the approved client for this non-fungible token, or null if there is none
 */
func (sc *ERC721Contract) GetApproved(ctx, tokenId) {
	const nft = await
	this._readNFT(ctx, tokenId)
	return nft.approved
}

/**
 * IsApprovedForAll returns if a client is an authorized operator for another client
 *
 * @param {Context} ctx the transaction context
 * @param {String} owner The client that owns the non-fungible tokens
 * @param {String} operator The client that acts on behalf of the owner
 * @returns {Boolean} Return true if the operator is an approved operator for the owner, false otherwise
 */
func (sc *ERC721Contract) IsApprovedForAll(ctx, owner, operator) {
	const approvalKey = ctx.stub.createCompositeKey(approvalPrefix,[owner, operator])
	const approvalBytes = await
	ctx.stub.getState(approvalKey)
	let
	approved
	if approvalBytes && approvalBytes.length > 0 {
		const approval = JSON.parse(approvalBytes.toString())
		approved = approval.approved
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
func (sc *ERC721Contract) Name(ctx) {
	const nameAsBytes = await
	ctx.stub.getState(nameKey)
	return nameAsBytes.toString()
}

/**
 * Symbol returns an abbreviated name for non-fungible tokens in this contract.
 *
 * @param {Context} ctx the transaction context
 * @returns {String} Returns the symbol of the token
 */
func (sc *ERC721Contract) Symbol(ctx) {
	const symbolAsBytes = await
	ctx.stub.getState(symbolKey)
	return symbolAsBytes.toString()
}

/**
 * TokenURI returns a distinct Uniform Resource Identifier (URI) for a given token.
 *
 * @param {Context} ctx the transaction context
 * @param {string} tokenId The identifier for a non-fungible token
 * @returns {String} Returns the URI of the token
 */
func (sc *ERC721Contract) TokenURI(ctx, tokenId) {
	const nft = await
	this._readNFT(ctx, tokenId)
	return nft.tokenURI
}

// ============== ERC721 enumeration extension ===============

/**
 * TotalSupply counts non-fungible tokens tracked by this contract.
 *
 * @param {Context} ctx the transaction context
 * @returns {Number} Returns a count of valid non-fungible tokens tracked by this contract,
 * where each one of them has an assigned and queryable owner.
 */
func (sc *ERC721Contract) TotalSupply(ctx) {
	// There is a key record for every non-fungible token in the format of nftPrefix.tokenId.
	// TotalSupply() queries for and counts all records matching nftPrefix.*
	const iterator = await
	ctx.stub.getStateByPartialCompositeKey(nftPrefix,[])

	// Count the number of returned composite keys
	let
	totalSupply = 0
	let
	result = await
	iterator.next()
	while(!result.done)
	{
		totalSupply++
		result = await
		iterator.next()
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
func (sc *ERC721Contract) SetOption(ctx, name, symbol) {

	// Check minter authorization - this sample assumes Org1 is the issuer with privilege to set the name and symbol
	const clientMSPID = ctx.clientIdentity.getMSPID()
	if (clientMSPID != = 'Org1MSP') {
		throw
		new
		Error('client is not authorized to set the name and symbol of the token')
	}

	await
	ctx.stub.putState(nameKey, Buffer.from(name))
	await
	ctx.stub.putState(symbolKey, Buffer.from(symbol))
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
func (sc *ERC721Contract) MintWithTokenURI(ctx, tokenId, tokenURI) {

	// Check minter authorization - this sample assumes Org1 is the issuer with privilege to mint a new token
	const clientMSPID = ctx.clientIdentity.getMSPID()
	if (clientMSPID != = 'Org1MSP') {
		throw
		new
		Error('client is not authorized to mint new tokens')
	}

	// Get ID of submitting client identity
	const minter = ctx.clientIdentity.getID()

	// Check if the token to be minted does not exist
	const exists = await
	this._nftExists(ctx, tokenId)
	if exists {
		throw
		new
		Error(`The token ${tokenId} is already minted.`)
	}

	// Add a non-fungible token
	const tokenIdInt = parseInt(tokenId)
	if isNaN(tokenIdInt) {
		throw
		new
		Error(`The tokenId ${tokenId} is invalid. tokenId must be an integer`)
	}
	const nft =
	{
	tokenId:
		tokenIdInt,
			owner: minter,
		tokenURI: tokenURI
	}
	const nftKey = ctx.stub.createCompositeKey(nftPrefix,[tokenId])
	await
	ctx.stub.putState(nftKey, Buffer.from(JSON.stringify(nft)))

	// A composite key would be balancePrefix.owner.tokenId, which enables partial
	// composite key query to find and count all records matching balance.owner.*
	// An empty value would represent a delete, so we simply insert the null character.
	const balanceKey = ctx.stub.createCompositeKey(balancePrefix,[minter, tokenId])
	await
	ctx.stub.putState(balanceKey, Buffer.from('\u0000'))

	// Emit the Transfer event
	const transferEvent =
	{
	from:
		'0x0', to: minter, tokenId: tokenIdInt
	}
	ctx.stub.setEvent('Transfer', Buffer.from(JSON.stringify(transferEvent)))

	return nft
}

/**
 * Burn a non-fungible token
 *
 * @param {Context} ctx the transaction context
 * @param {String} tokenId Unique ID of a non-fungible token
 * @returns {Boolean} Return whether the burn was successful or not
 */
func (sc *ERC721Contract) Burn(ctx, tokenId) {
	const owner = ctx.clientIdentity.getID()

	// Check if a caller is the owner of the non-fungible token
	const nft = await
	this._readNFT(ctx, tokenId)
	if (nft.owner != = owner) {
		throw
		new
		Error(`Non-fungible token ${tokenId} is not owned by ${owner}`)
	}

	// Delete the token
	const nftKey = ctx.stub.createCompositeKey(nftPrefix,[tokenId])
	await
	ctx.stub.deleteState(nftKey)

	// Remove a composite key from the balance of the owner
	const balanceKey = ctx.stub.createCompositeKey(balancePrefix,[owner, tokenId])
	await
	ctx.stub.deleteState(balanceKey)

	// Emit the Transfer event
	const tokenIdInt = parseInt(tokenId)
	const transferEvent =
	{
	from:
		owner, to: '0x0', tokenId: tokenIdInt
	}
	ctx.stub.setEvent('Transfer', Buffer.from(JSON.stringify(transferEvent)))

	return true
}

func (sc *ERC721Contract) _readNFT(ctx contractapi.TransactionContextInterface, tokenId string) (NFT, error) {
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
		return NFT{}, errors.New(`The tokenId ${tokenId} is invalid. It does not exist`)
	}
	var nft NFT
	err = json.Unmarshal(nftBytes, &nft)
	return nft, nil
}

func (sc *ERC721Contract) _nftExists(ctx, tokenId) {
	const nftKey = ctx.stub.createCompositeKey(nftPrefix,[tokenId])
	const nftBytes = await
	ctx.stub.getState(nftKey)
	return nftBytes && nftBytes.length > 0
}

/**
 * ClientAccountBalance returns the balance of the requesting client's account.
 *
 * @param {Context} ctx the transaction context
 * @returns {Number} Returns the account balance
 */
func (sc *ERC721Contract) ClientAccountBalance(ctx) {
	// Get ID of submitting client identity
	const clientAccountID = ctx.clientIdentity.getID()
	return this.BalanceOf(ctx, clientAccountID)
}

// ClientAccountID returns the id of the requesting client's account.
// In this implementation, the client account ID is the clientId itself.
// Users can use this function to get their own account id, which they can then give to others as the payment address
func (sc *ERC721Contract) ClientAccountID(ctx) {
	// Get ID of submitting client identity
	const clientAccountID = ctx.clientIdentity.getID()
	return clientAccountID
}
