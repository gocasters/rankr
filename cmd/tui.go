package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gocasters/rankr/pkg/ui"
	"os"
)

func main() {
	p := tea.NewProgram(ui.NewMainMenuModel())

	if err := p.Start(); err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
}
