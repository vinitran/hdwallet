package blockchain

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

const (
	RPC = "https://bsc-mainnet.nodereal.io/v1/3ac635da99d7432abf96f266693e38e4"
)

var address = "0x1111111254eeb25477b68fb85ed929f73a960582"
var contract = "0xe9e7cea3dedca5984780bafc599bd69add087d56"

func TestGetTransaction(t *testing.T) {
	var txs chan []Transaction
	tracking, err := NewTxTracking(RPC)
	if err != nil {
		t.Error(err)
	}
	tracking.GetTransaction(txs, []common.Address{common.HexToAddress("0xe9e7cea3dedca5984780bafc599bd69add087d56")})

	for {
		select {
		case msg := <-txs:
			fmt.Println(msg)
		}
	}
}
