package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"misc/internal/models"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	Init() error

	GetHabitRule(string) (*models.HabiticaHabitRule, error)
	GetTodoistHabiticaTextRules() ([]models.TodoistHabiticaTextRule, error)
	GetTodoistHabiticaProjectRule(string) (models.TodoistHabiticaProjectRule, error)
}

type service struct {
	db *sql.DB
}

var (
	dburl      = os.Getenv("DB_URL")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}

	dbInstance = &service{
		db: db,
	}

	if err := dbInstance.Init(); err != nil {
		log.Fatal(err)
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(fmt.Sprintf("db down: %v", err)) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", dburl)
	return s.db.Close()
}

// Create initial tables in the database
func (s *service) Init() error {
	_, err := s.db.Exec(
		`CREATE TABLE IF NOT EXISTS HabiticaHabitRule (
			id INTEGER PRIMARY KEY,
			name TEXT,
			habitId TEXT,
			dailyId TEXT,
			minScore INTEGER
		)`,
	)
	if err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	}

	_, err = s.db.Exec(
		`CREATE TABLE IF NOT EXISTS TodoistHabitTextRule (
			id INTEGER PRIMARY KEY,
			name TEXT,
			rule TEXT,
			habitId TEXT
		)`,
	)
	if err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	}

	_, err = s.db.Exec(
		`CREATE TABLE IF NOT EXISTS TodoistHabitProjectRule (
			id INTEGER PRIMARY KEY,
			name TEXT,
			todoistProjectId TEXT,
			habitId TEXT
		)`,
	)

	if err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	}

	_, err = s.db.Exec(
		`CREATE TABLE IF NOT EXISTS TodoistHabitTextRule (
			id INTEGER PRIMARY KEY,
			name TEXT,
			ruleText TEXT,
			habitId TEXT
		)`,
	)

	if err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	}
	return nil
}

func (s *service) GetHabitRule(habitId string) (*models.HabiticaHabitRule, error) {
	var rule models.HabiticaHabitRule
	row := s.db.QueryRow(
		`SELECT habitId, dailyId, minScore FROM HabiticaHabitRule WHERE habitId = ?`,
		habitId,
	)
	if err := row.Scan(&rule.HabitId, &rule.DailyId, &rule.MinScore); err != nil {
		return nil, fmt.Errorf("error retrieving habit rule: %w", err)
	}
	return &rule, nil
}

func (s *service) GetTodoistHabiticaTextRules() ([]models.TodoistHabiticaTextRule, error) {
	rules := make([]models.TodoistHabiticaTextRule, 0)
	rows, err := s.db.Query(`SELECT name, rule, habitId FROM TodoistHabitTextRule;`)
	if err != nil {
		return rules, fmt.Errorf("error creating text rule query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rule models.TodoistHabiticaTextRule
		err := rows.Scan(&rule.Name, &rule.Rule, &rule.HabitId)
		if err != nil {
			return rules, fmt.Errorf("error scanning text rule row: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (s *service) GetTodoistHabiticaProjectRule(projectId string) (models.TodoistHabiticaProjectRule, error) {
	var rule models.TodoistHabiticaProjectRule
	row := s.db.QueryRow(
		`SELECT name, todoistProjectId, habitId FROM TodoistHabitProjectRule WHERE todoistProjectId = ?`,
		projectId,
	)
	if err := row.Scan(&rule.Name, &rule.ProjectId, &rule.HabitId); err != nil {
		return rule, fmt.Errorf("error retrieving todoist project rule: %w", err)
	}
	return rule, nil
}
