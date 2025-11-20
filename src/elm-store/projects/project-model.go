package projects

import (
	"fmt"
	"tltui/src/common"
	"tltui/src/domain"
	"tltui/src/domain/repository"
	"tltui/src/render"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type ProjectsModel struct {
	Width  int
	Height int

	ActiveModal ProjectModal

	TableView common.TableView
	Projects  []domain.Project
	NextID    int
}

func NewProjectsModel() ProjectsModel {
	m := ProjectsModel{}

	projects, err := repository.GetAllProjectsFromDB()
	if err != nil {
		projects = []domain.Project{}
	}
	m.Projects = projects

	m.NextID = 1
	for _, p := range m.Projects {
		if p.ID >= m.NextID {
			m.NextID = p.ID + 1
		}
	}

	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Project Name", Width: 30},
		{Title: "Odoo ID", Width: 10},
	}

	rows := []table.Row{}
	for _, p := range m.Projects {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.ID),
			p.Name,
			fmt.Sprintf("%d", p.OdooID),
		})
	}

	m.TableView = common.NewTableView(columns, rows)
	m.TableView.Table.SetHeight(100)

	return m
}

func (m ProjectsModel) Init() tea.Cmd {
	return nil
}

func (m ProjectsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ProjectCreatedMsg:
		return m.handleProjectCreated(msg)

	case ProjectCreateCanceledMsg:
		m.ActiveModal = nil
		return m, nil

	case ProjectEditedMsg:
		return m.handleProjectEdited(msg)

	case ProjectEditCanceledMsg:
		m.ActiveModal = nil
		return m, nil

	case ProjectDeletedMsg:
		return m.handleProjectDeleted(msg)

	case ProjectDeleteCanceledMsg:
		m.ActiveModal = nil
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		verticalMargin := 12
		m.TableView.SetSize(msg.Width, msg.Height, verticalMargin)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Only close delete modal on 'q' (edit/create modals have text inputs where user might type 'q')
			if _, isDelete := m.ActiveModal.(ProjectDeleteModalWrapper); isDelete {
				m.ActiveModal = nil
				return m, nil
			}
			// If edit/create modal is open, don't intercept - let the modal/textinput handle it
			if m.ActiveModal != nil {
				break // Don't quit, let it pass through to modal forwarding
			}
			// No modal open, quit
			return m, tea.Quit

		case "n":
			if m.ActiveModal == nil {
				m.ActiveModal = ProjectCreateModalWrapper{NewProjectCreateModal()}
				return m, nil
			}

		case "d":
			if m.ActiveModal == nil {
				selectedProject := m.getSelectedProject()
				if selectedProject != nil {
					m.ActiveModal = ProjectDeleteModalWrapper{NewProjectDeleteModal(
						selectedProject.ID,
						selectedProject.Name,
					)}
					return m, nil
				}
			}

		case "enter":
			if m.ActiveModal == nil {
				selectedProject := m.getSelectedProject()
				if selectedProject != nil {
					m.ActiveModal = ProjectEditModalWrapper{NewProjectEditModal(
						selectedProject.ID,
						selectedProject.Name,
						selectedProject.OdooID,
					)}
					return m, nil
				}
			}
		}
	}

	if m.ActiveModal != nil {
		_, cmd := m.ActiveModal.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.TableView, cmd = m.TableView.Update(msg)
	return m, cmd
}

func (m ProjectsModel) View() string {
	helpText := render.RenderHelpText("↑/↓: navigate", "enter: edit", "n: new", "d: delete", "q: quit")

	if m.ActiveModal != nil {
		return m.ActiveModal.View(m.Width, m.Height)
	}

	return m.TableView.View() + "\n" + helpText
}

func (m ProjectsModel) getSelectedProject() *domain.Project {
	cursor := m.TableView.Cursor()
	if cursor >= 0 && cursor < len(m.Projects) {
		return &m.Projects[cursor]
	}
	return nil
}
