package calendar

import tea "github.com/charmbracelet/bubbletea"

// CalendarModal represents any modal that can be displayed in the calendar view
type CalendarModal interface {
	Update(msg tea.Msg) (CalendarModal, tea.Cmd)
	View(width, height int) string
}

// WorkhoursViewModalWrapper wraps WorkhoursViewModal to implement CalendarModal
type WorkhoursViewModalWrapper struct {
	modal *WorkhoursViewModal
}

func (w *WorkhoursViewModalWrapper) Update(msg tea.Msg) (CalendarModal, tea.Cmd) {
	if w.modal == nil {
		return nil, nil
	}
	updated, cmd := w.modal.Update(msg)
	w.modal = &updated
	return w, cmd
}

func (w *WorkhoursViewModalWrapper) View(width, height int) string {
	if w.modal == nil {
		return ""
	}
	return w.modal.View(width, height)
}

// WorkhourCreateModalWrapper wraps WorkhourCreateModal to implement CalendarModal
type WorkhourCreateModalWrapper struct {
	modal *WorkhourCreateModal
}

func (w *WorkhourCreateModalWrapper) Update(msg tea.Msg) (CalendarModal, tea.Cmd) {
	if w.modal == nil {
		return nil, nil
	}
	updated, cmd := w.modal.Update(msg)
	w.modal = &updated
	return w, cmd
}

func (w *WorkhourCreateModalWrapper) View(width, height int) string {
	if w.modal == nil {
		return ""
	}
	return w.modal.View(width, height)
}

// WorkhourEditModalWrapper wraps WorkhourEditModal to implement CalendarModal
type WorkhourEditModalWrapper struct {
	modal *WorkhourEditModal
}

func (w *WorkhourEditModalWrapper) Update(msg tea.Msg) (CalendarModal, tea.Cmd) {
	if w.modal == nil {
		return nil, nil
	}
	updated, cmd := w.modal.Update(msg)
	w.modal = &updated
	return w, cmd
}

func (w *WorkhourEditModalWrapper) View(width, height int) string {
	if w.modal == nil {
		return ""
	}
	return w.modal.View(width, height)
}

// WorkhourDeleteModalWrapper wraps WorkhourDeleteModal to implement CalendarModal
type WorkhourDeleteModalWrapper struct {
	modal *WorkhourDeleteModal
}

func (w *WorkhourDeleteModalWrapper) Update(msg tea.Msg) (CalendarModal, tea.Cmd) {
	if w.modal == nil {
		return nil, nil
	}
	updated, cmd := w.modal.Update(msg)
	w.modal = &updated
	return w, cmd
}

func (w *WorkhourDeleteModalWrapper) View(width, height int) string {
	if w.modal == nil {
		return ""
	}
	return w.modal.View(width, height)
}

// ReportGeneratorModalWrapper wraps ReportGeneratorModal to implement CalendarModal
type ReportGeneratorModalWrapper struct {
	modal *ReportGeneratorModal
}

func (w *ReportGeneratorModalWrapper) Update(msg tea.Msg) (CalendarModal, tea.Cmd) {
	if w.modal == nil {
		return nil, nil
	}
	updated, cmd := w.modal.Update(msg)
	w.modal = &updated
	return w, cmd
}

func (w *ReportGeneratorModalWrapper) View(width, height int) string {
	if w.modal == nil {
		return ""
	}
	return w.modal.View(width, height)
}
