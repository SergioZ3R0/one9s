package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type dsListModel struct {
	datastores  []dsRow
	cursor      int
	scroll      int
	scrollX     int
	width       int
	height      int
	lines       []string
	filter      textinput.Model
	filtering   bool
	filterStr   string
	filtered    []dsRow
	filterDirty bool
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
	ti := textinput.New()
	ti.Placeholder = "filter datastores..."
	ti.CharLimit = 64
	return dsListModel{
		filter:      ti,
		filterDirty: true,
	}
}

func (m dsListModel) Init() tea.Cmd { return nil }

func (m dsListModel) Update(msg tea.Msg) (dsListModel, tea.Cmd) {
	var cmds []tea.Cmd

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
		m.filterDirty = true
		m.rebuildLines()
		return m, nil

	case tea.KeyMsg:
		k := msg.String()

		if m.filtering {
			switch k {
			case "enter":
				m.filtering = false
				m.filterStr = m.filter.Value()
				m.filter.Blur()
				m.filterDirty = true
				m.cursor = 0
				m.scroll = 0
				m.rebuildLines()
				return m, nil
			case "esc":
				m.filtering = false
				m.filter.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			cmds = append(cmds, cmd)
			m.filterStr = m.filter.Value()
			m.filterDirty = true
			m.cursor = 0
			m.scroll = 0
			m.rebuildLines()
			return m, tea.Batch(cmds...)
		}

		viewH := m.viewHeight()
		maxC := len(m.getFiltered()) - 1
		if maxC < 0 {
			maxC = 0
		}

		switch k {
		case "down", "j":
			if m.cursor < maxC {
				m.cursor++
				if m.cursor >= m.scroll+viewH {
					m.scroll = m.cursor - viewH + 1
				}
				if m.cursor < m.scroll {
					m.scroll = m.cursor
				}
				m.rebuildLines()
				return m, nil
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scroll {
					m.scroll = m.cursor
				}
				m.rebuildLines()
				return m, nil
			}
		case "pgdown", "f":
			m.cursor = min(m.cursor+viewH, maxC)
			m.scroll = min(m.scroll+viewH, max(0, len(m.getFiltered())-viewH))
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
		case "/":
			m.filtering = true
			m.filter.Focus()
			cmds = append(cmds, textinput.Blink)
		case "left":
			if m.scrollX > 0 {
				m.scrollX--
			}
		case "right":
			m.scrollX++
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *dsListModel) getFiltered() []dsRow {
	if m.filterDirty {
		if m.filterStr == "" {
			m.filtered = m.datastores
		} else {
			needle := strings.ToLower(m.filterStr)
			m.filtered = make([]dsRow, 0)
			for _, d := range m.datastores {
				haystack := strings.ToLower(d.ID + " " + d.Name + " " + d.Type + " " + d.Total + " " + d.Used + " " + d.Free)
				if strings.Contains(haystack, needle) {
					m.filtered = append(m.filtered, d)
				}
			}
		}
		m.filterDirty = false
	}
	return m.filtered
}

func (m *dsListModel) viewHeight() int {
	h := m.height - 12
	if m.filtering || m.filterStr != "" {
		h--
	}
	if h < 1 {
		h = 1
	}
	return h
}

func (m *dsListModel) rebuildLines() {
	m.lines = make([]string, 0, len(m.getFiltered())+1)
	m.lines = append(m.lines, tableHeader.Render(
		fmt.Sprintf("  %-6s %-24s %-10s %-12s %-12s %-12s",
			"ID", "NAME", "TYPE", "TOTAL", "USED", "FREE"),
	))
	for i, d := range m.getFiltered() {
		row := fmt.Sprintf("%-6s %-24s %-10s %-12s %-12s %-12s",
			d.ID, truncate(d.Name, 23), d.Type, d.Total, d.Used, d.Free)
		if i == m.cursor {
			m.lines = append(m.lines, cursorStyle.Render(row))
		} else {
			m.lines = append(m.lines, row)
		}
	}
	viewH := m.viewHeight()
	if m.cursor >= len(m.getFiltered()) {
		m.cursor = max(0, len(m.getFiltered())-1)
	}
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+viewH {
		m.scroll = m.cursor - viewH + 1
	}
	maxScroll := max(0, len(m.getFiltered())-viewH)
	if m.scroll > maxScroll {
		m.scroll = maxScroll
	}
}

func (m dsListModel) View() string {
	var parts []string
	if m.filtering {
		parts = append(parts, filterStyle.Render("Filter: ")+m.filter.View())
	} else if m.filterStr != "" {
		parts = append(parts, filterStyle.Render("Filter: ")+m.filterStr+" [esc]")
	}
	// Sticky header: first line is always the table header
	if len(m.lines) > 0 {
		parts = append(parts, clipLine(scrollLine(m.lines[0], m.scrollX), m.width))
	}
	viewH := m.viewHeight()
	// Content rows from scroll offset
	start := m.scroll + 1
	end := min(start+viewH, len(m.lines))
	if start < len(m.lines) {
		visible := m.lines[start:end]
		for _, line := range visible {
			parts = append(parts, clipLine(scrollLine(line, m.scrollX), m.width))
		}
	}
	filtered := m.getFiltered()
	status := statusStyle.Render(fmt.Sprintf(" %d/%d Datastores  cursor:%d/%d", len(filtered), len(m.datastores), m.cursor, max(0, len(filtered)-1)))
	parts = append(parts, status)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
