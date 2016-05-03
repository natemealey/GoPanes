package gopanes

import (
	"github.com/nsf/termbox-go"
)

type ColorStruct struct {
}

// TODO does this belong in gopanes?
// a termbox-go compatible colored strring
type ColorStr struct {
	str string
	fg  termbox.Attribute //TODO this should be better
	bg  termbox.Attribute //TODO this should be better
}

type ColorRune struct {
	r  rune
	fg termbox.Attribute //TODO this should be better
	bg termbox.Attribute //TODO this should be better
}

var Color ColorStruct

func (c *ColorStruct) Default(str string) ColorStr {
	return ColorStr{str: str, fg: termbox.ColorDefault, bg: termbox.ColorDefault}
}
func (c *ColorStruct) Black(str string) ColorStr {
	return ColorStr{str: str, fg: termbox.ColorBlack, bg: termbox.ColorDefault}
}
func (c *ColorStruct) Red(str string) ColorStr {
	return ColorStr{str: str, fg: termbox.ColorRed, bg: termbox.ColorDefault}
}
func (c *ColorStruct) Green(str string) ColorStr {
	return ColorStr{str: str, fg: termbox.ColorGreen, bg: termbox.ColorDefault}
}
func (c *ColorStruct) Yellow(str string) ColorStr {
	return ColorStr{str: str, fg: termbox.ColorYellow, bg: termbox.ColorDefault}
}
func (c *ColorStruct) Blue(str string) ColorStr {
	return ColorStr{str: str, fg: termbox.ColorBlue, bg: termbox.ColorDefault}
}
func (c *ColorStruct) Magenta(str string) ColorStr {
	return ColorStr{str: str, fg: termbox.ColorMagenta, bg: termbox.ColorDefault}
}
func (c *ColorStruct) Cyan(str string) ColorStr {
	return ColorStr{str: str, fg: termbox.ColorCyan, bg: termbox.ColorDefault}
}
func (c *ColorStruct) White(str string) ColorStr {
	return ColorStr{str: str, fg: termbox.ColorWhite, bg: termbox.ColorDefault}
}
func (c *ColorStruct) DarkGray(str string) ColorStr {
	return ColorStr{str: str, fg: termbox.ColorBlack | termbox.AttrBold, bg: termbox.ColorDefault}
}
