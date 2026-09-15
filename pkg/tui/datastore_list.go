package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type dsListModel struct {
	datastores []dsRow
	cursor     int
	scroll     int
	width      int
	height     int
	lines      []string
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
	return dsListModel{}
}

func (m dsListModel) Init() tea.Cmd { return nil }

func (m dsListModel) Update(msg tea.Msg) (dsListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
		m.cursor = 0
		m.scroll = 0
		m.rebuildLines()
		return m, nil

	case tea.KeyMsg:
		viewH := m.viewHeight()
		maxC := len(m.datastores) - 1
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
			m.scroll = min(m.scroll+viewH, max(0, len(m.datastores)-viewH))
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

func (m *dsListModel) viewHeight() int {
	h := m.height - 3
	if h < 1 {
		h = 1
	}
	return h
}

func (m *dsListModel) rebuildLines() {
	m.lines = make([]string, 0, len(m.datastores)+1)
	m.lines = append(m.lines, tableHeader.Render(
		fmt.Sprintf("  %-6s %-24s %-10s %-12s %-12s %-12s",
			"ID", "NAME", "TYPE", "TOTAL", "USED", "FREE"),
	))
	for i, d := range m.datastores {
		row := fmt.Sprintf("%-6s %-24s %-10s %-12s %-12s %-12s",
			d.ID, truncate(d.Name, 23), d.Type, d.Total, d.Used, d.Free)
		if i == m.cursor {
			m.lines = append(m.lines, cursorStyle.Render("▸ "+row))
		} else {
			m.lines = append(m.lines, "  "+row)
		}
	}
	viewH := m.viewHeight()
	if m.cursor >= len(m.datastores) {
		m.cursor = max(0, len(m.datastores)-1)
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

func (m dsListModel) View() string {
	viewH := m.viewHeight()
	end := min(m.scroll+viewH, len(m.lines))
	visible := m.lines[m.scroll:end]
	status := statusStyle.Render(fmt.Sprintf(" %d Datastores  cursor:%d/%d", len(m.datastores), m.cursor, max(0, len(m.datastores)-1)))
	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(visible, "\n"), status)
}
