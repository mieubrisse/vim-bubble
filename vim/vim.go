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
