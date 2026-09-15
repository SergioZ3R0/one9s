package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// modalType distinguishes between y/n confirm and text input confirm
type modalType int

const (
	modalYN        modalType = iota // simple y/n
	modalTextInput                  // type "yes" or custom text
)

type modalState struct {
	active    bool
	modalType modalType
	title     string
	message   string
	cmd       tea.Cmd // action to run on confirm

	// For text input modals
	input     textinput.Model
	expecting string // expected input to confirm
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

func (m modalState) View() string {
	if !m.active {
		return ""
	}
	body := fmt.Sprintf(
		"%s\n\n%s\n",
		titleStyle.Render(m.title),
		m.message,
	)

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
	}

	return body
}
