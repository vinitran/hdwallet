package api

import (
	"ExchangeManager/databases"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"log"
	"os"
)

type GinEngine struct {
	g *gin.Engine
}

var db *bun.DB

func NewGin() *GinEngine {
	database, err := databases.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}

	db = database.Db
	return &GinEngine{g: gin.New()}
}

func (g *GinEngine) Run() {
	//gin.SetupRoutes()
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"PUT", "PATCH", "GET", "POST", "HEAD", "OPTIONS"},
		AllowHeaders: []string{"*"},
	}))
	router.GET("/", HelloWorld)
	router.POST("/user", NewUser)
	router.POST("/user/balance/withdrawal", Withdrawal)

	err := router.Run(fmt.Sprintf(":%s", os.Getenv("SV_PORT")))
	if err != nil {
		log.Fatal(err)
	}
}
