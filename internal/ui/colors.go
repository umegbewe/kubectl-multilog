package ui

import "github.com/gdamore/tcell/v2"

var colors = struct {
	Text       tcell.Color
	Background tcell.Color
	Sidebar    tcell.Color
	TopBar     tcell.Color
	Accent     tcell.Color
	Highlight  tcell.Color
	Button     tcell.Color
	NavButton  tcell.Color
}{

	Text:       tcell.NewHexColor(0xE0E0E0),
	Background: tcell.NewHexColor(0x1E1E1E),
	Sidebar:    tcell.NewHexColor(0x252526),
	TopBar:     tcell.NewHexColor(0x2D2D30),
	Accent:     tcell.NewHexColor(0x569CD6),
	Highlight:  tcell.NewHexColor(0x3C3C3C),
	Button:     tcell.NewHexColor(0x569CD6),
	NavButton:  tcell.ColorDarkGreen,
}
