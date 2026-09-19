package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type hostListModel struct {
	hosts       []hostRow
	cursor      int
	scroll      int
	width       int
	height      int
	lines       []string
	filter      textinput.Model
	filtering   bool
	filterStr   string
	filtered    []hostRow
	filterDirty bool
}

type hostRow struct {
	ID      string
	Name    string
	State   string
	CPU     string
	Memory  string
	VMs     string
	Cluster string
}

func newHostListModel() hostListModel {
	ti := textinput.New()
	ti.Placeholder = "filter hosts..."
	ti.CharLimit = 64
	return hostListModel{
		filter:      ti,
		filterDirty: true,
	}
}

func (m hostListModel) Init() tea.Cmd { return nil }

func (m hostListModel) Update(msg tea.Msg) (hostListModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case hostsFetchedMsg:
		if msg.err != nil {
			return m, nil
		}
		m.hosts = make([]hostRow, 0, len(msg.hosts))
		for _, h := range msg.hosts {
			m.hosts = append(m.hosts, hostRow{
				ID:      fmt.Sprintf("%d", h.ID),
				Name:    h.Name,
				State:   h.State,
				CPU:     h.CPU,
				Memory:  h.Memory,
				VMs:     h.VMs,
				Cluster: h.Cluster,
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
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *hostListModel) getFiltered() []hostRow {
	if m.filterDirty {
		if m.filterStr == "" {
			m.filtered = m.hosts
		} else {
			needle := strings.ToLower(m.filterStr)
			m.filtered = make([]hostRow, 0)
			for _, h := range m.hosts {
				haystack := strings.ToLower(h.ID + " " + h.Name + " " + h.State + " " + h.CPU + " " + h.Memory + " " + h.VMs + " " + h.Cluster)
				if strings.Contains(haystack, needle) {
					m.filtered = append(m.filtered, h)
				}
			}
		}
		m.filterDirty = false
	}
	return m.filtered
}

func (m *hostListModel) viewHeight() int {
	h := m.height - 7
	if h < 1 {
		h = 1
	}
	return h
}

func (m *hostListModel) rebuildLines() {
	m.lines = make([]string, 0, len(m.getFiltered())+1)
	m.lines = append(m.lines, tableHeader.Render(
		fmt.Sprintf("  %-6s %-24s %-12s %-8s %-8s %-10s %-8s",
			"ID", "NAME", "STATE", "CPU", "MEM", "VMs", "CLUSTER"),
	))
	for i, h := range m.getFiltered() {
		row := fmt.Sprintf("%-6s %-24s %-12s %-8s %-8s %-10s %-8s",
			h.ID, truncate(h.Name, 23), h.State, h.CPU, h.Memory, h.VMs, h.Cluster)
		if i == m.cursor {
			m.lines = append(m.lines, cursorStyle.Render(row))
		} else {
			m.lines = append(m.lines, hostStateStyle(h.State).Render(row))
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
	if m.scroll > 0 && m.scroll+viewH > len(m.lines) {
		m.scroll = max(0, len(m.lines)-viewH)
	}
}

func (m hostListModel) View() string {
	var parts []string
	if m.filtering {
		parts = append(parts, filterStyle.Render("Filter: ")+m.filter.View())
	} else if m.filterStr != "" {
		parts = append(parts, filterStyle.Render("Filter: ")+m.filterStr+" [esc]")
	}

	viewH := m.viewHeight()
	end := min(m.scroll+viewH, len(m.lines))
	visible := m.lines[m.scroll:end]
	parts = append(parts, strings.Join(visible, "\n"))

	filtered := m.getFiltered()
	status := statusStyle.Render(fmt.Sprintf(" %d/%d Hosts  cursor:%d/%d", len(filtered), len(m.hosts), m.cursor, max(0, len(filtered)-1)))
	parts = append(parts, status)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
