package gopanes

import (
	"fmt"
	"strconv"
)

/* how does a GoPane work?
well, a lot like tmux, but within go
a GoPane lets you split it horizontally or vertically
so a GoPane has either zero or two elements, depending on if it's split or not
if there's zero, then it's a single element, and thus a complete pane, and it
  can be printed to directly
if there are two, the pane is split
a print to a parent will print to the first element of each child until it reaches the bottom
all GoPanes have a width and height
the width and height of the topmost parent will be the width and height of the window
Conceptually, the GoPane structure is a complete binary tree

What's supported?
- ANSI terminals (just those for now)
- text wrapping, because we don't want to lose data
- title overflow
- vertical overflow (will scroll up and down)
	* TODO this is a later feature
- responsive resize detection (using polling)
	* TODO this is a later feature
	* when resized, panes will be resized by percentages
		- TODO unless otherwise specified?
- custom coloring (including dividers)

What's not supported?
- padding or margins or any such styles - this is not a window manager
- non-ANSI terminals - ANSI is enough for now!

Why does this exist?
- I don't like the existing solutions - I don't want a window manager or anything
	heavyweight
- I think tmux is incredibly useful and it'd be great to have inside my Go projects
- The terminal is great and this makes using a minimalistic terminal UI in Go very easy
- I think Go is fun and I like writing things in it
*/

type GoPaneUi struct {
	rootPane *GoPane
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

// splits are
type GoPane struct {
	first         *GoPane
	second        *GoPane
	isVertical    bool
	splitLocation int
	x             int
	y             int
	width         int
	height        int
	content       []string
}

func NewGoPane(x int, y int, width int, height int) *GoPane {
	return &GoPane{
		width:  width,
		height: height,
		x:      x,
		y:      y}
}

// Creates a new root UI. This should only be used once in a program,
// when initializing the GoPane UI
func NewGoPaneUi() *GoPaneUi {
	var newUi GoPaneUi

	newUi.rootPane = NewGoPane(newUi.getWindowWidth(), newUi.getWindowHeight(), 0, 0)

	return &newUi
}

func (gp *GoPane) isSplit() bool {
	return gp.first == nil && gp.second == nil
}

// splits a pane horizonally at the given line
func (gp *GoPane) horiz(y int) bool {
	// just refuse to split if it's invalid
	if y <= 0 || y >= gp.height {
		return false
	}
	gp.isVertical = false
	gp.first = NewGoPane(gp.width, y-1, gp.x, gp.y)
	gp.second = NewGoPane(gp.width, gp.height-y, gp.x, gp.y+y)
	// the first pane inherits the split content
	gp.first.content = gp.content
	gp.content = nil
	return true
}

// splits a pane vertically at the given line
func (gp *GoPane) vert(x int) bool {
	// just refuse to split if it's invalid
	if x <= 0 || x >= gp.width {
		return false
	}
	gp.isVertical = true
	gp.first = NewGoPane(x-1, gp.height, gp.x, gp.y)
	gp.second = NewGoPane(gp.width-x, gp.height, gp.x+x, gp.y)
	return true
}

// uses ANSI codes to move the cursor to the given point
func moveCursor(x int, y int) {
	fmt.Printf("\033[" + strconv.Itoa(x) + ";" + strconv.Itoa(y) + "H")
}

// rerenders all content in the given pane
func (gp *GoPane) render() {
	if gp.isSplit() {
		// in-order traversal of child panes
		gp.first.render()
		// render line
		if gp.isVertical {
			for y := gp.y; y <= gp.y+gp.height; y++ {
				moveCursor(gp.x+gp.splitLocation, y)
				fmt.Printf("|") //TODO custom border
			}
		} else {
			for x := gp.y; x <= gp.x+gp.width; x++ {
				moveCursor(gp.x+gp.splitLocation, x)
				fmt.Printf("-") //TODO custom border
			}
		}
		gp.second.render()
	} else {
		// it's a leaf pane, so render its content
		// first, create a byte buffer of the output
		maxRows := len(gp.content)
		buf := make([]string, maxRows)
		col := 0
		for row := 0; row < maxRows; row++ {
			for _, char := range gp.content[row] {
				// if it's not printable, don't add it to the width
				if strconv.IsPrint(char) {
					col++
				}
				// handle line wrapping
				if col > gp.width {
					col = 0
					row++
					maxRows++
					buf = append(buf, "")
				}
				buf[row] += string(char)
			}
		}
		// move the cursor to the output start
		moveCursor(gp.x, gp.y)
		// then, output the buffer (or at least, all that can fit)
		startRow := gp.height - len(buf)
		if startRow < 0 {
			startRow = 0
		}
		for row := startRow; row < len(buf); row++ {
			fmt.Println(buf[row])
		}
	}
}
