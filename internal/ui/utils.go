package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (t *App) setStatus(message string) {
	t.statusBar.SetText(message)
}

func (t *App) setStatusError(message string) {
	t.statusBar.SetText("[red]" + message + "[-]")
}

func (t *App) showLoading(message string) {
	t.App.QueueUpdateDraw(func() {
		t.statusBar.SetText(message + " Loading...")
	})

	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(500 * time.Millisecond)
			t.App.QueueUpdateDraw(func() {
				currentText := t.statusBar.GetText(false)
				t.statusBar.SetText(currentText + ".")
			})
		}
	}()
}
func createTreeNode(label string, isLeaf bool) *tview.TreeNode {
	node := tview.NewTreeNode("")

	if isLeaf {
		node.SetText(fmt.Sprintf("  %s", label))
	} else {
		node.SetText(fmt.Sprintf("▶ %s", label))
		node.SetExpanded(false)
	}

	return node
}

func setNodeWithToggleIcon(node *tview.TreeNode, label string, toggleFunc func()) {
	node.SetSelectedFunc(func() {
		if node.IsExpanded() {
			node.CollapseAll()
			node.SetText(fmt.Sprintf("▶ %s", label))
		} else {
			node.ExpandAll()
			node.SetText(fmt.Sprintf("▼ %s", label))
		}
		toggleFunc()
	})
}

func createButton(label string, bgColor tcell.Color, selectedFunc func()) *tview.Button {
	return tview.NewButton(label).
		SetLabelColor(colors.Text).
		SetStyle(tcell.StyleDefault.Background(bgColor)).
		SetSelectedFunc(selectedFunc)
}

func (t *App) updateScrollBar() {
	content := t.logTextView.GetText(true)

	lines := strings.Count(content, "\n")
	totalLines := lines
	if content != "" && content[len(content)-1] != '\n' {
		totalLines++
	}

	_, _, _, height := t.logTextView.GetInnerRect()
	firstVisibleLine, _ := t.logTextView.GetScrollOffset()

	t.scrollBar.SetTotalLines(totalLines)
	t.scrollBar.SetVisibleLines(height)
	t.scrollBar.SetCurrentLine(firstVisibleLine)
}
