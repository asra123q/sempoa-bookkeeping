package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
    "strings"

    "github.com/gin-gonic/gin/binding"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

// =========================
// Models
// =========================

type Journal struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Entry struct {
	ID               uuid.UUID `json:"id"`
	JournalID        uuid.UUID `json:"journal_id"`
	Description      string    `json:"description"`
	Debit            float32   `json:"debit"`
	Credit           float32   `json:"credit"`
	RemainingBalance float32   `json:"remaining_balance"`
	CreatedAt        time.Time `json:"created_at"`
}

// Request body for registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// Database representation of a user
type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Server contains dependencies used by handlers
type Server struct {
	DB *sql.DB
}

// =========================
// Validator
// =========================

var email validator.Func = func(fl validator.FieldLevel) bool {
	userEmail, ok := fl.Field().Interface().(string)

	if !ok {
		return false
	}

	// Basic email format validation.
	// The built-in "email" validator is already better for most cases,
	// so this custom validator isn't actually necessary.
	matched, err := regexp.MatchString(
		`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		userEmail,
	)

	return err == nil && matched
}

// =========================
// Main
// =========================

func main() {

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// PostgreSQL connection string
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// Verify database connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Create server with database dependency
	server := &Server{
		DB: db,
	}

	// =========================
	// Gin
	// =========================

	r := gin.Default()

	// Logger middleware
	r.Use(Logger())

	// =========================
	// Validation
	// =========================

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("customemail", email)
	}

	// =========================
	// Redis Session Store
	// =========================

	store, err := redis.NewStore(
		10,
		"tcp",
		"localhost:6379",
		"",
		os.Getenv("SECRET"),
	)

	if err != nil {
		log.Fatal(err)
	}

	r.Use(sessions.Sessions("mysession", store))

	// =========================
	// Routes
	// =========================

	v1 := r.Group("/v1")

	v1.POST("/login", server.login)
	v1.POST("/register", server.register)

	// =========================
	// Start Server
	// =========================

	if err := r.Run(":8003"); err != nil {
		log.Fatal(err)
	}
}

// =========================
// Middleware
// =========================

func Logger() gin.HandlerFunc {

	return func(c *gin.Context) {

		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf(
			"%s %s %d %v",
			c.Request.Method,
			c.Request.URL.Path,
			status,
			latency,
		)
	}
}

// =========================
// Handlers
// =========================

func (s *Server) login(c *gin.Context) {

	session := sessions.Default(c)

	session.Set("user_id", "example-user")

	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not save session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
	})
}

func (s *Server) register(c *gin.Context) {

	// -------------------------
	// Parse JSON
	// -------------------------

	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// -------------------------
	// Hash password
	// -------------------------

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not hash password",
		})
		return
	}

	// -------------------------
	// Create user ID
	// -------------------------

	userID := uuid.New()

	// -------------------------
	// Insert user
	// -------------------------

	_, err = s.DB.ExecContext(
		c.Request.Context(),
		`
		INSERT INTO users (
			id,
			username,
			email,
			password
		)
		VALUES ($1, $2, $3, $4)
		`,
		userID,
		req.Username,
		req.Email,
		string(hashedPassword),
	)

	if err != nil {

		// PostgreSQL duplicate violation
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "username or email already exists",
			})
			return
		}

		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not create user",
		})
		return
	}

	// -------------------------
	// Response
	// -------------------------

	c.JSON(http.StatusCreated, gin.H{
		"id":       userID,
		"username": req.Username,
		"email":    req.Email,
	})
}