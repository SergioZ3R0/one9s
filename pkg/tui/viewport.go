package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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
	viewAbout      viewName = "about"
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
	vmDetail    *client.VMDetail // nil when showing list
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

	if m.modal.active {
		switch m.modal.modalType {
		case modalYN:
			if k, ok := msg.(tea.KeyMsg); ok {
				switch k.String() {
				case "y":
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

		case modalTextInput:
			if k, ok := msg.(tea.KeyMsg); ok {
				switch k.String() {
				case "enter":
					input := m.modal.input.Value()
					m.modal.active = false
					if m.modal.expecting == "" {
						if input == "" {
							m.err = fmt.Errorf("input cannot be empty")
							return m, nil
						}
						hostID := m.getSelectedHostID()
						newName := input
						return m, func() tea.Msg {
							err := m.client.HostRename(m.ctx, hostID, newName)
							return hostActionResultMsg{hostID: hostID, action: "rename", err: err}
						}
					}
					if input == m.modal.expecting {
						hostID := m.getSelectedHostID()
						return m, func() tea.Msg {
							err := m.client.HostDelete(m.ctx, hostID)
							return hostActionResultMsg{hostID: hostID, action: "delete", err: err}
						}
					}
					m.err = fmt.Errorf("expected '%s', got '%s'", m.modal.expecting, input)
					return m, nil
				case "esc":
					m.modal.active = false
					return m, nil
				default:
					var cmd tea.Cmd
					m.modal.input, cmd = m.modal.input.Update(msg)
					cmds = append(cmds, cmd)
				}
			}
			return m, tea.Batch(cmds...)

		case modalForm:
			if k, ok := msg.(tea.KeyMsg); ok {
				switch k.String() {
				case "tab":
					for i := range m.modal.formFields {
						if m.modal.formFields[i].input.Focused() {
							m.modal.formFields[i].input.Blur()
							next := (i + 1) % len(m.modal.formFields)
							m.modal.formFields[next].input.Focus()
							break
						}
					}
					return m, nil
				case "shift+tab":
					for i := range m.modal.formFields {
						if m.modal.formFields[i].input.Focused() {
							m.modal.formFields[i].input.Blur()
							prev := (i - 1 + len(m.modal.formFields)) % len(m.modal.formFields)
							m.modal.formFields[prev].input.Focus()
							break
						}
					}
					return m, nil
				case "y":
					values := make(map[string]string)
					for _, f := range m.modal.formFields {
						values[f.key] = f.input.Value()
					}
					m.modal.active = false
					if m.modal.formCmd != nil {
						return m, m.modal.formCmd(values)
					}
					return m, nil
				case "n", "esc":
					m.modal.active = false
					return m, nil
				default:
					for i := range m.modal.formFields {
						if m.modal.formFields[i].input.Focused() {
							updated, cmd := m.modal.formFields[i].input.Update(msg)
							m.modal.formFields[i].input = updated
							cmds = append(cmds, cmd)
							break
						}
					}
				}
			}
			return m, tea.Batch(cmds...)

		case modalInfo:
			if _, ok := msg.(tea.KeyMsg); ok {
				m.modal.active = false
				return m, nil
			}
			return m, nil
		}
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
	case vmDetailFetchedMsg:
		m.vmDetail = &msg.vm

	case quotasFetchedMsg:
		m.fetching = false
		m.qList, _ = m.qList.Update(msg)

	case vmActionMsg:
		return m, m.executeVMAction(msg.id, msg.action)

	case actionResultMsg:
		if msg.err != nil {
			m.err = humanizeError(msg.action, msg.err)
		} else if msg.action == "update" && msg.resource == "quota" {
			m.err = fmt.Errorf("quota updated successfully")
		}
		switch msg.resource {
		case "quota":
			cmds = append(cmds, m.fetchQuotas())
		default:
			cmds = append(cmds, m.fetchVMs())
		}
		return m, tea.Batch(cmds...)

	case hostActionResultMsg:
		if msg.err != nil {
			m.err = humanizeError(msg.action, msg.err)
		}
		cmds = append(cmds, m.fetchHosts())
		return m, tea.Batch(cmds...)

	case errorMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		k := msg.String()
		m.lastKey = k

		// Global keys
		switch k {
		case "q", "ctrl+c":
			m.cancel()
			return m, tea.Quit
		case "tab":
			m.cycleView(1)
			return m, nil
		case "shift+tab":
			m.cycleView(-1)
			return m, nil
		case "?":
			if m.currentView == viewHelp {
				m.currentView = viewVMs
			} else {
				m.currentView = viewHelp
			}
			return m, nil
		case "ctrl+r":
			if !m.fetching {
				m.fetching = true
				return m, m.refreshView()
			}
			return m, nil
		case "ctrl+a":
			if m.modal.active {
				return m, nil
			}
			m.modal = modalState{
				active:    true,
				modalType: modalInfo,
				title:     "about",
				message:   aboutView(),
			}
			return m, nil
		}

		// VM tab: enter to show detail
		if m.currentView == viewVMs && m.vmDetail == nil && k == "enter" {
			vmID := m.getSelectedVMID()
			if vmID >= 0 {
				cmds = append(cmds, m.fetchVMDetail(vmID))
				return m, tea.Batch(cmds...)
			}
		}

		// VM detail: esc to go back
		if m.vmDetail != nil && k == "esc" {
			m.vmDetail = nil
			return m, nil
		}

		// VM tab: terminate modal
		if m.currentView == viewVMs && k == "d" {
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

		// Host tab actions
		if m.currentView == viewHosts {
			hostID := m.getSelectedHostID()
			hostName := m.getSelectedHostName()

			switch k {
			case "e":
				return m, m.executeHostAction(hostID, "enable")
			case "d":
				return m, m.executeHostAction(hostID, "disable")
			case "o":
				m.modal = newModal(
					"Offline Host",
					fmt.Sprintf("Set host '%s' offline?\nThis host will stop running VMs.", hostName),
					func() tea.Msg {
						err := m.client.HostAction(m.ctx, hostID, "offline")
						return hostActionResultMsg{hostID: hostID, action: "offline", err: err}
					},
				)
				return m, nil
			case "x":
				m.modal = newTextInputModal(
					"Delete Host",
					fmt.Sprintf("Delete host '%s' from OpenNebula?", hostName),
					"yes",
					func() tea.Msg {
						err := m.client.HostDelete(m.ctx, hostID)
						return hostActionResultMsg{hostID: hostID, action: "delete", err: err}
					},
				)
				return m, nil
			case "n":
				m.modal = newTextInputModal(
					"Rename Host",
					fmt.Sprintf("Enter new name for '%s':", hostName),
					"",
					func() tea.Msg {
						newName := m.modal.input.Value()
						err := m.client.HostRename(m.ctx, hostID, newName)
						return hostActionResultMsg{hostID: hostID, action: "rename", err: err}
					},
				)
				return m, nil
			}
		}

		// Quota tab actions
		if m.currentView == viewQuotas && k == "e" {
			q := m.getSelectedQuota()
			if q.UserID < 0 {
				return m, nil
			}
			fields := make([]formField, 7)
			for i := range fields {
				ti := textinput.New()
				ti.CharLimit = 12
				fields[i] = formField{input: ti}
			}
			fields[0].label = "VMs"
			fields[0].key = "vms"
			fields[1].label = "CPU"
			fields[1].key = "cpu"
			fields[2].label = "Memory (MB)"
			fields[2].key = "memory"
			fields[3].label = "Running VMs"
			fields[3].key = "running"
			fields[4].label = "Images"
			fields[4].key = "images"
			fields[5].label = "Size (MB)"
			fields[5].key = "size"
			fields[6].label = "Leases"
			fields[6].key = "leases"

			fields[0].input.SetValue(fmt.Sprintf("%d", q.VMsLimit))
			fields[1].input.SetValue(fmt.Sprintf("%d", q.CPULimit))
			fields[2].input.SetValue(fmt.Sprintf("%d", q.MemoryLimit))
			fields[3].input.SetValue(fmt.Sprintf("%d", q.RunningVMsLimit))
			fields[4].input.SetValue(fmt.Sprintf("%d", q.ImagesLimit))
			fields[5].input.SetValue(fmt.Sprintf("%d", q.SizeLimit))
			fields[6].input.SetValue(fmt.Sprintf("%d", q.LeasesLimit))

			userID := q.UserID
			m.modal = newFormModal(
				fmt.Sprintf("Edit Quota: %s", q.Entity),
				fields,
				func(values map[string]string) tea.Cmd {
					tpl := buildQuotaTemplate(values)
					return func() tea.Msg {
						err := m.client.QuotaUpdate(m.ctx, userID, tpl)
						return actionResultMsg{resource: "quota", id: userID, action: "update", err: err}
					}
				},
			)
			return m, nil
		}

		// Forward ALL keys to active sub-model
		switch m.currentView {
		case viewVMs:
			var cmd tea.Cmd
			m.vmList, cmd = m.vmList.Update(msg)
			cmds = append(cmds, cmd)
		case viewHosts:
			var cmd tea.Cmd
			m.hostList, cmd = m.hostList.Update(msg)
			cmds = append(cmds, cmd)
		case viewDatastores:
			var cmd tea.Cmd
			m.dsList, cmd = m.dsList.Update(msg)
			cmds = append(cmds, cmd)
		case viewACLs:
			var cmd tea.Cmd
			m.aclList, cmd = m.aclList.Update(msg)
			cmds = append(cmds, cmd)
		case viewQuotas:
			var cmd tea.Cmd
			m.qList, cmd = m.qList.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *rootModel) cycleView(dir int) {
	for i, v := range allViews {
		if v == m.currentView {
			m.currentView = allViews[(i+dir+len(allViews))%len(allViews)]
			m.vmDetail = nil
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

func (m *rootModel) getSelectedHostID() int {
	if m.hostList.cursor >= len(m.hostList.hosts) {
		return -1
	}
	h := m.hostList.hosts[m.hostList.cursor]
	id := 0
	_, _ = fmt.Sscanf(h.ID, "%d", &id)
	return id
}

func (m *rootModel) getSelectedHostName() string {
	if m.hostList.cursor >= len(m.hostList.hosts) {
		return ""
	}
	return m.hostList.hosts[m.hostList.cursor].Name
}

func (m *rootModel) getSelectedQuota() client.QuotaInfo {
	if m.qList.cursor >= len(m.qList.quotas) {
		return client.QuotaInfo{UserID: -1}
	}
	q := m.qList.quotas[m.qList.cursor]
	return client.QuotaInfo{
		UserID:          q.UserID,
		Entity:          q.Entity,
		VMsLimit:        q.VMsLimit,
		CPULimit:        q.CPULimit,
		MemoryLimit:     q.MemoryLimit,
		RunningVMsLimit: q.RunningVMsLimit,
		ImagesLimit:     q.ImagesLimit,
		SizeLimit:       q.SizeLimit,
		LeasesLimit:     q.LeasesLimit,
	}
}

func buildQuotaTemplate(values map[string]string) string {
	var vmParts []string
	if v := parseQuotaVal(values["vms"]); v != -1 {
		vmParts = append(vmParts, fmt.Sprintf("VMS = %d", v))
	}
	if v := parseQuotaVal(values["cpu"]); v != -1 {
		vmParts = append(vmParts, fmt.Sprintf("CPU = %d", v))
	}
	if v := parseQuotaVal(values["memory"]); v != -1 {
		vmParts = append(vmParts, fmt.Sprintf("MEMORY = %d", v))
	}
	if v := parseQuotaVal(values["running"]); v != -1 {
		vmParts = append(vmParts, fmt.Sprintf("RUNNING_VMS = %d", v))
	}

	var parts []string
	if len(vmParts) > 0 {
		parts = append(parts, fmt.Sprintf("VM = [\n  %s\n]", strings.Join(vmParts, ",\n  ")))
	}
	return strings.Join(parts, "\n")
}

func parseQuotaVal(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return -2
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return -2
	}
	return v
}

func (m rootModel) executeVMAction(id int, action string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.VMAction(m.ctx, id, action)
		return actionResultMsg{resource: "vm", id: id, action: action, err: err}
	}
}

func (m rootModel) fetchVMDetail(id int) tea.Cmd {
	return func() tea.Msg {
		vm, err := m.client.GetVMDetailInfo(m.ctx, id)
		if err != nil {
			return errorMsg{err: err}
		}
		return vmDetailFetchedMsg{vm: vm}
	}
}

func (m rootModel) executeHostAction(id int, action string) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch action {
		case "delete":
			err = m.client.HostDelete(m.ctx, id)
		default:
			err = m.client.HostAction(m.ctx, id, action)
		}
		return hostActionResultMsg{hostID: id, action: action, err: err}
	}
}

func (m rootModel) View() string {
	if m.modal.active {
		return m.renderModalView()
	}

	// --- Header ---
	line3 := separatorStyle.Render(strings.Repeat("─", max(0, m.width-2)))

	left := appNameStyle.Render("one9s")
	right := connectionStyle.Render("Connected • OpenNebula")
	gap := max(0, m.width-lipgloss.Width(left)-lipgloss.Width(right)-2)
	line1 := left + strings.Repeat(" ", gap) + right

	tabItems := []string{
		m.renderTab("1:VMs", m.currentView == viewVMs),
		m.renderTab("2:Hosts", m.currentView == viewHosts),
		m.renderTab("3:DS", m.currentView == viewDatastores),
		m.renderTab("4:ACLs", m.currentView == viewACLs),
		m.renderTab("5:Quotas", m.currentView == viewQuotas),
		m.renderTab("?:Help", m.currentView == viewHelp),
	}
	line2 := strings.Join(tabItems, "")

	header := lipgloss.JoinVertical(lipgloss.Left, line1, line3, line2, line3)

	// --- Error bar ---
	var errBar string
	if m.err != nil {
		msg := m.err.Error()
		if strings.HasPrefix(msg, "quota updated") {
			errBar = stateRunning.Render(" ✓ " + msg)
		} else {
			errBar = statePoweroff.Render(" ! " + msg)
		}
	}

	// --- VMs tab: split view (list + detail) ---
	if m.currentView == viewVMs {
		listW := m.width/2 - 2
		viewH := m.height - 7
		if viewH < 1 {
			viewH = 1
		}

		// Left: VM list
		listContent := m.vmList.View()
		listLines := strings.Split(listContent, "\n")
		var listVisible []string
		for i, line := range listLines {
			if i >= viewH {
				break
			}
			if lipgloss.Width(line) > listW {
				line = truncate(line, listW-1)
			}
			listVisible = append(listVisible, line)
		}
		leftPanel := panelStyle.Width(max(30, listW)).Render(strings.Join(listVisible, "\n"))

		// Right: detail pane
		detailW := m.width/2 - 2
		var detailContent string
		if m.vmDetail != nil {
			detailContent = vmDetailView(*m.vmDetail)
		} else {
			detailContent = filterStyle.Render("Select a VM") + statusStyle.Render("\n\nPress enter on a VM\nto view details")
		}
		detailLines := strings.Split(detailContent, "\n")
		var detailVisible []string
		for i, line := range detailLines {
			if i >= viewH {
				break
			}
			if lipgloss.Width(line) > detailW {
				line = truncate(line, detailW-1)
			}
			detailVisible = append(detailVisible, line)
		}
		rightPanel := panelStyle.Width(max(30, detailW)).Render(strings.Join(detailVisible, "\n"))

		split := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

		footerSep := separatorStyle.Render(strings.Repeat("─", max(0, m.width-2)))
		footerText := statusStyle.Render(" tab:switch • ↑↓:navigate • enter:detail • /:filter • ? help • q quit")
		footer := lipgloss.JoinVertical(lipgloss.Left, footerSep, footerText)

		inner := lipgloss.JoinVertical(lipgloss.Left, header, errBar, split, footer)
		return frameStyle.Render(inner)
	}

	// --- Other tabs: single panel ---
	var content string
	switch m.currentView {
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

	panelW := m.width - 4
	viewH := m.height - 7
	if viewH < 1 {
		viewH = 1
	}
	lines := strings.Split(content, "\n")
	var visible []string
	for i, line := range lines {
		if i >= viewH {
			break
		}
		maxW := panelW - 4
		if maxW < 10 {
			maxW = 10
		}
		if lipgloss.Width(line) > maxW {
			line = truncate(line, maxW-1)
		}
		visible = append(visible, line)
	}
	panel := panelStyle.Render(strings.Join(visible, "\n"))

	footerSep := separatorStyle.Render(strings.Repeat("─", max(0, m.width-2)))
	footerText := statusStyle.Render(" tab:switch • ↑↓:navigate • /:filter • ? help • q quit")
	footer := lipgloss.JoinVertical(lipgloss.Left, footerSep, footerText)

	inner := lipgloss.JoinVertical(lipgloss.Left, header, errBar, panel, footer)
	return frameStyle.Render(inner)
}

func (m rootModel) renderModalView() string {
	// Calculate modal width from content
	modalContent := m.modal.View()
	modalLines := strings.Split(modalContent, "\n")

	// Find widest line in content
	maxLineW := 0
	for _, line := range modalLines {
		w := lipgloss.Width(line)
		if w > maxLineW {
			maxLineW = w
		}
	}

	// Modal width = content width + 4 padding (2 left + 2 right)
	modalW := maxLineW + 4
	// Ensure minimum width and fit in terminal
	modalW = min(modalW, m.width-2)
	if modalW < 40 {
		modalW = 40
	}

	var box strings.Builder
	box.WriteString(strings.Repeat("─", modalW))
	box.WriteString("\n")
	for _, line := range modalLines {
		w := lipgloss.Width(line)
		padRight := modalW - w - 2 // 2 for left padding
		if padRight < 0 {
			padRight = 0
		}
		box.WriteString("  " + line + strings.Repeat(" ", padRight))
		box.WriteString("\n")
	}
	box.WriteString(strings.Repeat("─", modalW))

	boxStr := box.String()
	boxLines := strings.Split(boxStr, "\n")
	totalH := len(boxLines)
	yOffset := max(0, (m.height-totalH)/2)
	xOffset := max(0, (m.width-modalW)/2)

	var screen strings.Builder
	for i := 0; i < yOffset; i++ {
		screen.WriteString("\n")
	}
	for _, line := range boxLines {
		screen.WriteString(strings.Repeat(" ", xOffset))
		screen.WriteString(line)
		screen.WriteString("\n")
	}

	return screen.String()
}

func (m rootModel) helpView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" one9s - Keyboard Shortcuts"))
	b.WriteString("\n\n")

	b.WriteString(tableHeader.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString("  tab       Next view\n")
	b.WriteString("  shift+tab Previous view\n")
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
	b.WriteString("  r         Reboot (ACTIVE only)\n")
	b.WriteString("  s         Stop (ACTIVE only)\n")
	b.WriteString("  u         Resume/Start (POWEROFF/SUSPENDED/STOPPED/UNDEPLOYED)\n")
	b.WriteString("  x         Suspend (ACTIVE only)\n")
	b.WriteString("  d         Terminate (hard, any state)\n")
	b.WriteString("\n")

	b.WriteString(tableHeader.Render("VM State Filters (VMs tab only)"))
	b.WriteString("\n")
	b.WriteString("  a         Show all VMs\n")
	b.WriteString("  i         Active only\n")
	b.WriteString("  o         Stopped only\n")
	b.WriteString("  p         Poweroff only\n")
	b.WriteString("  e         Error only\n")
	b.WriteString("\n")

	b.WriteString(tableHeader.Render("Host Actions (Hosts tab only)"))
	b.WriteString("\n")
	b.WriteString("  e         Enable host\n")
	b.WriteString("  d         Disable host\n")
	b.WriteString("  o         Offline host (confirm)\n")
	b.WriteString("  x         Delete host (type 'yes')\n")
	b.WriteString("  n         Rename host (type new name)\n")
	b.WriteString("\n")

	b.WriteString(tableHeader.Render("Quota Actions (Quotas tab only)"))
	b.WriteString("\n")
	b.WriteString("  e         Edit user quota\n")
	b.WriteString("\n")

	b.WriteString(tableHeader.Render("General"))
	b.WriteString("\n")
	b.WriteString("  Ctrl+R    Refresh current view\n")
	b.WriteString("  Ctrl+A    About one9s\n")
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

func humanizeError(action string, err error) error {
	msg := err.Error()
	if strings.Contains(msg, "Not authorized") || strings.Contains(msg, "Authorization") {
		return fmt.Errorf("permission denied: requires %s", permissionNeeded(action))
	}
	return err
}

func permissionNeeded(action string) string {
	switch action {
	case "reboot", "stop", "suspend", "poweroff", "resume":
		return "VM:MANAGE"
	case "terminate-hard":
		return "VM:ADMIN"
	case "enable", "disable", "offline":
		return "HOST:ADMIN"
	case "delete":
		return "HOST:ADMIN"
	case "rename":
		return "HOST:ADMIN"
	case "update":
		return "Quota:ADMIN"
	}
	return "sufficient permissions"
}
