package projects

import (
	"fmt"
	"tltui/src/common"
	"tltui/src/domain"
	"tltui/src/domain/repository"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// Handler methods for ProjectsModel.Update
// These methods break up the Update function into logical chunks

// handleProjectCreated handles the ProjectCreatedMsg
func (m ProjectsModel) handleProjectCreated(msg ProjectCreatedMsg) (ProjectsModel, tea.Cmd) {
	newProject := domain.Project{
		ID:     m.NextID,
		Name:   msg.Name,
		OdooID: msg.OdooID,
	}

	err := repository.CreateProject(newProject)
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to create project: %v", err))
	}

	projects, err := repository.GetAllProjectsFromDB()
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to reload projects: %v", err))
	}
	m.Projects = projects
	m.NextID++

	m.updateTableRows()
	m.ActiveModal = nil
	return m, nil
}

// handleProjectEdited handles the ProjectEditedMsg
func (m ProjectsModel) handleProjectEdited(msg ProjectEditedMsg) (ProjectsModel, tea.Cmd) {
	updatedProject := domain.Project{
		ID:     msg.ProjectID,
		Name:   msg.Name,
		OdooID: msg.OdooID,
	}
	err := repository.UpdateProject(updatedProject)
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to update project: %v", err))
	}

	projects, err := repository.GetAllProjectsFromDB()
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to reload projects: %v", err))
	}
	m.Projects = projects

	m.updateTableRows()
	m.ActiveModal = nil
	return m, nil
}

// handleProjectDeleted handles the ProjectDeletedMsg
func (m ProjectsModel) handleProjectDeleted(msg ProjectDeletedMsg) (ProjectsModel, tea.Cmd) {
	err := repository.DeleteProject(msg.ProjectID)
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to delete project: %v", err))
	}

	projects, err := repository.GetAllProjectsFromDB()
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to reload projects: %v", err))
	}
	m.Projects = projects

	m.updateTableRows()
	m.ActiveModal = nil
	return m, nil
}

// updateTableRows updates the table with current projects data
func (m *ProjectsModel) updateTableRows() {
	rows := []table.Row{}
	for _, p := range m.Projects {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.ID),
			p.Name,
			fmt.Sprintf("%d", p.OdooID),
		})
	}
	m.ProjectsTable.SetRows(rows)
}
