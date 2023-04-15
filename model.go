package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

type appModel struct {
	textarea Model
}

func (model appModel) Init() tea.Cmd {
	return nil
}

func (model appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Allow quitting
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return model, tea.Quit
		// Unfocus the textarea when ESC is pressed in normal mode
		case "esc":
			if model.textarea.Focused() && model.textarea.mode == NormalMode {
				model.textarea.Blur()
				return model, nil
			}
		// Allow refocusing the textarea
		case "enter":
			if !model.textarea.Focused() {
				model.textarea.Focus()
				return model, nil
			}
		}
	}

	var cmd tea.Cmd
	model.textarea, cmd = model.textarea.Update(msg)
	return model, cmd
}

func (model appModel) View() string {
	return model.textarea.View() + fmt.Sprintf("\nfocused: %v", model.textarea.Focused())
}
