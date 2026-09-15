package tui

import (
	"fmt"
	"strings"

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
		m.cursor = 0
		m.scroll = 0
		m.rebuildLines()
		return m, nil

	case tea.KeyMsg:
		viewH := m.viewHeight()
		maxC := len(m.acls) - 1
		if maxC < 0 {
			maxC = 0
		}
		switch msg.String() {
		case "down", "j":
			if m.cursor < maxC {
				m.cursor++
			}
			if m.cursor >= m.scroll+viewH {
				m.scroll = m.cursor - viewH + 1
			}
			if m.cursor < m.scroll {
				m.scroll = m.cursor
			}
			m.rebuildLines()
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			if m.cursor < m.scroll {
				m.scroll = m.cursor
			}
			m.rebuildLines()
			return m, nil
		case "pgdown", "f":
			m.cursor = min(m.cursor+viewH, maxC)
			m.scroll = min(m.scroll+viewH, max(0, len(m.acls)-viewH))
			m.rebuildLines()
			return m, nil
		case "pgup", "b":
			m.cursor = max(m.cursor-viewH, 0)
			m.scroll = max(m.scroll-viewH, 0)
			m.rebuildLines()
			return m, nil
		case "g":
			m.cursor = 0
			m.scroll = 0
			m.rebuildLines()
			return m, nil
		case "G":
			m.cursor = maxC
			m.scroll = max(0, m.cursor-viewH+1)
			m.rebuildLines()
			return m, nil
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

	// Dynamic column widths based on terminal width
	avail := m.width - 6 // margins
	if avail < 40 {
		avail = 40
	}

	idW := min(6, avail/20)
	userW := min(14, avail/8)
	rightsW := min(25, avail/4)
	zoneW := max(6, avail/10)
	resW := avail - idW - userW - rightsW - zoneW - 10
	if resW < 10 {
		resW = 10
	}

	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s  %s",
		idW, "ID", userW, "USER", resW, "RESOURCE", rightsW, "RIGHTS", "ZONE")
	m.lines = append(m.lines, tableHeader.Render(header))

	for i, a := range m.acls {
		row := fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s  %s",
			idW, truncate(a.ID, idW),
			userW, truncate(a.User, userW),
			resW, truncate(a.Resource, resW),
			rightsW, truncate(a.Rights, rightsW),
			truncate(a.Zone, zoneW))
		if i == m.cursor {
			m.lines = append(m.lines, cursorStyle.Render("▸ "+row))
		} else {
			m.lines = append(m.lines, "  "+row)
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
	status := statusStyle.Render(fmt.Sprintf(" %d ACL Rules  cursor:%d/%d", len(m.acls), m.cursor, max(0, len(m.acls)-1)))
	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(visible, "\n"), status)
}
