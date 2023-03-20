package databases

import (
	"context"
	"github.com/uptrace/bun"
)

type UserInformation struct {
	bun.BaseModel `bun:"table:user_information,alias:ui"`
	Id            int      `bun:"id,pk,autoincrement" json:"id"`
	UserName      string   `bun:"user_name,pk,unique" json:"user_name"`
	Wallets       []Wallet `bun:"rel:has-many,join:id=id"`
}

type Wallet struct {
	bun.BaseModel `bun:"table:wallets,alias:wl"`
	Id            int    `bun:"id,pk,notnull" json:"id"`
	Contract      string `bun:"contract,pk,notnull" json:"contract"`
	SymbolToken   string `bun:"symbol_token,notnull" json:"symbol_token"`
	Address       string `bun:"address,pk,notnull" json:"address"`
	RawAmount     string `bun:"raw_amount,notnull" json:"raw_amount"`
}

func (database *Database) CreateTable() error {
	err := database.createUserInformation()
	if err != nil {
		return err
	}

	err = database.createUserWallet()
	if err != nil {
		return err
	}

	return nil
}

func (database *Database) createUserInformation() error {
	_, err := database.Db.NewCreateTable().
		Model((*UserInformation)(nil)).
		IfNotExists().
		Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (database *Database) createUserWallet() error {
	_, err := database.Db.NewCreateTable().
		Model((*Wallet)(nil)).
		IfNotExists().
		Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (database *Database) ChangeAmount(address, amount string) error {
	_, err := database.Db.NewCreateTable().
		Model((*Wallet)(nil)).
		IfNotExists().
		Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}
