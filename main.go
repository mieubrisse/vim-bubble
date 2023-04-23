package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mieubrisse/vim-bubble/vim"
	"os"
)

func main() {
	area := vim.New()
	area.Focus()
	area.SetValue("Four score, and seven years ago our founding fathers rocked out")

	model := appModel{
		vim: area,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
