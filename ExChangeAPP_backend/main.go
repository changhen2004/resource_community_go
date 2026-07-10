package main

import (
	"exchangeapp/config"
	"exchangeapp/router"
)

func main() {
	config.InitConfig()

	r := router.SetUpRouter()

	port := config.AppConfig.App.Port
	if port == "" {
		port = "3000"
	}

	r.Run(":" + port)
}
