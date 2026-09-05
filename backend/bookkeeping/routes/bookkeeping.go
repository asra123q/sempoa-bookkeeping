package routes

import (
	"github.com/asra123q/sempoa-bookkeeping/bookkeeping/controllers"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func bookkeepingGroupRouter(baseRouter *gin.RouterGroup, redisClient *redis.Client) {
	bookkeeping := baseRouter.Group("/bookkeeping")

	// Journal routes
	bookkeeping.GET("/:userId/journals", func(c *gin.Context) {
		controllers.GetJournals(c, redisClient)
	})

	bookkeeping.GET("/:userId/journal/:journalId", func(c *gin.Context) {
		controllers.GetJournal(c, redisClient)
	})

	bookkeeping.POST("/:userId/journal", func(c *gin.Context) {
		controllers.CreateJournal(c, redisClient)
	})

	bookkeeping.DELETE("/:userId/journal/:journalId", func(c *gin.Context) {
		controllers.DeleteJournal(c, redisClient)
	})

	// Entry routes
	bookkeeping.GET("/:userId/entries", func(c *gin.Context) {
		controllers.GetAllEntries(c, redisClient)
	})

	bookkeeping.GET("/:userId/journal/:journalId/entries", func(c *gin.Context) {
		controllers.GetJournalEntries(c, redisClient)
	})

	bookkeeping.GET("/:userId/journal/:journalId/entries/:entryId", func(c *gin.Context) {
		controllers.GetEntry(c, redisClient)
	})

	bookkeeping.POST("/:userId/journal/:journalId/entry", func(c *gin.Context) {
		controllers.CreateJournalEntry(c, redisClient)
	})

	bookkeeping.DELETE("/:userId/journal/:journalId/entry/:entryId", func(c *gin.Context) {
		controllers.DeleteJournalEntry(c, redisClient)
	})
}

func SetupRoutes(redisClient *redis.Client) *gin.Engine {
	r := gin.Default()

	versionRouter := r.Group("/api/v1")

	bookkeepingGroupRouter(versionRouter, redisClient)

	return r
}
