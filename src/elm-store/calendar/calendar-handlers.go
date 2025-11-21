package calendar

import (
	"fmt"
	"time"
	"tltui/src/common"
	"tltui/src/domain"
	"tltui/src/domain/repository"

	tea "github.com/charmbracelet/bubbletea"
)

func (m CalendarModel) handleWorkhourCreated(msg WorkhourCreateSubmittedMsg) (CalendarModel, tea.Cmd) {
	newWorkhour := domain.Workhour{
		Date:      msg.Date,
		DetailsID: msg.DetailsID,
		ProjectID: msg.ProjectID,
		Hours:     msg.Hours,
	}
	_, err := repository.CreateWorkhour(newWorkhour)
	if err != nil {
		return m, common.NotifyError("Failed to create workhour", err)
	}

	// Restore view modal and refresh data
	if m.ViewModalParent != nil && m.ViewModalParent.modal != nil {
		m.ViewModalParent.modal.Workhours = m.getWorkhoursForDate(m.ViewModalParent.modal.Date)
		m.ActiveModal = m.ViewModalParent
		m.ViewModalParent = nil
	} else {
		m.ActiveModal = nil
	}

	return m, nil
}

func (m CalendarModel) handleWorkhourEdited(msg WorkhourEditSubmittedMsg) (CalendarModel, tea.Cmd) {
	updatedWorkhour := domain.Workhour{
		Date:      msg.Date,
		DetailsID: msg.DetailsID,
		ProjectID: msg.ProjectID,
		Hours:     msg.Hours,
	}
	err := repository.UpdateWorkhour(msg.WorkhourID, updatedWorkhour)
	if err != nil {
		return m, common.NotifyError("Failed to update workhour", err)
	}

	// Restore view modal and refresh data
	if m.ViewModalParent != nil && m.ViewModalParent.modal != nil {
		m.ViewModalParent.modal.Workhours = m.getWorkhoursForDate(m.ViewModalParent.modal.Date)
		m.ActiveModal = m.ViewModalParent
		m.ViewModalParent = nil
	} else {
		m.ActiveModal = nil
	}

	return m, nil
}

func (m CalendarModel) handleWorkhourDeleted(msg WorkhourDeleteConfirmedMsg) (CalendarModel, tea.Cmd) {
	err := repository.DeleteWorkhour(msg.ID)
	if err != nil {
		return m, common.NotifyError("Failed to delete workhour", err)
	}

	// Restore view modal and refresh data
	if m.ViewModalParent != nil && m.ViewModalParent.modal != nil {
		m.ViewModalParent.modal.Workhours = m.getWorkhoursForDate(m.ViewModalParent.modal.Date)
		// Adjust selected index if needed
		if m.ViewModalParent.modal.SelectedWorkhourIndex >= len(m.ViewModalParent.modal.Workhours) && len(m.ViewModalParent.modal.Workhours) > 0 {
			m.ViewModalParent.modal.SelectedWorkhourIndex = len(m.ViewModalParent.modal.Workhours) - 1
		}
		m.ActiveModal = m.ViewModalParent
		m.ViewModalParent = nil
	} else {
		m.ActiveModal = nil
	}

	return m, nil
}

func (m CalendarModel) handleWorkhourCreateRequest(msg WorkhoursViewModalCreateRequestedMsg) (CalendarModel, tea.Cmd) {
	if viewWrapper, ok := m.ActiveModal.(*WorkhoursViewModalWrapper); ok {
		m.ViewModalParent = viewWrapper
	}

	workhourDetails, _ := repository.GetAllWorkhourDetailsFromDB()
	projects, _ := repository.GetAllProjectsFromDB()
	m.ActiveModal = &WorkhourCreateModalWrapper{
		modal: NewWorkhourCreateModal(msg.Date, workhourDetails, projects),
	}
	return m, nil
}

func (m CalendarModel) handleWorkhourEditRequest(msg WorkhoursViewModalEditRequestedMsg) (CalendarModel, tea.Cmd) {
	// Save current view modal
	if viewWrapper, ok := m.ActiveModal.(*WorkhoursViewModalWrapper); ok {
		m.ViewModalParent = viewWrapper
	}

	workhourDetails, _ := repository.GetAllWorkhourDetailsFromDB()
	projects, _ := repository.GetAllProjectsFromDB()
	workhours := m.getWorkhoursForDate(msg.Date)

	// Find the specific workhour
	var currentWorkhour *domain.Workhour
	for _, wh := range workhours {
		if wh.ID == msg.WorkhourID {
			currentWorkhour = &wh
			break
		}
	}

	if currentWorkhour != nil {
		m.ActiveModal = &WorkhourEditModalWrapper{
			modal: NewWorkhourEditModal(
				msg.WorkhourID,
				msg.Date,
				currentWorkhour.DetailsID,
				currentWorkhour.ProjectID,
				currentWorkhour.Hours,
				workhourDetails,
				projects,
			),
		}
	}
	return m, nil
}

func (m CalendarModel) handleWorkhourDeleteRequest(msg WorkhoursViewModalDeleteRequestedMsg) (CalendarModel, tea.Cmd) {
	// Save current view modal
	if viewWrapper, ok := m.ActiveModal.(*WorkhoursViewModalWrapper); ok {
		m.ViewModalParent = viewWrapper
	}

	workhours := m.getWorkhoursForDate(msg.Date)

	// Find the specific workhour
	var currentWorkhour *domain.Workhour
	for _, wh := range workhours {
		if wh.ID == msg.WorkhourID {
			currentWorkhour = &wh
			break
		}
	}

	if currentWorkhour != nil {
		workhourDetails, _ := repository.GetAllWorkhourDetailsFromDB()
		projects, _ := repository.GetAllProjectsFromDB()
		m.ActiveModal = &WorkhourDeleteModalWrapper{
			modal: NewWorkhourDeleteModal(
				msg.Date,
				*currentWorkhour,
				workhourDetails,
				projects,
			),
		}
	}
	return m, nil
}

func (m CalendarModel) handleYankWorkhours() (CalendarModel, tea.Cmd) {
	if len(m.getWorkhoursForDate(m.SelectedDate)) == 0 {
		return m, nil
	}
	m.YankedWorkhours = m.getWorkhoursForDate(m.SelectedDate)
	m.YankedFromDate = m.SelectedDate
	return m, common.NotifySuccess(fmt.Sprintf("📋 Copied %d workhour(s) from %s", len(m.YankedWorkhours), m.SelectedDate.Format("2006-01-02")))
}

func (m CalendarModel) handlePasteWorkhours() (CalendarModel, tea.Cmd) {
	if len(m.YankedWorkhours) == 0 {
		return m, nil
	}

	err := repository.DeleteWorkhoursByDate(m.SelectedDate)
	if err != nil {
		return m, common.NotifyError("Failed to clear existing workhours", err)
	}

	for _, wh := range m.YankedWorkhours {
		newWorkhour := domain.Workhour{
			Date:      m.SelectedDate,
			DetailsID: wh.DetailsID,
			ProjectID: wh.ProjectID,
			Hours:     wh.Hours,
		}
		_, err := repository.CreateWorkhour(newWorkhour)
		if err != nil {
			return m, common.NotifyError("Failed to paste workhour", err)
		}
	}
	return m, nil
}

func (m CalendarModel) handleDeleteWorkhours() (CalendarModel, tea.Cmd) {
	err := repository.DeleteWorkhoursByDate(m.SelectedDate)
	if err != nil {
		return m, common.NotifyError("Failed to delete workhours", err)
	}
	if m.isSameDay(m.SelectedDate, m.YankedFromDate) {
		m.YankedWorkhours = nil
		m.YankedFromDate = time.Time{}
	}
	return m, nil
}

func (m CalendarModel) handleOpenReportGenerator() (CalendarModel, tea.Cmd) {
	if m.ActiveModal == nil {
		m.ActiveModal = &ReportGeneratorModalWrapper{
			modal: NewReportGeneratorModal(m.ViewMonth, m.ViewYear),
		}
	}
	return m, nil
}

func (m CalendarModel) handleOpenDayView() (CalendarModel, tea.Cmd) {
	if m.ActiveModal == nil {
		workhours := m.getWorkhoursForDate(m.SelectedDate)
		workhourDetails, _ := repository.GetAllWorkhourDetailsFromDB()
		projects, _ := repository.GetAllProjectsFromDB()
		m.ActiveModal = &WorkhoursViewModalWrapper{
			modal: NewWorkhoursViewModal(
				m.SelectedDate,
				workhours,
				workhourDetails,
				projects,
			),
		}
	}
	return m, nil
}
