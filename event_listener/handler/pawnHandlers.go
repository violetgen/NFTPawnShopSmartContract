package handler

import (
	"fmt"
	pawningShop "khanh/contracts"
	"khanh/httpClient"
	"log"
)

type PawnStatus int

const (
	CREATED PawnStatus = iota
	CANCELLED
	DEAL
	LIQUIDATED
	REPAID
)

func PawnCreated(pawnCreated *pawningShop.ContractsPawnCreated, instance *pawningShop.Contracts, client *httpClient.Client) {
	fmt.Println(PawnCreatedName)
	pawn, err := instance.Pawns(nil, pawnCreated.PawnId)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(pawn)
	success, resBody := client.Pawn.InsertOne(
		pawnCreated.PawnId.String(),
		pawn.Creator.String(),
		pawn.ContractAddress.String(),
		pawn.TokenId.String(),
		pawn.Status,
	)
	log.Println("to api", PawnCreatedName, success)
	if success {
		success, _ := client.Notify.SendNotification(httpClient.Notification{
			Message:  "New pawn is created",
			Code:     PawnCreatedName,
			PawnID:   pawnCreated.PawnId.String(),
			Borrower: pawn.Creator.String(),
			Payload:  string(resBody),
		})
		log.Println("to notify", PawnCreatedName, success)
	}
}

func PawnCancelled(pawn *pawningShop.ContractsPawnCancelled, instance *pawningShop.Contracts, client *httpClient.Client) {
	log.Println(PawnCancelledName)
	success, resBody := client.Pawn.UpdateOne(pawn.PawnId.String(), int(CANCELLED), "")
	log.Println("to api", PawnCancelledName, success)
	if success {
		success, _ := client.Notify.SendNotification(httpClient.Notification{
			Message:  "A pawn is cancelled",
			Code:     PawnCancelledName,
			PawnID:   pawn.PawnId.String(),
			Borrower: pawn.Borrower.String(),
			Payload:  string(resBody),
		})
		log.Println("to notify", PawnCancelledName, success)
	}
}

func PawnRepaid(pawn *pawningShop.ContractsPawnRepaid, client *httpClient.Client) {
	log.Println(PawnRepaidName)
	success, resBody := client.Pawn.UpdateOne(pawn.PawnId.String(), int(REPAID), "")
	log.Println("to api", PawnRepaidName, success)
	if success {
		success, _ := client.Notify.SendNotification(httpClient.Notification{
			Message:  "A pawn is repaid",
			Code:     PawnRepaidName,
			PawnID:   pawn.PawnId.String(),
			BidID:    pawn.BidId.String(),
			Lender:   pawn.Lender.String(),
			Borrower: pawn.Borrower.String(),
			Payload:  resBody,
		})
		log.Println("to notify", PawnRepaidName, success)
	}
}

func PawnLiquidated(pawn *pawningShop.ContractsPawnLiquidated, client *httpClient.Client) {
	log.Println(PawnLiquidatedName)
	success, resBody := client.Pawn.UpdateOne(pawn.PawnId.String(), int(LIQUIDATED), "")
	log.Println("to api", PawnLiquidatedName, success)
	if success {
		success, _ := client.Notify.SendNotification(httpClient.Notification{
			Message:  "A pawn is liquidated",
			Code:     PawnLiquidatedName,
			PawnID:   pawn.PawnId.String(),
			BidID:    pawn.BidId.String(),
			Lender:   pawn.Lender.String(),
			Borrower: pawn.Borrower.String(),
			Payload:  resBody,
		})
		log.Println("to notify", PawnLiquidatedName, success)
	}
}
