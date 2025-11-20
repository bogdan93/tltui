package workhour_details

import (
	"fmt"
	"tltui/src/common"
	"tltui/src/domain"
	"tltui/src/domain/repository"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// Handler methods for WorkhourDetailsModel.Update
// These methods break up the Update function into logical chunks

// handleWorkhourDetailCreated handles the WorkhourDetailsCreatedMsg
func (m WorkhourDetailsModel) handleWorkhourDetailCreated(msg WorkhourDetailsCreatedMsg) (WorkhourDetailsModel, tea.Cmd) {
	newWorkhourDetail := domain.WorkhourDetails{
		ID:        m.NextID,
		Name:      msg.Name,
		ShortName: msg.ShortName,
		IsWork:    msg.IsWork,
	}

	err := repository.CreateWorkhourDetails(newWorkhourDetail)
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to create workhour detail: %v", err))
	}

	details, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to reload workhour details: %v", err))
	}
	m.WorkhourDetails = details
	m.NextID++

	m.updateTableRows()
	m.ActiveModal = nil
	return m, nil
}

// handleWorkhourDetailEdited handles the WorkhourDetailsEditedMsg
func (m WorkhourDetailsModel) handleWorkhourDetailEdited(msg WorkhourDetailsEditedMsg) (WorkhourDetailsModel, tea.Cmd) {
	updatedWorkhourDetail := domain.WorkhourDetails{
		ID:        msg.WorkhourDetailID,
		Name:      msg.Name,
		ShortName: msg.ShortName,
		IsWork:    msg.IsWork,
	}
	err := repository.UpdateWorkhourDetails(updatedWorkhourDetail)
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to update workhour detail: %v", err))
	}

	details, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to reload workhour details: %v", err))
	}
	m.WorkhourDetails = details

	m.updateTableRows()
	m.ActiveModal = nil
	return m, nil
}

// handleWorkhourDetailDeleted handles the WorkhourDetailsDeletedMsg
func (m WorkhourDetailsModel) handleWorkhourDetailDeleted(msg WorkhourDetailsDeletedMsg) (WorkhourDetailsModel, tea.Cmd) {
	err := repository.DeleteWorkhourDetails(msg.WorkhourDetailID)
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to delete workhour detail: %v", err))
	}

	details, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		m.ActiveModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to reload workhour details: %v", err))
	}
	m.WorkhourDetails = details

	m.updateTableRows()
	m.ActiveModal = nil
	return m, nil
}

// updateTableRows updates the table with current workhour details data
func (m *WorkhourDetailsModel) updateTableRows() {
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
}
