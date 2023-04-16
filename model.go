package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type appModel struct {
	area VimTextAreaModel
}

func (model appModel) Init() tea.Cmd {
	return nil
}

func (model appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Allow quitting
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return model, tea.Quit
		}
	case tea.WindowSizeMsg:
		model.area.Resize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	model.area, cmd = model.area.Update(msg)
	return model, cmd
}

func (model appModel) View() string {
	return model.area.View()
}
