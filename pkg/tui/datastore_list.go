package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type dsListModel struct {
	datastores []dsRow
	viewport   viewport.Model
	cursor     int
	width      int
	height     int
}

type dsRow struct {
	ID    string
	Name  string
	Type  string
	Total string
	Used  string
	Free  string
}

func newDSListModel() dsListModel {
	return dsListModel{
		viewport: viewport.New(80, 24),
	}
}

func (m dsListModel) Init() tea.Cmd { return nil }

func (m dsListModel) Update(msg tea.Msg) (dsListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 3
		return m, nil

	case datastoresFetchedMsg:
		if msg.err != nil {
			return m, nil
		}
		m.datastores = make([]dsRow, 0, len(msg.datastores))
		for _, d := range msg.datastores {
			m.datastores = append(m.datastores, dsRow{
				ID:    fmt.Sprintf("%d", d.ID),
				Name:  d.Name,
				Type:  d.Type,
				Total: d.Total,
				Used:  d.Used,
				Free:  d.Free,
			})
		}
		m.updateContent()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.datastores)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func (m *dsListModel) updateContent() {
	var sb strings.Builder
	cols := "%-6s %-24s %-10s %-12s %-12s %-12s"
	header := fmt.Sprintf(cols, "ID", "NAME", "TYPE", "TOTAL", "USED", "FREE")
	sb.WriteString(tableHeader.Render(header))
	sb.WriteString("\n")

	for i, d := range m.datastores {
		row := fmt.Sprintf(cols, d.ID, truncate(d.Name, 23), d.Type, d.Total, d.Used, d.Free)
		if i == m.cursor {
			sb.WriteString(lipgloss.NewStyle().Foreground(primary).Bold(true).Render(row))
		} else {
			sb.WriteString(row)
		}
		sb.WriteString("\n")
	}
	m.viewport.SetContent(sb.String())
}

func (m dsListModel) View() string {
	status := statusStyle.Render(fmt.Sprintf(" %d Datastores", len(m.datastores)))
	return lipgloss.JoinVertical(lipgloss.Left, m.viewport.View(), status)
}
