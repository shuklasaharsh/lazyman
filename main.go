package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Check for -S flag (search index feature - BETA)
	if len(os.Args) > 1 && os.Args[1] == "-S" {
		handleSearchIndex(os.Args[2:])
		return
	}

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

// handleSearchIndex handles the -S flag for indexing and searching
func handleSearchIndex(args []string) {
	betaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")).
		Bold(true)

	fmt.Println(betaStyle.Render("🧪 BETA FEATURE: Full-Text Search Index"))
	fmt.Println()

	if len(args) == 0 {
		// Build or rebuild the index
		if IndexExists() {
			fmt.Println("Refreshing existing search index...")
		} else {
			fmt.Println("Building search index for the first time...")
		}

		if err := IndexAllManPages(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		indexPath, _ := GetIndexPath()
		fmt.Printf("\n✓ Index stored at: %s\n", indexPath)
		fmt.Println("\nYou can now search with: lazyman -S <query>")
		return
	}

	// Search the index with TUI
	query := strings.Join(args, " ")

	if !IndexExists() {
		fmt.Println("Error: Search index not found.")
		fmt.Println("Run 'lazyman -S' first to build the index.")
		os.Exit(1)
	}

	// Perform search
	results, err := SearchIndexedManPages(query)
	if err != nil {
		fmt.Printf("Error performing search: %v\n", err)
		fmt.Printf("\nDebug info:\n")
		fmt.Printf("  Query: %s\n", query)
		fmt.Printf("  Index path: %s\n", indexPath)
		if absPath, err := GetIndexPath(); err == nil {
			fmt.Printf("  Absolute path: %s\n", absPath)
		}
		fmt.Printf("\nTry rebuilding the index with: lazyman -S\n")
		os.Exit(1)
	}

	// Convert search results to ManPages and store matches
	pages := make([]ManPage, 0, len(results))
	matchesMap := make(map[string][]string)

	for _, result := range results {
		pages = append(pages, result.ManPage)
		// Store matches for this page
		key := fmt.Sprintf("%s(%s)", result.ManPage.Name, result.ManPage.Section)
		matchesMap[key] = result.Matches

		// Debug: print stored matches
		if len(result.Matches) > 0 {
			fmt.Printf("DEBUG: Stored %d matches for %s\n", len(result.Matches), key)
		}
	}

	fmt.Printf("DEBUG: Total pages with matches: %d/%d\n", len(matchesMap), len(pages))

	// Launch existing TUI with search results
	model := InitialModel("")
	model.manPages = pages
	model.filteredPages = pages
	model.initialQuery = query
	model.searchInput.SetValue(query)
	model.searchResultMatches = matchesMap

	// If no results, show modal
	if len(pages) == 0 {
		model.noMatchSuggestions = model.findFuzzySuggestions(query, model.manPages)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running lazyman: %v\n", err)
		os.Exit(1)
	}
}
