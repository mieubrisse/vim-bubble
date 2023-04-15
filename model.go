package main

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type appModel struct {
	text string

	cursorColumn int
	cursorRow    int
}

func (model appModel) Init() tea.Cmd {
	textarea.Model{hh}
	return nil
}

func (model appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return model, tea.Quit
		}
		/*
			case tea.WindowSizeMsg:
				model = model.Resize(msg.Width, msg.Height)
		*/
	}
	return model, nil
}

func (model appModel) View() string {
	return model.text
}
