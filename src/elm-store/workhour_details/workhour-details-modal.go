package workhour_details

import tea "github.com/charmbracelet/bubbletea"

// WorkhourDetailsModal represents any modal in the WorkhourDetails view
type WorkhourDetailsModal interface {
	Update(tea.Msg) (WorkhourDetailsModal, tea.Cmd)
	View(width, height int) string
}

// WorkhourDetailsCreateModalWrapper wraps WorkhourDetailsCreateModal to implement WorkhourDetailsModal
type WorkhourDetailsCreateModalWrapper struct {
	*WorkhourDetailsCreateModal
}

func (w WorkhourDetailsCreateModalWrapper) Update(msg tea.Msg) (WorkhourDetailsModal, tea.Cmd) {
	_, cmd := w.WorkhourDetailsCreateModal.Update(msg)
	return w, cmd
}

func (w WorkhourDetailsCreateModalWrapper) View(width, height int) string {
	return w.WorkhourDetailsCreateModal.View(width, height)
}

// WorkhourDetailsEditModalWrapper wraps WorkhourDetailsEditModal to implement WorkhourDetailsModal
type WorkhourDetailsEditModalWrapper struct {
	*WorkhourDetailsEditModal
}

func (w WorkhourDetailsEditModalWrapper) Update(msg tea.Msg) (WorkhourDetailsModal, tea.Cmd) {
	_, cmd := w.WorkhourDetailsEditModal.Update(msg)
	return w, cmd
}

func (w WorkhourDetailsEditModalWrapper) View(width, height int) string {
	return w.WorkhourDetailsEditModal.View(width, height)
}

// WorkhourDetailsDeleteModalWrapper wraps WorkhourDetailsDeleteModal to implement WorkhourDetailsModal
type WorkhourDetailsDeleteModalWrapper struct {
	*WorkhourDetailsDeleteModal
}

func (w WorkhourDetailsDeleteModalWrapper) Update(msg tea.Msg) (WorkhourDetailsModal, tea.Cmd) {
	_, cmd := w.WorkhourDetailsDeleteModal.Update(msg)
	return w, cmd
}

func (w WorkhourDetailsDeleteModalWrapper) View(width, height int) string {
	return w.WorkhourDetailsDeleteModal.View(width, height)
}
