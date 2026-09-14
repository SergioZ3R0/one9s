package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/scabello/one9s/internal/client"
)

const pollInterval = 5 * time.Second

type viewName string

const (
	viewVMs        viewName = "vms"
	viewHosts      viewName = "hosts"
	viewDatastores viewName = "datastores"
	viewACLs       viewName = "acls"
	viewQuotas     viewName = "quotas"
)

var allViews = []viewName{viewVMs, viewHosts, viewDatastores, viewACLs, viewQuotas}

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
		m.tick(),
	)
}

func (m rootModel) tick() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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

	if m.modal.active {
		return m.updateModal(msg)
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

	case tickMsg:
		switch m.currentView {
		case viewVMs:
			cmds = append(cmds, m.fetchVMs())
		case viewHosts:
			cmds = append(cmds, m.fetchHosts())
		case viewDatastores:
			cmds = append(cmds, m.fetchDS())
		case viewACLs:
			cmds = append(cmds, m.fetchACLs())
		case viewQuotas:
			cmds = append(cmds, m.fetchQuotas())
		}
		cmds = append(cmds, m.tick())
		return m, tea.Batch(cmds...)

	case vmsFetchedMsg:
		m.vmList, _ = m.vmList.Update(msg)
	case hostsFetchedMsg:
		m.hostList, _ = m.hostList.Update(msg)
	case datastoresFetchedMsg:
		m.dsList, _ = m.dsList.Update(msg)
	case aclsFetchedMsg:
		m.aclList, _ = m.aclList.Update(msg)
	case quotasFetchedMsg:
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
		switch {
		case key.Matches(msg, keys.Quit):
			m.cancel()
			return m, tea.Quit

		case key.Matches(msg, keys.Tab):
			m.cycleView()
			return m, nil

		case key.Matches(msg, keys.Reboot):
			if m.currentView == viewVMs {
				return m, m.vmList.sendAction("reboot")
			}
		case key.Matches(msg, keys.Poweroff):
			if m.currentView == viewVMs {
				return m, m.vmList.sendAction("poweroff")
			}
		case key.Matches(msg, keys.Stop):
			if m.currentView == viewVMs {
				return m, m.vmList.sendAction("stop")
			}
		case key.Matches(msg, keys.Terminate):
			if m.currentView == viewVMs {
				vmID := m.getSelectedVMID()
				m.modal = newModal(
					"Terminate VM",
					"This will HARD delete the VM. Are you sure?",
					func() tea.Msg {
						return vmActionMsg{id: vmID, action: "terminate-hard"}
					},
				)
				return m, nil
			}
		case key.Matches(msg, keys.SSH):
			if m.currentView == viewVMs {
				return m, m.vmList.sshToVM()
			}
		}

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

func (m rootModel) updateModal(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m rootModel) View() string {
	tabs := []string{
		m.renderTab("1:VMs", m.currentView == viewVMs),
		m.renderTab("2:Hosts", m.currentView == viewHosts),
		m.renderTab("3:DS", m.currentView == viewDatastores),
		m.renderTab("4:ACLs", m.currentView == viewACLs),
		m.renderTab("5:Quotas", m.currentView == viewQuotas),
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
	}

	status := statusStyle.Render(" q:quit  tab:switch  /:filter")

	view := lipgloss.JoinVertical(lipgloss.Left, header, errBar, content, status)

	if m.modal.active {
		modalW := min(60, m.width-4)
		modalH := 8
		x := (m.width - modalW) / 2
		y := (m.height - modalH) / 2
		modalStr := modalBox.Width(modalW).Render(m.modal.View())
		lines := strings.Split(view, "\n")
		for i, line := range lines {
			if i >= y && i < y+modalH {
				modalLine := modalStr
				if i-y < len(strings.Split(modalStr, "\n")) {
					modalLine = strings.Split(modalStr, "\n")[i-y]
				}
				prefix := ""
				if x > 0 && x <= len(line) {
					prefix = line[:x]
				}
				suffix := ""
				mlWidth := lipgloss.Width(modalLine)
				if x+mlWidth < len(line) {
					suffix = line[x+mlWidth:]
				}
				lines[i] = prefix + modalLine + suffix
			}
		}
		view = strings.Join(lines, "\n")
	}

	return view
}

func (m rootModel) renderTab(label string, active bool) string {
	if active {
		return tabActive.Render(label)
	}
	return tabInactive.Render(label)
}
