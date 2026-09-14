package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type vmListModel struct {
	vms       []VMRow
	viewport  viewport.Model
	cursor    int
	filter    textinput.Model
	filtering bool
	filterStr string
	width     int
	height    int
}

type VMRow struct {
	ID       string
	Name     string
	State    string
	User     string
	CPU      string
	Memory   string
	IP       string
	Host     string
	DeployID string
}

func newVMListModel() vmListModel {
	ti := textinput.New()
	ti.Placeholder = "filter vms..."
	ti.CharLimit = 64
	return vmListModel{
		filter:   ti,
		viewport: viewport.New(80, 24),
	}
}

func (m vmListModel) Init() tea.Cmd {
	return nil
}

func (m vmListModel) Update(msg tea.Msg) (vmListModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4 // header + status + filter
		return m, nil

	case vmsFetchedMsg:
		if msg.err != nil {
			return m, nil
		}
		m.vms = make([]VMRow, 0, len(msg.vms))
		for _, v := range msg.vms {
			m.vms = append(m.vms, VMRow{
				ID:       fmt.Sprintf("%d", v.ID),
				Name:     v.Name,
				State:    v.State,
				User:     v.User,
				CPU:      v.CPU,
				Memory:   v.Memory,
				IP:       v.IP,
				Host:     v.Host,
				DeployID: v.DeployID,
			})
		}
		m.updateViewportContent()
		return m, nil

	case tea.KeyMsg:
		if m.filtering {
			switch msg.String() {
			case "enter":
				m.filtering = false
				m.filterStr = m.filter.Value()
				m.filter.Blur()
				m.updateViewportContent()
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
			m.updateViewportContent()
			return m, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.filteredVMs())-1 {
				m.cursor++
				m.ensureVisible()
			}
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
		case key.Matches(msg, keys.Search):
			m.filtering = true
			m.filter.Focus()
			cmds = append(cmds, textinput.Blink)
		case key.Matches(msg, keys.Reboot):
			return m, m.sendAction("reboot")
		case key.Matches(msg, keys.Poweroff):
			return m, m.sendAction("poweroff")
		case key.Matches(msg, keys.Stop):
			return m, m.sendAction("stop")
		case key.Matches(msg, keys.Terminate):
			return m, m.sendAction("terminate-hard")
		case key.Matches(msg, keys.SSH):
			return m, m.sshToVM()
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *vmListModel) sendAction(action string) tea.Cmd {
	return func() tea.Msg {
		if m.cursor >= len(m.filteredVMs()) {
			return nil
		}
		vm := m.filteredVMs()[m.cursor]
		id := 0
		fmt.Sscanf(vm.ID, "%d", &id)
		return vmActionMsg{id: id, action: action}
	}
}

func (m *vmListModel) sshToVM() tea.Cmd {
	return func() tea.Msg {
		if m.cursor >= len(m.filteredVMs()) {
			return nil
		}
		vm := m.filteredVMs()[m.cursor]
		if vm.IP == "" {
			return errorMsg{err: fmt.Errorf("VM %s has no IP", vm.Name)}
		}
		cmd := exec.Command("ssh", vm.IP)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
		return nil
	}
}

type vmActionMsg struct {
	id     int
	action string
}

func (m *vmListModel) filteredVMs() []VMRow {
	if m.filterStr == "" {
		return m.vms
	}
	needle := strings.ToLower(m.filterStr)
	var out []VMRow
	for _, v := range m.vms {
		haystack := strings.ToLower(
			strings.Join([]string{v.ID, v.Name, v.State, v.IP, v.Host, v.User, v.DeployID}, " "),
		)
		if strings.Contains(haystack, needle) {
			out = append(out, v)
		}
	}
	return out
}

func (m *vmListModel) updateViewportContent() {
	filtered := m.filteredVMs()
	var sb strings.Builder

	// Table header
	cols := []struct{ title, w string }{
		{"ID", "6"}, {"NAME", "24"}, {"STATE", "24"}, {"USER", "12"},
		{"CPU", "6"}, {"MEM", "8"}, {"IP", "16"}, {"HOST", "16"},
	}
	for _, c := range cols {
		sb.WriteString(tableHeader.Width(0).Render(
			lipgloss.NewStyle().Width(0).Render(
				fmt.Sprintf("%-6s", c.title),
			),
		))
		sb.WriteString(" ")
	}
	sb.WriteString("\n")

	for i, v := range filtered {
		row := fmt.Sprintf("%-6s %-24s %-24s %-12s %-6s %-8s %-16s %-16s",
			v.ID, truncate(v.Name, 23), truncate(v.State, 23),
			truncate(v.User, 11), v.CPU, v.Memory,
			truncate(v.IP, 15), truncate(v.Host, 15))
		if i == m.cursor {
			sb.WriteString(lipgloss.NewStyle().Foreground(primary).Bold(true).Render(row))
		} else {
			sb.WriteString(stateStyle(v.State).Render(row))
		}
		sb.WriteString("\n")
	}

	m.viewport.SetContent(sb.String())
	if m.cursor >= len(filtered) {
		m.cursor = max(0, len(filtered)-1)
	}
}

func (m *vmListModel) ensureVisible() {
	filtered := m.filteredVMs()
	if m.cursor >= len(filtered) {
		return
	}
	lineHeight := 1
	m.viewport.GotoTop()
	m.viewport.LineDown(m.cursor * lineHeight)
}

func (m vmListModel) View() string {
	var parts []string

	// Filter bar
	if m.filtering {
		parts = append(parts, filterStyle.Render("Filter: ")+m.filter.View())
	} else if m.filterStr != "" {
		parts = append(parts, filterStyle.Render("Filter: ")+m.filterStr+" [esc to clear]")
	}

	// Table
	parts = append(parts, m.viewport.View())

	// Status
	filtered := m.filteredVMs()
	status := statusStyle.Render(fmt.Sprintf(" %d/%d VMs", len(filtered), len(m.vms)))
	parts = append(parts, status)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
