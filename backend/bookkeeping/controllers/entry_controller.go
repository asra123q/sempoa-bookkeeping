package controllers

import (
	"net/http"
	"strconv"

	"github.com/asra123q/sempoa-bookkeeping/bookkeeping/models"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func GetAllEntries(c *gin.Context, redisClient *redis.Client) {
	userIdStr := c.Param("userId")

	userId64, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	// Fetch entries from the database
	entries, err := models.FetchAllEntries(uint(userId64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to fetch entries",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Entries fetched successfully",
		"data":    entries,
	})
}

func GetJournalEntries(c *gin.Context, redisClient *redis.Client) {
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

	// Fetch entries for the specific journal from the database
	entries, err := models.FetchEntriesByJournalId(uint(userId64), uint(journalId64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to fetch entries for the journal",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Entries for the journal fetched successfully",
		"data":    entries,
	})
}

func GetEntry(c *gin.Context, redisClient *redis.Client) {
	userIdStr := c.Param("userId")
	journalIdStr := c.Param("journalId")
	entryIdStr := c.Param("entryId")

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

	entryId64, err := strconv.ParseUint(entryIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid entry ID"})
		return
	}

	// Fetch the specific entry from the database
	entry, err := models.FetchEntryByID(uint(userId64), uint(journalId64), uint(entryId64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to fetch the entry",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Entry fetched successfully",
		"data":    entry,
	})
}

func CreateJournalEntry(c *gin.Context, redisClient *redis.Client) {
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

	var input models.Entry
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	input.OwnerID = uint(userId64)
	input.JournalID = uint(journalId64)

	savedEntry, err := input.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to create entry",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Entry created successfully",
		"data":    savedEntry,
	})
}

func DeleteJournalEntry(c *gin.Context, redisClient *redis.Client) {
	userIdStr := c.Param("userId")
	journalIdStr := c.Param("journalId")
	entryIdStr := c.Param("entryId")

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

	entryId64, err := strconv.ParseUint(entryIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid entry ID"})
		return
	}

	err = models.DeleteEntry(uint(userId64), uint(journalId64), uint(entryId64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to delete entry",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Entry deleted successfully",
	})
}
