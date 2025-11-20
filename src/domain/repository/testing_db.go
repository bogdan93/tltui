package repository

import (
	"database/sql"
	"testing"
)

// testDB holds test database instance
var testDB *sql.DB

// InitTestDB initializes an in-memory SQLite database for testing
func InitTestDB(t *testing.T) {
	var err error

	// Use in-memory database for tests (faster, isolated)
	testDB, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := testDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Set the global db to test DB
	db = testDB

	// Create schema
	if err := createSchema(); err != nil {
		t.Fatalf("failed to create test schema: %v", err)
	}
}

// CleanupTestDB closes the test database
func CleanupTestDB(t *testing.T) {
	if testDB != nil {
		testDB.Close()
	}
}

// ClearTestData removes all data from test tables
func ClearTestData(t *testing.T) {
	tables := []string{"workhours", "projects", "workhour_details"}
	for _, table := range tables {
		_, err := testDB.Exec("DELETE FROM " + table)
		if err != nil {
			t.Fatalf("failed to clear table %s: %v", table, err)
		}
	}
}
