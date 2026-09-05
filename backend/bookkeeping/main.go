package main

import (
	"github.com/asra123q/sempoa-bookkeeping/bookkeeping/models"
	"github.com/asra123q/sempoa-bookkeeping/bookkeeping/routes"
	"github.com/asra123q/sempoa-bookkeeping/bookkeeping/utils"
)

func main() {
	utils.LoadEnv()
	models.OpenDatabaseConnection()
	rdb := utils.ConnectRedis()
	models.AutoMigrateModels()
	router := routes.SetupRoutes(rdb)
	//middlewares.RegisterMiddlewares(router)
	router.Run(":8081")

}
