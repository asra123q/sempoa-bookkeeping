package routes

import (
	"github.com/asra123q/sempoa-bookkeeping/auth/controllers"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func authGroupRouter(baseRouter *gin.RouterGroup, redisClient *redis.Client) {
	auth := baseRouter.Group("/auth")

	auth.POST("/login", func(c *gin.Context) {
		controllers.Login(c, redisClient)
	})
	auth.POST("/register", func(c *gin.Context) {
		controllers.Register(c)
	})

}

func SetupRoutes(redisClient *redis.Client) *gin.Engine {
	r := gin.Default()

	versionRouter := r.Group("/api/v1")
	authGroupRouter(versionRouter, redisClient)

	return r
}
