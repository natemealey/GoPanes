// Modified version of github.com/nsf/termbox-go/_demos/editbox.go

package gopanes

import (
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"unicode/utf8"
)

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}

func fill(x, y, w, h int, cell termbox.Cell) {
	for ly := 0; ly < h; ly++ {
		for lx := 0; lx < w; lx++ {
			termbox.SetCell(x+lx, y+ly, cell.Ch, cell.Fg, cell.Bg)
		}
	}
}

func rune_advance_len(r rune, pos int) int {
	if r == '\t' {
		return tabstop_length - pos%tabstop_length
	}
	return runewidth.RuneWidth(r)
}

func voffset_coffset(text []byte, boffset int) (voffset, coffset int) {
	text = text[:boffset]
	for len(text) > 0 {
		r, size := utf8.DecodeRune(text)
		text = text[size:]
		coffset += 1
		voffset += rune_advance_len(r, voffset)
	}
	return
}

func byte_slice_grow(s []byte, desired_cap int) []byte {
	if cap(s) < desired_cap {
		ns := make([]byte, len(s), desired_cap)
		copy(ns, s)
		return ns
	}
	return s
}

func byte_slice_remove(text []byte, from, to int) []byte {
	size := to - from
	copy(text[from:], text[to:])
	text = text[:len(text)-size]
	return text
}

func byte_slice_insert(text []byte, offset int, what []byte) []byte {
	n := len(text) + len(what)
	text = byte_slice_grow(text, n)
	text = text[:n]
	copy(text[offset+len(what):], text[offset:])
	copy(text[offset:], what)
	return text
}

const preferred_horizontal_threshold = 5
const tabstop_length = 8

type EditBox struct {
	text              []byte
	output            chan []byte // TODO can this be merged with text?
	x                 int
	y                 int
	width             int
	height            int
	line_voffset      int
	cursor_boffset    int  // cursor offset in bytes
	cursor_voffset    int  // visual cursor offset in termbox cells
	cursor_coffset    int  // cursor offset in unicode code points
	received_kill_sig bool // if ESC is pressed, lets owner know to quit
}

// Draws the EditBox in the given location, 'h' is not used at the moment
func (eb *EditBox) Draw() {
	eb.AdjustVOffset(eb.width)

	const coldef = termbox.ColorDefault
	fill(eb.x, eb.y, eb.width, eb.height, termbox.Cell{Ch: ' '})

	t := eb.text
	lx := 0
	tabstop := 0
	for {
		rx := lx - eb.line_voffset
		if len(t) == 0 {
			break
		}

		if lx == tabstop {
			tabstop += tabstop_length
		}

		if rx >= eb.width {
			termbox.SetCell(eb.x+eb.width-1, eb.y, '→',
				coldef, coldef)
			break
		}

		r, size := utf8.DecodeRune(t)
		if r == '\t' {
			for ; lx < tabstop; lx++ {
				rx = lx - eb.line_voffset
				if rx >= eb.width {
					goto next
				}

				if rx >= 0 {
					termbox.SetCell(eb.x+rx, eb.y, ' ', coldef, coldef)
				}
			}
		} else {
			if rx >= 0 {
				termbox.SetCell(eb.x+rx, eb.y, r, coldef, coldef)
			}
			lx += runewidth.RuneWidth(r)
		}
	next:
		t = t[size:]
	}

	if eb.line_voffset != 0 {
		termbox.SetCell(eb.x, eb.y, '←', coldef, coldef)
	}
}

// Adjusts line visual offset to a proper value depending on width
func (eb *EditBox) AdjustVOffset(width int) {
	ht := preferred_horizontal_threshold
	max_h_threshold := (width - 1) / 2
	if ht > max_h_threshold {
		ht = max_h_threshold
	}

	threshold := width - 1
	if eb.line_voffset != 0 {
		threshold = width - ht
	}
	if eb.cursor_voffset-eb.line_voffset >= threshold {
		eb.line_voffset = eb.cursor_voffset + (ht - width + 1)
	}

	if eb.line_voffset != 0 && eb.cursor_voffset-eb.line_voffset < ht {
		eb.line_voffset = eb.cursor_voffset - ht
		if eb.line_voffset < 0 {
			eb.line_voffset = 0
		}
	}
}

func (eb *EditBox) MoveCursorTo(boffset int) {
	eb.cursor_boffset = boffset
	eb.cursor_voffset, eb.cursor_coffset = voffset_coffset(eb.text, boffset)
}

func (eb *EditBox) RuneUnderCursor() (rune, int) {
	return utf8.DecodeRune(eb.text[eb.cursor_boffset:])
}

func (eb *EditBox) RuneBeforeCursor() (rune, int) {
	return utf8.DecodeLastRune(eb.text[:eb.cursor_boffset])
}

func (eb *EditBox) MoveCursorOneRuneBackward() {
	if eb.cursor_boffset == 0 {
		return
	}
	_, size := eb.RuneBeforeCursor()
	eb.MoveCursorTo(eb.cursor_boffset - size)
}

func (eb *EditBox) MoveCursorOneRuneForward() {
	if eb.cursor_boffset == len(eb.text) {
		return
	}
	_, size := eb.RuneUnderCursor()
	eb.MoveCursorTo(eb.cursor_boffset + size)
}

func (eb *EditBox) MoveCursorToBeginningOfTheLine() {
	eb.MoveCursorTo(0)
}

func (eb *EditBox) MoveCursorToEndOfTheLine() {
	eb.MoveCursorTo(len(eb.text))
}

func (eb *EditBox) DeleteRuneBackward() {
	if eb.cursor_boffset == 0 {
		return
	}

	eb.MoveCursorOneRuneBackward()
	_, size := eb.RuneUnderCursor()
	eb.text = byte_slice_remove(eb.text, eb.cursor_boffset, eb.cursor_boffset+size)
}

func (eb *EditBox) DeleteRuneForward() {
	if eb.cursor_boffset == len(eb.text) {
		return
	}
	_, size := eb.RuneUnderCursor()
	eb.text = byte_slice_remove(eb.text, eb.cursor_boffset, eb.cursor_boffset+size)
}

func (eb *EditBox) DeleteTheRestOfTheLine() {
	eb.text = eb.text[:eb.cursor_boffset]
}

func (eb *EditBox) InsertRune(r rune) {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	eb.text = byte_slice_insert(eb.text, eb.cursor_boffset, buf[:n])
	eb.MoveCursorOneRuneForward()
}

// Please, keep in mind that cursor depends on the value of line_voffset, which
// is being set on Draw() call, so.. call this method after Draw() one.
func (eb *EditBox) CursorX() int {
	return eb.cursor_voffset - eb.line_voffset
}

func (eb *EditBox) SubmitLine() {
	eb.output <- eb.text
	eb.text = eb.text[:0]

}

func (eb *EditBox) GetLine() []byte {
	return <-eb.output
}

func (eb *EditBox) Kill() {
	eb.output <- nil
	eb.received_kill_sig = true
}

func (eb *EditBox) Alive() bool {
	return !eb.received_kill_sig
}

func (eb *EditBox) Listen() {
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				eb.Kill()
				break mainloop
			case termbox.KeyArrowLeft, termbox.KeyCtrlB:
				eb.MoveCursorOneRuneBackward()
			case termbox.KeyArrowRight, termbox.KeyCtrlF:
				eb.MoveCursorOneRuneForward()
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				eb.DeleteRuneBackward()
			case termbox.KeyDelete, termbox.KeyCtrlD:
				eb.DeleteRuneForward()
			case termbox.KeyTab:
				eb.InsertRune('\t')
			case termbox.KeySpace:
				eb.InsertRune(' ')
			case termbox.KeyCtrlK:
				eb.DeleteTheRestOfTheLine()
			case termbox.KeyHome, termbox.KeyCtrlA:
				eb.MoveCursorToBeginningOfTheLine()
			case termbox.KeyEnd, termbox.KeyCtrlE:
				eb.MoveCursorToEndOfTheLine()
			case termbox.KeyEnter:
				eb.SubmitLine()
				eb.MoveCursorToBeginningOfTheLine()
			default:
				if ev.Ch != 0 {
					eb.InsertRune(ev.Ch)
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
		eb.Draw()
		TermboxSafeFlush()
	}
}

// TODO allow multiple input boxes
var isInit bool

func NewEditBox(x, y, width, height int) *EditBox {
	termbox.SetInputMode(termbox.InputEsc)
	eb := EditBox{x: x, y: y, width: width, height: height, output: make(chan []byte)}
	// listen for input
	if !isInit {
		go eb.Listen()
	} else {
		panic("Can't initialize more than one EditBox!")
	}
	isInit = true
	return &eb
}
