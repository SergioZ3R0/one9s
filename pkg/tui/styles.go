package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	colorPrimary = lipgloss.Color("39")  // blue
	colorWhite   = lipgloss.Color("15")  // bright white
	colorGray    = lipgloss.Color("240") // dim gray
	colorGreen   = lipgloss.Color("10")  // green
	colorRed     = lipgloss.Color("9")   // red
	colorYellow  = lipgloss.Color("3")   // yellow
	colorCyan    = lipgloss.Color("14")  // cyan
	colorPurple  = lipgloss.Color("57")  // purple (tab active bg)
	colorDim     = lipgloss.Color("240") // frame gray

	// Header - minimal line
	headerStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Status bar
	statusStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	// Tabs - flat, no solid block
	tabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorPurple).
			Padding(0, 2)
	tabInactive = lipgloss.NewStyle().
			Foreground(colorGray).
			Padding(0, 2)

	// Table header
	tableHeader = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	// State colors
	stateRunning  = lipgloss.NewStyle().Foreground(colorGreen)
	statePoweroff = lipgloss.NewStyle().Foreground(colorRed)
	stateSuspend  = lipgloss.NewStyle().Foreground(colorYellow)
	stateDefault  = lipgloss.NewStyle().Foreground(colorGray)

	// Filter
	filterStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	// Panel with rounded border
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

	// Separator line
	separatorStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// App name in header
	appNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("87")).
			Bold(true)

	// Connection status in header
	connectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("114"))

	// Outer frame
	frameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Padding(0, 1)

	// Cursor - full row purple background
	cursorStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(colorPurple)

	// Title
	titleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Connection status
	connectedStyle = lipgloss.NewStyle().
			Foreground(colorGreen)
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

func hostStateStyle(state string) lipgloss.Style {
	switch state {
	case "MONITORED", "MONITORING":
		return lipgloss.NewStyle().Foreground(colorGreen)
	case "ERROR":
		return lipgloss.NewStyle().Foreground(colorRed)
	case "DISABLED", "OFFLINE":
		return lipgloss.NewStyle().Foreground(colorYellow)
	default:
		return lipgloss.NewStyle().Foreground(colorWhite)
	}
}
