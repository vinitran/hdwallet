package blockchain

import (
	"ExchangeManager/databases"
	"ExchangeManager/redis"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"strings"
	"time"
)

type TrackingTx struct {
	Client *ethclient.Client
}

type Tracking struct {
	Address  []common.Address
	Contract []common.Address
}

type Transaction struct {
	From        string    `json:"from"`
	To          string    `json:"to"`
	Contract    string    `json:"contract"`
	SymbolToken string    `json:"symbol_token"`
	RawAmount   string    `json:"raw_amount"`
	Hash        string    `json:"hash"`
	Time        time.Time `json:"time"`
}

var userAddresses []common.Address

var (
	transferEventHash = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")).Hex()
	AddressZero       = "0x0000000000000000000000000000000000000000"
)

func init() {
	db, err := databases.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}

	addresses, err := db.GetUserAddresses()
	if err != nil {
		log.Fatal(err)
	}

	for _, address := range addresses {
		userAddresses = append(userAddresses, common.HexToAddress(address))
	}

	go func() {
		rdb := redis.NewRedis()
		pubsub := rdb.Subscribe(context.Background(), "new_user_address")

		// Close the subscription when we are done.
		defer pubsub.Close()

		ch := pubsub.Channel()

		for msg := range ch {
			userAddresses = append(userAddresses, common.HexToAddress(msg.Payload))
		}
	}()
}

func NewTxTracking(rpc string) (*TrackingTx, error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, err
	}

	trackingTx := &TrackingTx{
		Client: client,
	}

	return trackingTx, nil
}

func (tracking *TrackingTx) GetTransaction(msg chan Transaction, contracts []common.Address) {
	block, err := tracking.LatestBlock()
	if err != nil {
		fmt.Println(err)
		return
	}

	chainID, err := tracking.Client.NetworkID(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		current, err := tracking.LatestBlock()
		if err != nil {
			fmt.Println(err)
			continue
		}

		if current.Cmp(block) == -1 {
			timeDelay := (block.Int64()-current.Int64())*3 + 1
			time.Sleep(time.Duration(timeDelay) * time.Second)
			continue
		}

		txBlock, err := tracking.TransactionByBlockNumber(chainID, block, userAddresses)
		if err != nil {
			fmt.Println(err)
			continue
		}

		txEvent, err := tracking.EventByBlockNumber(block, userAddresses, contracts)
		if err != nil {
			fmt.Println(err)
			continue
		}

		data := append(txBlock, txEvent...)
		fmt.Println(block)
		if len(data) > 0 {
			for _, dt := range data {
				msg <- dt
			}
		}

		block.Add(block, big.NewInt(1))
	}
}

func (tracking *TrackingTx) EventByBlockNumber(number *big.Int, addresses, contracts []common.Address) ([]Transaction, error) {
	// Assume that number is less than current block
	query := ethereum.FilterQuery{
		FromBlock: number,
		ToBlock:   number,
		Addresses: contracts,
	}

	logs, err := tracking.Client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}

	if len(logs) < 1 {
		return nil, err
	}

	var data []Transaction

	for _, vLog := range logs {
		switch vLog.Topics[0].Hex() {
		case transferEventHash:
			fmt.Println(vLog)
			if len(vLog.Topics) < 3 {
				continue
			}
			recipient := common.HexToAddress(vLog.Topics[2].Hex()).String()
			fmt.Println(recipient, addresses)
			if tracking.IsRecipient(recipient, addresses) == false {
				continue
			}

			timeStamp, err := tracking.TimeOfBlock(big.NewInt(int64(vLog.BlockNumber)))
			if err != nil {
				fmt.Println(err)
				continue
			}

			symbol, err := tracking.TokenSymbol(vLog.Address)
			if err != nil {
				println(err)
				continue
			}

			data = append(data, Transaction{
				From:        strings.ToLower(common.HexToAddress(vLog.Topics[1].Hex()).String()),
				To:          strings.ToLower(common.HexToAddress(vLog.Topics[2].Hex()).String()),
				Contract:    strings.ToLower(vLog.Address.String()),
				SymbolToken: symbol,
				RawAmount:   new(big.Int).SetBytes(vLog.Data).String(),
				Hash:        strings.ToLower(vLog.TxHash.String()),
				Time:        timeStamp,
			})
		}
	}

	return data, nil
}

func (tracking *TrackingTx) TransactionByBlockNumber(chainID *big.Int, number *big.Int, addresses []common.Address) ([]Transaction, error) {
	// Assume that number is less than current block
	block, err := tracking.Client.BlockByNumber(context.Background(), number)
	if err != nil {
		return nil, err
	}

	var data []Transaction
	for _, tx := range block.Transactions() {
		if tx.To() == nil {
			continue
		}

		isInTransaction := tracking.IsRecipient(tx.To().String(), addresses)
		if isInTransaction == false {
			continue
		}

		//Check amount > 0.
		if tx.Value().Cmp(big.NewInt(0)) == -1 {
			continue
		}

		msg, err := tx.AsMessage(types.NewEIP155Signer(chainID), tx.GasPrice())
		if err != nil {
			return nil, err
		}

		timestamp := time.Unix(int64(block.Time()), 0)
		data = append(data, Transaction{
			Contract:    strings.ToLower(AddressZero),
			SymbolToken: "BNB",
			From:        strings.ToLower(msg.From().String()),
			To:          strings.ToLower(tx.To().String()),
			RawAmount:   tx.Value().String(),
			Hash:        strings.ToLower(tx.Hash().Hex()),
			Time:        timestamp,
		})
	}
	return data, nil
}

func (tracking *TrackingTx) LatestBlock() (*big.Int, error) {
	header, err := tracking.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return header.Number, nil
}

func (tracking *TrackingTx) IsRecipient(recipient string, addresses []common.Address) bool {
	for _, addr := range addresses {
		if addr.String() != recipient {
			continue
		}
		return true
	}
	return false
}

func (tracking *TrackingTx) TimeOfBlock(block *big.Int) (time.Time, error) {
	header, err := tracking.Client.HeaderByNumber(context.Background(), block)
	if err != nil {
		return time.Time{}, err
	}

	timeStamp := time.Unix(int64(header.Time), 0)
	return timeStamp, nil
}
