package main

import (
	"log"

	k8s "github.com/umegbewe/kubectl-multilog/internal/k8sclient"
	"github.com/umegbewe/kubectl-multilog/internal/model"
	"github.com/umegbewe/kubectl-multilog/internal/ui"
)

func main() {
	k8sClient, err := k8s.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	model := model.NewModel(k8sClient)

	tui := ui.LogExplorerTUI(model)

	if err := tui.Run(); err != nil {
		log.Fatal(err)
	}
}
