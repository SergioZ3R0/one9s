package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/scabello/one9s/internal/client"
)

type viewName string

const (
	viewVMs        viewName = "vms"
	viewHosts      viewName = "hosts"
	viewDatastores viewName = "datastores"
	viewACLs       viewName = "acls"
	viewQuotas     viewName = "quotas"
	viewHelp       viewName = "help"
)

var allViews = []viewName{viewVMs, viewHosts, viewDatastores, viewACLs, viewQuotas, viewHelp}

type rootModel struct {
	client client.Client
	ctx    context.Context
	cancel context.CancelFunc

	vmList   vmListModel
	hostList hostListModel
	dsList   dsListModel
	aclList  aclListModel
	qList    quotaListModel

	currentView viewName
	modal       modalState
	fetching    bool
	lastKey     string
	width       int
	height      int
	err         error
}

func NewRootModel(c client.Client) rootModel {
	ctx, cancel := context.WithCancel(context.Background())
	return rootModel{
		client:      c,
		ctx:         ctx,
		cancel:      cancel,
		vmList:      newVMListModel(),
		hostList:    newHostListModel(),
		dsList:      newDSListModel(),
		aclList:     newACLListModel(),
		qList:       newQuotaListModel(),
		currentView: viewVMs,
	}
}

func (m rootModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchVMs(),
		m.fetchHosts(),
		m.fetchDS(),
		m.fetchACLs(),
		m.fetchQuotas(),
	)
}

func (m rootModel) refreshView() tea.Cmd {
	switch m.currentView {
	case viewVMs:
		return m.fetchVMs()
	case viewHosts:
		return m.fetchHosts()
	case viewDatastores:
		return m.fetchDS()
	case viewACLs:
		return m.fetchACLs()
	case viewQuotas:
		return m.fetchQuotas()
	}
	return nil
}

func (m rootModel) fetchVMs() tea.Cmd {
	return func() tea.Msg {
		vms, err := m.client.ListVMs(m.ctx)
		return vmsFetchedMsg{vms: vms, err: err}
	}
}

func (m rootModel) fetchHosts() tea.Cmd {
	return func() tea.Msg {
		hosts, err := m.client.ListHosts(m.ctx)
		return hostsFetchedMsg{hosts: hosts, err: err}
	}
}

func (m rootModel) fetchDS() tea.Cmd {
	return func() tea.Msg {
		ds, err := m.client.ListDatastores(m.ctx)
		return datastoresFetchedMsg{datastores: ds, err: err}
	}
}

func (m rootModel) fetchACLs() tea.Cmd {
	return func() tea.Msg {
		acls, err := m.client.ListACLs(m.ctx)
		return aclsFetchedMsg{acls: acls, err: err}
	}
}

func (m rootModel) fetchQuotas() tea.Cmd {
	return func() tea.Msg {
		q, err := m.client.ListQuotas(m.ctx)
		return quotasFetchedMsg{quotas: q, err: err}
	}
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Modal intercepts ALL keys
	if m.modal.active {
		if k, ok := msg.(tea.KeyMsg); ok {
			switch k.String() {
			case "y", "enter":
				cmd := m.modal.cmd
				m.modal.active = false
				if cmd != nil {
					return m, cmd
				}
				return m, nil
			case "n", "esc":
				m.modal.active = false
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.vmList, _ = m.vmList.Update(msg)
		m.hostList, _ = m.hostList.Update(msg)
		m.dsList, _ = m.dsList.Update(msg)
		m.aclList, _ = m.aclList.Update(msg)
		m.qList, _ = m.qList.Update(msg)
		return m, nil

	case vmsFetchedMsg:
		m.fetching = false
		m.vmList, _ = m.vmList.Update(msg)
	case hostsFetchedMsg:
		m.fetching = false
		m.hostList, _ = m.hostList.Update(msg)
	case datastoresFetchedMsg:
		m.fetching = false
		m.dsList, _ = m.dsList.Update(msg)
	case aclsFetchedMsg:
		m.fetching = false
		m.aclList, _ = m.aclList.Update(msg)
	case quotasFetchedMsg:
		m.fetching = false
		m.qList, _ = m.qList.Update(msg)

	case vmActionMsg:
		return m, m.executeAction(msg.id, msg.action)

	case actionResultMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		cmds = append(cmds, m.fetchVMs())
		return m, tea.Batch(cmds...)

	case errorMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		k := msg.String()
		// Store last keypress for diagnostics
		m.lastKey = k

		// Global keys
		switch k {
		case "q", "ctrl+c":
			m.cancel()
			return m, tea.Quit
		case "tab":
			m.cycleView()
			return m, nil
		case "?":
			if m.currentView == viewHelp {
				m.currentView = viewVMs
			} else {
				m.currentView = viewHelp
			}
			return m, nil
		case "F5":
			if !m.fetching {
				m.fetching = true
				return m, m.refreshView()
			}
			return m, nil
		}

		// VM-only keys
		if m.currentView == viewVMs {
			switch k {
			case "r":
				return m, m.vmList.sendAction("reboot")
			case "s":
				return m, m.vmList.sendAction("poweroff")
			case "x":
				return m, m.vmList.sendAction("stop")
			case "d":
				vmID := m.getSelectedVMID()
				m.modal = newModal(
					"Terminate VM",
					"This will HARD delete the VM. Are you sure?",
					func() tea.Msg {
						return vmActionMsg{id: vmID, action: "terminate-hard"}
					},
				)
				return m, nil
			case "c":
				return m, m.vmList.sshToVM()
			}
		}

		// Forward ALL keys to active sub-model
		switch m.currentView {
		case viewVMs:
			m.vmList, _ = m.vmList.Update(msg)
		case viewHosts:
			m.hostList, _ = m.hostList.Update(msg)
		case viewDatastores:
			m.dsList, _ = m.dsList.Update(msg)
		case viewACLs:
			m.aclList, _ = m.aclList.Update(msg)
		case viewQuotas:
			m.qList, _ = m.qList.Update(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *rootModel) cycleView() {
	for i, v := range allViews {
		if v == m.currentView {
			m.currentView = allViews[(i+1)%len(allViews)]
			return
		}
	}
}

func (m *rootModel) getSelectedVMID() int {
	if m.vmList.cursor >= len(m.vmList.getFiltered()) {
		return -1
	}
	vm := m.vmList.getFiltered()[m.vmList.cursor]
	id := 0
	_, _ = fmt.Sscanf(vm.ID, "%d", &id)
	return id
}

func (m rootModel) executeAction(id int, action string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.VMAction(m.ctx, id, action)
		return actionResultMsg{resource: "vm", id: id, action: action, err: err}
	}
}

func (m rootModel) View() string {
	tabs := []string{
		m.renderTab("1:VMs", m.currentView == viewVMs),
		m.renderTab("2:Hosts", m.currentView == viewHosts),
		m.renderTab("3:DS", m.currentView == viewDatastores),
		m.renderTab("4:ACLs", m.currentView == viewACLs),
		m.renderTab("5:Quotas", m.currentView == viewQuotas),
		m.renderTab("?:Help", m.currentView == viewHelp),
	}
	header := headerStyle.Width(m.width).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			titleStyle.Render(" one9s "),
			" ",
			strings.Join(tabs, " "),
		),
	)

	var errBar string
	if m.err != nil {
		errBar = statePoweroff.Render(" ! " + m.err.Error())
	}

	var content string
	switch m.currentView {
	case viewVMs:
		content = m.vmList.View()
	case viewHosts:
		content = m.hostList.View()
	case viewDatastores:
		content = m.dsList.View()
	case viewACLs:
		content = m.aclList.View()
	case viewQuotas:
		content = m.qList.View()
	case viewHelp:
		content = m.helpView()
	}

	status := statusStyle.Render(fmt.Sprintf(" q:quit tab:switch F5:refresh /:filter ?:help  [last: %s]", m.lastKey))

	return lipgloss.JoinVertical(lipgloss.Left, header, errBar, content, status)
}

func (m rootModel) helpView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" one9s - Keyboard Shortcuts"))
	b.WriteString("\n\n")

	b.WriteString(tableHeader.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString("  tab       Next view\n")
	b.WriteString("  ↑/k       Move up\n")
	b.WriteString("  ↓/j       Move down\n")
	b.WriteString("  pgup/b    Page up\n")
	b.WriteString("  pgdn/f    Page down\n")
	b.WriteString("  g         Go to top\n")
	b.WriteString("  G         Go to bottom\n")
	b.WriteString("  /         Search/filter\n")
	b.WriteString("  esc       Clear filter\n")
	b.WriteString("\n")

	b.WriteString(tableHeader.Render("VM Actions (VMs tab only)"))
	b.WriteString("\n")
	b.WriteString("  r         Reboot\n")
	b.WriteString("  s         Poweroff\n")
	b.WriteString("  x         Stop\n")
	b.WriteString("  d         Terminate (hard)\n")
	b.WriteString("\n")

	b.WriteString(tableHeader.Render("VM State Filters (VMs tab only)"))
	b.WriteString("\n")
	b.WriteString("  a         Show all VMs\n")
	b.WriteString("  u         Active only\n")
	b.WriteString("  o         Stopped only\n")
	b.WriteString("  p         Poweroff only\n")
	b.WriteString("  e         Error only\n")
	b.WriteString("\n")

	b.WriteString(tableHeader.Render("General"))
	b.WriteString("\n")
	b.WriteString("  F5        Refresh current view\n")
	b.WriteString("  ?         Toggle this help\n")
	b.WriteString("  q         Quit\n")

	return b.String()
}

func (m rootModel) renderTab(label string, active bool) string {
	if active {
		return tabActive.Render(label)
	}
	return tabInactive.Render(label)
}
