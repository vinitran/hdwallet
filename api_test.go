package main

import (
	"ExchangeManager/api"
	"github.com/joho/godotenv"
	"testing"
)

func TestApi(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Error(err)
	}

	gin := api.NewGin()
	gin.Run()
}
