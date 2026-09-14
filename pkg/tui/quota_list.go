package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
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
		m.rebuildLines()
		return m, nil

	case tea.KeyMsg:
		viewH := m.viewHeight()
		switch {
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.quotas)-1 {
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
			m.cursor = min(m.cursor+viewH, max(0, len(m.quotas)-1))
			m.scroll = min(m.scroll+viewH, max(0, len(m.quotas)-viewH))
		case key.Matches(msg, keys.PageUp):
			m.cursor = max(m.cursor-viewH, 0)
			m.scroll = max(m.scroll-viewH, 0)
		case key.Matches(msg, keys.Home):
			m.cursor = 0
			m.scroll = 0
		case key.Matches(msg, keys.End):
			m.cursor = max(0, len(m.quotas)-1)
			m.scroll = max(0, m.cursor-viewH+1)
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

	header := fmt.Sprintf("%-20s %-12s %-16s %-16s %-14s %-12s %-14s %-12s",
		"ENTITY", "VMs", "CPU", "MEMORY", "RUNNING", "IMAGES", "SIZE", "LEASES")
	m.lines = append(m.lines, tableHeader.Render(header))

	for i, q := range m.quotas {
		row := fmt.Sprintf("%-20s %-12s %-16s %-16s %-14s %-12s %-14s %-12s",
			truncate(q.Entity, 19), q.VMs, q.CPU, q.Memory,
			q.RunningVMs, q.Images, q.Size, q.Leases)
		if i == m.cursor {
			m.lines = append(m.lines, cursorStyle.Render(row))
		} else {
			m.lines = append(m.lines, row)
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

	status := statusStyle.Render(fmt.Sprintf(" %d User Quotas", len(m.quotas)))
	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(visible, "\n"), status)
}
