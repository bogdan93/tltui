package models

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// InitDB initializes the database connection and creates schema if needed
func InitDB() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	var dbDir string

	// Platform-specific config directory
	// macOS: ~/Library/Application Support/tltui
	// Linux: $XDG_CONFIG_HOME/tltui or ~/.config/tltui
	if runtime.GOOS == "darwin" {
		// macOS
		dbDir = filepath.Join(homeDir, "Library", "Application Support", "tltui")
	} else {
		// Linux/BSD - Use XDG Base Directory
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(homeDir, ".config")
		}
		dbDir = filepath.Join(configDir, "tltui")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	dbPath := filepath.Join(dbDir, "data.db")
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create schema
	if err := createSchema(); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return db
}

// createSchema creates the database schema if it doesn't exist
func createSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY,
		odoo_id TEXT NOT NULL,
		name TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS workhour_details (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		short_name TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS workhours (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		details_id INTEGER NOT NULL,
		project_id INTEGER NOT NULL,
		hours REAL NOT NULL,
		FOREIGN KEY (details_id) REFERENCES workhour_details(id),
		FOREIGN KEY (project_id) REFERENCES projects(id)
	);

	CREATE INDEX IF NOT EXISTS idx_workhours_date ON workhours(date);
	`

	_, err := db.Exec(schema)
	return err
}

// DateToString converts time.Time to ISO 8601 date string (YYYY-MM-DD)
func DateToString(t time.Time) string {
	return t.Format("2006-01-02")
}

// StringToDate converts ISO 8601 date string to time.Time
func StringToDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
