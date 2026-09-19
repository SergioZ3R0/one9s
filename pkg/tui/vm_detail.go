package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/scabello/one9s/internal/client"
)

func vmDetailView(vm client.VMDetail) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" VM Details"))
	b.WriteString("\n\n")

	sections := []struct {
		title  string
		fields [][]string
	}{
		{
			"General",
			[][]string{
				{"ID", fmt.Sprintf("%d", vm.ID)},
				{"Name", vm.Name},
				{"State", vm.State},
				{"User", vm.User},
				{"Group", vm.Group},
				{"Deploy ID", vm.DeployID},
			},
		},
		{
			"Capacity",
			[][]string{
				{"CPU", vm.CPU},
				{"Memory", vm.Memory},
				{"VCPU", vm.VCPU},
			},
		},
		{
			"Network",
			[][]string{
				{"IP", vm.IP},
				{"MAC", vm.MAC},
				{"Network", vm.Network},
				{"Bridge", vm.Bridge},
			},
		},
		{
			"Host",
			[][]string{
				{"Host", vm.Host},
				{"Cluster", vm.Cluster},
			},
		},
		{
			"Timing",
			[][]string{
				{"Start Time", vm.StartTime},
				{"End Time", vm.EndTime},
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

	b.WriteString(statusStyle.Render(" press esc to go back"))

	return b.String()
}
