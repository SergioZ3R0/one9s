package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type quotaListModel struct {
	quotas      []quotaRow
	cursor      int
	scroll      int
	scrollX     int
	width       int
	height      int
	lines       []string
	filter      textinput.Model
	filtering   bool
	filterStr   string
	filtered    []quotaRow
	filterDirty bool
}

type quotaRow struct {
	UserID          int
	Entity          string
	VMs             string
	CPU             string
	Memory          string
	RunningVMs      string
	Images          string
	Size            string
	Leases          string
	VMsLimit        int
	CPULimit        int
	MemoryLimit     int
	RunningVMsLimit int
	ImagesLimit     int
	SizeLimit       int
	LeasesLimit     int
}

func newQuotaListModel() quotaListModel {
	ti := textinput.New()
	ti.Placeholder = "filter quotas..."
	ti.CharLimit = 64
	return quotaListModel{
		filter:      ti,
		filterDirty: true,
	}
}

func (m quotaListModel) Init() tea.Cmd { return nil }

func (m quotaListModel) Update(msg tea.Msg) (quotaListModel, tea.Cmd) {
	var cmds []tea.Cmd

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
				UserID:          q.UserID,
				Entity:          q.Entity,
				VMs:             q.VMs,
				CPU:             q.CPU,
				Memory:          q.Memory,
				RunningVMs:      q.RunningVMs,
				Images:          q.Images,
				Size:            q.Size,
				Leases:          q.Leases,
				VMsLimit:        q.VMsLimit,
				CPULimit:        q.CPULimit,
				MemoryLimit:     q.MemoryLimit,
				RunningVMsLimit: q.RunningVMsLimit,
				ImagesLimit:     q.ImagesLimit,
				SizeLimit:       q.SizeLimit,
				LeasesLimit:     q.LeasesLimit,
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
				if m.scrollX < 0 {
					m.scrollX = 0
				}
			}
		case "right":
			m.scrollX++
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *quotaListModel) getFiltered() []quotaRow {
	if m.filterDirty {
		if m.filterStr == "" {
			m.filtered = m.quotas
		} else {
			needle := strings.ToLower(m.filterStr)
			m.filtered = make([]quotaRow, 0)
			for _, q := range m.quotas {
				haystack := strings.ToLower(q.Entity + " " + q.VMs + " " + q.CPU + " " + q.Memory)
				if strings.Contains(haystack, needle) {
					m.filtered = append(m.filtered, q)
				}
			}
		}
		m.filterDirty = false
	}
	return m.filtered
}

func (m *quotaListModel) viewHeight() int {
	h := m.height - 12
	if m.filtering || m.filterStr != "" {
		h--
	}
	if h < 1 {
		h = 1
	}
	return h
}

func (m *quotaListModel) rebuildLines() {
	m.lines = make([]string, 0, len(m.getFiltered())+1)

	w := m.width - 4
	if w < 30 {
		w = 30
	}
	eW := min(20, w/7)
	vW := min(10, w/10)
	cW := min(12, w/8)
	mW := min(14, w/7)
	rW := min(10, w/10)
	iW := min(8, w/12)

	m.lines = append(m.lines, tableHeader.Render(
		fmt.Sprintf("  %-*s %-*s %-*s %-*s %-*s %-*s %s",
			eW, "ENTITY", vW, "VMs", cW, "CPU", mW, "MEMORY", rW, "RUN", iW, "IMG", "LEASES"),
	))
	for i, q := range m.getFiltered() {
		row := fmt.Sprintf("  %-*s %-*s %-*s %-*s %-*s %-*s %s",
			eW, q.Entity, vW, q.VMs, cW, q.CPU, mW, q.Memory,
			rW, q.RunningVMs, iW, q.Images, q.Leases)
		if i == m.cursor {
			m.lines = append(m.lines, cursorStyle.Render(row))
		} else {
			m.lines = append(m.lines, row)
		}
	}
	viewH := m.viewHeight()
	filtered := m.getFiltered()
	if m.cursor >= len(filtered) {
		m.cursor = max(0, len(filtered)-1)
	}
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+viewH {
		m.scroll = m.cursor - viewH + 1
	}
	maxScroll := max(0, len(filtered)-viewH)
	if m.scroll > maxScroll {
		m.scroll = maxScroll
	}
}

func (m quotaListModel) View() string {
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
	status := statusStyle.Render(fmt.Sprintf(" %d/%d User Quotas  cursor:%d/%d", len(filtered), len(m.quotas), m.cursor, max(0, len(filtered)-1)))
	parts = append(parts, status)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
