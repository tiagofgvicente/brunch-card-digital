package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// ConnectDB initializes the connection to the Postgres cluster
func ConnectDB(host, user, password, dbname string) (*sql.DB, error) {
	// Connection string for Kubernetes service
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		host, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Check if the connection is alive
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// RunMigrations reads the local SQL file and executes it
// to ensure the database schema is up to date.
func RunMigrations(db *sql.DB, filepath string) error {
	// Read the SQL file
	query, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("could not read migration file: %w", err)
	}

	// Execute the migration queries
	_, err = db.Exec(string(query))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	fmt.Println("Database migrations applied successfully!")
	return nil
}
