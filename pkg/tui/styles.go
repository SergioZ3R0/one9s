package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// OpenNebula-inspired blue palette
	colorPrimary = lipgloss.Color("39")  // bright blue (OpenNebula brand)
	colorAccent  = lipgloss.Color("75")  // light blue
	colorWhite   = lipgloss.Color("15")  // bright white
	colorGray    = lipgloss.Color("8")   // dark gray
	colorDim     = lipgloss.Color("240") // frame gray
	colorGreen   = lipgloss.Color("10")  // green
	colorRed     = lipgloss.Color("9")   // red
	colorYellow  = lipgloss.Color("3")   // yellow
	colorCyan    = lipgloss.Color("14")  // cyan

	// Header bar
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorPrimary).
			Padding(0, 1)

	// Status bar
	statusStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Padding(0, 1)

	// Tab styles
	tabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorCyan).
			Padding(0, 2)
	tabInactive = lipgloss.NewStyle().
			Foreground(colorGray).
			Padding(0, 2)

	// Table
	tableHeader = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorDim)

	// State colors
	stateRunning  = lipgloss.NewStyle().Foreground(colorGreen)
	statePoweroff = lipgloss.NewStyle().Foreground(colorRed)
	stateSuspend  = lipgloss.NewStyle().Foreground(colorYellow)
	stateDefault  = lipgloss.NewStyle().Foreground(colorGray)

	// Filter input
	filterStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	// Title
	titleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// VM list cursor
	cursorStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(colorPrimary).
			Bold(true)
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
