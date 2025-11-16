package models

import (
	"fmt"
	"tltui/src/domain/repository"
	"tltui/src/render"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkhourDetailsModel struct {
	Width  int
	Height int

	WorkhourDetailsEditModal   *WorkhourDetailsEditModal
	WorkhourDetailsCreateModal *WorkhourDetailsCreateModal
	WorkhourDetailsDeleteModal *WorkhourDetailsDeleteModal

	WorkhourDetailsTable    table.Model
	WorkhourDetailsViewport viewport.Model
	WorkhourDetails         []WorkhourDetails
	NextID                  int // Track next available ID for new workhour details
}

func NewWorkhourDetailsModel() WorkhourDetailsModel {
	m := WorkhourDetailsModel{}

	// Load from database
	workhourDetails, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		// Fallback to empty if error
		workhourDetails = []WorkhourDetails{}
	}
	m.WorkhourDetails = workhourDetails

	// Calculate next available ID
	m.NextID = 1
	for _, wd := range m.WorkhourDetails {
		if wd.ID >= m.NextID {
			m.NextID = wd.ID + 1
		}
	}

	// Setup table columns
	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Name", Width: 25},
		{Title: "Short Name", Width: 15},
		{Title: "Is Work", Width: 10},
	}

	// Convert workhour details to table rows
	rows := []table.Row{}
	for _, wd := range m.WorkhourDetails {
		isWorkStr := "No"
		if wd.IsWork {
			isWorkStr = "Yes"
		}
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", wd.ID),
			wd.Name,
			wd.ShortName,
			isWorkStr,
		})
	}

	// Create table
	m.WorkhourDetailsTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
		table.WithWidth(60),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("39"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)

	m.WorkhourDetailsTable.SetStyles(s)
	m.WorkhourDetailsViewport = viewport.New(100, 100)
	m.WorkhourDetailsTable.SetHeight(100)

	return m
}

func (m WorkhourDetailsModel) Init() tea.Cmd {
	return nil
}

func (m WorkhourDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case WorkhourDetailsCreatedMsg:
		// Create new workhour detail with next available ID
		newWorkhourDetail := WorkhourDetails{
			ID:        m.NextID,
			Name:      msg.Name,
			ShortName: msg.ShortName,
			IsWork:    msg.IsWork,
		}

		// Save to database
		err := repository.CreateWorkhourDetails(newWorkhourDetail)
		if err != nil {
			m.WorkhourDetailsCreateModal = nil
			return m, DispatchErrorNotification(fmt.Sprintf("Failed to create workhour detail: %v", err))
		}

		// Reload from database
		details, err := repository.GetAllWorkhourDetailsFromDB()
		if err != nil {
			m.WorkhourDetailsCreateModal = nil
			return m, DispatchErrorNotification(fmt.Sprintf("Failed to reload workhour details: %v", err))
		}
		m.WorkhourDetails = details
		m.NextID++

		// Update table
		rows := []table.Row{}
		for _, wd := range m.WorkhourDetails {
			isWorkStr := "No"
			if wd.IsWork {
				isWorkStr = "Yes"
			}
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", wd.ID),
				wd.Name,
				wd.ShortName,
				isWorkStr,
			})
		}
		m.WorkhourDetailsTable.SetRows(rows)
		m.WorkhourDetailsCreateModal = nil
		return m, nil

	case WorkhourDetailsCreateCanceledMsg:
		m.WorkhourDetailsCreateModal = nil
		return m, nil

	case WorkhourDetailsDeletedMsg:
		// Delete from database
		err := repository.DeleteWorkhourDetails(msg.WorkhourDetailID)
		if err != nil {
			m.WorkhourDetailsDeleteModal = nil
			return m, DispatchErrorNotification(fmt.Sprintf("Failed to delete workhour detail: %v", err))
		}

		// Reload from database
		details, err := repository.GetAllWorkhourDetailsFromDB()
		if err != nil {
			m.WorkhourDetailsDeleteModal = nil
			return m, DispatchErrorNotification(fmt.Sprintf("Failed to reload workhour details: %v", err))
		}
		m.WorkhourDetails = details

		// Update table
		rows := []table.Row{}
		for _, wd := range m.WorkhourDetails {
			isWorkStr := "No"
			if wd.IsWork {
				isWorkStr = "Yes"
			}
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", wd.ID),
				wd.Name,
				wd.ShortName,
				isWorkStr,
			})
		}
		m.WorkhourDetailsTable.SetRows(rows)
		m.WorkhourDetailsDeleteModal = nil
		return m, nil

	case WorkhourDetailsDeleteCanceledMsg:
		m.WorkhourDetailsDeleteModal = nil
		return m, nil

	case WorkhourDetailsEditedMsg:
		// Update in database
		updatedWorkhourDetail := WorkhourDetails{
			ID:        msg.WorkhourDetailID,
			Name:      msg.Name,
			ShortName: msg.ShortName,
			IsWork:    msg.IsWork,
		}
		err := repository.UpdateWorkhourDetails(updatedWorkhourDetail)
		if err != nil {
			m.WorkhourDetailsEditModal = nil
			return m, DispatchErrorNotification(fmt.Sprintf("Failed to update workhour detail: %v", err))
		}

		// Reload from database
		details, err := repository.GetAllWorkhourDetailsFromDB()
		if err != nil {
			m.WorkhourDetailsEditModal = nil
			return m, DispatchErrorNotification(fmt.Sprintf("Failed to reload workhour details: %v", err))
		}
		m.WorkhourDetails = details

		rows := []table.Row{}
		for _, wd := range m.WorkhourDetails {
			isWorkStr := "No"
			if wd.IsWork {
				isWorkStr = "Yes"
			}
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", wd.ID),
				wd.Name,
				wd.ShortName,
				isWorkStr,
			})
		}
		m.WorkhourDetailsTable.SetRows(rows)
		m.WorkhourDetailsEditModal = nil
		return m, nil

	case WorkhourDetailsEditCanceledMsg:
		m.WorkhourDetailsEditModal = nil
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		verticalMargin := 12 // Increased to account for tab bar (4 lines) + padding
		tableHeight := msg.Height - verticalMargin
		m.WorkhourDetailsViewport.Width = msg.Width - 2
		m.WorkhourDetailsViewport.Height = tableHeight
		m.WorkhourDetailsTable.SetHeight(tableHeight)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Only close delete modal on 'q' (edit/create modals have text inputs where user might type 'q')
			if m.WorkhourDetailsDeleteModal != nil {
				m.WorkhourDetailsDeleteModal = nil
				return m, nil
			}
			// If edit/create modal is open, don't intercept - let the modal/textinput handle it
			if m.WorkhourDetailsEditModal != nil || m.WorkhourDetailsCreateModal != nil {
				break // Don't quit, let it pass through to modal forwarding
			}
			// No modal open, quit
			return m, tea.Quit

		case "n":
			if m.WorkhourDetailsEditModal == nil && m.WorkhourDetailsCreateModal == nil && m.WorkhourDetailsDeleteModal == nil {
				m.WorkhourDetailsCreateModal = NewWorkhourDetailsCreateModal()
				return m, nil
			}

		case "d":
			if m.WorkhourDetailsEditModal == nil && m.WorkhourDetailsCreateModal == nil && m.WorkhourDetailsDeleteModal == nil {
				selectedWorkhourDetail := m.getSelectedWorkhourDetail()
				if selectedWorkhourDetail != nil {
					m.WorkhourDetailsDeleteModal = NewWorkhourDetailsDeleteModal(
						selectedWorkhourDetail.ID,
						selectedWorkhourDetail.Name,
					)
					return m, nil
				}
			}

		case "enter":
			if m.WorkhourDetailsEditModal == nil && m.WorkhourDetailsCreateModal == nil && m.WorkhourDetailsDeleteModal == nil {
				selectedWorkhourDetail := m.getSelectedWorkhourDetail()
				if selectedWorkhourDetail != nil {
					m.WorkhourDetailsEditModal = NewWorkhourDetailsEditModal(
						selectedWorkhourDetail.ID,
						selectedWorkhourDetail.Name,
						selectedWorkhourDetail.ShortName,
						selectedWorkhourDetail.IsWork,
					)
					return m, nil
				}
			}
		}
	}

	if m.WorkhourDetailsEditModal != nil {
		_, cmd := m.WorkhourDetailsEditModal.Update(msg)
		return m, cmd
	}

	if m.WorkhourDetailsCreateModal != nil {
		_, cmd := m.WorkhourDetailsCreateModal.Update(msg)
		return m, cmd
	}

	if m.WorkhourDetailsDeleteModal != nil {
		_, cmd := m.WorkhourDetailsDeleteModal.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.WorkhourDetailsTable, cmd = m.WorkhourDetailsTable.Update(msg)
	cmds = append(cmds, cmd)

	m.WorkhourDetailsViewport, cmd = m.WorkhourDetailsViewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m WorkhourDetailsModel) View() string {
	helpText := render.RenderHelpText("↑/↓: navigate", "enter: edit", "n: new", "d: delete", "q: quit")
	m.WorkhourDetailsViewport.SetContent(m.WorkhourDetailsTable.View())

	if m.WorkhourDetailsEditModal != nil {
		return m.WorkhourDetailsEditModal.View(m.Width, m.Height)
	}

	if m.WorkhourDetailsCreateModal != nil {
		return m.WorkhourDetailsCreateModal.View(m.Width, m.Height)
	}

	if m.WorkhourDetailsDeleteModal != nil {
		return m.WorkhourDetailsDeleteModal.View(m.Width, m.Height)
	}

	return m.WorkhourDetailsViewport.View() + "\n" + helpText
}

// getSelectedWorkhourDetail returns the currently selected workhour detail from the table
func (m WorkhourDetailsModel) getSelectedWorkhourDetail() *WorkhourDetails {
	cursor := m.WorkhourDetailsTable.Cursor()
	if cursor >= 0 && cursor < len(m.WorkhourDetails) {
		return &m.WorkhourDetails[cursor]
	}
	return nil
}
