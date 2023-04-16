package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	area := NewVimTextArea()
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
