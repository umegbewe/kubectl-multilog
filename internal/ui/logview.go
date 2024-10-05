package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	k8s "github.com/umegbewe/kubectl-multilog/internal/k8sclient"
)

func (t *App) initLogView() *tview.TextView {
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

func (t *App) loadLogs(namespace, pod, container string) {
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
		t.logBuffer.Clear()
		for _, line := range strings.Split(logs, "\n") {
			if line != "" {
				t.processNewLogEntry(line)
			}
		}
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
			t.processNewLogEntry(logEntry)
		case <-ctx.Done():
			return
		}
	}
}

func (t *App) processNewLogEntry(logEntry string) {
    t.logBuffer.AddLine(logEntry)

    if t.searchResult != nil && t.searchResult.Term != "" {
        t.updateSearchForNewLogs()
    } else {
		fmt.Fprintf(t.logView, "%s\n", logEntry)
	}
}

func (t *App) processLiveLogs() {
	for logEntry := range t.logChan {
		t.logMutex.Lock()
		if logEntry.Timestamp.After(t.liveTailStartTime) {
			formattedLog := fmt.Sprintf("[%s] [%s/%s/%s] %s: %s\n",
				logEntry.Timestamp.Format(time.RFC3339),
				logEntry.Namespace,
				logEntry.Pod,
				logEntry.Container,
				logEntry.Level,
				logEntry.Message)
			t.processNewLogEntry(formattedLog)
		}
		t.logMutex.Unlock()
	}
}

func (t *App) clearLogView() {
	t.App.QueueUpdateDraw(func() {
		t.logView.Clear()
		t.logView.SetText("")
	})
}

func (t *App) toggleLiveTail() {
	if t.LiveTailActive {
		t.stopLiveTail()
		t.liveTailBtn.SetLabel("Start Live Tail").SetBackgroundColor(colors.TopBar)
	} else {
		t.startLiveTail()
		t.liveTailBtn.SetLabel("Stop Live Tail").SetBackgroundColor(colors.Accent)
	}
}

func (t *App) startLiveTail() {
	t.LiveTailActive = true
	t.logView.Clear()
	t.statusBar.SetText("Live tail active")

	t.liveTailStartTime = time.Now()
	t.liveTailCtx, t.liveTailCancel = context.WithCancel(context.Background())
	t.logChan = make(chan k8s.LogEntry, 100)

	go t.k8sClient.StreamAllLogs(t.liveTailCtx, t.logChan, t.liveTailStartTime)
	go t.processLiveLogs()
}

func (t *App) stopLiveTail() {
	t.LiveTailActive = false
	t.liveTailCancel()
	close(t.logChan)
	t.statusBar.SetText("Live tail stopped")
}
