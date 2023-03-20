package main

import (
	"ExchangeManager/transfer"
	"math/big"
	"testing"
)

func TestTransfer(t *testing.T) {
	//privateKey := "c14857a87233fb532b34bf4287894fc066d5a6b206f5fd82ca95e55d1349b094"
	privateKey := "de2f32a7dcc5530cd39654250aecb1a2ab4fe0bb5bb7f95f6fe6bbfd1b1c8368"
	tokenContract := "0x0C719E9E019C68BbE925E7DFc72B5f39c34b91Dd"
	amount := big.NewInt(1000000000000000000)
	to := "0x3979649A15f1Da3965d67BBC082194343737Ad5b"
	//transfer.TransferNativeToken()
	err := transfer.ERC20Token(privateKey, to, tokenContract, amount)
	if err != nil {
		t.Error(err)
	}
}
