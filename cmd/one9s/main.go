package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/scabello/one9s/internal/client"
	"github.com/scabello/one9s/internal/config"
	"github.com/scabello/one9s/pkg/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "one9s: config error: %v\n", err)
		os.Exit(1)
	}

	c := client.NewGOCA(cfg)

	p := tea.NewProgram(
		tui.NewRootModel(c),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "one9s: %v\n", err)
		os.Exit(1)
	}
}
