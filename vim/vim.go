package vim

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mieubrisse/vim-bubble/textarea"
	"strings"
)

type Mode string

const (
	NormalMode  Mode = "NORMAL"
	InsertMode  Mode = "INSERT"
	CommandMode Mode = "COMMAND"
)

const (
	shouldBindToLineWhenMovingRight = true

	normalModeColorHex  = "#defa51"
	insertModeColorHex  = "#61d4fa"
	commandModeColorHex = "#61d43f"

	maxNgraphPanelCharacters  = 5
	desiredNgraphPanelPadding = 1

	minModePanelCharacters = 1
	// TODO Make this dynamic by looking at the length of the mode strings!
	maxModePanelCharacters  = 6
	desiredModePanelPadding = 1

	numHistoryStepsToKeep = 20
)

type Model struct {
	mode Mode

	isFocused bool

	area textarea.Model

	// Buffer for storing N-graphs (e.g. digraphs, trigraphs, etc.)
	// TODO is this actually called an ngraph?
	nGraphBuffer string

	// TODO something about the written vs unwritten buffer

	// The undo history, which gets an entry every time we leave insert mode
	// New history entries are added to the back of this list
	undoHistory []string

	// The pointer within the history of what the text buffer is currently displaying (needed for redoing)
	historyPointer int

	// Corresponds to Vim's " register, which contains both yanked and deleted text
	commaRegister string

	width  int
	height int
}

func New() Model {
	area := textarea.New()
	area.SetValue("")
	area.Prompt = ""
	return Model{
		mode:        NormalMode,
		isFocused:   false,
		area:        area,
		undoHistory: []string{""},
	}
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var resultCmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch model.mode {
		case InsertMode:
			if msg.String() == "esc" {
				model.mode = NormalMode
				model.area.CharacterLeft(true)

				model.saveHistory()

				break
			}

			var cmd tea.Cmd
			model.area, cmd = model.area.Update(msg)
			resultCmds = append(resultCmds, cmd)
		case NormalMode:
			// TODO clean this whole thing up to make the processing of motion commands way better!

			// TODO handle ngraphs + motion keys (right now they just clear)
			switch msg.String() {
			case "esc":
				model.nGraphBuffer = ""
			case "a":
				// This is a deviation from Vim, but I'm fine with it
				model.nGraphBuffer = ""
				model.area.CharacterRight(false)
				model.mode = InsertMode
			case "A":
				model.area.CursorEnd(false)
				model.mode = InsertMode
			case "i":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.mode = InsertMode
				break
			case "I":
				model.area.CursorStart()
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
			case "b", "B":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.area.WordStartLeft()
			case "w", "W":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.area.WordStartRight()
			case "e", "E":
				// TODO handle repeats
				switch model.nGraphBuffer {
				case "":
					model.area.WordEndRight()
				case "g":
					model.area.WordEndLeft()
				}
				// I thiiink this is right??
				model.nGraphBuffer = ""
			case "^":
				switch model.nGraphBuffer {
				case "d":
					model.area.DeleteBeforeCursor()
					model.nGraphBuffer = ""
					// TODO extract this into something better!
					model.saveHistory()
				case "c":
					model.area.DeleteBeforeCursor()
					model.mode = InsertMode
					model.nGraphBuffer = ""
					// TODO extract this into something better!
					model.saveHistory()
				default:
					model.area.CursorStart()
				}
			case "$":
				switch model.nGraphBuffer {
				case "d":
					model.area.DeleteAfterCursor()
					model.nGraphBuffer = ""
					// TODO extract this into something better!
					model.saveHistory()
				case "c":
					model.area.DeleteAfterCursor()
					model.area.CharacterRight(false)
					model.mode = InsertMode
					model.nGraphBuffer = ""
				default:
					model.area.CursorEnd(true)
				}
			case "g":
				switch model.nGraphBuffer {
				case "":
					model.nGraphBuffer = msg.String()
				case "g":
					model.area.SetRow(0)
					model.nGraphBuffer = ""
				default:
					// TODO is this right?
					model.nGraphBuffer = ""
				}
			case "G":
				model.area.RowEnd()
			case "D":
				model.area.DeleteAfterCursor()
				// TODO extract this into something better!
				model.saveHistory()
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
					// TODO extract this into something better!
					model.saveHistory()
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
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				model.nGraphBuffer += msg.String()
			case "u":
				newHistoryPointer := max(0, model.historyPointer-1)
				if newHistoryPointer != model.historyPointer {
					model.area.SetValue(model.undoHistory[newHistoryPointer])
					model.historyPointer = newHistoryPointer
				}
			case "ctrl+r":
				newHistoryPointer := min(len(model.undoHistory)-1, model.historyPointer+1)
				if newHistoryPointer != model.historyPointer {
					model.area.SetValue(model.undoHistory[newHistoryPointer])
					model.historyPointer = newHistoryPointer
				}
				// TODO keep the cursor position when reinserting text
			case "x":
				deletedRunes := model.area.DeleteOnCursor()
				model.commaRegister = string(deletedRunes)
				// TODO extract this into something better!
				model.saveHistory()
			case "p":
				// TODO make this a better thing in the textarea class - it's a kind of hacky way to implement the "paste AFTER cursor location" logic of Vim
				model.area.CharacterRight(false)
				model.area.InsertString(model.commaRegister)
				model.area.CharacterLeft(false)
			}
			// TODO 't', 'f', ';', and ','
		}
	}
	return model, tea.Batch(resultCmds...)
}

func (model Model) View() string {
	resultBuilder := strings.Builder{}

	resultBuilder.WriteString(model.area.View())
	resultBuilder.WriteString("\n")
	resultBuilder.WriteString(model.renderStatusBar())

	return resultBuilder.String()
}

func (model *Model) Focus() {
	model.isFocused = true
	model.area.Focus()
}

func (model *Model) Blur() {
	model.isFocused = false
	model.area.Blur()
}

func (model Model) Focused() bool {
	return model.isFocused
}

func (model *Model) Resize(width int, height int) {
	model.width = width
	model.height = height

	// Leave space for the status bar
	// TODO Use max function with 0
	model.area.SetWidth(width - 1)
	model.area.SetHeight(height - 1)
}

func (model *Model) SetValue(str string) {
	model.area.SetValue(str)
}

func (model *Model) GetValue() string {
	return model.area.Value()
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func (model *Model) saveHistory() {
	// TODO don't save a history step if nothing new was written
	if model.area.Value() != model.undoHistory[len(model.undoHistory)-1] {
		// If the user has rewound, then we discard things they've rewound past
		preservedHistory := model.undoHistory[:model.historyPointer+1]
		newHistory := append(
			preservedHistory,
			model.area.Value(),
		)

		// Now discard down to the appropriate number of history steps
		subsliceStartIdx := max(0, len(newHistory)-numHistoryStepsToKeep)
		model.undoHistory = newHistory[subsliceStartIdx:]

		// We reset the steps-rewound because we've now thrown away the steps the user rewound past
		model.historyPointer = len(model.undoHistory) - 1
	}
}

func (model Model) renderStatusBar() string {
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
	case CommandMode:
		modePanelColorHex = commandModeColorHex
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

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
