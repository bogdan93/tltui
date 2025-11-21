package common

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FocusNextMsg struct{}
type FocusPrevMsg struct{}

func DispatchFocusNext() tea.Cmd {
	return func() tea.Msg {
		return FocusNextMsg{}
	}
}

func DispatchFocusPrev() tea.Cmd {
	return func() tea.Msg {
		return FocusPrevMsg{}
	}
}

type FormElement interface {
	Focus() tea.Cmd
	Blur()
	Update(msg tea.Msg) tea.Cmd
	View() string
}

type MixedForm struct {
	Elements     []FormElement
	FocusedIndex int
	ErrorMessage string
}

func NewMixedForm(elements ...FormElement) *MixedForm {
	form := &MixedForm{
		Elements:     elements,
		FocusedIndex: 0,
	}

	// Focus the first element
	if len(elements) > 0 {
		elements[0].Focus()
	}

	return form
}

func (f *MixedForm) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			f.NextField()
			return nil
		case "shift+tab":
			f.PrevField()
			return nil
		}
	}

	if f.FocusedIndex >= 0 && f.FocusedIndex < len(f.Elements) {
		cmd = f.Elements[f.FocusedIndex].Update(msg)
	}

	switch msg.(type) {
	case FocusNextMsg:
		f.NextField()

	case FocusPrevMsg:
		f.PrevField()
	}

	return cmd
}

func (f *MixedForm) NextField() {
	if len(f.Elements) == 0 {
		return
	}

	f.Elements[f.FocusedIndex].Blur()
	f.FocusedIndex = (f.FocusedIndex + 1) % len(f.Elements)
	f.Elements[f.FocusedIndex].Focus()
}

func (f *MixedForm) PrevField() {
	if len(f.Elements) == 0 {
		return
	}

	f.Elements[f.FocusedIndex].Blur()
	f.FocusedIndex = (f.FocusedIndex - 1 + len(f.Elements)) % len(f.Elements)
	f.Elements[f.FocusedIndex].Focus()
}

func (f *MixedForm) FocusField(index int) {
	if index < 0 || index >= len(f.Elements) {
		return
	}

	f.Elements[f.FocusedIndex].Blur()
	f.FocusedIndex = index
	f.Elements[f.FocusedIndex].Focus()
}

func (f *MixedForm) SetError(message string) {
	f.ErrorMessage = message
}

func (f *MixedForm) ClearError() {
	f.ErrorMessage = ""
}

func (f *MixedForm) GetField(index int) *FormField {
	if index < 0 || index >= len(f.Elements) {
		return nil
	}
	if field, ok := f.Elements[index].(*FormField); ok {
		return field
	}
	return nil
}

func (f *MixedForm) GetCheckbox(index int) *FormCheckbox {
	if index < 0 || index >= len(f.Elements) {
		return nil
	}
	if checkbox, ok := f.Elements[index].(*FormCheckbox); ok {
		return checkbox
	}
	return nil
}

func (f *MixedForm) GetSelect(index int) *FormSelect {
	if index < 0 || index >= len(f.Elements) {
		return nil
	}
	if select_, ok := f.Elements[index].(*FormSelect); ok {
		return select_
	}
	return nil
}

func (f *MixedForm) Validate() error {
	f.ErrorMessage = ""

	for i, element := range f.Elements {
		if field, ok := element.(*FormField); ok {
			if err := field.Validate(); err != nil {
				f.ErrorMessage = err.Error()
				f.FocusField(i)
				return err
			}
		}

		if select_, ok := element.(*FormSelect); ok {
			if err := select_.Validate(); err != nil {
				f.ErrorMessage = err.Error()
				f.FocusField(i)
				return err
			}
		}
	}

	return nil
}

func (f *MixedForm) View() string {
	var output string

	for _, element := range f.Elements {
		output += element.View()
		output += "\n"
	}

	if f.ErrorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		output += errorStyle.Render("⚠ " + f.ErrorMessage)
		output += "\n"
	}

	return output
}
