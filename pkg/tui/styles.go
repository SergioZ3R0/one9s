package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// ANSI 256 palette - readable on dark terminals
	primary = lipgloss.Color("129") // bright purple
	accent  = lipgloss.Color("75")  // bright cyan
	success = lipgloss.Color("114") // bright green
	warning = lipgloss.Color("221") // bright yellow
	danger  = lipgloss.Color("203") // bright red
	dim     = lipgloss.Color("245") // gray
	bright  = lipgloss.Color("231") // white

	// Header bar
	headerStyle = lipgloss.NewStyle().
			Background(primary).
			Foreground(bright).
			Bold(true).
			Padding(0, 1)

	// Status bar
	statusStyle = lipgloss.NewStyle().
			Foreground(dim).
			Padding(0, 1)

	// Tab styles
	tabActive = lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			Underline(true)
	tabInactive = lipgloss.NewStyle().
			Foreground(dim)

	// Table
	tableHeader = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(dim)

	// State colors
	stateRunning  = lipgloss.NewStyle().Foreground(success)
	statePoweroff = lipgloss.NewStyle().Foreground(danger)
	stateSuspend  = lipgloss.NewStyle().Foreground(warning)
	stateDefault  = lipgloss.NewStyle().Foreground(dim)

	// Filter input
	filterStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	// Title
	titleStyle = lipgloss.NewStyle().
			Foreground(primary).
			Bold(true)

	// VM list cursor
	cursorStyle = lipgloss.NewStyle().Foreground(accent).Bold(true).Reverse(true)
)

func stateStyle(state string) lipgloss.Style {
	switch {
	case strings.HasPrefix(state, "ACTIVE"):
		return stateRunning
	case state == "POWEROFF" || state == "UNDEPLOYED" || strings.Contains(state, "SHUTDOWN"):
		return statePoweroff
	case state == "STOPPED" || state == "SUSPENDED" || state == "HOLD":
		return stateSuspend
	case strings.Contains(state, "FAILURE") || strings.HasSuffix(state, "/UNKNOWN"):
		return statePoweroff
	default:
		return stateDefault
	}
}
