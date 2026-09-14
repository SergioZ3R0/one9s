package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type aclListModel struct {
	acls   []aclRow
	cursor int
	scroll int
	width  int
	height int
	lines  []string
}

type aclRow struct {
	ID       string
	User     string
	Resource string
	Rights   string
	Zone     string
}

func newACLListModel() aclListModel {
	return aclListModel{}
}

func (m aclListModel) Init() tea.Cmd { return nil }

func (m aclListModel) Update(msg tea.Msg) (aclListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case aclsFetchedMsg:
		if msg.err != nil {
			return m, nil
		}
		m.acls = make([]aclRow, 0, len(msg.acls))
		for _, a := range msg.acls {
			m.acls = append(m.acls, aclRow{
				ID:       fmt.Sprintf("%d", a.ID),
				User:     a.User,
				Resource: a.Resource,
				Rights:   a.Rights,
				Zone:     a.Zone,
			})
		}
		m.rebuildLines()
		return m, nil

	case tea.KeyMsg:
		viewH := m.viewHeight()
		switch {
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.acls)-1 {
				m.cursor++
				if m.cursor >= m.scroll+viewH {
					m.scroll = m.cursor - viewH + 1
				}
			}
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scroll {
					m.scroll = m.scroll - 1
				}
			}
		case key.Matches(msg, keys.PageDown):
			m.cursor = min(m.cursor+viewH, max(0, len(m.acls)-1))
			m.scroll = min(m.scroll+viewH, max(0, len(m.acls)-viewH))
		case key.Matches(msg, keys.PageUp):
			m.cursor = max(m.cursor-viewH, 0)
			m.scroll = max(m.scroll-viewH, 0)
		case key.Matches(msg, keys.Home):
			m.cursor = 0
			m.scroll = 0
		case key.Matches(msg, keys.End):
			m.cursor = max(0, len(m.acls)-1)
			m.scroll = max(0, m.cursor-viewH+1)
		}
	}

	return m, nil
}

func (m *aclListModel) viewHeight() int {
	h := m.height - 3
	if h < 1 {
		h = 1
	}
	return h
}

func (m *aclListModel) rebuildLines() {
	m.lines = make([]string, 0, len(m.acls)+1)

	header := fmt.Sprintf("%-6s %-20s %-24s %-20s %-12s",
		"ID", "USER", "RESOURCE", "RIGHTS", "ZONE")
	m.lines = append(m.lines, tableHeader.Render(header))

	for i, a := range m.acls {
		row := fmt.Sprintf("%-6s %-20s %-24s %-20s %-12s",
			a.ID, truncate(a.User, 19), truncate(a.Resource, 23),
			truncate(a.Rights, 19), a.Zone)
		if i == m.cursor {
			m.lines = append(m.lines, cursorStyle.Render(row))
		} else {
			m.lines = append(m.lines, row)
		}
	}

	viewH := m.viewHeight()
	if m.cursor >= len(m.acls) {
		m.cursor = max(0, len(m.acls)-1)
	}
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+viewH {
		m.scroll = m.cursor - viewH + 1
	}
	if m.scroll > 0 && m.scroll+viewH > len(m.lines) {
		m.scroll = max(0, len(m.lines)-viewH)
	}
}

func (m aclListModel) View() string {
	viewH := m.viewHeight()
	end := min(m.scroll+viewH, len(m.lines))
	visible := m.lines[m.scroll:end]

	status := statusStyle.Render(fmt.Sprintf(" %d ACL Rules", len(m.acls)))
	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(visible, "\n"), status)
}
