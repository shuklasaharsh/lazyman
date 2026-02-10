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
)

// Model represents the application state
type Model struct {
	mode           viewMode
	manPages       []ManPage
	filteredPages  []ManPage
	cursor         int
	viewport       viewport.Model
	searchInput    textinput.Model
	currentContent string
	width          int
	height         int
	err            error
	loading        bool
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
func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Search man pages..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	vp := viewport.New(80, 20)

	return Model{
		mode:          listView,
		manPages:      []ManPage{},
		filteredPages: []ManPage{},
		cursor:        0,
		viewport:      vp,
		searchInput:   ti,
		loading:       true,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
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

	case manPagesLoadedMsg:
		m.manPages = msg.pages
		m.filteredPages = msg.pages
		m.loading = false
		m.cursor = 0

	case manContentLoadedMsg:
		m.currentContent = msg.content
		m.viewport.SetContent(msg.content)
		m.mode = detailView
		m.viewport.GotoTop()

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
				}

			case "down", "j":
				if m.cursor < len(m.filteredPages)-1 {
					m.cursor++
				}

			case "enter":
				if len(m.filteredPages) > 0 {
					page := m.filteredPages[m.cursor]
					return m, loadManContent(page.Name, page.Section)
				}

			case "/":
				m.mode = searchView
				m.searchInput.SetValue("")
				return m, textinput.Blink

			case "r":
				m.loading = true
				return m, loadManPages
			}

		case detailView:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				m.mode = listView
				m.currentContent = ""

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
		}
	}

	return m, tea.Batch(cmds...)
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
	default:
		return ""
	}
}

func (m Model) renderListView() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render(" LazyMan - Manual Pages ")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Error display
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Error: %v\n\n", m.err)))
	}

	// Status
	status := statusStyle.Render(fmt.Sprintf("  Showing %d man pages", len(m.filteredPages)))
	b.WriteString(status)
	b.WriteString("\n\n")

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

		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render("▸ " + line))
		} else {
			b.WriteString(itemStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	help := helpStyle.Render(
		"↑/k up • ↓/j down • enter view • / search • r refresh • q quit",
	)
	b.WriteString(help)

	return b.String()
}

func (m Model) renderDetailView() string {
	var b strings.Builder

	// Title
	if len(m.filteredPages) > 0 && m.cursor < len(m.filteredPages) {
		page := m.filteredPages[m.cursor]
		title := titleStyle.Render(fmt.Sprintf(" %s(%s) ", page.Name, page.Section))
		b.WriteString(title)
		b.WriteString("\n\n")
	}

	// Content viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Help
	help := helpStyle.Render(
		"↑/k up • ↓/j down • g top • G bottom • u/d half page • q/esc back",
	)
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
