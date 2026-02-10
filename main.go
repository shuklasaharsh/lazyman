package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Check for command-line arguments
	var initialQuery string
	if len(os.Args) > 1 {
		initialQuery = strings.Join(os.Args[1:], " ")
	}

	p := tea.NewProgram(InitialModel(initialQuery), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running lazyman: %v\n", err)
		os.Exit(1)
	}
}
