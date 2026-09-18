package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const aboutASCII = `██████╗ ███╗   ██╗███████╗ █████╗ ███████╗
██╔═══██╗████╗  ██║██╔════╝██╔══██╗██╔════╝
██║   ██║██╔██╗ ██║█████╗  ╚██████║███████╗
██║   ██║██║╚██╗██║██╔══╝   ╚═══██║╚════██║
╚██████╔╝██║ ╚████║███████╗ █████╔╝███████║
╚═════╝ ╚═╝  ╚═══╝╚══════╝ ╚════╝ ╚══════╝`

func aboutView() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(aboutASCII))
	b.WriteString("\n\n")

	info := []struct{ k, v string }{
		{"Version", "0.1.2"},
		{"License", "Apache 2.0"},
		{"Author", "SergioZ3R0"},
		{"Repository", "github.com/SergioZ3R0/one9s"},
		{"Website", "one9s.scszero.com"},
		{"Docs", "one9s.scszero.com/docs.html"},
	}

	for _, i := range info {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			lipgloss.NewStyle().Foreground(colorGray).Width(14).Render(i.k),
			i.v,
		))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render(
		"  TUI for OpenNebula cluster management",
	))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render(
		"  Built with Bubble Tea + GOCA",
	))

	return b.String()
}
