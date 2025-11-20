package projects

import (
	"testing"
	"tltui/src/domain/repository"

	tea "github.com/charmbracelet/bubbletea"
)

func TestProjectsModel_HandleProjectCreated(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	tests := []struct {
		name              string
		msg               ProjectCreatedMsg
		wantProjectCount  int
		wantModalClosed   bool
	}{
		{
			name:             "creates project successfully",
			msg:              ProjectCreatedMsg{Name: "Test Project", OdooID: 123},
			wantProjectCount: 1,
			wantModalClosed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewProjectsModel()

			updatedModel, _ := m.handleProjectCreated(tt.msg)

			if len(updatedModel.Projects) != tt.wantProjectCount {
				t.Errorf("got %d projects, want %d", len(updatedModel.Projects), tt.wantProjectCount)
			}

			if tt.wantModalClosed && updatedModel.ActiveModal != nil {
				t.Error("expected modal to be closed after creation")
			}
		})
	}
}

func TestProjectsModel_HandleProjectEdited(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create initial project
	initialProject := repository.CreateTestProject(t, 1, "Old Name", 100)

	m := NewProjectsModel()

	msg := ProjectEditedMsg{
		ProjectID: initialProject.ID,
		Name:      "New Name",
		OdooID:    200,
	}

	updatedModel, _ := m.handleProjectEdited(msg)

	if updatedModel.ActiveModal != nil {
		t.Error("expected modal to be closed after edit")
	}

	// Verify project was updated in database
	projects, _ := repository.GetAllProjectsFromDB()
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	if projects[0].Name != "New Name" {
		t.Errorf("got name %q, want %q", projects[0].Name, "New Name")
	}
	if projects[0].OdooID != 200 {
		t.Errorf("got odooID %d, want %d", projects[0].OdooID, 200)
	}
}

func TestProjectsModel_HandleProjectDeleted(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create project to delete
	project := repository.CreateTestProject(t, 1, "Test Project", 100)

	m := NewProjectsModel()

	msg := ProjectDeletedMsg{ProjectID: project.ID}

	updatedModel, _ := m.handleProjectDeleted(msg)

	if updatedModel.ActiveModal != nil {
		t.Error("expected modal to be closed after delete")
	}

	// Verify project was deleted
	projects, _ := repository.GetAllProjectsFromDB()
	if len(projects) != 0 {
		t.Errorf("expected 0 projects after delete, got %d", len(projects))
	}
}

func TestProjectsModel_Update_WindowResize(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	m := NewProjectsModel()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	updatedModel, _ := m.Update(msg)
	pm := updatedModel.(ProjectsModel)

	if pm.Width != 100 {
		t.Errorf("got width %d, want 100", pm.Width)
	}
	if pm.Height != 50 {
		t.Errorf("got height %d, want 50", pm.Height)
	}
}

func TestProjectsModel_Update_OpenCreateModal(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	m := NewProjectsModel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}

	updatedModel, _ := m.Update(msg)
	pm := updatedModel.(ProjectsModel)

	if pm.ActiveModal == nil {
		t.Error("expected create modal to be open")
	}

	if _, ok := pm.ActiveModal.(ProjectCreateModalWrapper); !ok {
		t.Error("expected ProjectCreateModalWrapper")
	}
}

func TestProjectsModel_Update_OpenEditModal(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create a project to edit
	repository.CreateTestProject(t, 1, "Test Project", 100)

	m := NewProjectsModel()

	// Press enter to open edit modal
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	updatedModel, _ := m.Update(msg)
	pm := updatedModel.(ProjectsModel)

	if pm.ActiveModal == nil {
		t.Error("expected edit modal to be open")
	}

	if _, ok := pm.ActiveModal.(ProjectEditModalWrapper); !ok {
		t.Error("expected ProjectEditModalWrapper")
	}
}

func TestProjectsModel_Update_OpenDeleteModal(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create a project to delete
	repository.CreateTestProject(t, 1, "Test Project", 100)

	m := NewProjectsModel()

	// Press 'd' to open delete modal
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}

	updatedModel, _ := m.Update(msg)
	pm := updatedModel.(ProjectsModel)

	if pm.ActiveModal == nil {
		t.Error("expected delete modal to be open")
	}

	if _, ok := pm.ActiveModal.(ProjectDeleteModalWrapper); !ok {
		t.Error("expected ProjectDeleteModalWrapper")
	}
}

func TestProjectsModel_GetSelectedProject(t *testing.T) {
	cleanup := repository.SetupTest(t)
	defer cleanup()

	// Create some projects
	repository.CreateTestProject(t, 1, "Project 1", 100)
	repository.CreateTestProject(t, 2, "Project 2", 200)

	m := NewProjectsModel()

	// Should return first project (cursor at 0)
	selected := m.getSelectedProject()
	if selected == nil {
		t.Fatal("expected selected project, got nil")
	}

	if selected.Name != "Project 1" {
		t.Errorf("got name %q, want %q", selected.Name, "Project 1")
	}
}
