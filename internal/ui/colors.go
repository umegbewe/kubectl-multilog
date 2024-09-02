package ui

import "github.com/gdamore/tcell/v2"

var colors = struct {
	Text       tcell.Color
	Background tcell.Color
	Sidebar    tcell.Color
	TopBar     tcell.Color
	Accent     tcell.Color
	Highlight  tcell.Color
}{
	Text:       tcell.NewHexColor(0x569cdb),
	Background: tcell.ColorDefault,
	Sidebar:    tcell.NewRGBColor(30, 30, 30),
	TopBar:     tcell.NewHexColor(0x262626),
	Accent:     tcell.NewHexColor(0xffd700),
	Highlight:  tcell.NewHexColor(0x32cd32),
}
