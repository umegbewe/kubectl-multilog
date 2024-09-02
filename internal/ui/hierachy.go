package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

func (t *LogExplorerTUI) createMainArea() *tview.Flex {
	return tview.NewFlex().
		AddItem(t.hierarchy, 0, 1, true).
		AddItem(t.logView, 0, 5, false)
}

func (t *LogExplorerTUI) refreshHierarchy() {
	root := t.hierarchy.GetRoot()
	if root == nil {
		t.setStatus("Hierarchy is empty")
		return
	}

	root.ClearChildren()
	t.loadNamespaces(root)
}

func (t *LogExplorerTUI) loadNamespaces(root *tview.TreeNode) {
	namespaces, err := t.k8sClient.GetNamespaces()
	if err != nil {
		t.statusBar.SetText(fmt.Sprintf("Error fetching namespaces: %v", err))
		return
	}

	for _, ns := range namespaces {
		nsNode := createTreeNode(ns, false).SetReference(ns)
		setNodeWithToggleIcon(nsNode, ns, func() {
			nsNode.ClearChildren()
			go t.loadPods(nsNode)
		})
		root.AddChild(nsNode)
	}
}

func (t *LogExplorerTUI) loadPods(nsNode *tview.TreeNode) {
	namespace := nsNode.GetReference().(string)
	t.showLoading(fmt.Sprintf("Fetching pods for %s", namespace))
	pods, err := t.k8sClient.GetPods(namespace)
	if err != nil {
		t.App.QueueUpdateDraw(func() {
			t.statusBar.SetText(fmt.Sprintf("Error fetching pods for %s: %v", namespace, err))
		})
		return
	}

	t.App.QueueUpdateDraw(func() {
		nsNode.ClearChildren()
		for _, pod := range pods {
			podNode := createTreeNode(pod, false).SetReference(pod)
			setNodeWithToggleIcon(podNode, pod, func() {
				podNode.ClearChildren()
				// Use a closure to capture the correct pod name
				podName := podNode.GetReference().(string)
				go t.loadContainers(podNode, namespace, podName)
			})
			nsNode.AddChild(podNode)
		}
		t.statusBar.SetText(fmt.Sprintf("Loaded %d pods in namespace %s", len(pods), namespace))
	})
}

func (t *LogExplorerTUI) loadContainers(podNode *tview.TreeNode, namespace, pod string) {
	t.showLoading(fmt.Sprintf("Fetching containers for %s/%s", namespace, pod))
	containers, err := t.k8sClient.GetContainers(namespace, pod)
	if err != nil {
		t.App.QueueUpdateDraw(func() {
			t.statusBar.SetText(fmt.Sprintf("Error fetching containers for %s/%s: %v", namespace, pod, err))
		})
		return
	}

	t.App.QueueUpdateDraw(func() {
		podNode.ClearChildren() // Clear existing children to avoid duplicates
		for _, container := range containers {
			containerNode := tview.NewTreeNode(container).SetColor(colors.Text).SetReference(container)
			containerNode.SetSelectedFunc(func() {
				go t.loadLogs(namespace, pod, container)
			})
			podNode.AddChild(containerNode)
		}
		t.statusBar.SetText(fmt.Sprintf("Loaded %d containers for %s/%s", len(containers), namespace, pod))
	})
}
