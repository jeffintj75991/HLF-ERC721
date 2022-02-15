package main

import (
	"fmt"

	"encoding/json"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/common/flogging"
)

type ERC721TokenContract struct {
	contractapi.Contract
}

type ERC20Token struct {
	Symbol      string `json:"symbol"`
	TokenId     string `json:"tokenId"`
	Description string `json:"description"`
	Creator     string `json:"creator"`
}

var logger = flogging.MustGetLogger("erc20_cc")

const OwnerPrefix = "owner."

func (token *ERC721TokenContract) Mint(ctx contractapi.TransactionContextInterface) string {

	logger.Info("function:Mint")
	stub := ctx.GetStub()

	_, args := stub.GetFunctionAndParameters()

	if len(args) < 4 {
		return "Failed - incorrect number of parameters!! "
	}
	symbol := string(args[0])

	tokenId := string(args[1])

	if len(args[3]) == 0 {
		return errorResponse("Creator identity cannot be 0 length!!!", 3)
	}
	creator := string(args[3])

	var erc20 = ERC20Token{Symbol: symbol, TokenId: tokenId, Description: string(args[2]), Creator: creator}

	jsonERC20, _ := json.Marshal(erc20)
	ctx.GetStub().PutState("token", []byte(jsonERC20))

	key := OwnerPrefix + creator
	fmt.Println("Key=", key)
	err := ctx.GetStub().PutState(key, []byte(args[1]))

	if err != nil {
		return errorResponse(err.Error(), 4)
	}
	fmt.Println("func:TokenCreation END")
	return string([]byte(jsonERC20))
}

func (token *ERC721TokenContract) BalanceOf(ctx contractapi.TransactionContextInterface, OwnerIden string) string {

	logger.Info("function:BalanceOf")
	if len(OwnerIden) < 1 {
		return errorResponse("Needs OwnerID!!!", 6)
	}
	OwnerID := OwnerIden
	bytes, err := ctx.GetStub().GetState(OwnerPrefix + OwnerID)
	if err != nil {
		return errorResponse(err.Error(), 7)
	}

	response := balanceJSON(OwnerID, string(bytes))

	return successResponse("balance:" + response)
}

func (token *ERC721TokenContract) TransferFrom(ctx contractapi.TransactionContextInterface) string {

	stub := ctx.GetStub()

	_, args := stub.GetFunctionAndParameters()
	if len(args) < 3 {
		return errorResponse("Needs to, from & amount!!!", 700)
	}

	from := string(args[0])
	to := string(args[1])
	amount, err := strconv.Atoi(string(args[2]))

	fmt.Println("from", from, "to", to, "amount", amount)
	if err != nil {
		return errorResponse(err.Error(), 701)
	}
	if amount <= 0 {
		return errorResponse("Amount MUST be > 0!!!", 702)
	}

	bytes, _ := ctx.GetStub().GetState(OwnerPrefix + from)
	if len(bytes) == 0 {

		return errorResponse("Balance MUST be > 0!!!", 703)
	}
	fromBalance, _ := strconv.Atoi(string(bytes))
	if fromBalance < amount {
		return errorResponse("Insufficient balance to cover transfer!!!", 704)
	}

	fromBalance = fromBalance - amount

	bytes, _ = ctx.GetStub().GetState(OwnerPrefix + to)
	toBalance := 0
	if len(bytes) > 0 {
		toBalance, _ = strconv.Atoi(string(bytes))
	}
	toBalance += amount

	bytes = []byte(strconv.FormatInt(int64(fromBalance), 10))
	err = ctx.GetStub().PutState(OwnerPrefix+from, bytes)

	bytes = []byte(strconv.FormatInt(int64(toBalance), 10))
	err = ctx.GetStub().PutState(OwnerPrefix+to, bytes)
	stub = ctx.GetStub()

	eventPayload := "{\"from\":\"" + from + "\", \"to\":\"" + to + "\",\"amount\":" + strconv.FormatInt(int64(amount), 10) + "}"
	stub.SetEvent("transfer", []byte(eventPayload))
	return successResponse("Transfer Successful!!!")
}

func (token *ERC721TokenContract) SafeTransferFrom(ctx contractapi.TransactionContextInterface) string {

	stub := ctx.GetStub()

	_, args := stub.GetFunctionAndParameters()
	if len(args) < 3 {
		return errorResponse("Needs to, from & tokenId!!!", 700)
	}

	from := string(args[0])
	to := string(args[1])
	tokenId := string(args[2])

	fmt.Println("from", from, "to", to, "tokenId", tokenId)

	bytes, _ := ctx.GetStub().GetState(OwnerPrefix + from)

	bytes, _ = ctx.GetStub().GetState(OwnerPrefix + to)

	bytes = []byte(tokenId)
	err := ctx.GetStub().PutState(OwnerPrefix+from, bytes)

	if err != nil {
		return errorResponse(err.Error(), 705)
	}

	bytes = []byte(tokenId)
	err = ctx.GetStub().PutState(OwnerPrefix+to, bytes)
	stub = ctx.GetStub()

	eventPayload := "{\"from\":\"" + from + "\", \"to\":\"" + to + "\",\"tokenId\":" + tokenId + "}"
	stub.SetEvent("transfer", []byte(eventPayload))
	return successResponse("Transfer Successful!!!")
}

func balanceJSON(OwnerID, balance string) string {
	return "{\"owner\":\"" + OwnerID + "\", \"balance\":" + balance + "}"
}

func errorResponse(err string, code uint) string {
	codeStr := strconv.FormatUint(uint64(code), 10)
	// errorString := "{\"error\": \"" + err +"\", \"code\":"+codeStr+" \" }"
	errorString := "{\"error\":" + err + ", \"code\":" + codeStr + " \" }"
	return errorString
}

func successResponse(dat string) string {
	success := "{\"response\": " + dat + ", \"code\": 0 }"
	return string([]byte(success))
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(ERC721TokenContract))
	if err != nil {
		fmt.Printf("Error create ERC721TokenContract chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincodes: %s", err.Error())
	}
}
