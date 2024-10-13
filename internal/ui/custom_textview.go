package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ScrollableTextView struct {
	*tview.TextView
	App           *tview.Application
	ScrollBar     *ScrollBar
	updateHandler func()
}

func NewScrollableTextView(app *tview.Application, scrollBar *ScrollBar, updateHandler func()) *ScrollableTextView {
	return &ScrollableTextView{
		TextView:      tview.NewTextView(),
		App:           app,
		ScrollBar:     scrollBar,
		updateHandler: updateHandler,
	}
}

func (stv *ScrollableTextView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handler := stv.TextView.InputHandler()
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		oldOffset, _ := stv.GetScrollOffset()
		if handler != nil {
			handler(event, setFocus)
		}
		newOffset, _ := stv.GetScrollOffset()
		if oldOffset != newOffset && stv.updateHandler != nil {
			stv.updateHandler()
		}
	}
}

func (stv *ScrollableTextView) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	handler := stv.TextView.MouseHandler()
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
		oldOffset, _ := stv.GetScrollOffset()
		consumed, primitive := false, tview.Primitive(nil)
		if handler != nil {
			consumed, primitive = handler(action, event, setFocus)
		}
		newOffset, _ := stv.GetScrollOffset()
		if oldOffset != newOffset && stv.updateHandler != nil {
			stv.updateHandler()
		}
		return consumed, primitive
	}
}
