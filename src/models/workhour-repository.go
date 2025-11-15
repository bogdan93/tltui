package models

import (
	"fmt"
	"time"
)

// GetAllWorkhours retrieves all workhours from the database
func GetAllWorkhours() ([]Workhour, error) {
	rows, err := db.Query("SELECT id, date, details_id, project_id, hours FROM workhours ORDER BY date DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query workhours: %w", err)
	}
	defer rows.Close()

	var workhours []Workhour
	for rows.Next() {
		var wh Workhour
		var dateStr string
		if err := rows.Scan(&wh.ID, &dateStr, &wh.DetailsID, &wh.ProjectID, &wh.Hours); err != nil {
			return nil, fmt.Errorf("failed to scan workhour: %w", err)
		}

		// Convert date string to time.Time
		wh.Date, err = StringToDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}

		workhours = append(workhours, wh)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workhours: %w", err)
	}

	return workhours, nil
}

// GetWorkhoursByDate retrieves all workhours for a specific date
func GetWorkhoursByDate(date time.Time) ([]Workhour, error) {
	dateStr := DateToString(date)
	rows, err := db.Query(
		"SELECT id, date, details_id, project_id, hours FROM workhours WHERE date = ? ORDER BY id",
		dateStr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query workhours by date: %w", err)
	}
	defer rows.Close()

	var workhours []Workhour
	for rows.Next() {
		var wh Workhour
		var dbDateStr string
		if err := rows.Scan(&wh.ID, &dbDateStr, &wh.DetailsID, &wh.ProjectID, &wh.Hours); err != nil {
			return nil, fmt.Errorf("failed to scan workhour: %w", err)
		}

		// Convert date string to time.Time
		wh.Date, err = StringToDate(dbDateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}

		workhours = append(workhours, wh)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workhours: %w", err)
	}

	return workhours, nil
}

// GetWorkhoursByDateRange retrieves all workhours within a date range
func GetWorkhoursByDateRange(start, end time.Time) ([]Workhour, error) {
	startStr := DateToString(start)
	endStr := DateToString(end)

	rows, err := db.Query(
		"SELECT id, date, details_id, project_id, hours FROM workhours WHERE date BETWEEN ? AND ? ORDER BY date",
		startStr, endStr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query workhours by date range: %w", err)
	}
	defer rows.Close()

	var workhours []Workhour
	for rows.Next() {
		var wh Workhour
		var dateStr string
		if err := rows.Scan(&wh.ID, &dateStr, &wh.DetailsID, &wh.ProjectID, &wh.Hours); err != nil {
			return nil, fmt.Errorf("failed to scan workhour: %w", err)
		}

		// Convert date string to time.Time
		wh.Date, err = StringToDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}

		workhours = append(workhours, wh)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workhours: %w", err)
	}

	return workhours, nil
}

// CreateWorkhour inserts a new workhour into the database
func CreateWorkhour(workhour Workhour) (int, error) {
	dateStr := DateToString(workhour.Date)

	result, err := db.Exec(
		"INSERT INTO workhours (date, details_id, project_id, hours) VALUES (?, ?, ?, ?)",
		dateStr, workhour.DetailsID, workhour.ProjectID, workhour.Hours,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create workhour: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return int(id), nil
}

// UpdateWorkhour updates an existing workhour
func UpdateWorkhour(id int, workhour Workhour) error {
	dateStr := DateToString(workhour.Date)

	result, err := db.Exec(
		"UPDATE workhours SET date = ?, details_id = ?, project_id = ?, hours = ? WHERE id = ?",
		dateStr, workhour.DetailsID, workhour.ProjectID, workhour.Hours, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update workhour: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("workhour not found")
	}

	return nil
}

// DeleteWorkhour deletes a workhour by ID
func DeleteWorkhour(id int) error {
	result, err := db.Exec("DELETE FROM workhours WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete workhour: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("workhour not found")
	}

	return nil
}

// DeleteWorkhoursByDate deletes all workhours for a specific date
func DeleteWorkhoursByDate(date time.Time) error {
	dateStr := DateToString(date)
	_, err := db.Exec("DELETE FROM workhours WHERE date = ?", dateStr)
	if err != nil {
		return fmt.Errorf("failed to delete workhours by date: %w", err)
	}
	return nil
}
