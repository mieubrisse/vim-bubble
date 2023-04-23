package vim

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mieubrisse/vim-bubble/textarea"
	"strings"
)

type Mode string

const (
	NormalMode Mode = "NORMAL"
	InsertMode Mode = "INSERT"
)

const (
	shouldBindToLineWhenMovingRight = true

	maxNgraphPanelCharacters  = 5
	desiredNgraphPanelPadding = 1

	minModePlacardCharacters = 1
	// TODO Make this dynamic by looking at the length of the mode strings!
	maxModePlacardCharacters  = 6
	desiredModePlacardPadding = 1

	numHistoryStepsToKeep = 20
)

var defaultNormalModePlacardStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#defa51")).
	Foreground(lipgloss.Color("#000000"))

var defaultInsertModePlacardStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#61d4fa")).
	Foreground(lipgloss.Color("#000000"))

var unknownModePlacardStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#FFFFFF")).
	Foreground(lipgloss.Color("#000000"))

type Model struct {
	NormalModePlacardStyle lipgloss.Style

	InsertModePlacardStyle lipgloss.Style

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
		NormalModePlacardStyle: defaultNormalModePlacardStyle,
		InsertModePlacardStyle: defaultInsertModePlacardStyle,
		mode:                   NormalMode,
		isFocused:              false,
		area:                   area,
		nGraphBuffer:           "",
		undoHistory:            []string{""},
		historyPointer:         0,
		commaRegister:          "",
		width:                  0,
		height:                 0,
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
				model.area.MoveCursorLeftOneRune()

				model.CheckpointHistory()

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
				model.area.MoveCursorRightOneRune(false)
				model.mode = InsertMode
			case "A":
				model.area.MoveCursorToLineEnd(false)
				model.mode = InsertMode
			case "i":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.mode = InsertMode
				break
			case "I":
				model.area.MoveCursorToLineStart()
				model.mode = InsertMode
			case "h":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.area.MoveCursorLeftOneRune()
			case "j":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				// We want line-binding because we're in normal mode, so we shouldn't have the cursor beyond the end of the line
				model.area.MoveCursorDown(true)
			case "k":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				// We want line-binding because we're in normal mode, so we shouldn't have the cursor beyond the end of the line
				model.area.MoveCursorUp(true)
			case "l":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.area.MoveCursorRightOneRune(shouldBindToLineWhenMovingRight)
			case "b", "B":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.area.MoveCursorByWord(textarea.CursorMovementDirection_Left, textarea.WordwiseMovementStopPosition_Terminus)
			case "w", "W":
				// TODO handle movement commands with numbers
				model.nGraphBuffer = ""
				model.area.MoveCursorByWord(textarea.CursorMovementDirection_Right, textarea.WordwiseMovementStopPosition_Incidence)
			case "e", "E":
				// TODO handle repeats
				switch model.nGraphBuffer {
				case "":
					model.area.MoveCursorByWord(textarea.CursorMovementDirection_Right, textarea.WordwiseMovementStopPosition_Terminus)
				case "g":
					model.area.MoveCursorByWord(textarea.CursorMovementDirection_Left, textarea.WordwiseMovementStopPosition_Incidence)
				}
				// I thiiink this is right??
				model.nGraphBuffer = ""
			case "^":
				switch model.nGraphBuffer {
				case "d":
					model.area.DeleteBeforeCursor()
					model.nGraphBuffer = ""
					// TODO extract this into something better!
					model.CheckpointHistory()
				case "c":
					model.area.DeleteBeforeCursor()
					model.mode = InsertMode
					model.nGraphBuffer = ""
					// TODO extract this into something better!
					model.CheckpointHistory()
				default:
					model.area.MoveCursorToLineStart()
				}
			case "$":
				switch model.nGraphBuffer {
				case "d":
					model.area.DeleteAfterCursor()
					model.nGraphBuffer = ""
					// TODO extract this into something better!
					model.CheckpointHistory()
				case "c":
					model.area.DeleteAfterCursor()
					model.area.MoveCursorRightOneRune(false)
					model.mode = InsertMode
					model.nGraphBuffer = ""
				default:
					model.area.MoveCursorToLineEnd(true)
				}
			case "g":
				switch model.nGraphBuffer {
				case "":
					model.nGraphBuffer = msg.String()
				case "g":
					model.area.SetCursorRow(0)
					model.nGraphBuffer = ""
				default:
					// TODO is this right?
					model.nGraphBuffer = ""
				}
			case "f":
				switch model.nGraphBuffer {
				case "":
					model.nGraphBuffer = msg.String()
				case "f":
				}
			case "G":
				model.area.MoveCursorToLastRow()
			case "D":
				model.area.DeleteAfterCursor()
				// TODO extract this into something better!
				model.CheckpointHistory()
			case "C":
				model.area.DeleteAfterCursor()
				model.area.MoveCursorRightOneRune(false)
				model.mode = InsertMode
			case "o":
				model.area.InsertLineBelow()
				model.area.MoveCursorDown(true)
				model.mode = InsertMode
			case "O":
				model.area.InsertLineAbove()
				model.area.MoveCursorUp(true)
				model.mode = InsertMode
			case "d":
				switch model.nGraphBuffer {
				case "":
					model.nGraphBuffer = msg.String()
				case "d":
					model.nGraphBuffer = ""
					model.area.DeleteLine()
					// TODO extract this into something better!
					model.CheckpointHistory()
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
					model.area.MoveCursorToLineStart()
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
				model.CheckpointHistory()
			case "p":
				// TODO make this a better thing in the textarea class - it's a kind of hacky way to implement the "paste AFTER cursor location" logic of Vim
				model.area.MoveCursorRightOneRune(false)
				model.area.InsertString(model.commaRegister)
				model.area.MoveCursorLeftOneRune()
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

func (model Model) GetWidth() int {
	return model.width
}

func (model Model) GetHeight() int {
	return model.height
}

func (model *Model) SetValue(str string) {
	model.area.SetValue(str)
}

func (model *Model) GetValue() string {
	return model.area.GetValue()
}

func (model Model) SetMode(mode Mode) Model {
	model.mode = mode
	return model
}

func (model Model) GetMode() Mode {
	return model.mode
}

func (model Model) GetCursorRow() int {
	return model.area.GetRow()
}

// TODO this is a nasty hack, that I'm exposing purely to allow for tab completion in the app that needs this
// TODO the ideal would be some standard way to programmatically manipulate the Vim buffer
func (model Model) ReplaceLine(newContents string) Model {
	model.area.ClearLine()
	model.area.InsertString(newContents)
	return model
}

// Forces a checkpoint in the Vim buffer's history, for undo
func (model *Model) CheckpointHistory() {
	if model.area.GetValue() == model.undoHistory[len(model.undoHistory)-1] {
		return
	}

	// If the user has rewound, then we discard things they've rewound past
	preservedHistory := model.undoHistory[:model.historyPointer+1]
	newHistory := append(
		preservedHistory,
		model.area.GetValue(),
	)

	// Now discard down to the appropriate number of history steps
	subsliceStartIdx := max(0, len(newHistory)-numHistoryStepsToKeep)
	model.undoHistory = newHistory[subsliceStartIdx:]

	// We reset the steps-rewound because we've now thrown away the steps the user rewound past
	model.historyPointer = len(model.undoHistory) - 1
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func (model Model) renderStatusBar() string {
	if !model.isFocused {
		return strings.Repeat(" ", model.width)
	}

	// First calculate the ngraph panel size, leaving room for at least one char of mode panel
	// This means the digraph panel will be the first to get space when the window expands, up to its limit
	ngraphPanelSize := clamp(model.width-minModePlacardCharacters, 0, maxNgraphPanelCharacters+2*desiredNgraphPanelPadding)

	// Next calculate the mode placard size, up to its max
	// This means the mode placard will get extra space second
	modePlacardSize := clamp(model.width-ngraphPanelSize, minModePlacardCharacters, maxModePlacardCharacters+2*desiredModePlacardPadding)

	// Finally, pad any extra space
	numPads := max(0, model.width-modePlacardSize-ngraphPanelSize)
	padStr := strings.Repeat(" ", numPads)

	var modePlacardStyle lipgloss.Style
	switch model.mode {
	case InsertMode:
		modePlacardStyle = model.InsertModePlacardStyle
	case NormalMode:
		modePlacardStyle = model.NormalModePlacardStyle
	default:
		modePlacardStyle = unknownModePlacardStyle
	}
	modePlacardStr := coerceToWidth(string(model.mode), modePlacardSize, true)
	modePlacardStr = modePlacardStyle.Render(modePlacardStr)

	// TODO get rid of magic consts
	ngraphPanelStr := coerceToWidth(model.nGraphBuffer, ngraphPanelSize, false)

	return modePlacardStr + padStr + ngraphPanelStr
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
