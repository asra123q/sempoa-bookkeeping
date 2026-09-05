package controllers

import (
	"net/http"
	"time"

	"github.com/asra123q/sempoa-bookkeeping/auth/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func Login(c *gin.Context, redisClient *redis.Client) {
	var input models.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	user, err := models.FetchUserByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid email or password",
			"data":    nil,
		})
		return
	}

	if !models.CheckPasswordHash(input.Password, user.PasswordHash) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid email or password",
			"data":    nil,
		})
		return
	}

	// Generate a session ID
	sessionID := uuid.New().String()

	// Store session in Redis
	err = redisClient.Set(
		c.Request.Context(),
		"session:"+sessionID,
		user.ID,
		24*time.Hour,
	).Err()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to create session",
			"data":    nil,
		})
		return
	}

	// Send session ID as an HttpOnly cookie
	c.SetCookie(
		"session_id",
		sessionID,
		86400,
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"data":    user,
	})
}

func Register(c *gin.Context) {
	var input models.RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	hashedPassword, err := models.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to hash password",
			"data":    nil,
		})
		return
	}

	// Converts the request into a User
	user := models.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: hashedPassword,
	}

	savedUser, err := user.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "User registered successfully",
		"data": gin.H{
			"id":       savedUser.ID,
			"username": savedUser.Username,
			"email":    savedUser.Email,
		},
	})
}
