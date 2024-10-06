package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	k8s "github.com/umegbewe/kubectl-multilog/internal/k8sclient"
	"github.com/umegbewe/kubectl-multilog/internal/search"
)

type Model struct {
	K8sClient         *k8s.Client
	LogBuffer         *search.LogBuffer
	LiveTailActive    bool
	LiveTailCtx       context.Context
	LiveTailCancel    context.CancelFunc
	LogChan           chan k8s.LogEntry
	LogMutex          sync.Mutex
	LiveTailStartTime time.Time
	LogStreamCancel   context.CancelFunc
	SearchResult      *search.SearchResult
	CurrentMatchIndex int
	SearchOptions     search.SearchOptions
}

func NewModel(k8sClient *k8s.Client) *Model {
	return &Model{
		K8sClient: k8sClient,
		LogBuffer: search.NewLogBuffer(),
	}
}

func (m *Model) GetClusterNames() []string {
	return m.K8sClient.GetClusterNames()
}

func (m *Model) GetCurrentContext() string {
	return m.K8sClient.GetCurrentContext()
}

func (m *Model) GetNamespaces() ([]string, error) {
	return m.K8sClient.GetNamespaces()
}

func (m *Model) GetPods(namespace string) ([]string, error) {
	return m.K8sClient.GetPods(namespace)
}

func (m *Model) GetContainers(namespace, pod string) ([]string, error) {
	return m.K8sClient.GetContainers(namespace, pod)
}

func (m *Model) SwitchCluster(contextName string) error {
	err := m.K8sClient.SwitchCluster(contextName)
	if err != nil {
		return fmt.Errorf("failed to switch cluster: %w", err)
	}

	return nil
}

func (m *Model) PerformSearch(term string, options search.SearchOptions) (*search.SearchResult, error) {
	lines := m.LogBuffer.GetLinesContent()
	return search.PerformSearch(lines, term, options)
}
