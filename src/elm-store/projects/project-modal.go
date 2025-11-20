package projects

import tea "github.com/charmbracelet/bubbletea"

// ProjectModal represents any modal in the Projects view
type ProjectModal interface {
	Update(tea.Msg) (ProjectModal, tea.Cmd)
	View(width, height int) string
}

// ProjectCreateModalWrapper wraps ProjectCreateModal to implement ProjectModal
type ProjectCreateModalWrapper struct {
	*ProjectCreateModal
}

func (w ProjectCreateModalWrapper) Update(msg tea.Msg) (ProjectModal, tea.Cmd) {
	_, cmd := w.ProjectCreateModal.Update(msg)
	return w, cmd
}

func (w ProjectCreateModalWrapper) View(width, height int) string {
	return w.ProjectCreateModal.View(width, height)
}

// ProjectEditModalWrapper wraps ProjectEditModal to implement ProjectModal
type ProjectEditModalWrapper struct {
	*ProjectEditModal
}

func (w ProjectEditModalWrapper) Update(msg tea.Msg) (ProjectModal, tea.Cmd) {
	_, cmd := w.ProjectEditModal.Update(msg)
	return w, cmd
}

func (w ProjectEditModalWrapper) View(width, height int) string {
	return w.ProjectEditModal.View(width, height)
}

// ProjectDeleteModalWrapper wraps ProjectDeleteModal to implement ProjectModal
type ProjectDeleteModalWrapper struct {
	*ProjectDeleteModal
}

func (w ProjectDeleteModalWrapper) Update(msg tea.Msg) (ProjectModal, tea.Cmd) {
	_, cmd := w.ProjectDeleteModal.Update(msg)
	return w, cmd
}

func (w ProjectDeleteModalWrapper) View(width, height int) string {
	return w.ProjectDeleteModal.View(width, height)
}
