package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type modalType int

const (
	modalYN modalType = iota
	modalTextInput
	modalForm
)

type modalState struct {
	active    bool
	modalType modalType
	title     string
	message   string
	cmd       tea.Cmd

	input     textinput.Model
	expecting string

	formFields []formField
	formCmd    func(values map[string]string) tea.Cmd
}

type formField struct {
	label string
	input textinput.Model
	key   string
}

func newModal(title, msg string, cmd tea.Cmd) modalState {
	return modalState{
		active:    true,
		modalType: modalYN,
		title:     title,
		message:   msg,
		cmd:       cmd,
	}
}

func newTextInputModal(title, msg, expecting string, cmd tea.Cmd) modalState {
	ti := textinput.New()
	ti.Placeholder = expecting
	ti.CharLimit = 64
	ti.Focus()
	return modalState{
		active:    true,
		modalType: modalTextInput,
		title:     title,
		message:   msg,
		cmd:       cmd,
		input:     ti,
		expecting: expecting,
	}
}

func newFormModal(title string, fields []formField, cmd func(values map[string]string) tea.Cmd) modalState {
	for i := range fields {
		fields[i].input.CharLimit = 12
		if i == 0 {
			fields[i].input.Focus()
		}
	}
	return modalState{
		active:     true,
		modalType:  modalForm,
		title:      title,
		formFields: fields,
		formCmd:    cmd,
	}
}

func (m modalState) View() string {
	if !m.active {
		return ""
	}
	body := fmt.Sprintf("%s\n", titleStyle.Render(m.title))

	if m.message != "" {
		body += fmt.Sprintf("%s\n", m.message)
	}

	switch m.modalType {
	case modalYN:
		body += fmt.Sprintf("\n%s",
			lipgloss.JoinHorizontal(lipgloss.Top,
				stateRunning.Render("[y] Yes"),
				"  ",
				statePoweroff.Render("[n] No"),
			),
		)
	case modalTextInput:
		body += fmt.Sprintf("\n%s\n\n%s",
			filterStyle.Render("Type '"+m.expecting+"' to confirm:"),
			m.input.View(),
		)
	case modalForm:
		for _, f := range m.formFields {
			body += fmt.Sprintf("\n  %s %s", filterStyle.Render(f.label+":"), f.input.View())
		}
		body += fmt.Sprintf("\n\n%s",
			lipgloss.JoinHorizontal(lipgloss.Top,
				stateRunning.Render("[y] Apply"),
				"  ",
				statePoweroff.Render("[n] Cancel"),
			),
		)
	}

	return body
}
