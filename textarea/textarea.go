package textarea

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/runeutil"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	rw "github.com/mattn/go-runewidth"
)

type CursorMovementDirection int

const (
	CursorMovementDirection_Left  CursorMovementDirection = -1
	CursorMovementDirection_Right CursorMovementDirection = 1
)

// The position in the word that the cursor will stop at *IN THE DIRECTION OF CURSOR TRAVEL*
// E.g.
type WordwiseMovementStopPosition int

const (
	// These integers actually represent the index offset (relative to the direction of cursor travel)
	// where we'll look for a word boundary
	// If we're going right, "incident" = beginning of word and "terminus" = end of word
	// If we're going left, "incident" = *end* of word, and "terminus" = beginning of word
	// We used "incident" and "terminus" to avoid confusion with "start" and "end" when going left

	// WordwiseMovementStopPosition_Incidence tells the cursor to stop as soon as hit a word (in direction of cursor travel)
	WordwiseMovementStopPosition_Incidence WordwiseMovementStopPosition = -1

	// WordwiseMovementStopPosition_Terminus tells the cursor to stop as soon as it would leave a word (in direction of cursor travel)
	WordwiseMovementStopPosition_Terminus WordwiseMovementStopPosition = 1
)

// When moving the cursor by a given character, the position where the cursor will stop relative to the character
type CharacterwiseMovementStopPosition int

const (
	// Stop on the character (corresponds to 'f' in Vim)
	CharacterwiseMovementStopPosition_On = 0

	// Stop just before the character (corresponds to 't' in Vim)
	CharacterwiseMovementStopPosition_Before = 1
)

const (
	minHeight        = 1
	minWidth         = 2
	defaultHeight    = 6
	defaultWidth     = 40
	defaultCharLimit = 400
	maxHeight        = 99
	maxWidth         = 500

	lineNumberColorHex = "#5d5d5d"
)

// Paste is a tea.Cmd for pasting from the clipboard into the text input.
func Paste() tea.Msg {
	str, err := clipboard.ReadAll()
	if err != nil {
		return pasteErrMsg{err}
	}
	return pasteMsg(str)
}

// Internal messages for clipboard operations.
type pasteMsg string
type pasteErrMsg struct{ error }

// Model is the Bubble Tea model for this text area element.
type Model struct {
	Err error

	// General settings.

	// Prompt is printed at the beginning of each line.
	//
	// When changing the value of Prompt after the model has been
	// initialized, ensure that SetWidth() gets called afterwards.
	//
	// See also SetPromptFunc().
	Prompt string

	// Placeholder is the text displayed when the user
	// hasn't entered anything yet.
	Placeholder string

	// ShowLineNumbers, if enabled, causes line numbers to be printed
	// after the prompt.
	ShowLineNumbers bool

	// EndOfBufferCharacter is displayed at the end of the input.
	EndOfBufferCharacter rune

	// KeyMap encodes the keybindings recognized by the widget.
	KeyMap KeyMap

	// Styling. FocusedStyle and BlurredStyle are used to style the textarea in
	// focused and blurred states.
	FocusedStyle Style
	BlurredStyle Style
	// style is the current styling to use.
	// It is used to abstract the differences in focus state when styling the
	// model, since we can simply assign the set of styles to this variable
	// when switching focus states.
	style *Style

	// Cursor is the text area cursor.
	Cursor cursor.Model

	// CharLimit is the maximum number of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	// If promptFunc is set, it replaces Prompt as a generator for
	// prompt strings at the beginning of each line.
	promptFunc func(line int) string

	// promptWidth is the width of the prompt.
	promptWidth int

	// width is the maximum number of characters that can be displayed at once.
	// If 0 or less this setting is ignored.
	width int

	// height is the maximum number of lines that can be displayed at once. It
	// essentially treats the text field like a vertically scrolling viewport
	// if there are more lines than the permitted height.
	height int

	// Underlying text value.
	value [][]rune

	// focus indicates whether user input focus should be on this input
	// component. When false, ignore keyboard input and hide the cursor.
	focus bool

	// Cursor column in the 'value' rune grid
	col int

	// Cursor row in the 'value' rune grid
	row int

	// Last character offset, used to maintain state when the cursor is moved
	// vertically such that we can maintain the same navigating position.
	lastCharOffset int

	// lineNumberFormat is the format string used to display line numbers.
	lineNumberFormat string

	// viewport is the vertically-scrollable viewport of the multi-line text
	// input.
	viewport *viewport.Model

	// rune sanitizer for input.
	rsan runeutil.Sanitizer
}

// New creates a new model with default settings.
func New() Model {
	vp := viewport.New(0, 0)
	vp.KeyMap = viewport.KeyMap{}
	cur := cursor.New()

	focusedStyle, blurredStyle := DefaultStyles()

	m := Model{
		CharLimit:            defaultCharLimit,
		Prompt:               lipgloss.ThickBorder().Left + " ",
		style:                &blurredStyle,
		FocusedStyle:         focusedStyle,
		BlurredStyle:         blurredStyle,
		EndOfBufferCharacter: '~',
		ShowLineNumbers:      true,
		Cursor:               cur,
		KeyMap:               DefaultKeyMap,

		value:            make([][]rune, minHeight, maxHeight),
		focus:            false,
		col:              0,
		row:              0,
		lineNumberFormat: "%2v ",

		viewport: &vp,
	}

	m.SetHeight(defaultHeight)
	m.SetWidth(defaultWidth)

	return m
}

// GetValue returns the value of the text input.
func (m Model) GetValue() string {
	if m.value == nil {
		return ""
	}

	var v strings.Builder
	for _, l := range m.value {
		v.WriteString(string(l))
		v.WriteByte('\n')
	}

	return strings.TrimSuffix(v.String(), "\n")
}

// SetValue sets the value of the text input.
func (m *Model) SetValue(s string) {
	m.Reset()
	m.InsertString(s)
	m.MoveCursorLeftOneRune()
}

// InsertString inserts a string at the cursor position.
func (m *Model) InsertString(s string) {
	m.insertRunesFromUserInput([]rune(s))
}

// InsertRune inserts a rune at the cursor position.
func (m *Model) InsertRune(r rune) {
	m.insertRunesFromUserInput([]rune{r})
}

// GetLength returns the number of characters currently in the text input.
func (m *Model) GetLength() int {
	var l int
	for _, row := range m.value {
		l += rw.StringWidth(string(row))
	}
	// We add len(m.value) to include the newline characters.
	return l + len(m.value) - 1
}

// GetNumRows returns the number of lines that are currently in the text input.
func (m *Model) GetNumRows() int {
	return len(m.value)
}

// MoveCursorDown moves the cursor down by one line.
// Returns whether or not the cursor blink should be reset.
func (m *Model) MoveCursorDown(bindToLine bool) {
	li := m.GetLineInfo()
	charOffset := max(m.lastCharOffset, li.CharOffset)
	m.lastCharOffset = charOffset

	if li.RowOffset+1 >= li.Height && m.row < len(m.value)-1 {
		m.row++
		m.col = 0
	} else {
		// Move the cursor to the start of the next line. So that we can get
		// the line information. We need to add 2 columns to account for the
		// trailing space wrapping.
		m.col = min(li.StartColumn+li.Width+2, len(m.value[m.row])-1)
	}

	nli := m.GetLineInfo()
	m.col = nli.StartColumn

	if nli.Width <= 0 {
		return
	}

	stopThreshold := len(m.value[m.row])
	if bindToLine {
		stopThreshold--
	}

	offset := 0
	for offset < charOffset {
		// TODO TESTING
		if m.col >= stopThreshold || offset >= nli.CharWidth-1 {
			break
		}
		offset += rw.RuneWidth(m.value[m.row][m.col])
		m.col++
	}

	m.repositionView()
}

// MoveCursorUp moves the cursor up by one line.
// If bindToLine is set, the cursor will not move past the last character of the line
func (m *Model) MoveCursorUp(bindToLine bool) {
	li := m.GetLineInfo()
	charOffset := max(m.lastCharOffset, li.CharOffset)
	m.lastCharOffset = charOffset

	if li.RowOffset <= 0 && m.row > 0 {
		m.row--
		m.col = len(m.value[m.row])
	} else {
		// Move the cursor to the end of the previous line.
		// This can be done by moving the cursor to the start of the line and
		// then subtracting 2 to account for the trailing space we keep on
		// soft-wrapped lines.
		m.col = li.StartColumn - 2
	}

	nli := m.GetLineInfo()
	m.col = nli.StartColumn

	if nli.Width <= 0 {
		return
	}

	stopThreshold := len(m.value[m.row])
	if bindToLine {
		stopThreshold--
	}

	offset := 0
	for offset < charOffset {
		if m.col >= stopThreshold || offset >= nli.CharWidth-1 {
			break
		}
		offset += rw.RuneWidth(m.value[m.row][m.col])
		m.col++
	}

	m.repositionView()
}

// GetCursorColumn gets the column within the rune grid where the cursor is currently at
// Note that the cursor can be beyond the right-hand end of the rune grid!
func (m Model) GetCursorColumn() int {
	return m.col
}

// SetCursorColumn moves the cursor to the given position. If the position is
// out of bounds the cursor will be moved to the start or end accordingly.
func (m *Model) SetCursorColumn(col int) {
	m.col = clamp(col, 0, len(m.value[m.row]))
	// Any time that we move the cursor horizontally we need to reset the last
	// offset so that the horizontal position when navigating is adjusted.
	m.lastCharOffset = 0
}

// MoveCursorToLineStart moves the cursor to the start of the input field.
func (m *Model) MoveCursorToLineStart() {
	m.SetCursorColumn(0)
}

// MoveCursorToLineEnd moves the cursor to the end of the input field.
// If bindToLine is set, only allow going to the last char of the line
func (m *Model) MoveCursorToLineEnd(bindToLine bool) {
	newPosition := len(m.value[m.row])
	if bindToLine {
		newPosition--
	}
	m.SetCursorColumn(newPosition)
}

func (m Model) GetRow() int {
	return m.row
}

func (m *Model) SetCursorRow(targetRow int) {
	targetRow = clamp(targetRow, 0, len(m.value)-1)

	adjustmentNeeded := targetRow - m.row
	if adjustmentNeeded > 0 {
		for i := 0; i < adjustmentNeeded; i++ {
			m.MoveCursorDown(true)
		}
	} else {
		for i := 0; i > adjustmentNeeded; i-- {
			m.MoveCursorUp(true)
		}
	}

	m.repositionView()
}

func (m *Model) MoveCursorToFirstRow() {
	m.SetCursorRow(0)
}

func (m *Model) MoveCursorToLastRow() {
	m.SetCursorRow(len(m.value) - 1)
}

// IsFocused returns the focus state on the model.
func (m Model) IsFocused() bool {
	return m.focus
}

// Focus sets the focus state on the model. When the model is in focus it can
// receive keyboard input and the cursor will be hidden.
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	m.style = &m.FocusedStyle
	return m.Cursor.Focus()
}

// Blur removes the focus state on the model.  When the model is blurred it can
// not receive keyboard input and the cursor will be hidden.
func (m *Model) Blur() {
	m.focus = false
	m.style = &m.BlurredStyle
	m.Cursor.Blur()
}

// Reset sets the input to its default state with no input.
func (m *Model) Reset() {
	m.value = make([][]rune, minHeight, maxHeight)
	m.col = 0
	m.row = 0
	m.viewport.GotoTop()
	m.SetCursorColumn(0)
}

// DeleteBeforeCursor deletes all text before the cursor. Returns whether or
// not the cursor blink should be reset.
func (m *Model) DeleteBeforeCursor() {
	m.value[m.row] = m.value[m.row][m.col:]
	m.SetCursorColumn(0)
}

// DeleteAfterCursor deletes all text after the cursor. Returns whether or not
// the cursor blink should be reset. If input is masked delete everything after
// the cursor so as not to reveal word breaks in the masked input.
func (m *Model) DeleteAfterCursor() {
	m.value[m.row] = m.value[m.row][:m.col]
	m.SetCursorColumn(len(m.value[m.row]) - 1)
}

// Deletes the single character on the cursor
// Returnst he rune that was deleted (if any)
func (m *Model) DeleteOnCursor() []rune {
	currentRow := m.value[m.row]
	if len(currentRow) == 0 {
		return make([]rune, 0)
	}

	deletedChar := []rune{currentRow[m.col]}
	newRow := currentRow[:m.col]
	if m.col < len(currentRow)-1 {
		// i.e., there are more characters after the cursor
		newRow = append(newRow, currentRow[m.col+1:]...)
	}
	m.value[m.row] = newRow

	newCol := min(m.col, len(newRow)-1)
	m.SetCursorColumn(newCol)

	return deletedChar
}

// MoveCursorRightOneRune moves the cursor one character to the right.
// If bindToLine is set, the cursor will not move psat the last character of the line
func (m *Model) MoveCursorRightOneRune(bindToLine bool) {
	moveRightLimit := len(m.value[m.row])
	if bindToLine {
		moveRightLimit--
	}
	if m.col < moveRightLimit {
		m.SetCursorColumn(m.col + 1)
	}
}

// MoveCursorLeftOneRune moves the cursor one character to the left.
// If insideLine is set, the cursor is moved to the last
// character in the previous line, instead of one past that.
func (m *Model) MoveCursorLeftOneRune() {
	if m.col > 0 {
		m.SetCursorColumn(m.col - 1)
	}
}

func (m *Model) MoveCursorByWord(direction CursorMovementDirection, stopPosition WordwiseMovementStopPosition) {
	m.doWordwiseMovement(direction, stopPosition)
}

// Moves the cursor in the direction of travel to the specified character
func (m *Model) MoveCursorByCharacter(direction CursorMovementDirection, char rune) {

}

func (m *Model) InsertLineAbove() {
	newValue := make([][]rune, 0, maxHeight)

	preCursorLines := m.value[0:m.row]
	newValue = append(newValue, preCursorLines...)

	newValue = append(newValue, make([]rune, 0))

	cursorLineAndAfter := m.value[m.row:]
	newValue = append(newValue, cursorLineAndAfter...)

	m.row++
	m.value = newValue
}

func (m *Model) InsertLineBelow() {
	newValue := make([][]rune, 0, maxHeight)

	cursorLineAndPrevious := m.value[0 : m.row+1]
	newValue = append(newValue, cursorLineAndPrevious...)

	newValue = append(newValue, make([]rune, 0))

	postCursorLines := m.value[m.row+1:]
	newValue = append(newValue, postCursorLines...)

	m.value = newValue
}

func (m *Model) DeleteLine() {
	if len(m.value) <= 1 {
		m.value = make([][]rune, minHeight, maxHeight)
		m.SetCursorColumn(0)
		return
	}

	preCursorLines := m.value[:m.row]
	postCursorLines := make([][]rune, 0)
	if m.row < len(m.value)-1 {
		postCursorLines = m.value[m.row+1:]
	}

	newValue := make([][]rune, 0, maxHeight)
	newValue = append(newValue, preCursorLines...)
	newValue = append(newValue, postCursorLines...)

	m.value = newValue

	m.row = clamp(m.row, 0, len(m.value)-1)
}

func (m *Model) ClearLine() {
	m.value[m.row] = make([]rune, 0, maxWidth)
	m.SetCursorColumn(0)
}

// LineInfo returns the number of characters from the start of the
// (soft-wrapped) line and the (soft-wrapped) line width.
func (m Model) GetLineInfo() LineInfo {
	grid := wrap(m.value[m.row], m.width)

	// Find out which line we are currently on. This can be determined by the
	// m.col and counting the number of runes that we need to skip.
	var counter int
	for i, line := range grid {
		// We've found the line that we are on
		if counter+len(line) == m.col && i+1 < len(grid) {
			// We wrap around to the next line if we are at the end of the
			// previous line so that we can be at the very beginning of the row
			return LineInfo{
				CharOffset:   0,
				ColumnOffset: 0,
				Height:       len(grid),
				RowOffset:    i + 1,
				StartColumn:  m.col,
				Width:        len(grid[i+1]),
				CharWidth:    rw.StringWidth(string(line)),
			}
		}

		if counter+len(line) >= m.col {
			return LineInfo{
				CharOffset:   rw.StringWidth(string(line[:max(0, m.col-counter)])),
				ColumnOffset: m.col - counter,
				Height:       len(grid),
				RowOffset:    i,
				StartColumn:  counter,
				Width:        len(line),
				CharWidth:    rw.StringWidth(string(line)),
			}
		}

		counter += len(line)
	}
	return LineInfo{}
}

// SetWidth sets the width of the textarea to fit exactly within the given width.
// This means that the textarea will account for the width of the prompt and
// whether or not line numbers are being shown.
//
// Ensure that SetWidth is called after setting the Prompt and ShowLineNumbers,
// If it important that the width of the textarea be exactly the given width
// and no more.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = clamp(w, minWidth, maxWidth)

	// Since the width of the textarea input is dependant on the width of the
	// prompt and line numbers, we need to calculate it by subtracting.
	inputWidth := w
	if m.ShowLineNumbers {
		inputWidth -= rw.StringWidth(fmt.Sprintf(m.lineNumberFormat, 0))
	}

	// Account for base style borders and padding.
	inputWidth -= m.style.Base.GetHorizontalFrameSize()

	if m.promptFunc == nil {
		m.promptWidth = rw.StringWidth(m.Prompt)
	}

	inputWidth -= m.promptWidth
	m.width = clamp(inputWidth, minWidth, maxWidth)
}

// GetWidth returns the width of the textarea.
func (m Model) GetWidth() int {
	return m.width
}

// SetPromptFunc supersedes the Prompt field and sets a dynamic prompt
// instead.
// If the function returns a prompt that is shorter than the
// specified promptWidth, it will be padded to the left.
// If it returns a prompt that is longer, display artifacts
// may occur; the caller is responsible for computing an adequate
// promptWidth.
func (m *Model) SetPromptFunc(promptWidth int, fn func(lineIdx int) string) {
	m.promptFunc = fn
	m.promptWidth = promptWidth
}

// GetHeight returns the current height of the textarea.
func (m Model) GetHeight() int {
	return m.height
}

// SetHeight sets the height of the textarea.
func (m *Model) SetHeight(h int) {
	m.height = clamp(h, minHeight, maxHeight)

	// Leave room for the extra mode display line
	m.viewport.Height = clamp(h, minHeight, maxHeight)
}

// Update is the Bubble Tea update loop.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.focus {
		m.Cursor.Blur()
		return nil
	}

	// Used to determine if the cursor should blink.
	oldRow, oldCol := m.cursorLineNumber(), m.col

	var cmds []tea.Cmd

	if m.value[m.row] == nil {
		m.value[m.row] = make([]rune, 0)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.DeleteAfterCursor):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
			m.DeleteAfterCursor()
		case key.Matches(msg, m.KeyMap.DeleteBeforeCursor):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			m.DeleteBeforeCursor()
		case key.Matches(msg, m.KeyMap.DeleteCharacterBackward):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			if len(m.value[m.row]) > 0 {
				m.value[m.row] = append(m.value[m.row][:max(0, m.col-1)], m.value[m.row][m.col:]...)
				if m.col > 0 {
					m.SetCursorColumn(m.col - 1)
				}
			}
		case key.Matches(msg, m.KeyMap.DeleteCharacterForward):
			if len(m.value[m.row]) > 0 && m.col < len(m.value[m.row]) {
				m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][m.col+1:]...)
			}
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
		case key.Matches(msg, m.KeyMap.DeleteWordBackward):
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			m.deleteWordLeft()
		case key.Matches(msg, m.KeyMap.DeleteWordForward):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
			m.deleteWordRight()
		case key.Matches(msg, m.KeyMap.InsertNewline):
			if len(m.value) >= maxHeight {
				return nil
			}
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			m.splitLine(m.row, m.col)
		case key.Matches(msg, m.KeyMap.LineEnd):
			// If the user is going to the end of the line, do allow them to go off the end of the line
			m.MoveCursorToLineEnd(false)
		case key.Matches(msg, m.KeyMap.LineStart):
			m.MoveCursorToLineStart()
		case key.Matches(msg, m.KeyMap.CharacterForward):
			// When we're editing normally, we DO want to allow moving off the end of the line
			m.MoveCursorRightOneRune(false)
		case key.Matches(msg, m.KeyMap.LineNext):
			// If the user is using the arrow keys, we don't want to be binding to the end of the line because they're
			// almost definitely in insert mode
			m.MoveCursorDown(false)
		case key.Matches(msg, m.KeyMap.WordForward):
			m.MoveCursorByWord(CursorMovementDirection_Right, WordwiseMovementStopPosition_Incidence)
		case key.Matches(msg, m.KeyMap.Paste):
			return Paste
		case key.Matches(msg, m.KeyMap.CharacterBackward):
			m.MoveCursorLeftOneRune()
		case key.Matches(msg, m.KeyMap.LinePrevious):
			// If the user is using the arrow keys, we don't want to be binding to the end of the line because they're
			// almost definitely in insert mode
			m.MoveCursorUp(false)
		case key.Matches(msg, m.KeyMap.WordBackward):
			// Note that "End" here is actually the start of the word
			m.MoveCursorByWord(CursorMovementDirection_Left, WordwiseMovementStopPosition_Terminus)
		case key.Matches(msg, m.KeyMap.InputBegin):
			m.moveToBegin()
		case key.Matches(msg, m.KeyMap.InputEnd):
			m.moveToEnd()

		default:
			m.insertRunesFromUserInput(msg.Runes)
		}

	case pasteMsg:
		m.insertRunesFromUserInput([]rune(msg))

	case pasteErrMsg:
		m.Err = msg
	}

	vp, cmd := m.viewport.Update(msg)
	m.viewport = &vp
	cmds = append(cmds, cmd)

	newRow, newCol := m.cursorLineNumber(), m.col
	m.Cursor, cmd = m.Cursor.Update(msg)
	if newRow != oldRow || newCol != oldCol {
		m.Cursor.Blink = false
		cmd = m.Cursor.BlinkCmd()
	}
	cmds = append(cmds, cmd)

	m.repositionView()

	return tea.Batch(cmds...)
}

// View renders the text area in its current state.
func (m Model) View() string {
	if m.GetValue() == "" && m.row == 0 && m.col == 0 && m.Placeholder != "" {
		return m.placeholderView()
	}
	m.Cursor.TextStyle = m.style.CursorLine

	var s strings.Builder
	var style lipgloss.Style
	lineInfo := m.GetLineInfo()

	var newLines int

	displayLine := 0
	for l, line := range m.value {
		wrappedLines := wrap(line, m.width)

		if m.row == l {
			style = m.style.CursorLine
		} else {
			style = m.style.Text
		}

		for wl, wrappedLine := range wrappedLines {
			prompt := m.getPromptString(displayLine)
			prompt = m.style.Prompt.Render(prompt)
			s.WriteString(style.Render(prompt))
			displayLine++

			if m.ShowLineNumbers {
				if wl == 0 {
					if m.row == l {
						s.WriteString(style.Render(m.style.CursorLineNumber.Render(fmt.Sprintf(m.lineNumberFormat, l+1))))
					} else {
						s.WriteString(style.Render(m.style.LineNumber.Render(fmt.Sprintf(m.lineNumberFormat, l+1))))
					}
				} else {
					s.WriteString(m.style.LineNumber.Render(style.Render("   ")))
				}
			}

			strwidth := rw.StringWidth(string(wrappedLine))
			padding := m.width - strwidth
			// If the trailing space causes the line to be wider than the
			// width, we should not draw it to the screen since it will result
			// in an extra space at the end of the line which can look off when
			// the cursor line is showing.
			if strwidth > m.width {
				// The character causing the line to be wider than the width is
				// guaranteed to be a space since any other character would
				// have been wrapped.
				wrappedLine = []rune(strings.TrimSuffix(string(wrappedLine), " "))
				padding -= m.width - strwidth
			}
			if m.row == l && lineInfo.RowOffset == wl {
				s.WriteString(style.Render(string(wrappedLine[:lineInfo.ColumnOffset])))
				if m.col >= len(line) && lineInfo.CharOffset >= m.width {
					m.Cursor.SetChar(" ")
					s.WriteString(m.Cursor.View())
				} else {
					m.Cursor.SetChar(string(wrappedLine[lineInfo.ColumnOffset]))
					s.WriteString(style.Render(m.Cursor.View()))
					s.WriteString(style.Render(string(wrappedLine[lineInfo.ColumnOffset+1:])))
				}
			} else {
				s.WriteString(style.Render(string(wrappedLine)))
			}
			s.WriteString(style.Render(strings.Repeat(" ", max(0, padding))))
			s.WriteRune('\n')
			newLines++
		}
	}

	// Always show at least `m.GetHeight` lines at all times.
	// To do this we can simply pad out a few extra new lines in the view.
	for i := 0; i < m.height; i++ {
		prompt := m.getPromptString(displayLine)
		prompt = m.style.Prompt.Render(prompt)
		s.WriteString(prompt)
		displayLine++

		if m.ShowLineNumbers {
			lineNumber := m.style.EndOfBuffer.Render((fmt.Sprintf(m.lineNumberFormat, string(m.EndOfBufferCharacter))))
			s.WriteString(lineNumber)
		}
		s.WriteRune('\n')
	}

	m.viewport.SetContent(s.String())
	return m.style.Base.Render(m.viewport.View())
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func (m *Model) doWordwiseMovement(direction CursorMovementDirection, stopPosition WordwiseMovementStopPosition) {
	// This function utilizes the insight that the textarea string can be thought of as a "tape" of words, joined by whitespace
	// With this insight, we can handle both (left,right) and (word_start,word_end) by simply sliding along the tape in
	// the appropriate direction looking for the sequence we want

	// If no lines, abort immediately
	if len(m.value) == 0 {
		return
	}

	// Factor applied to index calculations to account for the desired direction of cursor travel
	directionMultiplier := int(direction)

	// Factor applied to index calculations to account for stopPosition, where "incidence" means "look at the
	// character BEHIND the cursor to figure out if you should stop" and "terminus" means "look at the character
	// AHEAD of the cursor to figure out if you should stop" (where "ahead" and "behind" are measured by direction of
	// cursor travel)
	stopPositionMultiplier := int(stopPosition)

	// Figure out what the row index of each end of the tape is
	limitRowIndex := len(m.value) - 1
	if direction == CursorMovementDirection_Left {
		limitRowIndex = 0
	}

	// Our column might be off the right edge of the line; ensure we account for that
	sanitizedColIdx := min(m.col, len(m.value[m.row])-1)

	// At some point the proposed new column might be off either end of the line
	// Therefore, let's first calculate the boundary beyond which we know
	// that the proposed column is off the edge of the line
	limitColIndex := len(m.value[m.row]) - 1
	if direction == CursorMovementDirection_Left {
		limitColIndex = 0
	}

	// To start the algo off, shove the cursor column one character in the direction of travel
	// This prevents this being a noop if you're already at a word start/end
	nextColIdx := sanitizedColIdx + directionMultiplier
	for {
		// Now, check if the proposed column would indeed put us off the line
		remainingColsBeforeLimit := directionMultiplier * (limitColIndex - nextColIdx)
		if remainingColsBeforeLimit < 0 {
			// We're off the end; can we go to another line to keep going?
			nextRowIdx := m.row + directionMultiplier
			remainingRowsBeforeLimit := directionMultiplier * (limitRowIndex - nextRowIdx)
			if remainingRowsBeforeLimit < 0 {
				// We're at the end of the "tape"; nothing to do
				return
			}

			// We still have at least one more line, so let's use it (which means we're crossing a newline char which
			// means we're crossing a word boundary)
			m.row += directionMultiplier
			if direction == CursorMovementDirection_Right {
				nextColIdx = 0
			} else {
				nextColIdx = len(m.value[m.row]) - 1
			}
		}

		// Now that we know we're moving the column index to a valid line, move it
		m.SetCursorColumn(nextColIdx)
		nextColIdx = m.col + directionMultiplier // Prep for next iteration

		// We still might have moved the cursor to an empty line, making the cursor location invalid!
		// Vim will stop on these empty lines, so we try to as well
		if len(m.value[m.row]) == 0 {
			return
		}

		cursorChar := m.value[m.row][m.col]

		// Grab a comparison column which must be whitespace to stop the algorithm
		// The stopPosition multiplier means that "incident" will require whitespace to be *behind* the cursor in the direction
		// of algorithm travel, whereas "terminus" will require whitespace to be *ahead* of the cursor
		// in the direction of algorithm travel
		cursorAdjacentColIdx := m.col + stopPositionMultiplier*directionMultiplier

		// Depending on stopPosition, the adjacent col might actually be *behind* the cursor, meaning we need to use the opposite boundary
		// as the direction of algo travel along the "tape"
		// This happens if we're moving right but we're stopping at word incidence, OR if we're moving left and stopping at word terminus
		// NOTE: This is actually an XOR
		var adjacentColLimitIndex int
		if ((direction == CursorMovementDirection_Right) && (stopPosition == WordwiseMovementStopPosition_Terminus)) || ((direction == CursorMovementDirection_Left) && (stopPosition == WordwiseMovementStopPosition_Incidence)) {
			// The adjacent col limit index will be the list's right limit
			adjacentColLimitIndex = len(m.value[m.row]) - 1
		} else {
			adjacentColLimitIndex = 0
		}

		// Now grab a character that may or may not be whitespace
		// If the col we're grabbing is off the end of the line, it's automatically a newline
		remainingColsBeforeCursorAdjacentColIsOff := stopPositionMultiplier * directionMultiplier * (adjacentColLimitIndex - cursorAdjacentColIdx)
		var candidateWhitespaceChar rune
		if remainingColsBeforeCursorAdjacentColIsOff < 0 {
			candidateWhitespaceChar = '\n'
		} else {
			candidateWhitespaceChar = m.value[m.row][cursorAdjacentColIdx]
		}

		// Evaluate if we reached our target
		if !unicode.IsSpace(cursorChar) && unicode.IsSpace(candidateWhitespaceChar) {
			return
		}

		// We're not done, so prep for next iteration
	}
}

func (m *Model) doCharacterwiseMovement(targetChar rune, direction CursorMovementDirection, stopPosition CharacterwiseMovementStopPosition) {
	directionMultiplier := int(direction)

	stopPositionOffset := int(stopPosition)

	// To start the algo off, shove the proposed column one character in the direction of travel
	// This prevents this being a noop if you're already on the character you're looking for
	newColIdx := m.col + directionMultiplier

	// Based on the stop position, we need to examine different cells (either on the cursor, or ahead of it)
	examinationColIdx := newColIdx + stopPositionOffset

	for {
		// If our examintion column is out-of-bounds, abort; we haven't found anything
		if examinationColIdx < 0 || examinationColIdx >= len(m.value[m.row]) {
			return
		}

		if m.value[m.row][examinationColIdx] == targetChar {
			m.SetCursorColumn(newColIdx)
			return
		}

		newColIdx += directionMultiplier
		examinationColIdx += directionMultiplier
	}
}

// rsan initializes or retrieves the rune sanitizer.
func (m *Model) san() runeutil.Sanitizer {
	if m.rsan == nil {
		// Textinput has all its input on a single line so collapse
		// newlines/tabs to single spaces.
		m.rsan = runeutil.NewSanitizer()
	}
	return m.rsan
}

// insertRunesFromUserInput inserts runes at the current cursor position.
func (m *Model) insertRunesFromUserInput(runes []rune) {
	// Clean up any special characters in the input provided by the
	// clipboard. This avoids bugs due to e.g. tab characters and
	// whatnot.
	runes = m.san().Sanitize(runes)

	var availSpace int
	if m.CharLimit > 0 {
		availSpace = m.CharLimit - m.GetLength()
		// If the char limit's been reached, cancel.
		if availSpace <= 0 {
			return
		}
		// If there's not enough space to paste the whole thing cut the pasted
		// runes down so they'll fit.
		if availSpace < len(runes) {
			runes = runes[:len(runes)-availSpace]
		}
	}

	// Split the input into lines.
	var lines [][]rune
	lstart := 0
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\n' {
			lines = append(lines, runes[lstart:i])
			lstart = i + 1
		}
	}
	if lstart < len(runes) {
		// The last line did not end with a newline character.
		// Take it now.
		lines = append(lines, runes[lstart:])
	}

	// Obey the maximum height limit.
	if len(m.value)+len(lines)-1 > maxHeight {
		allowedHeight := max(0, maxHeight-len(m.value)+1)
		lines = lines[:allowedHeight]
	}

	if len(lines) == 0 {
		// Nothing left to insert.
		return
	}

	// Save the reminder of the original line at the current
	// cursor position.
	tail := make([]rune, len(m.value[m.row][m.col:]))
	copy(tail, m.value[m.row][m.col:])

	// Paste the first line at the current cursor position.
	m.value[m.row] = append(m.value[m.row][:m.col], lines[0]...)
	m.col += len(lines[0])

	if numExtraLines := len(lines) - 1; numExtraLines > 0 {
		// Add the new lines.
		// We try to reuse the slice if there's already space.
		var newGrid [][]rune
		if cap(m.value) >= len(m.value)+numExtraLines {
			// Can reuse the extra space.
			newGrid = m.value[:len(m.value)+numExtraLines]
		} else {
			// No space left; need a new slice.
			newGrid = make([][]rune, len(m.value)+numExtraLines)
			copy(newGrid, m.value[:m.row+1])
		}
		// Add all the rows that were after the cursor in the original
		// grid at the end of the new grid.
		copy(newGrid[m.row+1+numExtraLines:], m.value[m.row+1:])
		m.value = newGrid
		// Insert all the new lines in the middle.
		for _, l := range lines[1:] {
			m.row++
			m.value[m.row] = l
			m.col = len(l)
		}
	}

	// Finally add the tail at the end of the last line inserted.
	m.value[m.row] = append(m.value[m.row], tail...)

	m.SetCursorColumn(m.col)
}

// deleteWordLeft deletes the word left to the cursor. Returns whether or not
// the cursor blink should be reset.
func (m *Model) deleteWordLeft() {
	if m.col == 0 || len(m.value[m.row]) == 0 {
		return
	}

	// Linter note: it's critical that we acquire the initial cursor position
	// here prior to altering it via SetCursorColumn() below. As such, moving this
	// call into the corresponding if clause does not apply here.
	oldCol := m.col //nolint:ifshort

	m.SetCursorColumn(m.col - 1)
	for unicode.IsSpace(m.value[m.row][m.col]) {
		if m.col <= 0 {
			break
		}
		// ignore series of whitespace before cursor
		m.SetCursorColumn(m.col - 1)
	}

	for m.col > 0 {
		if !unicode.IsSpace(m.value[m.row][m.col]) {
			m.SetCursorColumn(m.col - 1)
		} else {
			if m.col > 0 {
				// keep the previous space
				m.SetCursorColumn(m.col + 1)
			}
			break
		}
	}

	if oldCol > len(m.value[m.row]) {
		m.value[m.row] = m.value[m.row][:m.col]
	} else {
		m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][oldCol:]...)
	}
}

// deleteWordRight deletes the word right to the cursor.
func (m *Model) deleteWordRight() {
	if m.col >= len(m.value[m.row]) || len(m.value[m.row]) == 0 {
		return
	}

	oldCol := m.col
	m.SetCursorColumn(m.col + 1)
	for unicode.IsSpace(m.value[m.row][m.col]) {
		// ignore series of whitespace after cursor
		m.SetCursorColumn(m.col + 1)

		if m.col >= len(m.value[m.row]) {
			break
		}
	}

	for m.col < len(m.value[m.row]) {
		if !unicode.IsSpace(m.value[m.row][m.col]) {
			m.SetCursorColumn(m.col + 1)
		} else {
			break
		}
	}

	if m.col > len(m.value[m.row]) {
		m.value[m.row] = m.value[m.row][:oldCol]
	} else {
		m.value[m.row] = append(m.value[m.row][:oldCol], m.value[m.row][m.col:]...)
	}

	m.SetCursorColumn(oldCol)
}

func (m *Model) doWordRight(fn func(charIdx int, pos int)) {
	haveEncounteredWhitespace := false
	for {
		// If we're at (or beyond) the last char of the line (which may be empty)...
		if m.col >= len(m.value[m.row])-1 {
			// ...and there are no more lines, we're done
			if m.row == len(m.value)-1 {
				return
			}

			// ...and the next line is empty, then move to the next line and we're done
			// This is a bit odd, but it's Vim behaviour
			if len(m.value[m.row+1]) == 0 {
				// Copied from MoveCursorRightOneRune
				m.MoveCursorToLineStart()
				m.row++
				return
			}

			// ...and the next line is not empty, prep to stop on the next word
			m.row++
			m.MoveCursorToLineStart()
			haveEncounteredWhitespace = true
			continue
		}

		charUnderCursor := m.value[m.row][m.col]

		// We've already left a word and found another; we're done
		if haveEncounteredWhitespace && !unicode.IsSpace(charUnderCursor) {
			return
		}

		// We've seen a space, so we're ready to stop
		if unicode.IsSpace(charUnderCursor) {
			haveEncounteredWhitespace = true
		}

		// We don't want the cursor to move off the end of everything
		m.MoveCursorRightOneRune(true)
	}

	/*
		charIdx := 0
		for m.col < len(m.value[m.row]) {
			if unicode.IsSpace(m.value[m.row][m.col]) {
				break
			}
			fn(charIdx, m.col)
			m.SetCursorColumn(m.col + 1)
			charIdx++
		}

	*/
}

// repositionView repositions the view of the viewport based on the defined
// scrolling behavior.
func (m *Model) repositionView() {
	min := m.viewport.YOffset
	max := min + m.viewport.Height - 1

	if row := m.cursorLineNumber(); row < min {
		m.viewport.LineUp(min - row)
	} else if row > max {
		m.viewport.LineDown(row - max)
	}
}

// moveToBegin moves the cursor to the beginning of the input.
func (m *Model) moveToBegin() {
	m.row = 0
	m.SetCursorColumn(0)
}

// moveToEnd moves the cursor to the end of the input.
func (m *Model) moveToEnd() {
	m.row = len(m.value) - 1
	m.SetCursorColumn(len(m.value[m.row]))
}

func (m Model) getPromptString(displayLine int) (prompt string) {
	prompt = m.Prompt
	if m.promptFunc == nil {
		return prompt
	}
	prompt = m.promptFunc(displayLine)
	pl := rw.StringWidth(prompt)
	if pl < m.promptWidth {
		prompt = fmt.Sprintf("%*s%s", m.promptWidth-pl, "", prompt)
	}
	return prompt
}

// placeholderView returns the prompt and placeholder view, if any.
func (m Model) placeholderView() string {
	var (
		s     strings.Builder
		p     = rw.Truncate(m.Placeholder, m.width, "...")
		style = m.style.Placeholder.Inline(true)
	)

	prompt := m.getPromptString(0)
	prompt = m.style.Prompt.Render(prompt)
	s.WriteString(m.style.CursorLine.Render(prompt))

	if m.ShowLineNumbers {
		s.WriteString(m.style.CursorLine.Render(m.style.CursorLineNumber.Render((fmt.Sprintf(m.lineNumberFormat, 1)))))
	}

	m.Cursor.TextStyle = m.style.Placeholder
	m.Cursor.SetChar(string(p[0]))
	s.WriteString(m.style.CursorLine.Render(m.Cursor.View()))

	// The rest of the placeholder text
	s.WriteString(m.style.CursorLine.Render(style.Render(p[1:] + strings.Repeat(" ", max(0, m.width-rw.StringWidth(p))))))

	// The rest of the new lines
	for i := 1; i < m.height; i++ {
		s.WriteRune('\n')
		prompt := m.getPromptString(i)
		prompt = m.style.Prompt.Render(prompt)
		s.WriteString(prompt)

		if m.ShowLineNumbers {
			eob := m.style.EndOfBuffer.Render((fmt.Sprintf(m.lineNumberFormat, string(m.EndOfBufferCharacter))))
			s.WriteString(eob)
		}
	}

	m.viewport.SetContent(s.String())
	return m.style.Base.Render(m.viewport.View())
}

// Blink returns the blink command for the cursor.
func Blink() tea.Msg {
	return cursor.Blink()
}

// cursorLineNumber returns the line number that the cursor is on.
// This accounts for soft wrapped lines.
func (m Model) cursorLineNumber() int {
	line := 0
	for i := 0; i < m.row; i++ {
		// Calculate the number of lines that the current line will be split
		// into.
		line += len(wrap(m.value[i], m.width))
	}
	line += m.GetLineInfo().RowOffset
	return line
}

// mergeLineBelow merges the current line with the line below.
func (m *Model) mergeLineBelow(row int) {
	if row >= len(m.value)-1 {
		return
	}

	// To perform a merge, we will need to combine the two lines and then
	m.value[row] = append(m.value[row], m.value[row+1]...)

	// Shift all lines up by one
	for i := row + 1; i < len(m.value)-1; i++ {
		m.value[i] = m.value[i+1]
	}

	// And, remove the last line
	if len(m.value) > 0 {
		m.value = m.value[:len(m.value)-1]
	}
}

// mergeLineAbove merges the current line the cursor is on with the line above.
func (m *Model) mergeLineAbove(row int) {
	if row <= 0 {
		return
	}

	m.col = len(m.value[row-1])
	m.row = m.row - 1

	// To perform a merge, we will need to combine the two lines and then
	m.value[row-1] = append(m.value[row-1], m.value[row]...)

	// Shift all lines up by one
	for i := row; i < len(m.value)-1; i++ {
		m.value[i] = m.value[i+1]
	}

	// And, remove the last line
	if len(m.value) > 0 {
		m.value = m.value[:len(m.value)-1]
	}
}

func (m *Model) splitLine(row, col int) {
	// To perform a split, take the current line and keep the content before
	// the cursor, take the content after the cursor and make it the content of
	// the line underneath, and shift the remaining lines down by one
	head, tailSrc := m.value[row][:col], m.value[row][col:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	m.value = append(m.value[:row+1], m.value[row:]...)

	m.value[row] = head
	m.value[row+1] = tail

	m.col = 0
	m.row++
}

// Wrap a rune string into an array of rune strings
func wrap(runes []rune, width int) [][]rune {
	var (
		lines  = [][]rune{{}}
		word   = []rune{}
		row    int
		spaces int
	)

	// Word wrap the runes
	for _, r := range runes {
		if unicode.IsSpace(r) {
			spaces++
		} else {
			word = append(word, r)
		}

		if spaces > 0 {
			if rw.StringWidth(string(lines[row]))+rw.StringWidth(string(word))+spaces > width {
				row++
				lines = append(lines, []rune{})
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatSpaces(spaces)...)
				spaces = 0
				word = nil
			} else {
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatSpaces(spaces)...)
				spaces = 0
				word = nil
			}
		} else {
			// If the last character is a double-width rune, then we may not be able to add it to this line
			// as it might cause us to go past the width.
			lastCharLen := rw.RuneWidth(word[len(word)-1])
			if rw.StringWidth(string(word))+lastCharLen > width {
				// If the current line has any content, let's move to the next
				// line because the current word fills up the entire line.
				if len(lines[row]) > 0 {
					row++
					lines = append(lines, []rune{})
				}
				lines[row] = append(lines[row], word...)
				word = nil
			}
		}
	}

	if rw.StringWidth(string(lines[row]))+rw.StringWidth(string(word))+spaces >= width {
		lines = append(lines, []rune{})
		lines[row+1] = append(lines[row+1], word...)
		// We add an extra space at the end of the line to account for the
		// trailing space at the end of the previous soft-wrapped lines so that
		// behaviour when navigating is consistent and so that we don't need to
		// continually add edges to handle the last line of the wrapped input.
		spaces++
		lines[row+1] = append(lines[row+1], repeatSpaces(spaces)...)
	} else {
		lines[row] = append(lines[row], word...)
		spaces++
		lines[row] = append(lines[row], repeatSpaces(spaces)...)
	}

	return lines
}

func repeatSpaces(n int) []rune {
	return []rune(strings.Repeat(string(' '), n))
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
