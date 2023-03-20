package databases

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"os"
)

type Database struct {
	Db *bun.DB `json:"db"`
}

func ConnectDatabase() (*Database, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	password := os.Getenv("DB_PASSWORD")
	dsn := os.Getenv("DNS")

	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(dsn),
		pgdriver.WithPassword(password),
	))

	db := Database{Db: bun.NewDB(sqldb, pgdialect.New())}
	fmt.Println("connected to database")

	err = db.CreateTable()
	if err != nil {
		return nil, err
	}

	return &db, nil
}
