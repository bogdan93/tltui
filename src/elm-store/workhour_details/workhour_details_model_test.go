package workhour_details

import (
	"testing"
	"tltui/src/domain/repository"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWorkhourDetailsModel_HandleWorkhourDetailCreated(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	tests := []struct {
		name          string
		msg           WorkhourDetailsCreatedMsg
		wantCount     int
		wantModalClosed bool
	}{
		{
			name:            "creates workhour detail successfully",
			msg:             WorkhourDetailsCreatedMsg{Name: "Test Detail", ShortName: "TD", IsWork: true},
			wantCount:       1,
			wantModalClosed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewWorkhourDetailsModel()

			updatedModel, _ := m.handleWorkhourDetailCreated(tt.msg)

			if len(updatedModel.WorkhourDetails) != tt.wantCount {
				t.Errorf("got %d workhour details, want %d", len(updatedModel.WorkhourDetails), tt.wantCount)
			}

			if tt.wantModalClosed && updatedModel.ActiveModal != nil {
				t.Error("expected modal to be closed after creation")
			}
		})
	}
}

func TestWorkhourDetailsModel_HandleWorkhourDetailEdited(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create initial workhour detail
	initialDetail := repository.CreateTestWorkhourDetails(t, 1, "Old Name", "ON", true)

	m := NewWorkhourDetailsModel()

	msg := WorkhourDetailsEditedMsg{
		WorkhourDetailID: initialDetail.ID,
		Name:             "New Name",
		ShortName:        "NN",
		IsWork:           false,
	}

	updatedModel, _ := m.handleWorkhourDetailEdited(msg)

	if updatedModel.ActiveModal != nil {
		t.Error("expected modal to be closed after edit")
	}

	// Verify workhour detail was updated in database
	details, _ := repository.GetAllWorkhourDetailsFromDB()
	if len(details) != 1 {
		t.Fatalf("expected 1 workhour detail, got %d", len(details))
	}

	if details[0].Name != "New Name" {
		t.Errorf("got name %q, want %q", details[0].Name, "New Name")
	}
	if details[0].ShortName != "NN" {
		t.Errorf("got short name %q, want %q", details[0].ShortName, "NN")
	}
	if details[0].IsWork != false {
		t.Errorf("got IsWork %v, want %v", details[0].IsWork, false)
	}
}

func TestWorkhourDetailsModel_HandleWorkhourDetailDeleted(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create workhour detail to delete
	detail := repository.CreateTestWorkhourDetails(t, 1, "Test Detail", "TD", true)

	m := NewWorkhourDetailsModel()

	msg := WorkhourDetailsDeletedMsg{WorkhourDetailID: detail.ID}

	updatedModel, _ := m.handleWorkhourDetailDeleted(msg)

	if updatedModel.ActiveModal != nil {
		t.Error("expected modal to be closed after delete")
	}

	// Verify workhour detail was deleted
	details, _ := repository.GetAllWorkhourDetailsFromDB()
	if len(details) != 0 {
		t.Errorf("expected 0 workhour details after delete, got %d", len(details))
	}
}

func TestWorkhourDetailsModel_Update_WindowResize(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	m := NewWorkhourDetailsModel()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	updatedModel, _ := m.Update(msg)
	wdm := updatedModel.(WorkhourDetailsModel)

	if wdm.Width != 100 {
		t.Errorf("got width %d, want 100", wdm.Width)
	}
	if wdm.Height != 50 {
		t.Errorf("got height %d, want 50", wdm.Height)
	}
}

func TestWorkhourDetailsModel_Update_OpenCreateModal(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	m := NewWorkhourDetailsModel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}

	updatedModel, _ := m.Update(msg)
	wdm := updatedModel.(WorkhourDetailsModel)

	if wdm.ActiveModal == nil {
		t.Error("expected create modal to be open")
	}

	if _, ok := wdm.ActiveModal.(WorkhourDetailsCreateModalWrapper); !ok {
		t.Error("expected WorkhourDetailsCreateModalWrapper")
	}
}

func TestWorkhourDetailsModel_Update_OpenEditModal(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create a workhour detail to edit
	repository.CreateTestWorkhourDetails(t, 1, "Test Detail", "TD", true)

	m := NewWorkhourDetailsModel()

	// Press enter to open edit modal
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	updatedModel, _ := m.Update(msg)
	wdm := updatedModel.(WorkhourDetailsModel)

	if wdm.ActiveModal == nil {
		t.Error("expected edit modal to be open")
	}

	if _, ok := wdm.ActiveModal.(WorkhourDetailsEditModalWrapper); !ok {
		t.Error("expected WorkhourDetailsEditModalWrapper")
	}
}

func TestWorkhourDetailsModel_Update_OpenDeleteModal(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create a workhour detail to delete
	repository.CreateTestWorkhourDetails(t, 1, "Test Detail", "TD", true)

	m := NewWorkhourDetailsModel()

	// Press 'd' to open delete modal
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}

	updatedModel, _ := m.Update(msg)
	wdm := updatedModel.(WorkhourDetailsModel)

	if wdm.ActiveModal == nil {
		t.Error("expected delete modal to be open")
	}

	if _, ok := wdm.ActiveModal.(WorkhourDetailsDeleteModalWrapper); !ok {
		t.Error("expected WorkhourDetailsDeleteModalWrapper")
	}
}

func TestWorkhourDetailsModel_GetSelectedWorkhourDetail(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create some workhour details
	repository.CreateTestWorkhourDetails(t, 1, "Detail 1", "D1", true)
	repository.CreateTestWorkhourDetails(t, 2, "Detail 2", "D2", false)

	m := NewWorkhourDetailsModel()

	// Should return first workhour detail (cursor at 0)
	selected := m.getSelectedWorkhourDetail()
	if selected == nil {
		t.Fatal("expected selected workhour detail, got nil")
	}

	if selected.Name != "Detail 1" {
		t.Errorf("got name %q, want %q", selected.Name, "Detail 1")
	}
}
