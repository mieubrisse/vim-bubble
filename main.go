package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mieubrisse/vim-textarea-testing/vim_textarea"
	"os"
)

func main() {
	area := vim_textarea.NewVimTextArea()
	area.Focus()

	model := appModel{
		area: area,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
