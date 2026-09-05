package middlewares

import "github.com/gin-gonic/gin"

// This middleware authentication is minial
// It cehcks if the `x-api-key` heeader is empty.
// You can update the middleware and make it work with your own
// authentication philosophy

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")
		if apiKey == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized to perform request. Please get a valid API key"})
			return
		}
		c.Next()
	}
}

// Register middleware on the base router
func RegisterMiddlewares(router *gin.Engine) {
	router.Use(AuthMiddleware())
}
