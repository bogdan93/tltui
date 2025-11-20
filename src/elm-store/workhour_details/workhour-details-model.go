package workhour_details

import (
	"fmt"
	"tltui/src/common"
	"tltui/src/domain"
	"tltui/src/domain/repository"
	"tltui/src/render"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type WorkhourDetailsModel struct {
	Width  int
	Height int

	ActiveModal WorkhourDetailsModal

	TableView       common.TableView
	WorkhourDetails []domain.WorkhourDetails
	NextID          int
}

func NewWorkhourDetailsModel() WorkhourDetailsModel {
	m := WorkhourDetailsModel{}

	workhourDetails, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		workhourDetails = []domain.WorkhourDetails{}
	}
	m.WorkhourDetails = workhourDetails

	m.NextID = 1
	for _, wd := range m.WorkhourDetails {
		if wd.ID >= m.NextID {
			m.NextID = wd.ID + 1
		}
	}

	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Name", Width: 25},
		{Title: "Short Name", Width: 15},
		{Title: "Is Work", Width: 10},
	}

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

	m.TableView = common.NewTableView(columns, rows)
	m.TableView.Table.SetHeight(100)

	return m
}

func (m WorkhourDetailsModel) Init() tea.Cmd {
	return nil
}

func (m WorkhourDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case WorkhourDetailsCreatedMsg:
		return m.handleWorkhourDetailCreated(msg)

	case WorkhourDetailsCreateCanceledMsg:
		m.ActiveModal = nil
		return m, nil

	case WorkhourDetailsDeletedMsg:
		return m.handleWorkhourDetailDeleted(msg)

	case WorkhourDetailsDeleteCanceledMsg:
		m.ActiveModal = nil
		return m, nil

	case WorkhourDetailsEditedMsg:
		return m.handleWorkhourDetailEdited(msg)

	case WorkhourDetailsEditCanceledMsg:
		m.ActiveModal = nil
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		verticalMargin := 12 // Increased to account for tab bar (4 lines) + padding
		m.TableView.SetSize(msg.Width, msg.Height, verticalMargin)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Only close delete modal on 'q' (edit/create modals have text inputs where user might type 'q')
			if _, isDelete := m.ActiveModal.(WorkhourDetailsDeleteModalWrapper); isDelete {
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
				m.ActiveModal = WorkhourDetailsCreateModalWrapper{NewWorkhourDetailsCreateModal()}
				return m, nil
			}

		case "d":
			if m.ActiveModal == nil {
				selectedWorkhourDetail := m.getSelectedWorkhourDetail()
				if selectedWorkhourDetail != nil {
					m.ActiveModal = WorkhourDetailsDeleteModalWrapper{NewWorkhourDetailsDeleteModal(
						selectedWorkhourDetail.ID,
						selectedWorkhourDetail.Name,
					)}
					return m, nil
				}
			}

		case "enter":
			if m.ActiveModal == nil {
				selectedWorkhourDetail := m.getSelectedWorkhourDetail()
				if selectedWorkhourDetail != nil {
					m.ActiveModal = WorkhourDetailsEditModalWrapper{NewWorkhourDetailsEditModal(
						selectedWorkhourDetail.ID,
						selectedWorkhourDetail.Name,
						selectedWorkhourDetail.ShortName,
						selectedWorkhourDetail.IsWork,
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

func (m WorkhourDetailsModel) View() string {
	helpText := render.RenderHelpText("↑/↓: navigate", "enter: edit", "n: new", "d: delete", "q: quit")

	if m.ActiveModal != nil {
		return m.ActiveModal.View(m.Width, m.Height)
	}

	return m.TableView.View() + "\n" + helpText
}

func (m WorkhourDetailsModel) getSelectedWorkhourDetail() *domain.WorkhourDetails {
	cursor := m.TableView.Cursor()
	if cursor >= 0 && cursor < len(m.WorkhourDetails) {
		return &m.WorkhourDetails[cursor]
	}
	return nil
}
