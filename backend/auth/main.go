package main

import (
	"github.com/asra123q/sempoa-bookkeeping/auth/models"
	"github.com/asra123q/sempoa-bookkeeping/auth/routes"
	"github.com/asra123q/sempoa-bookkeeping/auth/utils"
)

func main() {
	utils.LoadEnv()
	models.OpenDatabaseConnection()
	rdb := utils.ConnectRedis()
	models.AutoMigrateModels()
	router := routes.SetupRoutes(rdb)
	//middlewares.RegisterMiddlewares(router)
	router.Run(":8080")

}
