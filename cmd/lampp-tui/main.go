package main

import (
	"fmt"
	"os"
	"xampp-tui/internal/tui"

	tea "charm.land/bubbletea/v2"
)

func main() {
	p := tea.NewProgram(tui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
