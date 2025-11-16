package repository

import (
	"database/sql"
	"fmt"
	"tltui/src/domain"
)

func GetAllProjectsFromDB() ([]domain.Project, error) {
	rows, err := db.Query("SELECT id, odoo_id, name FROM projects ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.OdooID, &p.Name); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	return projects, nil
}

func GetProjectByID(id int) (*domain.Project, error) {
	var p domain.Project
	err := db.QueryRow("SELECT id, odoo_id, name FROM projects WHERE id = ?", id).
		Scan(&p.ID, &p.OdooID, &p.Name)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &p, nil
}

func CreateProject(project domain.Project) error {
	_, err := db.Exec(
		"INSERT INTO projects (id, odoo_id, name) VALUES (?, ?, ?)",
		project.ID, project.OdooID, project.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

func UpdateProject(project domain.Project) error {
	result, err := db.Exec(
		"UPDATE projects SET odoo_id = ?, name = ? WHERE id = ?",
		project.OdooID, project.Name, project.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

func DeleteProject(id int) error {
	result, err := db.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

func SeedProjects() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check projects: %w", err)
	}

	if count > 0 {
		return nil
	}

	projects := fetchAllProjects()
	for _, p := range projects {
		if err := CreateProject(p); err != nil {
			return fmt.Errorf("failed to seed project: %w", err)
		}
	}

	return nil
}
