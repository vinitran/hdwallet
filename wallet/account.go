package wallet

import (
	"encoding/json"
	"fmt"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

var ChainToChainID = map[string]int{
	"ETH": 60,
	"TRX": 195,
}

type HdWallet struct {
	HdWallet *hdwallet.Wallet `json:"hd_wallet"`
}

type Wallet struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}

func (hdWallet *HdWallet) String() (string, error) {
	wallet, err := json.Marshal(hdWallet)
	if err != nil {
		return "", err
	}

	return string(wallet), nil
}

func (wl *Wallet) String() (string, error) {
	wallet, err := json.Marshal(wl)
	if err != nil {
		return "", err
	}

	return string(wallet), nil
}

func NewHdWallet() (*HdWallet, error) {
	entropy, err := hdwallet.NewEntropy(128)
	if err != nil {
		return nil, err
	}

	mnemonic, err := hdwallet.NewMnemonicFromEntropy(entropy)
	if err != nil {
		return nil, err
	}

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	hdWallet := HdWallet{
		HdWallet: wallet,
	}

	return &hdWallet, nil
}

func HdWalletByMnemonic(mnemonic string) (*HdWallet, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	hdWallet := HdWallet{
		HdWallet: wallet,
	}

	return &hdWallet, nil
}

func (hdWallet *HdWallet) WalletByDerivationPath(derivationPath string) (*Wallet, error) {
	path := hdwallet.MustParseDerivationPath(derivationPath)
	account, err := hdWallet.HdWallet.Derive(path, false)
	if err != nil {
		return nil, err
	}

	privateKey, err := hdWallet.HdWallet.PrivateKeyHex(account)
	if err != nil {
		return nil, err
	}

	publicKey, err := hdWallet.HdWallet.PublicKeyHex(account)
	if err != nil {
		return nil, err
	}

	address, err := hdWallet.HdWallet.AddressHex(account)
	if err != nil {
		return nil, err
	}

	wl := Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}
	return &wl, nil
}

func (hdWallet *HdWallet) Wallet(chainSymbol string, account, id int) (*Wallet, error) {
	derivationPath := fmt.Sprintf("m/44'/%d'/%d'/0/%d", ChainToChainID[chainSymbol], account, id)
	wl, err := hdWallet.WalletByDerivationPath(derivationPath)
	if err != nil {
		return nil, err
	}
	return wl, nil
}

func (hdWallet *HdWallet) PrivateKey(chainSymbol string, account, id int) (string, error) {
	wallet, err := hdWallet.Wallet(chainSymbol, account, id)
	if err != nil {
		return "", err
	}

	return wallet.PrivateKey, nil
}

func (hdWallet *HdWallet) Address(chainSymbol string, account, id int) (string, error) {
	wallet, err := hdWallet.Wallet(chainSymbol, account, id)
	if err != nil {
		return "", err
	}
	return wallet.Address, nil
}
