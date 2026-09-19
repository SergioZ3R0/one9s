package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type hostListModel struct {
	hosts  []hostRow
	cursor int
	scroll int
	width  int
	height int
	lines  []string
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
	return hostListModel{}
}

func (m hostListModel) Init() tea.Cmd { return nil }

func (m hostListModel) Update(msg tea.Msg) (hostListModel, tea.Cmd) {
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
		m.rebuildLines()
		return m, nil

	case tea.KeyMsg:
		viewH := m.viewHeight()
		maxC := len(m.hosts) - 1
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
			m.scroll = min(m.scroll+viewH, max(0, len(m.hosts)-viewH))
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

func (m *hostListModel) viewHeight() int {
	h := m.height - 9
	if h < 1 {
		h = 1
	}
	return h
}

func (m *hostListModel) rebuildLines() {
	m.lines = make([]string, 0, len(m.hosts)+1)
	m.lines = append(m.lines, tableHeader.Render(
		fmt.Sprintf("  %-6s %-24s %-12s %-8s %-8s %-10s %-8s",
			"ID", "NAME", "STATE", "CPU", "MEM", "VMs", "CLUSTER"),
	))
	for i, h := range m.hosts {
		row := fmt.Sprintf("%-6s %-24s %-12s %-8s %-8s %-10s %-8s",
			h.ID, truncate(h.Name, 23), h.State, h.CPU, h.Memory, h.VMs, h.Cluster)
		if i == m.cursor {
			m.lines = append(m.lines, cursorStyle.Render(row))
		} else {
			m.lines = append(m.lines, hostStateStyle(h.State).Render(row))
		}
	}
	viewH := m.viewHeight()
	if m.cursor >= len(m.hosts) {
		m.cursor = max(0, len(m.hosts)-1)
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
	viewH := m.viewHeight()
	end := min(m.scroll+viewH, len(m.lines))
	visible := m.lines[m.scroll:end]
	status := statusStyle.Render(fmt.Sprintf(" %d Hosts  cursor:%d/%d", len(m.hosts), m.cursor, max(0, len(m.hosts)-1)))
	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(visible, "\n"), status)
}
