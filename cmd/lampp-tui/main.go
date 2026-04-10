package main

import (
	"fmt"
	"os"
	"xampp-tui/internal/tui"

	tea "charm.land/bubbletea/v2"
)

// version is set at build time via -ldflags="-X main.version=<tag>".
var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("xampp-tui", version)
		return
	}

	p := tea.NewProgram(tui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
