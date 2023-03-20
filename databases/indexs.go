package databases

import (
	"context"
	"github.com/uptrace/bun"
)

var _ bun.AfterCreateTableHook = (*UserInformation)(nil)
var _ bun.AfterCreateTableHook = (*Wallet)(nil)

func (userInf *UserInformation) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	q := query.DB().NewCreateIndex().
		Model((*UserInformation)(nil))
	err := addIndex(q, "username_ind", "user_name")
	if err != nil {
		return err
	}

	_, err = q.IfNotExists().Exec(ctx)
	return err
}

func (wl *Wallet) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	q := query.DB().NewCreateIndex().
		Model((*Wallet)(nil))
	err := addIndex(q, "address_idx", "address")
	if err != nil {
		return err
	}

	_, err = q.IfNotExists().Exec(ctx)
	return err
}

func addIndex(query *bun.CreateIndexQuery, index, column string) error {
	query = query.Index(index).Column(column)
	return nil
}
