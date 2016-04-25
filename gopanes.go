package gopanes

import (
	"fmt"
	"strconv"
	"strings"
)

// from github.com/buger/goterm
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

type GoPaneUi struct {
	Root *GoPane
}

func (gu *GoPaneUi) getWindowWidth() int {
	ws, err := getWinsize()

	if err != nil {
		return -1
	}

	return int(ws.Col)
}

func (gu *GoPaneUi) getWindowHeight() int {
	ws, err := getWinsize()

	if err != nil {
		return -1
	}

	return int(ws.Row)
}

func (gu *GoPaneUi) Refresh() {
	gu.Root.Refresh()
}

func (gu *GoPaneUi) Clear() {
	fmt.Printf("\033[H\033[2J")
}

// splits are
type GoPane struct {
	First         *GoPane
	Second        *GoPane
	isVertical    bool
	splitLocation int
	x             int
	y             int
	width         int
	height        int
	content       []string
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

// Creates a new root UI. This should only be used once in a program,
// when initializing the GoPane UI
// TODO should this be a singleton?
func NewGoPaneUi() *GoPaneUi {
	var newUi GoPaneUi

	newUi.Root = NewGoPane(newUi.getWindowWidth(), newUi.getWindowHeight(), 0, 0)

	return &newUi
}

func (gp *GoPane) isSplit() bool {
	return gp.First != nil && gp.Second != nil
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
	// negative values split from bottom
	// TODO make this work better when resizing
	if y < 0 {
		y = gp.height + y
	}
	// just refuse to split if it's invalid
	if y <= 0 || y >= gp.height {
		return false
	}
	gp.isVertical = false
	gp.splitLocation = y
	gp.First = NewGoPane(gp.width, y-1, gp.x, gp.y)
	gp.Second = NewGoPane(gp.width, gp.height-(y+1), gp.x, gp.y+y)
	// the First pane inherits the split content
	gp.First.content = gp.content
	gp.content = nil
	return true
}

// splits a pane vertically at the given line
func (gp *GoPane) Vert(x int) bool {
	// negative values split from right
	// TODO make this work better when resizing
	if x < 0 {
		x = gp.width + x
	}
	// just refuse to split if it's invalid
	if x <= 0 || x >= gp.width {
		return false
	}
	gp.isVertical = true
	gp.splitLocation = x
	gp.First = NewGoPane(x, gp.height, gp.x, gp.y)
	gp.Second = NewGoPane(gp.width-(x+1), gp.height, gp.x+x+1, gp.y)
	// the First pane inherits the split content
	gp.First.content = gp.content
	gp.content = nil
	return true
}

func (gp *GoPane) AddLine(line string) {
	if gp.isSplit() {
		gp.First.AddLine(line)
	} else {
		gp.content = append(gp.content, line)
	}
}

// deletes all content
func (gp *GoPane) Clear() {
	gp.content = nil
}

// uses ANSI codes to move the cursor to the given point
func moveCursor(x int, y int) {
	// TODO for 1-indexed ANSI rows/cols
	x++
	y++
	fmt.Printf("\033[" + strconv.Itoa(y) + ";" + strconv.Itoa(x) + "H")
}

func (gp *GoPane) getCursorPosition() (int, int) {

	return 0, 0
}

func (gp *GoPane) Focus() {
	// move the cursor to the output start
	moveCursor(gp.x, gp.y)
}

// TODO move somewhere useful possibly
func getOutputWidth(src string) int {
	width := 0
	for _, char := range src {
		if strconv.IsPrint(char) {
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
		// TODO support more terminals
		fmt.Printf("\0337")
		// render line
		if gp.isVertical {
			for y := gp.y; y <= gp.y+gp.height; y++ {
				moveCursor(gp.x+gp.splitLocation-1, y)
				fmt.Printf("|") //TODO custom border
			}
		} else {
			for x := gp.x; x <= gp.x+gp.width; x++ {
				moveCursor(x, gp.y+gp.splitLocation-1)
				fmt.Printf("-") //TODO custom border
			}
		}
		// restore cursor position
		fmt.Printf("\0338")
		gp.Second.Refresh()
	} else {
		// it's a leaf pane, so render its content
		// First, create a byte buffer of the output
		buf := make([]string, len(gp.content))
		col := 0
		bufRow := 0
		for _, row := range gp.content {
			for _, char := range row {
				if strconv.IsPrint(char) {
					// if it's not printable, don't add it to the width
					col++
				}
				// handle line wrapping
				if col >= gp.width || char == '\n' {
					col = 0
					bufRow++
					buf = append(buf, "")
				}
				if char != '\n' {
					buf[bufRow] += string(char)
				}
			}
			col = 0
			bufRow++
		}
		// save original position
		// TODO support more terminals
		fmt.Printf("\0337")
		// move the cursor to the output start
		moveCursor(gp.x, gp.y)
		// then, output the buffer (or at least, all that can fit)
		startRow := len(buf) - gp.height
		if startRow < 0 {
			startRow = 0
		}
		for _, row := range buf[startRow:] {
			spaces := ""

			for spaceStart := getOutputWidth(row); spaceStart < gp.width; spaceStart++ {
				spaces += " "
			}
			fmt.Println(row + spaces)
		}
		// output all empty rows as spaces
		spaceRow := strings.Repeat(" ", gp.width)
		for row := 0; row < (gp.height - len(buf)); row++ {
			fmt.Println(spaceRow)
		}
		// restore cursor position
		fmt.Printf("\0338")
	}
}
