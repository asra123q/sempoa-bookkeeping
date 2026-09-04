package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

// ============================================================
// Models
// ============================================================

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
	Debit            float64   `json:"debit"`
	Credit           float64   `json:"credit"`
	RemainingBalance float64   `json:"remaining_balance"`
	CreatedAt        time.Time `json:"created_at"`
}

type Server struct {
	DB *sql.DB
}

// ============================================================
// Request Models
// ============================================================

type CreateJournalRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateEntryRequest struct {
	Description string  `json:"description" binding:"required"`
	Debit       float64 `json:"debit" binding:"gte=0"`
	Credit      float64 `json:"credit" binding:"gte=0"`
}

// ============================================================
// Main
// ============================================================

func main() {

	// --------------------------------------------------------
	// Environment
	// --------------------------------------------------------

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// --------------------------------------------------------
	// PostgreSQL
	// --------------------------------------------------------

	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	server := &Server{
		DB: db,
	}

	// --------------------------------------------------------
	// Gin
	// --------------------------------------------------------

	r := gin.Default()

	r.Use(Logger())

	// --------------------------------------------------------
	// Routes
	// --------------------------------------------------------

	v1 := r.Group("/v1")

	// Everything under /v1 requires authentication.
	v1.Use(AuthMiddleware())

	{
		// Journals
		v1.GET("/journals", server.getJournals)
		v1.POST("/journals", server.createJournal)
		v1.DELETE("/journals/:journalID", server.deleteJournal)

		// Entries
		v1.GET(
			"/journals/:journalID/entries",
			server.getEntriesFromAJournal,
		)

		v1.POST(
			"/journals/:journalID/entries",
			server.enterTransaction,
		)

		v1.DELETE(
			"/entries/:entryID",
			server.deleteEntry,
		)
	}

	// --------------------------------------------------------
	// Start
	// --------------------------------------------------------

	if err := r.Run(":8005"); err != nil {
		log.Fatal(err)
	}
}

// ============================================================
// Middleware
// ============================================================

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

// ============================================================
// JWT Authentication Middleware
// ============================================================

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		secret := os.Getenv("JWT_SECRET")

		if secret == "" {
			log.Println("JWT_SECRET is not configured")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "server configuration error",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(
			tokenString,
			func(token *jwt.Token) (interface{}, error) {

				// Prevent someone from changing the algorithm
				// to something unexpected.
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}

				return []byte(secret), nil
			},
		)

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// ----------------------------------------------------
		// Extract user ID from JWT
		// ----------------------------------------------------

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token claims",
			})
			c.Abort()
			return
		}

		userIDString, ok := claims["user_id"].(string)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user ID missing from token",
			})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDString)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user ID",
			})
			c.Abort()
			return
		}

		// Make user ID available to handlers.
		c.Set("userID", userID)

		c.Next()
	}
}

// ============================================================
// Handlers - Journals
// ============================================================

func (s *Server) getJournals(c *gin.Context) {

	userID := c.MustGet("userID").(uuid.UUID)

	rows, err := s.DB.QueryContext(
		c.Request.Context(),
		`
		SELECT
			id,
			user_id,
			name,
			created_at,
			updated_at
		FROM journals
		WHERE user_id = $1
		ORDER BY created_at DESC
		`,
		userID,
	)

	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	defer rows.Close()

	journals := []Journal{}

	for rows.Next() {

		var journal Journal

		err := rows.Scan(
			&journal.ID,
			&journal.UserID,
			&journal.Name,
			&journal.CreatedAt,
			&journal.UpdatedAt,
		)

		if err != nil {
			log.Println(err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database error",
			})
			return
		}

		journals = append(journals, journal)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"journals": journals,
	})
}

func (s *Server) createJournal(c *gin.Context) {

	userID := c.MustGet("userID").(uuid.UUID)

	var req CreateJournalRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	journalID := uuid.New()

	var journal Journal

	err := s.DB.QueryRowContext(
		c.Request.Context(),
		`
		INSERT INTO journals (
			id,
			user_id,
			name,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, user_id, name, created_at, updated_at
		`,
		journalID,
		userID,
		req.Name,
	).Scan(
		&journal.ID,
		&journal.UserID,
		&journal.Name,
		&journal.CreatedAt,
		&journal.UpdatedAt,
	)

	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not create journal",
		})
		return
	}

	c.JSON(http.StatusCreated, journal)
}

func (s *Server) deleteJournal(c *gin.Context) {

	userID := c.MustGet("userID").(uuid.UUID)

	journalID, err := uuid.Parse(
		c.Param("journalID"),
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid journal ID",
		})
		return
	}

	result, err := s.DB.ExecContext(
		c.Request.Context(),
		`
		DELETE FROM journals
		WHERE id = $1
		AND user_id = $2
		`,
		journalID,
		userID,
	)

	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "journal not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ============================================================
// Handlers - Entries
// ============================================================

func (s *Server) getEntriesFromAJournal(c *gin.Context) {

	userID := c.MustGet("userID").(uuid.UUID)

	journalID, err := uuid.Parse(
		c.Param("journalID"),
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid journal ID",
		})
		return
	}

	// First make sure this journal belongs to the user.
	var exists bool

	err = s.DB.QueryRowContext(
		c.Request.Context(),
		`
		SELECT EXISTS (
			SELECT 1
			FROM journals
			WHERE id = $1
			AND user_id = $2
		)
		`,
		journalID,
		userID,
	).Scan(&exists)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "journal not found",
		})
		return
	}

	rows, err := s.DB.QueryContext(
		c.Request.Context(),
		`
		SELECT
			id,
			journal_id,
			description,
			debit,
			credit,
			remaining_balance,
			created_at
		FROM entries
		WHERE journal_id = $1
		ORDER BY created_at ASC
		`,
		journalID,
	)

	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	defer rows.Close()

	entries := []Entry{}

	for rows.Next() {

		var entry Entry

		err := rows.Scan(
			&entry.ID,
			&entry.JournalID,
			&entry.Description,
			&entry.Debit,
			&entry.Credit,
			&entry.RemainingBalance,
			&entry.CreatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database error",
			})
			return
		}

		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
	})
}

func (s *Server) enterTransaction(c *gin.Context) {

	userID := c.MustGet("userID").(uuid.UUID)

	journalID, err := uuid.Parse(
		c.Param("journalID"),
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid journal ID",
		})
		return
	}

	var req CreateEntryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// A transaction should not normally have both debit
	// and credit simultaneously.
	if req.Debit > 0 && req.Credit > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "entry cannot have both debit and credit",
		})
		return
	}

	if req.Debit == 0 && req.Credit == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "entry must have a debit or credit",
		})
		return
	}

	// Make sure journal belongs to the authenticated user.
	var exists bool

	err = s.DB.QueryRowContext(
		c.Request.Context(),
		`
		SELECT EXISTS (
			SELECT 1
			FROM journals
			WHERE id = $1
			AND user_id = $2
		)
		`,
		journalID,
		userID,
	).Scan(&exists)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "journal not found",
		})
		return
	}

	// --------------------------------------------------------
	// Calculate balance
	// --------------------------------------------------------

	var previousBalance float64

	err = s.DB.QueryRowContext(
		c.Request.Context(),
		`
		SELECT COALESCE(
			(
				SELECT remaining_balance
				FROM entries
				WHERE journal_id = $1
				ORDER BY created_at DESC
				LIMIT 1
			),
			0
		)
		`,
		journalID,
	).Scan(&previousBalance)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not calculate balance",
		})
		return
	}

	newBalance := previousBalance + req.Debit - req.Credit

	// --------------------------------------------------------
	// Insert entry
	// --------------------------------------------------------

	entryID := uuid.New()

	var entry Entry

	err = s.DB.QueryRowContext(
		c.Request.Context(),
		`
		INSERT INTO entries (
			id,
			journal_id,
			description,
			debit,
			credit,
			remaining_balance,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING
			id,
			journal_id,
			description,
			debit,
			credit,
			remaining_balance,
			created_at
		`,
		entryID,
		journalID,
		req.Description,
		req.Debit,
		req.Credit,
		newBalance,
	).Scan(
		&entry.ID,
		&entry.JournalID,
		&entry.Description,
		&entry.Debit,
		&entry.Credit,
		&entry.RemainingBalance,
		&entry.CreatedAt,
	)

	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not create entry",
		})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func (s *Server) deleteEntry(c *gin.Context) {

	userID := c.MustGet("userID").(uuid.UUID)

	entryID, err := uuid.Parse(
		c.Param("entryID"),
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid entry ID",
		})
		return
	}

	// Only delete entries belonging to one of the user's journals.
	result, err := s.DB.ExecContext(
		c.Request.Context(),
		`
		DELETE FROM entries
		WHERE id = $1
		AND journal_id IN (
			SELECT id
			FROM journals
			WHERE user_id = $2
		)
		`,
		entryID,
		userID,
	)

	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "entry not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}