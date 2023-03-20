package blockchain

import (
	"ExchangeManager/tokens"
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (tracking *TrackingTx) NativeTokenAmount(address string) (string, error) {
	account := common.HexToAddress(address)
	balance, err := tracking.Client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return "", err
	}

	return balance.String(), nil
}

func (tracking *TrackingTx) TokenAmount(address, contract string) (string, error) {
	instance, err := tokens.NewToken(common.HexToAddress(contract), tracking.Client)
	if err != nil {
		return "", err
	}

	balance, err := instance.BalanceOf(&bind.CallOpts{}, common.HexToAddress(address))
	if err != nil {
		return "", err
	}

	return balance.String(), nil
}

func (tracking *TrackingTx) TokenSymbol(contract common.Address) (string, error) {
	instance, err := tokens.NewToken(contract, tracking.Client)
	if err != nil {
		return "", err
	}

	symbol, err := instance.Symbol(&bind.CallOpts{})
	if err != nil {
		return "", err
	}

	return symbol, nil
}
