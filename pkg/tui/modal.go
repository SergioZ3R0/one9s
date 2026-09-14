package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type modalState struct {
	active  bool
	title   string
	message string
	cmd     tea.Cmd // action to run on confirm
}

func newModal(title, msg string, cmd tea.Cmd) modalState {
	return modalState{
		active:  true,
		title:   title,
		message: msg,
		cmd:     cmd,
	}
}

func (m modalState) View() string {
	if !m.active {
		return ""
	}
	body := fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		titleStyle.Render(m.title),
		m.message,
		lipgloss.JoinHorizontal(lipgloss.Top,
			stateRunning.Render("[y] Yes"),
			"  ",
			statePoweroff.Render("[n] No"),
		),
	)
	return body
}
