package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/scabello/one9s/internal/client"
)

func progressBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	empty := width - filled
	if filled < 0 {
		filled = 0
	}
	if empty < 0 {
		empty = 0
	}

	var bar string
	if pct >= 80 {
		bar = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(strings.Repeat("█", filled))
	} else if pct >= 50 {
		bar = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(strings.Repeat("█", filled))
	} else {
		bar = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(strings.Repeat("█", filled))
	}
	bar += lipgloss.NewStyle().Foreground(colorGray).Render(strings.Repeat("░", empty))
	bar += fmt.Sprintf(" %3.0f%%", pct)
	return bar
}

func hostDetailView(h client.HostDetail) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" Host Details"))
	b.WriteString("\n\n")

	sections := []struct {
		title  string
		fields [][]string
	}{
		{
			"General",
			[][]string{
				{"ID", fmt.Sprintf("%d", h.ID)},
				{"Name", h.Name},
				{"State", h.State},
				{"IM_MAD", h.IMMAD},
				{"VM_MAD", h.VMMAD},
			},
		},
		{
			"Cluster",
			[][]string{
				{"Cluster ID", fmt.Sprintf("%d", h.ClusterID)},
				{"Cluster Name", h.ClusterName},
			},
		},
	}

	for _, section := range sections {
		hasContent := false
		for _, f := range section.fields {
			if f[1] != "" && f[1] != "0" && f[1] != "-" {
				hasContent = true
				break
			}
		}
		if !hasContent {
			continue
		}

		b.WriteString(filterStyle.Render(section.title))
		b.WriteString("\n")
		for _, f := range section.fields {
			if f[1] == "" || f[1] == "0" || f[1] == "-" {
				continue
			}
			b.WriteString(fmt.Sprintf("  %s %s\n",
				lipgloss.NewStyle().Foreground(colorGray).Width(16).Render(f[0]),
				f[1],
			))
		}
		b.WriteString("\n")
	}

	// Resources with progress bars
	b.WriteString(filterStyle.Render("Resources"))
	b.WriteString("\n")

	cpuPct := 0.0
	if h.TotalCPU > 0 {
		cpuPct = float64(h.ShareCPU) / float64(h.TotalCPU) * 100
	}
	memPct := 0.0
	if h.TotalMem > 0 {
		memPct = float64(h.ShareMem) / float64(h.TotalMem) * 100
	}

	barW := 20
	b.WriteString(fmt.Sprintf("  %s %s\n",
		lipgloss.NewStyle().Foreground(colorGray).Width(16).Render("CPU"),
		progressBar(cpuPct, barW),
	))
	b.WriteString(fmt.Sprintf("  %s %s\n",
		lipgloss.NewStyle().Foreground(colorGray).Width(16).Render("Memory"),
		progressBar(memPct, barW),
	))
	b.WriteString(fmt.Sprintf("  %s %d\n",
		lipgloss.NewStyle().Foreground(colorGray).Width(16).Render("Running VMs"),
		h.RunningVMs,
	))

	b.WriteString("\n")
	b.WriteString(statusStyle.Render(" press esc to go back"))

	return b.String()
}
