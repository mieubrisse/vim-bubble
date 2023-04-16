package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type VimTextAreaMode string

const (
	NormalMode VimTextAreaMode = "NORMAL"
	InsertMode VimTextAreaMode = "INSERT"
)

// Rename?
type VimTextAreaModel struct {
	mode VimTextAreaMode

	isFocused bool

	area Model

	// Buffer for storing N-graphs (e.g. digraphs, trigraphs, etc.)
	nGraphBuffer string

	width  int
	height int
}

// TODO rename
func NewVimTextArea() VimTextAreaModel {
	return VimTextAreaModel{
		mode:      NormalMode,
		isFocused: false,
		area:      New(),
	}
}

func (model VimTextAreaModel) Init() tea.Cmd {
	return nil
}

func (model VimTextAreaModel) Update(msg tea.Msg) (VimTextAreaModel, tea.Cmd) {
	// TODO handle focus/blur logic here?
	var resultCmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch model.mode {
		case InsertMode:
			if msg.String() == "esc" {
				model.mode = NormalMode
				break
			}

			var cmd tea.Cmd
			model.area, cmd = model.area.Update(msg)
			resultCmds = append(resultCmds, cmd)
		case NormalMode:
			switch msg.String() {
			case "h":
				model.area.CharacterLeft(true)
			case "j":
				model.area.CursorDown()
			case "k":
				model.area.CursorUp()
			case "l":
				model.area.CharacterRight()
			case "b":
				model.area.WordLeft()
			case "w":
				model.area.WordRight()
			case "i":
				model.mode = InsertMode
				break
			}
		}
	}
	return model, tea.Batch(resultCmds...)
}

func (model VimTextAreaModel) View() string {
	resultBuilder := strings.Builder{}

	resultBuilder.WriteString(model.area.View())
	resultBuilder.WriteString("\n")

	statusLine := ""
	if model.isFocused {
		statusLine = string(model.mode)
		// TODO add n-graphs to status line
	}
	resultBuilder.WriteString(statusLine)

	return resultBuilder.String()
}

func (model *VimTextAreaModel) Focus() {
	model.isFocused = true
	model.area.Focus()
}

func (model *VimTextAreaModel) Blur() {
	model.isFocused = false
	model.area.Blur()
}

func (model VimTextAreaModel) Focused() bool {
	return model.isFocused
}

func (model *VimTextAreaModel) Resize(width int, height int) {
	model.width = width
	model.height = height

	// Leave space for the status bar
	// TODO Use max function with 0
	model.area.SetWidth(width - 1)
	model.area.SetHeight(height - 1)
}
