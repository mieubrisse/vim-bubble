package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mieubrisse/vim-textarea-testing/vim"
	"os"
)

func main() {
	area := vim.New()
	area.Focus()

	model := appModel{
		vim: area,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
