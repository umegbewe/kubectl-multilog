package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	k8s "github.com/umegbewe/kubectl-multilog/internal/k8s"
)

func (t *LogExplorerTUI) createLogView() *tview.TextView {
	logView := tview.NewTextView()
	logView.SetDynamicColors(true)
	logView.SetRegions(true)
	logView.SetScrollable(true)
	logView.SetWordWrap(true)
	logView.SetBackgroundColor(colors.Background)
	logView.SetTitle("Logs")
	logView.SetTitleColor(colors.Accent)
	logView.SetBorder(true)
	logView.SetBorderColor(colors.TopBar)
	logView.SetBorderAttributes(tcell.AttrDim)
	return logView
}

func (t *LogExplorerTUI) loadLogs(namespace, pod, container string) {
	if t.logStreamCancel != nil {
		t.logStreamCancel()
	}

	t.showLoading(fmt.Sprintf("Loading logs for %s/%s/%s", namespace, pod, container))

	tail := int64(150)
	logs, logChan, err := t.k8sClient.GetLogs(namespace, pod, container, true, &tail)
	if err != nil {
		t.App.QueueUpdateDraw(func() {
			t.statusBar.SetText(fmt.Sprintf("Error fetching logs: %v", err))
		})
		return
	}

	t.App.QueueUpdateDraw(func() {
		t.logView.Clear()
		t.logView.SetText(logs)
		t.logView.ScrollToEnd()
		t.statusBar.SetText(fmt.Sprintf("logs loaded for %s/%s/%s", namespace, pod, container))
	})

	var ctx context.Context
	ctx, t.logStreamCancel = context.WithCancel(context.Background())

	for {
		select {
		case logEntry, ok := <-logChan:
			if !ok {
				return
			}
			t.App.QueueUpdateDraw(func() {
				fmt.Fprintf(t.logView, "%s", logEntry)
			})
		case <-ctx.Done():
			return
		}
	}
}

func (t *LogExplorerTUI) processLiveLogs() {
	for logEntry := range t.logChan {
		t.App.QueueUpdateDraw(func() {
			t.logMutex.Lock()
			defer t.logMutex.Unlock()

			if logEntry.Timestamp.After(t.liveTailStartTime) {
				formattedLog := fmt.Sprintf("[%s] [%s/%s/%s] %s: %s\n",
					logEntry.Timestamp.Format(time.RFC3339),
					logEntry.Namespace,
					logEntry.Pod,
					logEntry.Container,
					logEntry.Level,
					logEntry.Message)

				fmt.Fprintf(t.logView, "%s", formattedLog)
				t.logView.ScrollToEnd()
			}
		})
	}
}

func (t *LogExplorerTUI) clearLogView() {
    t.App.QueueUpdateDraw(func() {
        t.logView.Clear()
        t.logView.SetText("")
    })
}

func (t *LogExplorerTUI) toggleLiveTail() {
	if t.isLiveTailActive {
		t.stopLiveTail()
		t.liveTailBtn.SetLabel("Start Live Tail").SetBackgroundColor(colors.TopBar)
	} else {
		t.startLiveTail()
		t.liveTailBtn.SetLabel("Stop Live Tail").SetBackgroundColor(colors.Accent)
	}
}

func (t *LogExplorerTUI) startLiveTail() {
	t.isLiveTailActive = true
	t.logView.Clear()
	t.statusBar.SetText("Live tail active")

	t.liveTailStartTime = time.Now()
	t.liveTailCtx, t.liveTailCancel = context.WithCancel(context.Background())
	t.logChan = make(chan k8s.LogEntry, 100)

	go t.k8sClient.StreamAllLogs(t.liveTailCtx, t.logChan, t.liveTailStartTime)
	go t.processLiveLogs()
}

func (t *LogExplorerTUI) stopLiveTail() {
	t.isLiveTailActive = false
	t.liveTailCancel()
	close(t.logChan)
	t.statusBar.SetText("Live tail stopped")
}
