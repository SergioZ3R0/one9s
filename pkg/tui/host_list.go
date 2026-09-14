package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type hostListModel struct {
	hosts    []hostRow
	viewport viewport.Model
	cursor   int
	width    int
	height   int
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
	return hostListModel{
		viewport: viewport.New(80, 24),
	}
}

func (m hostListModel) Init() tea.Cmd { return nil }

func (m hostListModel) Update(msg tea.Msg) (hostListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 3
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
		m.updateContent()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.hosts)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func (m *hostListModel) updateContent() {
	var sb strings.Builder
	cols := "%-6s %-24s %-12s %-8s %-8s %-10s %-8s"
	header := fmt.Sprintf(cols, "ID", "NAME", "STATE", "CPU", "MEM", "VMs", "CLUSTER")
	sb.WriteString(tableHeader.Render(header))
	sb.WriteString("\n")

	for i, h := range m.hosts {
		row := fmt.Sprintf(cols, h.ID, truncate(h.Name, 23), h.State, h.CPU, h.Memory, h.VMs, h.Cluster)
		if i == m.cursor {
			sb.WriteString(lipgloss.NewStyle().Foreground(primary).Bold(true).Render(row))
		} else {
			sb.WriteString(stateStyle(h.State).Render(row))
		}
		sb.WriteString("\n")
	}
	m.viewport.SetContent(sb.String())
}

func (m hostListModel) View() string {
	status := statusStyle.Render(fmt.Sprintf(" %d Hosts", len(m.hosts)))
	return lipgloss.JoinVertical(lipgloss.Left, m.viewport.View(), status)
}
