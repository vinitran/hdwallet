package api

import (
	"ExchangeManager/tokens"
	"ExchangeManager/transfer"
	"ExchangeManager/wallet"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"os"
)

type Contract struct {
	Token *tokens.Token `json:"token"`
}

var feeDefault = "50000000000000000" //0.05 nativetoken

func NewToken(rpc, contract string) (*Contract, error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, err
	}

	instance, err := tokens.NewToken(common.HexToAddress(contract), client)
	if err != nil {
		return nil, err
	}

	return &Contract{Token: instance}, nil
}

func GetUserAddress(id int, chainSymbol string) (string, error) {
	hdWallet, err := wallet.HdWalletByMnemonic(os.Getenv("MNEMONIC"))
	if err != nil {
		return "", err
	}

	address, err := hdWallet.Address(chainSymbol, 0, id)
	if err != nil {
		return "", err
	}

	return address, nil
}

func GetUserWallet(id int, chainSymbol string) (*wallet.Wallet, error) {
	hdWallet, err := wallet.HdWalletByMnemonic(os.Getenv("MNEMONIC"))
	if err != nil {
		return nil, err
	}

	wl, err := hdWallet.Wallet(chainSymbol, 0, id)
	if err != nil {
		return nil, err
	}

	return wl, nil
}

func SendAllNativeTokenToManager(id int, chainSymbol string) error {
	wl, err := GetUserWallet(id, chainSymbol)
	if err != nil {
		return err
	}
	fmt.Println(1)

	nativeAmountStr, err := GetNativeTokenAmount(os.Getenv("RPC"), wl.Address)
	if err != nil {
		return err
	}
	fmt.Println(12)

	nativeAmount, ok := new(big.Int).SetString(nativeAmountStr, 10)
	if !ok {
		return fmt.Errorf("error: setstring")
	}
	fmt.Println(13)

	feeAmount, ok := new(big.Int).SetString(feeDefault, 10)
	if !ok {
		return fmt.Errorf("error: setstring")
	}
	fmt.Println(14)

	if nativeAmount.Cmp(feeAmount) != 1 {
		return nil
	}
	fmt.Println(15)

	err = transfer.NativeTokenViaUserId(id, nativeAmount.Sub(nativeAmount, feeAmount))
	fmt.Println("Send all native token")
	return err
}

func SendAllErc20TokenToManager(id int, token *Contract, contract, chainSymbol string) error {
	wl, err := GetUserWallet(id, chainSymbol)
	if err != nil {
		return err
	}

	tokenAmountStr, err := token.Amount(wl.Address)

	tokenAmount, ok := new(big.Int).SetString(tokenAmountStr, 10)
	if !ok {
		return fmt.Errorf("error: setstring")
	}

	feeAmount, ok := new(big.Int).SetString(feeDefault, 10)
	if !ok {
		return fmt.Errorf("error: setstring")
	}

	if tokenAmount.Cmp(feeAmount) != 1 {
		return nil
	}

	err = transfer.Erc20TokenViaUserId(id, contract, tokenAmount.Sub(tokenAmount, feeAmount))
	fmt.Println("Send all erc20 token")
	return err
}

//
//func GetUserPrivateKey(id int, chainSymbol string) (string, error) {
//	hdWallet, err := wallet.HdWalletByMnemonic(os.Getenv("MNEMONIC"))
//	if err != nil {
//		return "", err
//	}
//
//	privateKey, err := hdWallet.PrivateKey(chainSymbol, 0, id)
//	if err != nil {
//		return "", err
//	}
//	return privateKey, nil
//}

func (token *Contract) Amount(address string) (string, error) {
	balance, err := token.Token.BalanceOf(&bind.CallOpts{}, common.HexToAddress(address))
	if err != nil {
		return "", err
	}

	return balance.String(), nil
}

func (token *Contract) Symbol() (string, error) {
	symbol, err := token.Token.Symbol(&bind.CallOpts{})
	if err != nil {
		return "", err
	}

	return symbol, nil
}

func GetNativeTokenAmount(rpc, address string) (string, error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return "", err
	}

	account := common.HexToAddress(address)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return "", err
	}

	return balance.String(), nil
}

// CompareAmount compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func CompareAmount(amount1Str, amount2Str string) (int, error) {
	amount1, ok := new(big.Int).SetString(amount1Str, 10)
	if !ok {
		return -2, fmt.Errorf("SetString: error")
	}

	amount2, ok := new(big.Int).SetString(amount2Str, 10)
	if !ok {
		return -2, fmt.Errorf("SetString: error")
	}

	return amount1.Cmp(amount2), nil
}
