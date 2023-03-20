package main

import (
	"ExchangeManager/blockchain"
	"ExchangeManager/databases"
	"ExchangeManager/transfer"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	"log"
	"os"
	"testing"
	"time"
)

func TestBlockchain(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Error(err)
	}

	db, err := databases.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}

	txs := make(chan blockchain.Transaction)
	tracking, err := blockchain.NewTxTracking(os.Getenv("RPC"))
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		contractToken := os.Getenv("TOKEN_CONTRACT")
		tracking.GetTransaction(txs, []common.Address{common.HexToAddress(contractToken)})
	}()

	for {
		select {
		case msg := <-txs:
			err = db.IncreaseAmount(msg.RawAmount, msg.To, msg.Contract)
			if err != nil {
				fmt.Println(err)
				continue
			}

			var userId int
			_, err = db.Db.NewSelect().Model(new(databases.Wallet)).
				ColumnExpr("id").
				Where("address = ?", msg.To).
				Where("contract = ?", msg.Contract).
				Exec(context.Background(), &userId)
			if err != nil {
				fmt.Println(err)
				continue
			}

			err = transfer.AllTokenViaUserId(userId, msg.Contract, msg.RawAmount)
			if err != nil {
				fmt.Println(err)
				continue
			}
		default:
			time.Sleep(time.Second)
		}
	}
}
