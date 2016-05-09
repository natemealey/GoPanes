package gopanes

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"strconv"
	"sync"
)

const tcd = termbox.ColorDefault

// TODO is there a better place to put this?
var termboxMutex = &sync.Mutex{}

func TermboxSafeFlush() {
	termboxMutex.Lock()
	termbox.Flush()
	termboxMutex.Unlock()
}

type GoPaneUi struct {
	Root *GoPane
}

func (gu *GoPaneUi) getWindowWidth() int {
	x, _ := termbox.Size()

	return x
}

func (gu *GoPaneUi) getWindowHeight() int {
	_, y := termbox.Size()

	return y
}

func (gu *GoPaneUi) Refresh() {
	gu.Root.Refresh()
}

// Close MUST be called on program exit to clean up after termbox
func (gu *GoPaneUi) Close() {
	termbox.Close()
}

// splits are
type GoPane struct {
	First         *GoPane
	Second        *GoPane
	isVertical    bool
	isFocused     bool // only unsplit (leaf) panes can be focused
	splitLocation int
	x             int
	y             int
	width         int
	height        int
	content       [][]ColorStr
	editBox       *EditBox
}

func NewGoPane(width int, height int, x int, y int) *GoPane {
	return &GoPane{
		width:  width,
		height: height,
		x:      x,
		y:      y,
		First:  nil,
		Second: nil}
}

func (gu *GoPaneUi) GetFocusedPane() *GoPane {
	return gu.Root.GetFocusedChild()
}

func (gu *GoPaneUi) MoveUp(target *GoPane) {
	// find the pane above the given pane
	gu.FocusPane(gu.Root.First)
}

func (gu *GoPaneUi) MoveDown(target *GoPane) {
	// find the pane above the given pane
	gu.FocusPane(gu.Root.Second)
}

func (gu *GoPaneUi) MoveLeft(target *GoPane) {}

func (gu *GoPaneUi) MoveRight(target *GoPane) {}

func (gu *GoPaneUi) ResizeUp(target *GoPane) {}

func (gu *GoPaneUi) ResizeDown(target *GoPane) {}

func (gu *GoPaneUi) ResizeLeft(target *GoPane) {}

func (gu *GoPaneUi) ResizeRight(target *GoPane) {}

// all possible functions to do after a prefix
// TODO make these customizable
func (gu *GoPaneUi) HandleCommand(ev termbox.Event) {
	target := gu.GetFocusedPane()
	if target == nil {
		return // TODO error message?
	}
	switch ev.Key {
	case termbox.KeyArrowUp: // move up one pane
		gu.MoveUp(target)
	case termbox.KeyArrowDown: // move down one pane
		gu.MoveDown(target)
	case termbox.KeyArrowLeft: // move left one pane
		gu.MoveLeft(target)
	case termbox.KeyArrowRight: // move right one pane
		gu.MoveRight(target)
	default:
		switch ev.Ch {
		case 'k':
			gu.MoveUp(target)
		case 'j':
			gu.MoveDown(target)
		case 'l':
			gu.MoveLeft(target)
		case 'h':
			gu.MoveRight(target)
		case 'K':
			gu.ResizeUp(target)
		case 'J':
			gu.ResizeDown(target)
		case 'L':
			gu.ResizeLeft(target)
		case 'H':
			gu.ResizeRight(target)
		}
	}
}

func (gu *GoPaneUi) Listen() {
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	var prefix bool
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventMouse:
			// focus on clicked panes
			if ev.Key == termbox.MouseRelease {
				target := gu.GetTargetPane(ev.MouseX, ev.MouseY)
				if target != nil {
					gu.FocusPane(target)
				}
			}
		case termbox.EventKey:
			// TODO if it's kill signal, just quit
			if prefix {
				gu.HandleCommand(ev)
				prefix = false
			} else if ev.Key == termbox.KeyCtrlG { // TODO custom prefix
				prefix = true
			} else {
				// get target pane
				target := gu.GetFocusedPane()
				// call pane event handler
				// TODO filter out mouse event escape sequences
				if target != nil {
					target.HandleEvent(ev)
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}

}

// This is the ONLY function that should be used to focus a pane
//  Using an individual pane's focus() will cause inconsistent state, since
//  it could allow multiple panes to be focused
func (gu *GoPaneUi) FocusPane(gp *GoPane) {
	gu.Root.focusChild(gp)
}

func (gu *GoPaneUi) GetTargetPane(x, y int) *GoPane {
	return gu.Root.childInBounds(x, y)
}

// Creates a new root UI. This should only be used once in a program,
//   when initializing the GoPane UI
// TODO should this be a singleton?
func NewGoPaneUi() *GoPaneUi {
	var newUi GoPaneUi

	termbox.Init()

	newUi.Root = NewGoPane(newUi.getWindowWidth(), newUi.getWindowHeight(), 0, 0)

	go newUi.Listen()

	return &newUi
}

func (gp *GoPane) isSplit() bool {
	return gp.First != nil && gp.Second != nil
}

// returns the child that is in the given bounds, or nil if none exists
func (gp *GoPane) childInBounds(x, y int) *GoPane {
	if gp.isSplit() {
		first := gp.First.childInBounds(x, y)
		if first != nil {
			return first
		}
		return gp.Second.childInBounds(x, y)
	} else {
		if gp.IsInBounds(x, y) {
			return gp
		}
		return nil
	}
}

func (gp *GoPane) IsInBounds(x, y int) bool {
	return (x >= gp.x && x < gp.x+gp.width) && (y >= gp.y && y < gp.y+gp.height)
}

func (gp *GoPane) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyArrowUp:
	case termbox.KeyArrowDown:
	case termbox.KeyArrowLeft:
	case termbox.KeyArrowRight:
	}
}

func (gp *GoPane) HandleEvent(ev termbox.Event) {
	if gp.IsEditable() {
		gp.editBox.HandleEvent(ev)
	} else {
		gp.HandleKey(ev.Key)
	}
}

func (gp *GoPane) IsEditable() bool {
	return gp.editBox != nil
}

func (gp *GoPane) MakeEditable() {
	gp.editBox = NewEditBox(gp.x, gp.y, gp.width, gp.height, nil)
}

func (gp *GoPane) IsAlive() bool {
	// if it's editable, return edit alive state, otherwise true
	return (gp.IsEditable() && gp.editBox.Alive()) || !gp.IsEditable()
}

// If the goPane is editable, get the next line from it
func (gp *GoPane) GetLine() string {
	if gp.IsEditable() {
		return string(gp.editBox.GetLine())
	}
	// TODO should this have a better failure mode?
	return ""
}

func (gp *GoPane) ChangePrompt(colorStrs []ColorStr) {
	if gp.IsEditable() {
		gp.editBox.ChangePrompt(colorStrs)
	}
}
func (gp *GoPane) Info() {
	fmt.Printf("Pane is ")
	if !gp.isVertical {
		fmt.Printf("not ")
	}
	fmt.Printf("vertical, split at", gp.splitLocation, "at coords", gp.x, gp.y)
	fmt.Printf(" and is", gp.width, "wide and", gp.height, "tall\n")
}

// splits a pane horizonally at the given line
func (gp *GoPane) Horiz(y int) bool {
	// note: negative values split from bottom
	// just refuse to split if it's invalid
	if y == 0 || y >= gp.height || y*-1 > gp.height {
		return false
	}
	gp.isVertical = false
	gp.splitLocation = y
	absSplit := y
	if absSplit < 0 {
		absSplit += gp.height
	}
	gp.First = NewGoPane(gp.width, absSplit-1, gp.x, gp.y)
	gp.Second = NewGoPane(gp.width, gp.height-(absSplit+1), gp.x, gp.y+absSplit)
	// the First pane inherits the split content
	gp.First.content = gp.content
	gp.content = nil
	return true
}

// splits a pane vertically at the given line
func (gp *GoPane) Vert(x int) bool {
	// note: negative values split from right
	// just refuse to split if it's invalid
	if x == 0 || x >= gp.width || x*-1 > gp.width {
		return false
	}
	gp.isVertical = true
	gp.splitLocation = x
	absSplit := x
	if absSplit < 0 {
		absSplit += gp.width
	}
	gp.First = NewGoPane(absSplit, gp.height, gp.x, gp.y)
	gp.Second = NewGoPane(gp.width-(absSplit+1), gp.height, gp.x+absSplit+1, gp.y)
	// the First pane inherits the split content
	gp.First.content = gp.content
	gp.content = nil
	return true
}

func (gp *GoPane) AddLine(colorStrs []ColorStr) {
	if gp.isSplit() {
		gp.First.AddLine(colorStrs)
	} else {
		gp.content = append(gp.content, colorStrs)
	}
}

// deletes all content
func (gp *GoPane) Clear() {
	gp.content = nil
}

func (gp *GoPane) GetFocusedChild() *GoPane {
	if gp.isSplit() {
		first := gp.First.GetFocusedChild()
		if first != nil {
			return first
		}
		return gp.Second.GetFocusedChild()
	}
	if gp.isFocused {
		return gp
	}
	return nil
}

// unfocuses all leaf nodes of gp, and focuses on only the target pane
func (gp *GoPane) focusChild(target *GoPane) {
	if gp.isSplit() {
		gp.First.focusChild(target)
		gp.Second.focusChild(target)
	} else if gp == target {
		gp.focus()
	} else {
		gp.unfocus()
	}
}

// The focus, focusChild, and unfocus functions are ONLY for the GoPaneUi class to manipulate
func (gp *GoPane) focus() {
	// move the cursor to the output start
	if gp.isSplit() {
		gp.First.focus()
		gp.Second.unfocus()
	} else if gp.IsEditable() {
		gp.editBox.Focus()
		gp.editBox.Refresh()
	} else {
		termbox.SetCursor(gp.x, gp.y)
	}
	gp.isFocused = true
}

func (gp *GoPane) unfocus() {
	// move the cursor to the output start
	if gp.IsEditable() {
		gp.editBox.UnFocus()
		gp.editBox.Refresh()
	}
	gp.isFocused = false
}

// TODO move somewhere useful possibly
func getOutputWidth(src []ColorRune) int {
	width := 0
	for _, colorRune := range src {
		// TODO tab handling
		if strconv.IsPrint(colorRune.r) {
			width++
		}
	}
	return width
}

// rerenders all content in the given pane
func (gp *GoPane) Refresh() {
	if gp.isSplit() {
		// in-order traversal of child panes
		gp.First.Refresh()
		// TODO custom borders
		if gp.isVertical {
			absSplit := gp.splitLocation
			if gp.splitLocation < 0 {
				absSplit = gp.splitLocation + gp.width
			}
			for y := gp.y; y <= gp.y+gp.height; y++ {
				termbox.SetCell(gp.x+absSplit-1, y, '|', termbox.ColorWhite, tcd)
			}
		} else {
			absSplit := gp.splitLocation
			if gp.splitLocation < 0 {
				absSplit = gp.splitLocation + gp.height
			}
			for x := gp.x; x <= gp.x+gp.width; x++ {
				termbox.SetCell(x, gp.y+absSplit-1, '─', termbox.ColorWhite, tcd)
			}
		}
		gp.Second.Refresh()
	} else {
		// it's a leaf pane, so render its content
		// First, create a byte buffer of the output
		buf := make([][]ColorRune, len(gp.content))
		col := 0
		bufRow := 0
		for _, row := range gp.content {
			for _, colorStr := range row {
				for _, char := range colorStr.str {
					if strconv.IsPrint(char) {
						// if it's not printable, don't add it to the width
						col++
					}
					// handle line wrapping
					if col >= gp.width || char == '\n' {
						col = 0
						bufRow++
						buf = append(buf, make([]ColorRune, 0))
					}
					if char != '\n' {
						buf[bufRow] = append(buf[bufRow], ColorRune{
							r: char, fg: colorStr.fg, bg: colorStr.bg})
					}
				}
			}
			col = 0
			bufRow++
		}
		// set the cells in the termbox buffer (or at least, all that can fit)
		startRow := len(buf) - gp.height
		if startRow < 0 {
			startRow = 0
		}
		for rownum, row := range buf[startRow:] {
			for charnum, colorRune := range row {
				termbox.SetCell(gp.x+charnum, gp.y+rownum, colorRune.r, colorRune.fg, colorRune.bg)
			}
			for spaceStart := getOutputWidth(row); spaceStart < gp.width; spaceStart++ {
				termbox.SetCell(gp.x+spaceStart, gp.y+rownum, ' ', tcd, tcd)
			}
		}
		// set all empty rows as spaces
		for row := len(buf); row < gp.height; row++ {
			for idx := 0; idx < gp.width; idx++ {
				termbox.SetCell(gp.x+idx, gp.y+row, ' ', tcd, tcd)
			}
		}
	}
	TermboxSafeFlush()
}
