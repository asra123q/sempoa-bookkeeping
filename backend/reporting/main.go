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

type Server struct {
	DB *sql.DB
}

// Monthly financial summary
type MonthlyReport struct {
	Month              string  `json:"month"`
	TotalIncome        float64 `json:"total_income"`
	TotalExpenditure   float64 `json:"total_expenditure"`
	NetBalance         float64 `json:"net_balance"`
	TransactionCount   int     `json:"transaction_count"`
}

// Average monthly expenditure
type AverageExpenditure struct {
	AverageMonthlyExpenditure float64 `json:"average_monthly_expenditure"`
	MonthsIncluded            int     `json:"months_included"`
}

// Spending summary by journal
type JournalSummary struct {
	JournalID        uuid.UUID `json:"journal_id"`
	JournalName      string    `json:"journal_name"`
	TotalExpenditure float64   `json:"total_expenditure"`
	TotalIncome      float64   `json:"total_income"`
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

	// Reporting is read-only but still authenticated.
	v1.Use(AuthMiddleware())

	{
		// Current / requested month's report
		v1.GET("/reports/monthly", server.monthlyReport)

		// Average expenditure over a period
		v1.GET("/reports/average-expenditure", server.averageExpenditure)

		// Income vs expenditure for each journal
		v1.GET("/reports/journals", server.journalSummary)
	}

	// --------------------------------------------------------
	// Start
	// --------------------------------------------------------

	if err := r.Run(":8007"); err != nil {
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
// JWT Authentication
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
		// Get user ID from JWT
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

		c.Set("userID", userID)

		c.Next()
	}
}

// ============================================================
// Monthly Report
// ============================================================
//
// GET /v1/reports/monthly?year=2026&month=9
//
// If year/month aren't provided, the current month is used.
//
// ============================================================

func (s *Server) monthlyReport(c *gin.Context) {

	userID := c.MustGet("userID").(uuid.UUID)

	year := time.Now().Year()
	month := int(time.Now().Month())

	// Optional year
	if value := c.Query("year"); value != "" {
		_, err := fmt.Sscanf(value, "%d", &year)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid year",
			})
			return
		}
	}

	// Optional month
	if value := c.Query("month"); value != "" {
		_, err := fmt.Sscanf(value, "%d", &month)

		if err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid month",
			})
			return
		}
	}

	// First day of requested month
	startDate := time.Date(
		year,
		time.Month(month),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	// First day of next month
	endDate := startDate.AddDate(0, 1, 0)

	var report MonthlyReport

	report.Month = startDate.Format("2006-01")

	err := s.DB.QueryRowContext(
		c.Request.Context(),
		`
		SELECT
			COALESCE(SUM(debit), 0),
			COALESCE(SUM(credit), 0),
			COALESCE(SUM(debit - credit), 0),
			COUNT(*)
		FROM entries e
		JOIN journals j
			ON e.journal_id = j.id
		WHERE j.user_id = $1
		AND e.created_at >= $2
		AND e.created_at < $3
		`,
		userID,
		startDate,
		endDate,
	).Scan(
		&report.TotalIncome,
		&report.TotalExpenditure,
		&report.NetBalance,
		&report.TransactionCount,
	)

	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// ============================================================
// Average Expenditure
// ============================================================
//
// GET /v1/reports/average-expenditure?months=6
//
// Defaults to the previous 6 months.
//
// ============================================================

func (s *Server) averageExpenditure(c *gin.Context) {

	userID := c.MustGet("userID").(uuid.UUID)

	months := 6

	if value := c.Query("months"); value != "" {

		_, err := fmt.Sscanf(value, "%d", &months)

		if err != nil || months < 1 || months > 120 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "months must be between 1 and 120",
			})
			return
		}
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)

	var result AverageExpenditure

	err := s.DB.QueryRowContext(
		c.Request.Context(),
		`
		SELECT
			COALESCE(AVG(monthly_expenditure), 0),
			COUNT(*)
		FROM (
			SELECT
				DATE_TRUNC('month', e.created_at) AS month,
				SUM(e.credit) AS monthly_expenditure
			FROM entries e
			JOIN journals j
				ON e.journal_id = j.id
			WHERE j.user_id = $1
			AND e.created_at >= $2
			AND e.created_at < $3
			GROUP BY DATE_TRUNC('month', e.created_at)
		) monthly
		`,
		userID,
		startDate,
		endDate,
	).Scan(
		&result.AverageMonthlyExpenditure,
		&result.MonthsIncluded,
	)

	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ============================================================
// Journal Summary
// ============================================================
//
// GET /v1/reports/journals
//
// Returns income/expenditure grouped by journal.
//
// ============================================================

func (s *Server) journalSummary(c *gin.Context) {

	userID := c.MustGet("userID").(uuid.UUID)

	rows, err := s.DB.QueryContext(
		c.Request.Context(),
		`
		SELECT
			j.id,
			j.name,
			COALESCE(SUM(e.credit), 0),
			COALESCE(SUM(e.debit), 0)
		FROM journals j
		LEFT JOIN entries e
			ON e.journal_id = j.id
		WHERE j.user_id = $1
		GROUP BY j.id, j.name
		ORDER BY j.name
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

	summaries := []JournalSummary{}

	for rows.Next() {

		var summary JournalSummary

		err := rows.Scan(
			&summary.JournalID,
			&summary.JournalName,
			&summary.TotalExpenditure,
			&summary.TotalIncome,
		)

		if err != nil {
			log.Println(err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database error",
			})
			return
		}

		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"journals": summaries,
	})
}
