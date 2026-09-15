package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type quotaListModel struct {
	quotas []quotaRow
	cursor int
	scroll int
	width  int
	height int
	lines  []string
}

type quotaRow struct {
	Entity     string
	VMs        string
	CPU        string
	Memory     string
	RunningVMs string
	Images     string
	Size       string
	Leases     string
}

func newQuotaListModel() quotaListModel {
	return quotaListModel{}
}

func (m quotaListModel) Init() tea.Cmd { return nil }

func (m quotaListModel) Update(msg tea.Msg) (quotaListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case quotasFetchedMsg:
		if msg.err != nil {
			return m, nil
		}
		m.quotas = make([]quotaRow, 0, len(msg.quotas))
		for _, q := range msg.quotas {
			m.quotas = append(m.quotas, quotaRow{
				Entity:     q.Entity,
				VMs:        q.VMs,
				CPU:        q.CPU,
				Memory:     q.Memory,
				RunningVMs: q.RunningVMs,
				Images:     q.Images,
				Size:       q.Size,
				Leases:     q.Leases,
			})
		}
		m.cursor = 0
		m.scroll = 0
		m.rebuildLines()
		return m, nil

	case tea.KeyMsg:
		viewH := m.viewHeight()
		maxC := len(m.quotas) - 1
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
			m.scroll = min(m.scroll+viewH, max(0, len(m.quotas)-viewH))
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

func (m *quotaListModel) viewHeight() int {
	h := m.height - 3
	if h < 1 {
		h = 1
	}
	return h
}

func (m *quotaListModel) rebuildLines() {
	m.lines = make([]string, 0, len(m.quotas)+1)
	m.lines = append(m.lines, tableHeader.Render(
		fmt.Sprintf("  %-20s %-12s %-16s %-16s %-14s %-12s %-14s %-12s",
			"ENTITY", "VMs", "CPU", "MEMORY", "RUNNING", "IMAGES", "SIZE", "LEASES"),
	))
	for i, q := range m.quotas {
		row := fmt.Sprintf("%-20s %-12s %-16s %-16s %-14s %-12s %-14s %-12s",
			truncate(q.Entity, 19), q.VMs, q.CPU, q.Memory,
			q.RunningVMs, q.Images, q.Size, q.Leases)
		if i == m.cursor {
			m.lines = append(m.lines, cursorStyle.Render("▸ "+row))
		} else {
			m.lines = append(m.lines, "  "+row)
		}
	}
	viewH := m.viewHeight()
	if m.cursor >= len(m.quotas) {
		m.cursor = max(0, len(m.quotas)-1)
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

func (m quotaListModel) View() string {
	viewH := m.viewHeight()
	end := min(m.scroll+viewH, len(m.lines))
	visible := m.lines[m.scroll:end]
	status := statusStyle.Render(fmt.Sprintf(" %d User Quotas  cursor:%d/%d", len(m.quotas), m.cursor, max(0, len(m.quotas)-1)))
	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(visible, "\n"), status)
}
