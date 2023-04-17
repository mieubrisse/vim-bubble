package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mieubrisse/vim-textarea-testing/vim"
)

type appModel struct {
	vim vim.Model
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
		model.vim.Resize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	model.vim, cmd = model.vim.Update(msg)
	return model, cmd
}

func (model appModel) View() string {
	return model.vim.View()
}
