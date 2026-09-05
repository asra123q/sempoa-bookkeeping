package controllers

import (
	"net/http"
	"strconv"

	"github.com/asra123q/sempoa-bookkeeping/bookkeeping/models"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func GetJournals(c *gin.Context, redisClient *redis.Client) {
	userIdStr := c.Param("userId")

	userId64, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}
	// Fetch journals from the database
	journals, err := models.FetchAllJournals(uint(userId64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to fetch journals",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Journals fetched successfully",
		"data":    journals,
	})
}

func GetJournal(c *gin.Context, redisClient *redis.Client) {
	userIdStr := c.Param("userId")
	journalIdStr := c.Param("journalId")

	userId64, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	journalId64, err := strconv.ParseUint(journalIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid journal ID"})
		return
	}

	// Fetch journal from the database
	journal, err := models.FetchJournalByID(uint(userId64), uint(journalId64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to fetch journal",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Journal fetched successfully",
		"data":    journal,
	})
}

func CreateJournal(c *gin.Context, redisClient *redis.Client) {
	var input models.Journal

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	userIdStr := c.Param("userId")
	userId64, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	input.OwnerID = uint(userId64)

	savedJournal, err := input.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to save journal",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Journal created successfully",
		"data":    savedJournal,
	})
}

func DeleteJournal(c *gin.Context, redisClient *redis.Client) {
	userIdStr := c.Param("userId")
	journalIdStr := c.Param("journalId")

	userId64, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	journalId64, err := strconv.ParseUint(journalIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid journal ID"})
		return
	}

	err = models.DeleteJournal(uint(userId64), uint(journalId64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to delete journal",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Journal deleted successfully",
		"data":    nil,
	})
}
