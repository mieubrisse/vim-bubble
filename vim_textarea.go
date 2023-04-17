package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type VimTextAreaMode string

const (
	NormalMode VimTextAreaMode = "NORMAL"
	InsertMode VimTextAreaMode = "INSERT"
)

const (
	shouldBindToLineWhenMovingRight = true

	normalModeColorHex = "#defa51"
	insertModeColorHex = "#61d4fa"

	maxNgraphPanelCharacters  = 5
	desiredNgraphPanelPadding = 1

	minModePanelCharacters = 1
	// TODO Make this dynamic by looking at the length of the mode strings!
	maxModePanelCharacters  = 6
	desiredModePanelPadding = 1
)

// Rename?
type VimTextAreaModel struct {
	mode VimTextAreaMode

	isFocused bool

	area Model

	// Buffer for storing N-graphs (e.g. digraphs, trigraphs, etc.)
	nGraphBuffer string

	// TODO something about the written vs unwritten buffer

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
			// TODO c^ digraph
			// TODO c$ digraph
			// TODO handle ngraphs + motion keys (right now they just clear)
			switch msg.String() {
			case "esc":
				model.nGraphBuffer = ""
			case "a":
				// This is a deviation from Vim, but I'm fine with it
				model.nGraphBuffer = ""
				model.area.CharacterRight(false)
				model.mode = InsertMode
			case "h":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.area.CharacterLeft(true)
			case "j":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				// We want line-binding because we're in normal mode, so we shouldn't have the cursor beyond the end of the line
				model.area.CursorDown(true)
			case "k":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				// We want line-binding because we're in normal mode, so we shouldn't have the cursor beyond the end of the line
				model.area.CursorUp(true)
			case "l":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.area.CharacterRight(shouldBindToLineWhenMovingRight)
			case "b":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.area.WordStartLeft()
			case "w":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.area.WordStartRight()
			case "e":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.area.WordEndRight()
			case "i":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.mode = InsertMode
				break
			case "^":
				switch model.nGraphBuffer {
				case "d":
					model.area.DeleteBeforeCursor()
				case "c":
					model.area.DeleteBeforeCursor()
					model.mode = InsertMode
				default:
					model.area.CursorStart()
				}
			case "$":
				switch model.nGraphBuffer {
				case "d":
					model.area.DeleteAfterCursor()
				case "c":
					model.area.DeleteAfterCursor()
					model.area.CharacterRight(false)
					model.mode = InsertMode
				default:
					model.area.CursorEnd(true)
				}
			case "D":
				model.area.DeleteAfterCursor()
			case "C":
				model.area.DeleteAfterCursor()
				model.area.CharacterRight(false)
				model.mode = InsertMode
			case "o":
				model.area.InsertLineBelow()
				model.area.CursorDown(true)
				model.mode = InsertMode
			case "O":
				model.area.InsertLineAbove()
				model.area.CursorUp(true)
				model.mode = InsertMode
			case "d":
				switch model.nGraphBuffer {
				case "":
					model.nGraphBuffer = msg.String()
				case "d":
					model.nGraphBuffer = ""
					model.area.DeleteLine()
				default:
					model.nGraphBuffer = ""
				}
			case "c":
				switch model.nGraphBuffer {
				case "":
					model.nGraphBuffer = msg.String()
				case "c":
					model.nGraphBuffer = ""
					model.area.ClearLine()
					model.mode = InsertMode
				default:
					model.nGraphBuffer = ""
				}
			case "0":
				if model.nGraphBuffer == "" {
					model.area.CursorStart()
				} else {
					model.nGraphBuffer += msg.String()
				}
			case "1":
			case "2":
			case "3":
			case "4":
			case "5":
			case "6":
			case "7":
			case "8":
			case "9":
				model.nGraphBuffer += msg.String()
			}
		}
	}
	return model, tea.Batch(resultCmds...)
}

func (model VimTextAreaModel) View() string {
	resultBuilder := strings.Builder{}

	resultBuilder.WriteString(model.area.View())
	resultBuilder.WriteString("\n")
	resultBuilder.WriteString(model.renderStatusBar())

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

func (model VimTextAreaModel) renderStatusBar() string {
	if !model.isFocused {
		return strings.Repeat(" ", model.width)
	}

	// First calculate the ngraph panel size, leaving room for at least one char of mode panel
	// This means the digraph panel will be the first to get space when the window expands, up to its limit
	ngraphPanelSize := clamp(model.width-minModePanelCharacters, 0, maxNgraphPanelCharacters+2*desiredNgraphPanelPadding)

	// Next calculate the mode panel size, up to its max
	// This means the mode panel will get extra space second
	modePanelSize := clamp(model.width-ngraphPanelSize, minModePanelCharacters, maxModePanelCharacters+2*desiredModePanelPadding)

	// Finally, pad any extra space
	numPads := max(0, model.width-modePanelSize-ngraphPanelSize)
	padStr := strings.Repeat(" ", numPads)

	// TODO get rid of magic consts
	var modePanelColorHex string
	switch model.mode {
	case InsertMode:
		modePanelColorHex = insertModeColorHex
	default:
		modePanelColorHex = normalModeColorHex
	}
	modePanelStr := coerceToWidth(string(model.mode), modePanelSize, true)
	modePanelStr = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color(modePanelColorHex)).
		Render(modePanelStr)

	// TODO get rid of magic consts
	ngraphPanelStr := coerceToWidth(model.nGraphBuffer, ngraphPanelSize, false)

	return modePanelStr + padStr + ngraphPanelStr
}

// Takes the given string, centers it, truncating as needed, and adds padds if the desired size is bigger than
// the string itself
// If shouldTruncateWithFirstChars is set, truncating of the string will use the first N characters; if not, the last N
func coerceToWidth(input string, totalLength int, shouldTruncateWithFirstChars bool) string {
	// For some reason MaxWidth isn't actually truncating the string when it gets small, so we have to do it
	var truncatedStr string
	if shouldTruncateWithFirstChars {
		truncatedStr = input[:min(len(input), totalLength)]
	} else {
		firstIdx := max(0, len(input)-totalLength)
		truncatedStr = input[firstIdx:]
	}
	return lipgloss.NewStyle().
		Width(totalLength).
		Align(lipgloss.Center).
		Render(truncatedStr)
}
