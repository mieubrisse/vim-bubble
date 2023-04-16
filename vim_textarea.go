package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type VimTextAreaMode string

const (
	NormalMode VimTextAreaMode = ""
	InsertMode VimTextAreaMode = "-- INSERT --"
)

const (
	shouldBindToLineWhenMovingRight = true
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
	area := New()
	// TODO remove this
	area.SetValue("four score and seven years ago our founding fathers did a really cool thing that's really long\n\nthis is a thing")
	area.CursorEnd(true) // This is a bit of a hack; SetValue should really do this right
	return VimTextAreaModel{
		mode:      NormalMode,
		isFocused: false,
		area:      area,
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
				model.area.CharacterLeft(true)
				// TODO if the cursor is off the end of the line, move it back
				break
			}

			var cmd tea.Cmd
			model.area, cmd = model.area.Update(msg)
			resultCmds = append(resultCmds, cmd)
		case NormalMode:
			// TODO d^ digraph
			// TODO d$ digraph
			// TODO dd digraph
			// TODO c^ digraph
			// TODO c$ digraph
			switch msg.String() {
			case "a":
				model.area.CharacterRight(false)
				model.mode = InsertMode
			case "h":
				model.area.CharacterLeft(true)
			case "j":
				// We want line-binding because we're in normal mode, so we shouldn't have the cursor beyond the end of the line
				model.area.CursorDown(true)
			case "k":
				// We want line-binding because we're in normal mode, so we shouldn't have the cursor beyond the end of the line
				model.area.CursorUp(true)
			case "l":
				model.area.CharacterRight(shouldBindToLineWhenMovingRight)
			case "b":
				model.area.WordStartLeft()
			case "w":
				model.area.WordStartRight()
			case "e":
				model.area.WordEndRight()
			case "i":
				model.mode = InsertMode
				break
			case "^":
				model.area.CursorStart()
			case "$":
				model.area.CursorEnd(true)
			case "D":
				model.area.DeleteAfterCursor()
			case "C":
				model.area.DeleteAfterCursor()
				model.area.CharacterRight(false)
				model.mode = InsertMode
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
