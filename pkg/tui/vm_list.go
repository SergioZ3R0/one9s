package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type vmListModel struct {
	vms         []VMRow
	cursor      int
	scroll      int
	filter      textinput.Model
	filtering   bool
	filterStr   string
	stateFilter string
	width       int
	height      int

	// pre-rendered lines (with ANSI styles applied)
	lines       []string
	filtered    []VMRow
	filterDirty bool
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

var (
	cursorStyle = lipgloss.NewStyle().Foreground(accent).Bold(true).Reverse(true)
)

func newVMListModel() vmListModel {
	ti := textinput.New()
	ti.Placeholder = "filter vms..."
	ti.CharLimit = 64
	return vmListModel{
		filter:      ti,
		filterDirty: true,
	}
}

func (m vmListModel) Init() tea.Cmd { return nil }

func (m vmListModel) Update(msg tea.Msg) (vmListModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
		m.filterDirty = true
		m.rebuildLines()
		return m, nil

	case tea.KeyMsg:
		if m.filtering {
			switch msg.String() {
			case "enter":
				m.filtering = false
				m.filterStr = m.filter.Value()
				m.filter.Blur()
				m.filterDirty = true
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
			m.rebuildLines()
			return m, tea.Batch(cmds...)
		}

		viewH := m.viewHeight()

		switch {
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.getFiltered())-1 {
				m.cursor++
				if m.cursor >= m.scroll+viewH {
					m.scroll = m.cursor - viewH + 1
				}
			}
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scroll {
					m.scroll = m.cursor
				}
			}
		case key.Matches(msg, keys.PageDown):
			m.cursor = min(m.cursor+viewH, len(m.getFiltered())-1)
			m.scroll = min(m.scroll+viewH, max(0, len(m.getFiltered())-viewH))
		case key.Matches(msg, keys.PageUp):
			m.cursor = max(m.cursor-viewH, 0)
			m.scroll = max(m.scroll-viewH, 0)
		case key.Matches(msg, keys.Home):
			m.cursor = 0
			m.scroll = 0
		case key.Matches(msg, keys.End):
			m.cursor = max(0, len(m.getFiltered())-1)
			m.scroll = max(0, m.cursor-viewH+1)

		case key.Matches(msg, keys.Search):
			m.filtering = true
			m.filter.Focus()
			cmds = append(cmds, textinput.Blink)

		// State filters
		case key.Matches(msg, keys.FilterAll):
			m.stateFilter = ""
			m.filterDirty = true
			m.cursor = 0
			m.scroll = 0
			m.rebuildLines()
		case key.Matches(msg, keys.FilterActive):
			m.stateFilter = "active"
			m.filterDirty = true
			m.cursor = 0
			m.scroll = 0
			m.rebuildLines()
		case key.Matches(msg, keys.FilterStopped):
			m.stateFilter = "stopped"
			m.filterDirty = true
			m.cursor = 0
			m.scroll = 0
			m.rebuildLines()
		case key.Matches(msg, keys.FilterPoweroff):
			m.stateFilter = "poweroff"
			m.filterDirty = true
			m.cursor = 0
			m.scroll = 0
			m.rebuildLines()
		case key.Matches(msg, keys.FilterError):
			m.stateFilter = "error"
			m.filterDirty = true
			m.cursor = 0
			m.scroll = 0
			m.rebuildLines()

		// Actions
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

func (m *vmListModel) viewHeight() int {
	h := m.height - 4 // header + status + filter
	if h < 1 {
		h = 1
	}
	return h
}

func (m *vmListModel) getFiltered() []VMRow {
	if m.filterDirty {
		m.filtered = m.computeFiltered()
		m.filterDirty = false
	}
	return m.filtered
}

func matchesStateFilter(state, filter string) bool {
	if filter == "" {
		return true
	}
	switch filter {
	case "active":
		return strings.HasPrefix(state, "ACTIVE")
	case "stopped":
		return state == "STOPPED" || state == "SUSPENDED" || state == "HOLD"
	case "poweroff":
		return strings.Contains(state, "POWEROFF") || strings.Contains(state, "SHUTDOWN") || strings.Contains(state, "UNDEPLOYED")
	case "error":
		return strings.Contains(state, "FAILURE") || strings.Contains(state, "UNKNOWN") ||
			state == "CLONING_FAILURE" || state == "INIT"
	default:
		return true
	}
}

func (m *vmListModel) computeFiltered() []VMRow {
	out := make([]VMRow, 0, len(m.vms))
	for _, v := range m.vms {
		if !matchesStateFilter(v.State, m.stateFilter) {
			continue
		}
		if m.filterStr != "" {
			needle := strings.ToLower(m.filterStr)
			haystack := strings.ToLower(
				v.ID + " " + v.Name + " " + v.State + " " +
					v.IP + " " + v.Host + " " + v.User + " " + v.DeployID,
			)
			if !strings.Contains(haystack, needle) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func (m *vmListModel) rebuildLines() {
	filtered := m.getFiltered()
	m.lines = make([]string, 0, len(filtered)+1)

	// Header
	m.lines = append(m.lines, tableHeader.Render(
		fmt.Sprintf("%-6s %-24s %-24s %-12s %-6s %-8s %-16s %-16s",
			"ID", "NAME", "STATE", "USER", "CPU", "MEM", "IP", "HOST"),
	))

	// Data rows - pre-render with styles
	for i, v := range filtered {
		row := fmt.Sprintf("%-6s %-24s %-24s %-12s %-6s %-8s %-16s %-16s",
			v.ID, truncate(v.Name, 23), truncate(v.State, 23),
			truncate(v.User, 11), v.CPU, v.Memory,
			truncate(v.IP, 15), truncate(v.Host, 15))
		if i == m.cursor {
			m.lines = append(m.lines, cursorStyle.Render(row))
		} else {
			m.lines = append(m.lines, stateStyle(v.State).Render(row))
		}
	}

	// Clamp scroll
	viewH := m.viewHeight()
	if m.cursor >= len(filtered) {
		m.cursor = max(0, len(filtered)-1)
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

func (m vmListModel) sendAction(action string) tea.Cmd {
	return func() tea.Msg {
		filtered := m.getFiltered()
		if m.cursor >= len(filtered) {
			return nil
		}
		vm := filtered[m.cursor]
		id := 0
		_, _ = fmt.Sscanf(vm.ID, "%d", &id)
		return vmActionMsg{id: id, action: action}
	}
}

func (m vmListModel) sshToVM() tea.Cmd {
	return func() tea.Msg {
		filtered := m.getFiltered()
		if m.cursor >= len(filtered) {
			return nil
		}
		vm := filtered[m.cursor]
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

func stateFilterLabel(f string) string {
	switch f {
	case "active":
		return "Active"
	case "stopped":
		return "Stopped"
	case "poweroff":
		return "Poweroff"
	case "error":
		return "Error"
	default:
		return "All"
	}
}

func (m vmListModel) View() string {
	var parts []string

	// Filter bar
	if m.filtering {
		parts = append(parts, filterStyle.Render("Filter: ")+m.filter.View())
	} else if m.filterStr != "" {
		parts = append(parts, filterStyle.Render("Filter: ")+m.filterStr+" [esc]")
	}

	// Visible window of pre-rendered lines
	viewH := m.viewHeight()
	end := min(m.scroll+viewH, len(m.lines))
	visible := m.lines[m.scroll:end]
	parts = append(parts, strings.Join(visible, "\n"))

	// Status bar with state filter keys
	filtered := m.getFiltered()
	stateLbl := stateFilterLabel(m.stateFilter)
	status := fmt.Sprintf(" %d/%d VMs  [%s]  a:all u:active o:stop p:off e:error",
		len(filtered), len(m.vms), stateLbl)
	parts = append(parts, statusStyle.Render(status))

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
