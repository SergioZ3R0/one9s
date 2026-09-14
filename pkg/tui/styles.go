package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Base colors
	primary   = lipgloss.Color("#7C3AED")
	accent    = lipgloss.Color("#06B6D4")
	success   = lipgloss.Color("#22C55E")
	warning   = lipgloss.Color("#F59E0B")
	danger    = lipgloss.Color("#EF4444")
	dim       = lipgloss.Color("#6B7280")
	bright    = lipgloss.Color("#F9FAFB")

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

	// Modal
	modalBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(danger).
			Padding(1, 2)

	// Filter input
	filterStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	// Title
	titleStyle = lipgloss.NewStyle().
			Foreground(primary).
			Bold(true)
)

func stateStyle(state string) lipgloss.Style {
	switch {
	case strings.Contains(state, "RUNNING") || strings.Contains(state, "ACTIVE"):
		return stateRunning
	case strings.Contains(state, "POWEROFF") || strings.Contains(state, "SHUTDOWN"):
		return statePoweroff
	case strings.Contains(state, "SUSPENDED") || strings.Contains(state, "STOPPED"):
		return stateSuspend
	default:
		return stateDefault
	}
}
