package workhour_details

import (
	"fmt"
	"tltui/src/common"
	"tltui/src/domain"
	"tltui/src/domain/repository"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

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
		return m, common.NotifyError("Failed to create workhour detail", err)
	}

	details, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		m.ActiveModal = nil
		return m, common.NotifyError("Failed to reload workhour details", err)
	}
	m.WorkhourDetails = details
	m.NextID++

	m.updateTableRows()
	m.ActiveModal = nil
	return m, nil
}

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
		return m, common.NotifyError("Failed to update workhour detail", err)
	}

	details, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		m.ActiveModal = nil
		return m, common.NotifyError("Failed to reload workhour details", err)
	}
	m.WorkhourDetails = details

	m.updateTableRows()
	m.ActiveModal = nil
	return m, nil
}

func (m WorkhourDetailsModel) handleWorkhourDetailDeleted(msg WorkhourDetailsDeletedMsg) (WorkhourDetailsModel, tea.Cmd) {
	err := repository.DeleteWorkhourDetails(msg.WorkhourDetailID)
	if err != nil {
		m.ActiveModal = nil
		return m, common.NotifyError("Failed to delete workhour detail", err)
	}

	details, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		m.ActiveModal = nil
		return m, common.NotifyError("Failed to reload workhour details", err)
	}
	m.WorkhourDetails = details

	m.updateTableRows()
	m.ActiveModal = nil
	return m, nil
}

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
	m.TableView.SetRows(rows)
}
