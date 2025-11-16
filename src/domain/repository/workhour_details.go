package repository

import (
	"database/sql"
	"fmt"
	"tltui/src/domain"
)

// GetAllWorkhourDetailsFromDB retrieves all workhour details from the database
func GetAllWorkhourDetailsFromDB() ([]domain.WorkhourDetails, error) {
	rows, err := db.Query("SELECT id, name, short_name, is_work FROM workhour_details ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to query workhour details: %w", err)
	}
	defer rows.Close()

	var details []domain.WorkhourDetails
	for rows.Next() {
		var d domain.WorkhourDetails
		if err := rows.Scan(&d.ID, &d.Name, &d.ShortName, &d.IsWork); err != nil {
			return nil, fmt.Errorf("failed to scan workhour details: %w", err)
		}
		details = append(details, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workhour details: %w", err)
	}

	return details, nil
}

// GetWorkhourDetailsByID retrieves a single workhour details by ID
func GetWorkhourDetailsByID(id int) (*domain.WorkhourDetails, error) {
	var d domain.WorkhourDetails
	err := db.QueryRow("SELECT id, name, short_name, is_work FROM workhour_details WHERE id = ?", id).
		Scan(&d.ID, &d.Name, &d.ShortName, &d.IsWork)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workhour details: %w", err)
	}

	return &d, nil
}

// CreateWorkhourDetails inserts new workhour details into the database
func CreateWorkhourDetails(details domain.WorkhourDetails) error {
	_, err := db.Exec(
		"INSERT INTO workhour_details (id, name, short_name, is_work) VALUES (?, ?, ?, ?)",
		details.ID, details.Name, details.ShortName, details.IsWork,
	)
	if err != nil {
		return fmt.Errorf("failed to create workhour details: %w", err)
	}
	return nil
}

// UpdateWorkhourDetails updates existing workhour details
func UpdateWorkhourDetails(details domain.WorkhourDetails) error {
	result, err := db.Exec(
		"UPDATE workhour_details SET name = ?, short_name = ?, is_work = ? WHERE id = ?",
		details.Name, details.ShortName, details.IsWork, details.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update workhour details: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("workhour details not found")
	}

	return nil
}

// DeleteWorkhourDetails deletes workhour details by ID
func DeleteWorkhourDetails(id int) error {
	result, err := db.Exec("DELETE FROM workhour_details WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete workhour details: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("workhour details not found")
	}

	return nil
}

// SeedWorkhourDetails seeds the initial workhour details into the database
func SeedWorkhourDetails() error {
	// Check if workhour details already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM workhour_details").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check workhour details: %w", err)
	}

	// Only seed if table is empty
	if count > 0 {
		return nil
	}

	details := fetchAllWorkhourDetails()
	for _, d := range details {
		if err := CreateWorkhourDetails(d); err != nil {
			return fmt.Errorf("failed to seed workhour details: %w", err)
		}
	}

	return nil
}
