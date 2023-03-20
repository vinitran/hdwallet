package wallet

import (
	"testing"
)

const mnemonic = "credit prosper brown exotic remove acid truck depend muffin traffic random airport"

var (
	testChainSymbol = "ETH"
	testAccount     = 0
	testIndex       = 0
	testPrivateKey  = "de2f32a7dcc5530cd39654250aecb1a2ab4fe0bb5bb7f95f6fe6bbfd1b1c8368"
	testPublicKey   = "3c4a18507a592aaa49e68307e19c8d423e5aa9f1777f1bf3e648a9447d785b11062b0ece98b337e7292292f4a9438236bace0590d5252ea62f61e49028461954"
	testAddress     = "0xD1C80e25CDb409b3F3cB9340a8e35f511A7EbE1F"
)

func TestGetWallet(t *testing.T) {
	hdWallet, err := HdWalletByMnemonic(mnemonic)
	if err != nil {
		t.Error(err)
	}

	wallet, err := hdWallet.Wallet(testChainSymbol, testAccount, testIndex)
	if err != nil {
		t.Error(err)
	}

	if wallet.PrivateKey != testPrivateKey {
		t.Errorf("error: private does not match with test")
	}

	if wallet.Address != testAddress {
		t.Errorf("error: address does not match with test")
	}

	if wallet.PublicKey != testPublicKey {
		t.Errorf("error: publicKey does not match with test")
	}
}

func TestGetAddress(t *testing.T) {
	hdWallet, err := HdWalletByMnemonic(mnemonic)
	if err != nil {
		t.Error(err)
	}
	address, err := hdWallet.Address(testChainSymbol, testAccount, testIndex)
	if err != nil {
		t.Error(err)
	}

	if address != testAddress {
		t.Errorf("error: address dont match with test")
	}
}
