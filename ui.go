package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View modes
type viewMode int

const (
	listView viewMode = iota
	detailView
	searchView
	detailSearchView
)

// SectionFilter represents manual section filters
type SectionFilter struct {
	Section string
	Name    string
	Enabled bool
}

// Model represents the application state
type Model struct {
	mode               viewMode
	manPages           []ManPage
	filteredPages      []ManPage
	cursor             int
	viewport           viewport.Model
	previewPort        viewport.Model
	searchInput        textinput.Model
	detailSearchInput  textinput.Model
	currentContent     string
	previewContent     string
	searchQuery        string
	searchMatches      []int // line numbers with matches
	currentMatch       int   // index in searchMatches
	sectionFilters     []SectionFilter
	initialQuery       string
	noMatchSuggestions []ManPage
	width              int
	height             int
	err                error
	loading            bool
	loadingPreview     bool
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true).
				PaddingLeft(2)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0, 0, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

// InitialModel creates the initial model
func InitialModel(initialQuery string) Model {
	ti := textinput.New()
	ti.Placeholder = "Search man pages..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	dsi := textinput.New()
	dsi.Placeholder = "Search in document..."
	dsi.CharLimit = 156
	dsi.Width = 50

	vp := viewport.New(80, 20)
	pp := viewport.New(40, 20)

	// Initialize section filters (all enabled by default)
	filters := []SectionFilter{
		{Section: "1", Name: "General Commands", Enabled: true},
		{Section: "2", Name: "System Calls", Enabled: true},
		{Section: "3", Name: "Library Functions", Enabled: true},
		{Section: "4", Name: "Kernel Interfaces", Enabled: true},
		{Section: "5", Name: "File Formats", Enabled: true},
		{Section: "6", Name: "Games", Enabled: true},
		{Section: "7", Name: "Miscellaneous", Enabled: true},
		{Section: "8", Name: "System Manager's", Enabled: true},
		{Section: "9", Name: "Kernel Developer's", Enabled: true},
	}

	return Model{
		mode:              listView,
		manPages:          []ManPage{},
		filteredPages:     []ManPage{},
		cursor:            0,
		viewport:          vp,
		previewPort:       pp,
		searchInput:       ti,
		detailSearchInput: dsi,
		sectionFilters:    filters,
		initialQuery:      initialQuery,
		loading:           true,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	if m.initialQuery != "" {
		return tea.Batch(
			tea.EnterAltScreen,
			searchManPages(m.initialQuery),
		)
	}
	return tea.Batch(
		tea.EnterAltScreen,
		loadManPages,
	)
}

// Messages
type manPagesLoadedMsg struct {
	pages []ManPage
}

type manContentLoadedMsg struct {
	content string
}

type previewLoadedMsg struct {
	content string
}

type errMsg struct {
	err error
}

// Commands
func loadManPages() tea.Msg {
	pages, err := SearchManPages(".")
	if err != nil {
		return errMsg{err}
	}
	return manPagesLoadedMsg{pages: pages}
}

func loadManContent(name, section string) tea.Cmd {
	return func() tea.Msg {
		content, err := GetManContent(name, section)
		if err != nil {
			return errMsg{err}
		}
		return manContentLoadedMsg{content: content}
	}
}

func searchManPages(query string) tea.Cmd {
	return func() tea.Msg {
		pages, err := SearchManPages(query)
		if err != nil {
			return errMsg{err}
		}
		return manPagesLoadedMsg{pages: pages}
	}
}

func loadPreview(name, section string) tea.Cmd {
	return func() tea.Msg {
		content, err := GetManContent(name, section)
		if err != nil {
			return previewLoadedMsg{content: fmt.Sprintf("Error loading preview: %v", err)}
		}
		return previewLoadedMsg{content: content}
	}
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5
		// Split view: left 60%, right 40%
		listWidth := int(float64(msg.Width) * 0.6)
		previewWidth := msg.Width - listWidth - 2 // -2 for border
		m.previewPort.Width = previewWidth
		m.previewPort.Height = msg.Height - 5

	case manPagesLoadedMsg:
		m.manPages = msg.pages
		m.filteredPages = m.applyFilters(msg.pages)
		m.loading = false
		m.cursor = 0

		// Handle initial query behavior
		if m.initialQuery != "" {
			if len(m.filteredPages) == 0 {
				// No matches - find fuzzy suggestions
				m.noMatchSuggestions = m.findFuzzySuggestions(m.initialQuery, m.manPages)
				// Load preview for first suggestion
				if len(m.noMatchSuggestions) > 0 {
					page := m.noMatchSuggestions[0]
					cmds = append(cmds, loadPreview(page.Name, page.Section))
				}
			} else if len(m.filteredPages) == 1 {
				// Single match - auto-open
				page := m.filteredPages[0]
				m.initialQuery = "" // Clear so we don't re-trigger
				return m, loadManContent(page.Name, page.Section)
			}
			// Multiple matches - show list (normal behavior)
			m.initialQuery = "" // Clear so we don't re-trigger
		}

		// Load preview for first item
		if len(m.filteredPages) > 0 {
			page := m.filteredPages[0]
			cmds = append(cmds, loadPreview(page.Name, page.Section))
		}

	case manContentLoadedMsg:
		m.currentContent = msg.content
		m.viewport.SetContent(msg.content)
		m.mode = detailView
		m.viewport.GotoTop()

	case previewLoadedMsg:
		m.previewContent = msg.content
		m.previewPort.SetContent(msg.content)
		m.previewPort.GotoTop()
		m.loadingPreview = false

	case errMsg:
		m.err = msg.err
		m.loading = false

	case tea.KeyMsg:
		switch m.mode {
		case listView:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit

			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					// Load preview for new cursor position
					pages := m.filteredPages
					if len(pages) == 0 && len(m.noMatchSuggestions) > 0 {
						pages = m.noMatchSuggestions
					}
					if len(pages) > 0 {
						page := pages[m.cursor]
						m.loadingPreview = true
						cmds = append(cmds, loadPreview(page.Name, page.Section))
					}
				}

			case "down", "j":
				maxLen := len(m.filteredPages)
				if len(m.filteredPages) == 0 && len(m.noMatchSuggestions) > 0 {
					maxLen = len(m.noMatchSuggestions)
				}
				if m.cursor < maxLen-1 {
					m.cursor++
					// Load preview for new cursor position
					pages := m.filteredPages
					if len(pages) == 0 && len(m.noMatchSuggestions) > 0 {
						pages = m.noMatchSuggestions
					}
					if len(pages) > 0 {
						page := pages[m.cursor]
						m.loadingPreview = true
						cmds = append(cmds, loadPreview(page.Name, page.Section))
					}
				}

			case "enter":
				pages := m.filteredPages
				if len(pages) == 0 && len(m.noMatchSuggestions) > 0 {
					pages = m.noMatchSuggestions
				}
				if len(pages) > 0 && m.cursor < len(pages) {
					page := pages[m.cursor]
					return m, loadManContent(page.Name, page.Section)
				}

			case "/":
				m.mode = searchView
				m.searchInput.SetValue("")
				return m, textinput.Blink

			case "r":
				m.loading = true
				return m, loadManPages

			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				// Toggle filter for this section
				section := msg.String()
				for i := range m.sectionFilters {
					if m.sectionFilters[i].Section == section {
						m.sectionFilters[i].Enabled = !m.sectionFilters[i].Enabled
						break
					}
				}
				// Reapply filters
				m.filteredPages = m.applyFilters(m.manPages)
				if m.cursor >= len(m.filteredPages) {
					m.cursor = len(m.filteredPages) - 1
				}
				if m.cursor < 0 {
					m.cursor = 0
				}
				// Load preview for current cursor position
				if len(m.filteredPages) > 0 {
					page := m.filteredPages[m.cursor]
					m.loadingPreview = true
					cmds = append(cmds, loadPreview(page.Name, page.Section))
				}
			}

		case detailView:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				m.mode = listView
				m.currentContent = ""
				m.searchQuery = ""
				m.searchMatches = nil

			case "up", "k":
				m.viewport.LineUp(1)

			case "down", "j":
				m.viewport.LineDown(1)

			case "g":
				m.viewport.GotoTop()

			case "G":
				m.viewport.GotoBottom()

			case "u":
				m.viewport.HalfViewUp()

			case "d":
				m.viewport.HalfViewDown()

			case "/":
				m.mode = detailSearchView
				m.detailSearchInput.SetValue("")
				m.detailSearchInput.Focus()
				return m, textinput.Blink

			case "n":
				// Next match
				if len(m.searchMatches) > 0 {
					m.currentMatch = (m.currentMatch + 1) % len(m.searchMatches)
					m.viewport.SetYOffset(m.searchMatches[m.currentMatch])
				}

			case "N":
				// Previous match
				if len(m.searchMatches) > 0 {
					m.currentMatch--
					if m.currentMatch < 0 {
						m.currentMatch = len(m.searchMatches) - 1
					}
					m.viewport.SetYOffset(m.searchMatches[m.currentMatch])
				}
			}
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)

		case searchView:
			switch msg.String() {
			case "esc":
				m.mode = listView

			case "enter":
				query := m.searchInput.Value()
				if query != "" {
					m.mode = listView
					m.loading = true
					return m, searchManPages(query)
				}
				m.mode = listView

			default:
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)
			}

		case detailSearchView:
			switch msg.String() {
			case "esc":
				m.mode = detailView
				m.detailSearchInput.Blur()

			case "enter":
				query := m.detailSearchInput.Value()
				if query != "" {
					m.searchQuery = query
					m.searchMatches = m.findMatches(query)
					m.currentMatch = 0
					if len(m.searchMatches) > 0 {
						m.viewport.SetYOffset(m.searchMatches[0])
					}
				}
				m.mode = detailView
				m.detailSearchInput.Blur()

			default:
				m.detailSearchInput, cmd = m.detailSearchInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// applyFilters filters manual pages based on enabled section filters
func (m Model) applyFilters(pages []ManPage) []ManPage {
	filtered := []ManPage{}
	for _, page := range pages {
		// Check if this section is enabled
		for _, filter := range m.sectionFilters {
			if filter.Section == page.Section && filter.Enabled {
				filtered = append(filtered, page)
				break
			}
		}
	}
	return filtered
}

// levenshteinDistance calculates edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	s1Lower := strings.ToLower(s1)
	s2Lower := strings.ToLower(s2)

	if len(s1Lower) == 0 {
		return len(s2Lower)
	}
	if len(s2Lower) == 0 {
		return len(s1Lower)
	}

	// Create matrix
	matrix := make([][]int, len(s1Lower)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2Lower)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1Lower); i++ {
		for j := 1; j <= len(s2Lower); j++ {
			cost := 0
			if s1Lower[i-1] != s2Lower[j-1] {
				cost = 1
			}

			min := matrix[i-1][j] + 1 // deletion
			if matrix[i][j-1]+1 < min {
				min = matrix[i][j-1] + 1 // insertion
			}
			if matrix[i-1][j-1]+cost < min {
				min = matrix[i-1][j-1] + cost // substitution
			}

			matrix[i][j] = min
		}
	}

	return matrix[len(s1Lower)][len(s2Lower)]
}

// findFuzzySuggestions finds similar manual pages using fuzzy matching
func (m Model) findFuzzySuggestions(query string, allPages []ManPage) []ManPage {
	type scoredPage struct {
		page  ManPage
		score int
	}

	var scored []scoredPage
	queryLower := strings.ToLower(query)

	for _, page := range allPages {
		pageLower := strings.ToLower(page.Name)

		// Calculate similarity score (lower is better)
		distance := levenshteinDistance(query, page.Name)

		// Also check if query is a substring (boost score)
		if strings.Contains(pageLower, queryLower) {
			distance -= 5 // Bonus for substring match
		}

		// Only include reasonable matches
		if distance <= 5 {
			scored = append(scored, scoredPage{page: page, score: distance})
		}
	}

	// Sort by score (best matches first)
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score < scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Return top 10 suggestions
	result := []ManPage{}
	max := 10
	if len(scored) < max {
		max = len(scored)
	}
	for i := 0; i < max; i++ {
		result = append(result, scored[i].page)
	}

	return result
}

// findMatches searches for query in current content and returns line numbers
func (m Model) findMatches(query string) []int {
	if query == "" || m.currentContent == "" {
		return nil
	}

	lines := strings.Split(m.currentContent, "\n")
	matches := []int{}
	searchLower := strings.ToLower(query)

	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), searchLower) {
			matches = append(matches, i)
		}
	}

	return matches
}

// View renders the UI
func (m Model) View() string {
	if m.loading {
		return "\n  Loading man pages...\n\n"
	}

	switch m.mode {
	case listView:
		return m.renderListView()
	case detailView:
		return m.renderDetailView()
	case searchView:
		return m.renderSearchView()
	case detailSearchView:
		return m.renderDetailSearchView()
	default:
		return ""
	}
}

func (m Model) renderFilterBar() string {
	var b strings.Builder

	enabledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	disabledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Strikethrough(true)

	b.WriteString("  Sections: ")

	for i, filter := range m.sectionFilters {
		if i > 0 {
			b.WriteString(" ")
		}

		label := fmt.Sprintf("[%s]%s", filter.Section, filter.Name)

		if filter.Enabled {
			b.WriteString(enabledStyle.Render(label))
		} else {
			b.WriteString(disabledStyle.Render(label))
		}
	}

	return b.String()
}

func (m Model) renderListView() string {
	// Calculate widths for split view
	listWidth := int(float64(m.width) * 0.6)
	if listWidth == 0 {
		listWidth = 80 // default
	}

	// Build left panel (list)
	var leftPanel strings.Builder

	// Title
	title := titleStyle.Render(" LazyMan - Manual Pages ")
	leftPanel.WriteString(title)
	leftPanel.WriteString("\n\n")

	// Filter bar
	filterBar := m.renderFilterBar()
	leftPanel.WriteString(filterBar)
	leftPanel.WriteString("\n")

	// Error display
	if m.err != nil {
		leftPanel.WriteString(errorStyle.Render(fmt.Sprintf("  Error: %v\n\n", m.err)))
	}

	// Status or no matches message
	if len(m.filteredPages) == 0 && len(m.noMatchSuggestions) > 0 {
		noMatchStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true)
		leftPanel.WriteString(noMatchStyle.Render("  No exact matches found.\n"))
		leftPanel.WriteString(statusStyle.Render("  Did you mean:\n\n"))

		// Show suggestions
		for i, page := range m.noMatchSuggestions {
			line := fmt.Sprintf("%s(%s)", page.Name, page.Section)
			if page.Description != "" {
				line += " - " + page.Description
			}

			// Truncate line if too long for left panel
			if len(line) > listWidth-6 {
				line = line[:listWidth-9] + "..."
			}

			if i == m.cursor && m.cursor < len(m.noMatchSuggestions) {
				leftPanel.WriteString(selectedItemStyle.Render("▸ " + line))
			} else {
				leftPanel.WriteString(itemStyle.Render(line))
			}
			leftPanel.WriteString("\n")
		}
	} else {
		status := statusStyle.Render(fmt.Sprintf("  Showing %d man pages", len(m.filteredPages)))
		leftPanel.WriteString(status)
		leftPanel.WriteString("\n\n")

		// Man pages list
		start := m.cursor - 10
		if start < 0 {
			start = 0
		}
		end := start + 20
		if end > len(m.filteredPages) {
			end = len(m.filteredPages)
		}

		for i := start; i < end; i++ {
			page := m.filteredPages[i]
			line := fmt.Sprintf("%s(%s)", page.Name, page.Section)
			if page.Description != "" {
				line += " - " + page.Description
			}

			// Truncate line if too long for left panel
			if len(line) > listWidth-6 {
				line = line[:listWidth-9] + "..."
			}

			if i == m.cursor {
				leftPanel.WriteString(selectedItemStyle.Render("▸ " + line))
			} else {
				leftPanel.WriteString(itemStyle.Render(line))
			}
			leftPanel.WriteString("\n")
		}
	}

	// Help
	leftPanel.WriteString("\n")
	help := helpStyle.Render(
		"↑/k up • ↓/j down • enter view • / search • 1-9 toggle filter • r refresh • q quit",
	)
	leftPanel.WriteString(help)

	// Build right panel (preview)
	var rightPanel strings.Builder
	previewTitle := titleStyle.Render(" Preview ")
	rightPanel.WriteString(previewTitle)
	rightPanel.WriteString("\n\n")

	if m.loadingPreview {
		rightPanel.WriteString("  Loading preview...")
	} else if m.previewContent != "" {
		rightPanel.WriteString(m.previewPort.View())
	} else {
		rightPanel.WriteString("  No preview available")
	}

	// Combine left and right panels
	leftLines := strings.Split(leftPanel.String(), "\n")
	rightLines := strings.Split(rightPanel.String(), "\n")

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	var result strings.Builder
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	for i := 0; i < maxLines; i++ {
		// Left side
		if i < len(leftLines) {
			line := leftLines[i]
			// Pad or truncate to listWidth
			visualLen := lipgloss.Width(line)
			if visualLen < listWidth {
				line += strings.Repeat(" ", listWidth-visualLen)
			} else if visualLen > listWidth {
				line = line[:listWidth]
			}
			result.WriteString(line)
		} else {
			result.WriteString(strings.Repeat(" ", listWidth))
		}

		// Border
		result.WriteString(borderStyle.Render(" │ "))

		// Right side
		if i < len(rightLines) {
			result.WriteString(rightLines[i])
		}

		result.WriteString("\n")
	}

	return result.String()
}

func (m Model) renderDetailView() string {
	var b strings.Builder

	// Title
	if len(m.filteredPages) > 0 && m.cursor < len(m.filteredPages) {
		page := m.filteredPages[m.cursor]
		title := titleStyle.Render(fmt.Sprintf(" %s(%s) ", page.Name, page.Section))
		b.WriteString(title)

		// Show search info if active
		if m.searchQuery != "" {
			searchInfo := statusStyle.Render(fmt.Sprintf("  [Search: %s - Match %d/%d]",
				m.searchQuery, m.currentMatch+1, len(m.searchMatches)))
			b.WriteString(searchInfo)
		}
		b.WriteString("\n\n")
	}

	// Content viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Help
	var helpText string
	if m.searchQuery != "" {
		helpText = "↑/k up • ↓/j down • n next match • N prev match • / search • q/esc back"
	} else {
		helpText = "↑/k up • ↓/j down • g top • G bottom • u/d half page • / search • q/esc back"
	}
	help := helpStyle.Render(helpText)
	b.WriteString(help)

	return b.String()
}

func (m Model) renderSearchView() string {
	var b strings.Builder

	title := titleStyle.Render(" Search Man Pages ")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString("  ")
	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")

	help := helpStyle.Render("enter search • esc cancel")
	b.WriteString(help)

	return b.String()
}

func (m Model) renderDetailSearchView() string {
	var b strings.Builder

	title := titleStyle.Render(" Search in Document ")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString("  ")
	b.WriteString(m.detailSearchInput.View())
	b.WriteString("\n\n")

	help := helpStyle.Render("enter search • esc cancel")
	b.WriteString(help)

	return b.String()
}
