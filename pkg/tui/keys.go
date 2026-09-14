package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up             key.Binding
	Down           key.Binding
	PageUp         key.Binding
	PageDown       key.Binding
	Home           key.Binding
	End            key.Binding
	Quit           key.Binding
	Tab            key.Binding
	Enter          key.Binding
	Search         key.Binding
	Refresh        key.Binding
	Reboot         key.Binding
	Poweroff       key.Binding
	Terminate      key.Binding
	Stop           key.Binding
	Migrate        key.Binding
	Logs           key.Binding
	SSH            key.Binding
	Confirm        key.Binding
	Cancel         key.Binding
	Help           key.Binding
	FilterAll      key.Binding
	FilterActive   key.Binding
	FilterStopped  key.Binding
	FilterPoweroff key.Binding
	FilterError    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Search, k.Reboot, k.Poweroff, k.Terminate}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit, k.Tab, k.Help},
		{k.Up, k.Down, k.PageUp, k.PageDown, k.Home, k.End},
		{k.Reboot, k.Poweroff, k.Stop, k.Terminate, k.Migrate, k.Logs},
		{k.FilterAll, k.FilterActive, k.FilterStopped, k.FilterPoweroff, k.FilterError},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "b"),
		key.WithHelp("pgup/b", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "f"),
		key.WithHelp("pgdn/f", "page down"),
	),
	Home: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "top"),
	),
	End: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "bottom"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next view"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "refresh"),
	),
	Reboot: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reboot"),
	),
	Poweroff: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "poweroff"),
	),
	Terminate: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "terminate"),
	),
	Stop: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "stop"),
	),
	Migrate: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "migrate"),
	),
	Logs: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "logs"),
	),
	SSH: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "ssh"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("y", "enter"),
		key.WithHelp("y/enter", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n", "esc"),
		key.WithHelp("n/esc", "cancel"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	FilterAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "all VMs"),
	),
	FilterActive: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "active only"),
	),
	FilterStopped: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "stopped only"),
	),
	FilterPoweroff: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "poweroff only"),
	),
	FilterError: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "error only"),
	),
}
