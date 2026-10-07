package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type netListModel struct {
	networks    []netRow
	cursor      int
	scroll      int
	scrollX     int
	width       int
	height      int
	lines       []string
	filter      textinput.Model
	filtering   bool
	filterStr   string
	filtered    []netRow
	filterDirty bool
}

type netRow struct {
	ID       string
	Name     string
	Bridge   string
	VNMad    string
	Used     string
	Total    string
	Usage    string
	Owner    string
	UsedInt  int
	TotalInt int
	RawID    int
}

func newNetListModel() netListModel {
	ti := textinput.New()
	ti.Placeholder = "filter networks..."
	ti.CharLimit = 64
	return netListModel{
		filter:      ti,
		filterDirty: true,
	}
}

func (m netListModel) Init() tea.Cmd { return nil }

func (m netListModel) Update(msg tea.Msg) (netListModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case networksFetchedMsg:
		if msg.err != nil {
			return m, nil
		}
		m.networks = make([]netRow, 0, len(msg.networks))
		for _, n := range msg.networks {
			var usage string
			if n.Total > 0 {
				usage = fmt.Sprintf("%d/%d", n.Used, n.Total)
			} else {
				usage = fmt.Sprintf("%d/-", n.Used)
			}
			m.networks = append(m.networks, netRow{
				ID:       fmt.Sprintf("%d", n.ID),
				Name:     n.Name,
				Bridge:   n.Bridge,
				VNMad:    n.VNMad,
				Used:     n.UsedText,
				Total:    n.TotalText,
				Usage:    usage,
				Owner:    n.Owner,
				UsedInt:  n.Used,
				TotalInt: n.Total,
				RawID:    n.ID,
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

func (m *netListModel) getFiltered() []netRow {
	if m.filterDirty {
		if m.filterStr == "" {
			m.filtered = m.networks
		} else {
			needle := strings.ToLower(m.filterStr)
			m.filtered = make([]netRow, 0)
			for _, n := range m.networks {
				haystack := strings.ToLower(n.ID + " " + n.Name + " " + n.Bridge + " " + n.VNMad + " " + n.Usage + " " + n.Owner)
				if strings.Contains(haystack, needle) {
					m.filtered = append(m.filtered, n)
				}
			}
		}
		m.filterDirty = false
	}
	return m.filtered
}

func (m *netListModel) viewHeight() int {
	h := m.height - 12
	if m.filtering || m.filterStr != "" {
		h--
	}
	if h < 1 {
		h = 1
	}
	return h
}

func (m *netListModel) rebuildLines() {
	m.lines = make([]string, 0, len(m.getFiltered())+1)
	m.lines = append(m.lines, tableHeader.Render(
		fmt.Sprintf("  %-6s %-22s %-12s %-8s %-10s %-12s %s",
			"ID", "NAME", "BRIDGE", "VN_MAD", "LEASES", "CAPACITY", "OWNER"),
	))
	for i, n := range m.getFiltered() {
		row := fmt.Sprintf("%-6s %-22s %-12s %-8s %-10s %-12s %s",
			n.ID, truncate(n.Name, 21), truncate(n.Bridge, 11), n.VNMad,
			n.Usage, n.Total, truncate(n.Owner, 16))
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

func (m netListModel) View() string {
	var parts []string
	if m.filtering {
		parts = append(parts, filterStyle.Render("Filter: ")+m.filter.View())
	} else if m.filterStr != "" {
		parts = append(parts, filterStyle.Render("Filter: ")+m.filterStr+" [esc]")
	}
	if len(m.lines) > 0 {
		parts = append(parts, clipLine(scrollLine(m.lines[0], m.scrollX), m.width))
	}
	viewH := m.viewHeight()
	start := m.scroll + 1
	end := min(start+viewH, len(m.lines))
	if start < len(m.lines) {
		visible := m.lines[start:end]
		for _, line := range visible {
			parts = append(parts, clipLine(scrollLine(line, m.scrollX), m.width))
		}
	}
	filtered := m.getFiltered()
	status := statusStyle.Render(fmt.Sprintf(" %d/%d Networks  cursor:%d/%d  leases used/total", len(filtered), len(m.networks), m.cursor, max(0, len(filtered)-1)))
	parts = append(parts, status)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
