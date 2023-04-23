package vim

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mieubrisse/vim-bubble/textarea"
	"golang.org/x/text/runes"
	"strings"
)

// When these characters are in the
type characterReducers struct {


}

func (model Model) handleNormalModeKeypress(msg tea.KeyMsg) Model {
	// 2. process the buffer
	// 3. potentially execute an action


	// 1. add to the buffer
	sanitizedRunes :=
	model.nGraphBuffer += msg.Runes

	// 2. process the buffer
	if strings.HasSuffix(model.nGraphBuffer, "G") {
		// "G" is a little special - unlike the other motion commands, adding a number in front of it won't just redo
		// the same command N times

	}

	// Rest of the motion commands take N in front of them






	// This is the only weird one, where repetition doesn't work
	if msg.String() == "G" {

	}

	// 3 types:
	// - immediate action (e.g. j, k, etc.)
	// - thing that adds to the ngraph buffer for the future

	// If this is a known thing,

	// Next,


	// Otherwise, send the input to the text area





	willDoCharacterwiseMovement := strings.HasSuffix(model.nGraphBuffer, "f") ||
		strings.HasSuffix(model.nGraphBuffer, "t") ||

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

	return model
}

func (model Model) parseCommandStr(str string) Model {
	// Strip off numbers first
	numbersEndIdxInclusive := 1
	for i := 0; i <  {

	}
}

// Parse a command string that's guaranteed not to have leading numbers
func (model Model) parseNumberlessCommandStr() func(model Model) Model {

}

func isNgraphSuffixCharacterwiseMovement(ngraphStr string) {
	return strings.HasSuffix(ngraphStr, "f") ||
		strings.HasSuffix(ngraphStr, "t") ||
		strings.HasSuffix(ngraphStr, "F") ||
		strings.HasSuffix(ngraphStr)
}

