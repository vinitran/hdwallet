package transfer

import (
	"ExchangeManager/tokens"
	"ExchangeManager/wallet"
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"os"
)

var addressZero = "0x0000000000000000000000000000000000000000"

func NewClient() *ethclient.Client {
	client, err := ethclient.Dial(os.Getenv("RPC_TESTNET"))
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func GetManagerPrivateKey() (string, error) {
	hdWallet, err := wallet.HdWalletByMnemonic(os.Getenv("MNEMONIC"))
	if err != nil {
		return "", err
	}

	privateKey, err := hdWallet.PrivateKey(os.Getenv("CHAIN_SYMBOL"), 1, 0)
	if err != nil {
		return "", err
	}

	return privateKey, nil
}

func Token(privateKey, toAddress, contract, amount string) error {
	value, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return fmt.Errorf("SetString: error")
	}

	if contract == addressZero {
		err := NativeToken(privateKey, toAddress, value)
		if err != nil {
			return err
		}
		return nil
	}

	err := ERC20Token(privateKey, toAddress, contract, value)
	return err
}

func AllTokenViaUserId(userId int, contract, amount string) error {
	hdWallet, err := wallet.HdWalletByMnemonic(os.Getenv("MNEMONIC"))
	if err != nil {
		return err
	}

	privateKey, err := hdWallet.PrivateKey(os.Getenv("CHAIN_SYMBOL"), 0, userId)
	if err != nil {
		return err
	}

	toAddress, err := hdWallet.Address(os.Getenv("CHAIN_SYMBOL"), 1, 0)
	if err != nil {
		return err
	}

	value, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return fmt.Errorf("SetString: error")
	}

	if contract == addressZero {
		err = NativeToken(privateKey, toAddress, value)
		return err
	}

	err = ERC20Token(privateKey, toAddress, contract, value)
	return err
}

func NativeTokenViaUserId(userId int, amount *big.Int) error {
	hdWallet, err := wallet.HdWalletByMnemonic(os.Getenv("MNEMONIC"))
	if err != nil {
		return err
	}

	privateKey, err := hdWallet.PrivateKey(os.Getenv("CHAIN_SYMBOL"), 0, userId)
	if err != nil {
		return err
	}

	toAddress, err := hdWallet.Address(os.Getenv("CHAIN_SYMBOL"), 1, 0)
	if err != nil {
		return err
	}

	err = NativeToken(privateKey, toAddress, amount)
	return err
}

func Erc20TokenViaUserId(userId int, contract string, amount *big.Int) error {
	hdWallet, err := wallet.HdWalletByMnemonic(os.Getenv("MNEMONIC"))
	if err != nil {
		return err
	}

	privateKey, err := hdWallet.PrivateKey(os.Getenv("CHAIN_SYMBOL"), 0, userId)
	if err != nil {
		return err
	}

	toAddress, err := hdWallet.Address(os.Getenv("CHAIN_SYMBOL"), 1, 0)
	if err != nil {
		return err
	}

	err = ERC20Token(privateKey, toAddress, contract, amount)
	return err
}

func NativeToken(privateK, to string, amount *big.Int) error {
	client := NewClient()

	privateKey, err := crypto.HexToECDSA(privateK)
	if err != nil {
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	toAddress := common.HexToAddress(to)
	var data []byte
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    amount,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}

	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
	return nil
}

func ERC20Token(privateK, to, contract string, amount *big.Int) error {
	client := NewClient()

	privateKey, err := crypto.HexToECDSA(privateK)
	if err != nil {
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return err
	}

	toAddress := common.HexToAddress(to)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	tokenContract := common.HexToAddress(contract)
	instance, err := tokens.NewToken(tokenContract, client)
	if err != nil {
		return err
	}

	tx, err := instance.Transfer(auth, toAddress, amount)
	if err != nil {
		return err
	}

	fmt.Printf("tx sent: %s", tx.Hash().Hex()) // tx sent: 0x8d490e535678e9a24360e955d75b27ad307bdfb97a1dca51d0f3035dcee3e870
	return nil
}
