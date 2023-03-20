package databases

import (
	"context"
	"math/big"
)

func (database *Database) GetUserAddresses() ([]string, error) {
	var address []string
	_, err := database.Db.NewSelect().Model(new([]Wallet)).ColumnExpr("address").GroupExpr("address").Exec(context.Background(), &address)
	if err != nil {
		return nil, err
	}
	return address, nil
}

func (database *Database) IncreaseAmount(IncreaseAmountStr, address, contract string) error {
	var firstAmountStr string
	_, err := database.Db.NewSelect().Model(new(Wallet)).
		ColumnExpr("raw_amount").
		Where("address = ?", address).
		Where("contract = ?", contract).
		Exec(context.Background(), &firstAmountStr)
	if err != nil {
		return err
	}

	firstAmount, ok := new(big.Int).SetString(firstAmountStr, 10)
	if !ok {
		return err
	}

	increaseAmount, ok := new(big.Int).SetString(IncreaseAmountStr, 10)
	if !ok {
		return err
	}

	currentAmount := firstAmount.Add(firstAmount, increaseAmount).String()

	_, err = database.Db.NewUpdate().Model(&Wallet{RawAmount: currentAmount}).
		Column("raw_amount").
		Where("address = ?", address).
		Where("contract = ?", contract).
		Exec(context.Background())

	return err
}
