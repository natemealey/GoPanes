package gopanes

import (
	"testing"
)

// TODO exact output testing with mocks (or whatever those are in Go)
func TestSplit(t *testing.T) {
	ui := NewGoPaneUi()
	ui.rootPane.addLine("Test line!")
	ui.refresh()
}
