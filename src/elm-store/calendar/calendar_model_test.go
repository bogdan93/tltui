package calendar

import (
	"testing"
	"time"
	"tltui/src/domain/repository"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCalendarModel_HandleWorkhourCreated(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create test data
	detail := repository.CreateTestWorkhourDetails(t, 1, "Test Detail", "TD", true)
	project := repository.CreateTestProject(t, 1, "Test Project", 100)

	m := NewCalendarModel()
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	msg := WorkhourCreateSubmittedMsg{
		Date:      date,
		DetailsID: detail.ID,
		ProjectID: project.ID,
		Hours:     8.0,
	}

	updatedModel, _ := m.handleWorkhourCreated(msg)

	// Verify workhour was created in database
	workhours := updatedModel.getWorkhoursForDate(date)
	if len(workhours) != 1 {
		t.Fatalf("expected 1 workhour, got %d", len(workhours))
	}

	if workhours[0].Hours != 8.0 {
		t.Errorf("got hours %f, want 8.0", workhours[0].Hours)
	}
}

func TestCalendarModel_HandleWorkhourEdited(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create test data
	detail := repository.CreateTestWorkhourDetails(t, 1, "Test Detail", "TD", true)
	project := repository.CreateTestProject(t, 1, "Test Project", 100)
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	workhour := repository.CreateTestWorkhour(t, date, detail.ID, project.ID, 5.0)

	m := NewCalendarModel()

	msg := WorkhourEditSubmittedMsg{
		WorkhourID: workhour.ID,
		Date:       date,
		DetailsID:  detail.ID,
		ProjectID:  project.ID,
		Hours:      8.5,
	}

	updatedModel, _ := m.handleWorkhourEdited(msg)

	// Verify workhour was updated in database
	workhours := updatedModel.getWorkhoursForDate(date)
	if len(workhours) != 1 {
		t.Fatalf("expected 1 workhour, got %d", len(workhours))
	}

	if workhours[0].Hours != 8.5 {
		t.Errorf("got hours %f, want 8.5", workhours[0].Hours)
	}
}

func TestCalendarModel_HandleWorkhourDeleted(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create test data
	detail := repository.CreateTestWorkhourDetails(t, 1, "Test Detail", "TD", true)
	project := repository.CreateTestProject(t, 1, "Test Project", 100)
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	workhour := repository.CreateTestWorkhour(t, date, detail.ID, project.ID, 8.0)

	m := NewCalendarModel()

	msg := WorkhourDeleteConfirmedMsg{ID: workhour.ID}

	updatedModel, _ := m.handleWorkhourDeleted(msg)

	// Verify workhour was deleted
	workhours := updatedModel.getWorkhoursForDate(date)
	if len(workhours) != 0 {
		t.Errorf("expected 0 workhours after delete, got %d", len(workhours))
	}
}

func TestCalendarModel_HandleYankWorkhours(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create test data
	detail := repository.CreateTestWorkhourDetails(t, 1, "Test Detail", "TD", true)
	project := repository.CreateTestProject(t, 1, "Test Project", 100)
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	repository.CreateTestWorkhour(t, date, detail.ID, project.ID, 8.0)

	m := NewCalendarModel()
	m.SelectedDate = date

	updatedModel, _ := m.handleYankWorkhours()

	if len(updatedModel.YankedWorkhours) != 1 {
		t.Errorf("expected 1 yanked workhour, got %d", len(updatedModel.YankedWorkhours))
	}

	if !updatedModel.YankedFromDate.Equal(date) {
		t.Errorf("YankedFromDate should be %v, got %v", date, updatedModel.YankedFromDate)
	}
}

func TestCalendarModel_HandlePasteWorkhours(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create test data
	detail := repository.CreateTestWorkhourDetails(t, 1, "Test Detail", "TD", true)
	project := repository.CreateTestProject(t, 1, "Test Project", 100)
	sourceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	targetDate := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)
	repository.CreateTestWorkhour(t, sourceDate, detail.ID, project.ID, 8.0)

	m := NewCalendarModel()
	m.SelectedDate = sourceDate

	// First yank
	m, _ = m.handleYankWorkhours()

	// Then move to target date and paste
	m.SelectedDate = targetDate
	updatedModel, _ := m.handlePasteWorkhours()

	// Verify workhours were pasted to target date
	workhours := updatedModel.getWorkhoursForDate(targetDate)
	if len(workhours) != 1 {
		t.Fatalf("expected 1 workhour on target date, got %d", len(workhours))
	}

	if workhours[0].Hours != 8.0 {
		t.Errorf("got hours %f, want 8.0", workhours[0].Hours)
	}
}

func TestCalendarModel_HandleDeleteWorkhours(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create test data
	detail := repository.CreateTestWorkhourDetails(t, 1, "Test Detail", "TD", true)
	project := repository.CreateTestProject(t, 1, "Test Project", 100)
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	repository.CreateTestWorkhour(t, date, detail.ID, project.ID, 8.0)

	m := NewCalendarModel()
	m.SelectedDate = date

	updatedModel, _ := m.handleDeleteWorkhours()

	// Verify workhours were deleted
	workhours := updatedModel.getWorkhoursForDate(date)
	if len(workhours) != 0 {
		t.Errorf("expected 0 workhours after delete, got %d", len(workhours))
	}
}

func TestCalendarModel_Update_WindowResize(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	m := NewCalendarModel()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	updatedModel, _ := m.Update(msg)
	cm := updatedModel.(CalendarModel)

	if cm.Width != 100 {
		t.Errorf("got width %d, want 100", cm.Width)
	}
	if cm.Height != 50 {
		t.Errorf("got height %d, want 50", cm.Height)
	}
}

func TestCalendarModel_Update_NavigationLeft(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	m := NewCalendarModel()
	initialDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	m.SelectedDate = initialDate
	m.ViewMonth = 1
	m.ViewYear = 2024

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}

	updatedModel, _ := m.Update(msg)
	cm := updatedModel.(CalendarModel)

	expectedDate := initialDate.AddDate(0, 0, -1)
	if !cm.SelectedDate.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, cm.SelectedDate)
	}
}

func TestCalendarModel_Update_NavigationRight(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	m := NewCalendarModel()
	initialDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	m.SelectedDate = initialDate
	m.ViewMonth = 1
	m.ViewYear = 2024

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}

	updatedModel, _ := m.Update(msg)
	cm := updatedModel.(CalendarModel)

	expectedDate := initialDate.AddDate(0, 0, 1)
	if !cm.SelectedDate.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, cm.SelectedDate)
	}
}

func TestCalendarModel_Update_NavigationUp(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	m := NewCalendarModel()
	initialDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	m.SelectedDate = initialDate
	m.ViewMonth = 1
	m.ViewYear = 2024

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}

	updatedModel, _ := m.Update(msg)
	cm := updatedModel.(CalendarModel)

	expectedDate := initialDate.AddDate(0, 0, -7)
	if !cm.SelectedDate.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, cm.SelectedDate)
	}
}

func TestCalendarModel_Update_NavigationDown(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	m := NewCalendarModel()
	initialDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	m.SelectedDate = initialDate
	m.ViewMonth = 1
	m.ViewYear = 2024

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}

	updatedModel, _ := m.Update(msg)
	cm := updatedModel.(CalendarModel)

	expectedDate := initialDate.AddDate(0, 0, 7)
	if !cm.SelectedDate.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, cm.SelectedDate)
	}
}

func TestCalendarModel_Update_HelpToggle(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	m := NewCalendarModel()

	// Open help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updatedModel, _ := m.Update(msg)
	cm := updatedModel.(CalendarModel)

	if !cm.ShowHelp {
		t.Error("expected ShowHelp to be true")
	}

	// Close help
	updatedModel, _ = cm.Update(msg)
	cm = updatedModel.(CalendarModel)

	if cm.ShowHelp {
		t.Error("expected ShowHelp to be false")
	}
}

func TestCalendarModel_ResetToCurrentMonth(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	m := NewCalendarModel()
	// Set to a different month
	m.SelectedDate = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	m.ViewMonth = 1
	m.ViewYear = 2024

	m.ResetToCurrentMonth()

	now := time.Now()
	if m.ViewMonth != int(now.Month()) {
		t.Errorf("expected ViewMonth %d, got %d", int(now.Month()), m.ViewMonth)
	}
	if m.ViewYear != now.Year() {
		t.Errorf("expected ViewYear %d, got %d", now.Year(), m.ViewYear)
	}
}
