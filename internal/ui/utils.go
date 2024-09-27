package ui

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

func (t *LogExplorerTUI) setStatus(message string) {
	t.statusBar.SetText(message)
}

func (t *LogExplorerTUI) setStatusError(message string) {
	t.statusBar.SetText("[red]" + message + "[-]")
}

func (t *LogExplorerTUI) showLoading(message string) {
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
