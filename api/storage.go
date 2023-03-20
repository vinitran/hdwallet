package api

import (
	"ExchangeManager/databases"
	"ExchangeManager/redis"
	"ExchangeManager/transfer"
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/uptrace/bun"
	"math/big"
	"net/http"
	"os"
	"strings"
)

var addressZero = "0x0000000000000000000000000000000000000000"

func HelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, render.JSON{Data: "Hello World"})
	return
}

func NewUser(c *gin.Context) {
	type InforReq struct {
		UserName string `json:"user_name"`
	}

	var userInforReq InforReq
	err := c.BindJSON(&userInforReq)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	user := databases.UserInformation{
		UserName: userInforReq.UserName,
	}

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	_, err = tx.NewInsert().
		Model(&user).
		Exec(context.Background())
	if err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	chainSymbol := os.Getenv("CHAIN_SYMBOL")
	address, err := GetUserAddress(user.Id, chainSymbol)
	if err != nil {
		_ = tx.Rollback()
		fmt.Println("address")
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}
	address = strings.ToLower(address)
	err = redis.Publish(address, "new_user_address")
	if err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	err = SendAllNativeTokenToManager(user.Id, chainSymbol)
	if err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	nativeTokenAmount, err := GetNativeTokenAmount(os.Getenv("RPC"), address)
	if err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	_, err = tx.NewInsert().
		Model(&databases.Wallet{
			Id:          user.Id,
			Contract:    strings.ToLower(addressZero),
			SymbolToken: chainSymbol,
			Address:     strings.ToLower(address),
			RawAmount:   nativeTokenAmount,
		}).
		Exec(context.Background())
	if err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	contract := strings.ToLower(os.Getenv("TOKEN_CONTRACT"))
	token, err := NewToken(os.Getenv("RPC"), contract)
	if err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	err = SendAllErc20TokenToManager(user.Id, token, contract, chainSymbol)
	if err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	tokenAmount, err := token.Amount(address)
	if err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	tokenSymbol, err := token.Symbol()
	if err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	_, err = tx.NewInsert().
		Model(&databases.Wallet{
			Id:          user.Id,
			Contract:    contract,
			SymbolToken: tokenSymbol,
			Address:     address,
			RawAmount:   tokenAmount,
		}).
		Exec(context.Background())
	if err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	err = tx.Commit()
	c.JSON(http.StatusOK, render.JSON{Data: user.Id})
	return
}

func Withdrawal(c *gin.Context) {
	type InforReq struct {
		UserName string `json:"user_name"`
		Contract string `json:"contract"`
		Amount   string `json:"amount"`
		Address  string `json:"address"`
	}

	var userInforReq InforReq
	err := c.BindJSON(&userInforReq)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	userInfor := new(databases.UserInformation)
	err = db.NewSelect().Model(userInfor).
		Relation("Wallets", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("contract = ?", userInforReq.Contract)
		}).
		Where("user_name = ?", userInforReq.UserName).
		Scan(context.Background())
	if err != nil {
		fmt.Println("error", err.Error())
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	currentAmount, ok := new(big.Int).SetString(userInfor.Wallets[0].RawAmount, 10)
	if !ok {
		err = fmt.Errorf("SetString: error")
		fmt.Println("error", err.Error())
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	withdrawAmount, ok := new(big.Int).SetString(userInforReq.Amount, 10)
	if !ok {
		err = fmt.Errorf("SetString: error")
		fmt.Println("error", err.Error())

		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	if currentAmount.Cmp(withdrawAmount) == -1 {
		err = fmt.Errorf("error: amount must be less than current balance")
		fmt.Println("error", err.Error())

		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	managerPrivateKey, err := transfer.GetManagerPrivateKey()
	if err != nil {
		fmt.Println("error", err.Error())
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	err = db.RunInTx(context.Background(), &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		amountAfterWithdrawal := currentAmount.Sub(currentAmount, withdrawAmount).String()

		_, err = tx.NewUpdate().
			Model(&databases.Wallet{RawAmount: amountAfterWithdrawal}).
			Where("id = ?", userInfor.Id).
			Where("contract = ?", userInforReq.Contract).
			Exec(context.Background())

		err = transfer.Token(managerPrivateKey, userInforReq.Address, userInforReq.Contract, userInforReq.Amount)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		fmt.Println("error", err.Error())
		c.JSON(http.StatusBadRequest, render.JSON{Data: fmt.Sprintf("%s", err)})
		return
	}

	fmt.Println("success")
	c.JSON(http.StatusOK, render.JSON{Data: "Withdraw successfully"})
	return

}
