package api

func (g *GinEngine) SetupRoutes() {
	client := g.g.Group("/api")
	{
		client.GET("/", HelloWorld)
		client.POST("/user", NewUser)
		client.POST("/user/balance/withdrawal", Withdrawal)
	}
}
