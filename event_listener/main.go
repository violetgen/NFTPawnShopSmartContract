package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"khanh/config"
	pawningShop "khanh/contracts"
	"khanh/handler"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	env, err := config.LoadEnv()
	if err != nil {
		log.Panic(err)
	}
	client, err := ethclient.Dial(env.NETWORK_ADDRESS)
	if err != nil {
		log.Fatal(err)
	}
	contractAddress := common.HexToAddress(env.CONTRACT_ADDRESS)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}

	contractAbi, err := abi.JSON(strings.NewReader(string(pawningShop.ContractsABI)))
	if err != nil {
		log.Panic(err)
	}

	instance, _ := pawningShop.NewContracts(common.HexToAddress(env.CONTRACT_ADDRESS), client)
	if err != nil {
		log.Panic(err)
	}

	pawnRepaidChannel := make(chan *pawningShop.ContractsPawnRepaid)
	pawnRepayChannelErr, err := instance.WatchPawnRepaid(nil, pawnRepaidChannel)
	if err != nil {
		log.Panic(err)
	}

	pawnLiquidatedChannel := make(chan *pawningShop.ContractsPawnLiquidated)
	pawnLiquidatedChannelErr, err := instance.WatchPawnLiquidated(nil, pawnLiquidatedChannel)
	if err != nil {
		log.Panic(err)
	}

	log.Println("started to listen to ", env.CONTRACT_ADDRESS)

	for {
		select {
		case err := <-sub.Err():
			log.Panic(err)
		case vLog := <-logs:
			CategorizeEvent(vLog, contractAbi, instance, env)
		case err := <-pawnRepayChannelErr.Err():
			log.Panic(err)
		case repay := <-pawnRepaidChannel:
			handler.PawnRepaid(repay.PawnId.String(), env)
		case err := <-pawnLiquidatedChannelErr.Err():
			log.Panic(err)
		case liquidated := <-pawnLiquidatedChannel:
			handler.PawnLiquidated(liquidated.PawnId.String(), env)
		}
	}

}

func CategorizeEvent(log types.Log, abi abi.ABI, instance *pawningShop.Contracts, env *config.Env) {
	incommingEventHash := log.Topics[0]
	fmt.Println("incomming event hash: ", incommingEventHash)
	switch incommingEventHash {
	case Hash(handler.PawnCreatedSignature):
		handler.PawnCreated(log, abi, instance, env)
	case Hash(handler.PawnCancelledSignature):
		handler.PawnCancelled(log, abi, instance, env)
	case Hash(handler.WhiteListAddedSignature):
		handler.WhiteListAdded(log, abi, instance)
	case Hash(handler.WhiteListRemovedSignature):
		handler.WhiteListRemoved(log, abi, instance)
	case Hash(handler.BidCreatedNameSignature):
		handler.BidCreated(log, abi, instance, env)
	case Hash(handler.BidCancelledNameSignature):
		handler.BidCancelled(log, abi, instance, env)
	case Hash(handler.BidAcceptedNameSignature):
		handler.BidAccepted(log, abi, instance, env)
	}
}

func Hash(signature string) common.Hash {
	return crypto.Keccak256Hash([]byte(signature))
}
