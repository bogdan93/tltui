package repository

import (
	"testing"
	"time"
	"tltui/src/domain"
)

// SetupTest initializes test database and returns cleanup function
func SetupTest(t *testing.T) func() {
	InitTestDB(t)
	return func() {
		ClearTestData(t)
		CleanupTestDB(t)
	}
}

// CreateTestProject creates a project for testing
func CreateTestProject(t *testing.T, id int, name string, odooID int) domain.Project {
	p := domain.Project{
		ID:     id,
		Name:   name,
		OdooID: odooID,
	}
	if err := CreateProject(p); err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}
	return p
}

// CreateTestWorkhourDetails creates workhour details for testing
func CreateTestWorkhourDetails(t *testing.T, id int, name, shortName string, isWork bool) domain.WorkhourDetails {
	wd := domain.WorkhourDetails{
		ID:        id,
		Name:      name,
		ShortName: shortName,
		IsWork:    isWork,
	}
	if err := CreateWorkhourDetails(wd); err != nil {
		t.Fatalf("failed to create test workhour details: %v", err)
	}
	return wd
}

// CreateTestWorkhour creates a workhour for testing
func CreateTestWorkhour(t *testing.T, date time.Time, detailsID, projectID int, hours float64) domain.Workhour {
	wh := domain.Workhour{
		Date:      date,
		DetailsID: detailsID,
		ProjectID: projectID,
		Hours:     hours,
	}
	id, err := CreateWorkhour(wh)
	if err != nil {
		t.Fatalf("failed to create test workhour: %v", err)
	}
	wh.ID = id
	return wh
}
