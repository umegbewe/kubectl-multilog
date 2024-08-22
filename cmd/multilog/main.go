package main

import (
	"log"
	"github.com/umegbewe/kubectl-multilog/internal/ui"
)

func main() {
	tui, err := ui.NewLogExplorerTUI()
	if err != nil {
		log.Fatalf("Failed to create TUI: %v", err)
	}

	if err := tui.Run(); err != nil {
		log.Fatal(err)
	}
}
