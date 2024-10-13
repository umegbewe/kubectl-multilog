package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ScrollBar struct {
	*tview.Box

	totalLines     int
	visibleLines   int
	currentLine    int
	dragging       bool
	dragStartY     int
	dragStartLine  int
	scrollCallback func(line int)
}

func NewScrollBar() *ScrollBar {
	return &ScrollBar{
		Box: tview.NewBox(),
	}
}

func (s *ScrollBar) Draw(screen tcell.Screen) {

	if s.totalLines <= s.visibleLines {
		return
	}

	s.Box.DrawForSubclass(screen, s)
	x, y, width, height := s.GetInnerRect()

	thumbHeight := s.calculateThumbHeight(height)
	thumbY := s.calculateThumbPosition(y, height, thumbHeight)

	for i := y; i < y+height; i++ {
		screen.SetContent(x+width-1, i, ' ', nil, tcell.StyleDefault.Background(colors.Sidebar))
	}

	for i := thumbY; i < thumbY+thumbHeight && i < y+height; i++ {
		screen.SetContent(x+width-1, i, '█', nil, tcell.StyleDefault.Background(colors.Highlight))
	}
}

func (s *ScrollBar) calculateThumbHeight(height int) int {
	thumbHeight := s.visibleLines * height / s.totalLines
	if thumbHeight < 1 {
		thumbHeight = 1
	}
	if thumbHeight > height {
		thumbHeight = height
	}
	return thumbHeight
}

func (s *ScrollBar) calculateThumbPosition(y, height, thumbHeight int) int {
	maxScroll := s.totalLines - s.visibleLines
	if maxScroll <= 0 {
		return y
	}
	thumbY := y + (s.currentLine * (height - thumbHeight) / maxScroll)
	if thumbY < y {
		thumbY = y
	}
	if thumbY > y+height-thumbHeight {
		thumbY = y + height - thumbHeight
	}
	return thumbY
}

func (s *ScrollBar) SetTotalLines(total int) {
	s.totalLines = total
}

func (s *ScrollBar) SetVisibleLines(visible int) {
	s.visibleLines = visible
}

func (s *ScrollBar) SetCurrentLine(current int) {
	s.currentLine = current
}

func (s *ScrollBar) SetScrollCallback(callback func(line int)) {
	s.scrollCallback = callback
}

func (s *ScrollBar) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
		x, y := event.Position()
		if !s.InRect(x, y) {
			return false, nil
		}

		switch action {
		case tview.MouseLeftDown:
			setFocus(s)
			s.dragging = true
			s.dragStartY = y
			s.dragStartLine = s.currentLine
			return true, s
		case tview.MouseMove:
			if s.dragging {
				_, _, _, height := s.GetInnerRect()

				deltaY := y - s.dragStartY
				maxScroll := s.totalLines - s.visibleLines
				if maxScroll > 0 {
					scrollRatio := float64(deltaY) / float64(height)
					scrollLines := int(scrollRatio * float64(maxScroll))
					newLine := s.dragStartLine + scrollLines

					if newLine < 0 {
						newLine = 0
					} else if newLine > maxScroll {
						newLine = maxScroll
					}

					if newLine != s.currentLine {
						s.currentLine = newLine
						if s.scrollCallback != nil {
							s.scrollCallback(s.currentLine)
						}
					}
				}
				return true, s
			}
		case tview.MouseLeftUp:
			s.dragging = false
			return true, s
		case tview.MouseScrollUp:
			if s.currentLine > 0 {
				s.currentLine--
				if s.scrollCallback != nil {
					s.scrollCallback(s.currentLine)
				}
				return true, s
			}
		case tview.MouseScrollDown:
			maxScroll := s.totalLines - s.visibleLines
			if s.currentLine < maxScroll {
				s.currentLine++
				if s.scrollCallback != nil {
					s.scrollCallback(s.currentLine)
				}
				return true, s
			}
		}

		return false, nil
	}
}
