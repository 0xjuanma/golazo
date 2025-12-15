// Package app implements the main application model and view navigation logic.
package app

import (
	"context"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/footballdata"
	"github.com/0xjuanma/golazo/internal/fotmob"
	"github.com/0xjuanma/golazo/internal/ui"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type view int

const (
	viewMain view = iota
	viewLiveMatches
	viewStats
)

type model struct {
	width              int
	height             int
	currentView        view
	matches            []ui.MatchDisplay
	selected           int
	matchDetails       *api.MatchDetails
	liveUpdates        []string
	spinner            spinner.Model
	randomSpinner      *ui.RandomCharSpinner
	loading            bool
	mainViewLoading    bool
	liveViewLoading    bool
	statsViewLoading   bool
	useMockData        bool
	fotmobClient       *fotmob.Client
	footballDataClient *footballdata.Client
	apiKeyMissing      bool // Track if API key is missing
	parser             *fotmob.LiveUpdateParser
	lastEvents         []api.MatchEvent
	polling            bool
	liveMatchesList    list.Model
	statsMatchesList   list.Model
	statsDateRange     int // 1, 3, or 7 days (default: 1)
}

// NewModel creates a new application model with default values.
// useMockData determines whether to use mock data instead of real API data.
func NewModel(useMockData bool) model {
	s := spinner.New()
	s.Spinner = spinner.Line // More prominent spinner animation
	s.Style = ui.SpinnerStyle()

	// Initialize random character spinner for main view
	randomSpinner := ui.NewRandomCharSpinner()
	randomSpinner.SetWidth(30) // Wider spinner for more characters

	// Initialize API-Sports.io client if API key is available
	var footballDataClient *footballdata.Client
	apiKeyMissing := false
	if apiKey, err := data.FootballDataAPIKey(); err == nil {
		footballDataClient = footballdata.NewClient(apiKey)
	} else {
		apiKeyMissing = true
	}

	// Initialize list models with custom delegate
	delegate := ui.NewMatchListDelegate()

	liveList := list.New([]list.Item{}, delegate, 0, 0)
	liveList.Title = "Live Matches"
	liveList.SetShowStatusBar(false)
	liveList.SetFilteringEnabled(false)

	statsList := list.New([]list.Item{}, delegate, 0, 0)
	statsList.Title = "Finished Matches"
	statsList.SetShowStatusBar(false)
	statsList.SetFilteringEnabled(false)

	return model{
		currentView:        viewMain,
		selected:           0,
		spinner:            s,
		randomSpinner:      randomSpinner,
		liveUpdates:        []string{},
		useMockData:        useMockData,
		fotmobClient:       fotmob.NewClient(),
		footballDataClient: footballDataClient,
		apiKeyMissing:      apiKeyMissing,
		parser:             fotmob.NewLiveUpdateParser(),
		lastEvents:         []api.MatchEvent{},
		liveMatchesList:    liveList,
		statsMatchesList:   statsList,
		statsDateRange:     1, // Default to 1 day
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.randomSpinner.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update list sizes when window size changes
		if m.currentView == viewLiveMatches {
			leftWidth := m.width * 35 / 100
			if leftWidth < 25 {
				leftWidth = 25
			}
			h, v := 2, 2 // Approximate frame size
			titleHeight := 3
			spinnerHeight := 3 // Reserved space at top for spinner
			availableWidth := leftWidth - h*2
			availableHeight := m.height - v*2 - titleHeight - spinnerHeight
			if availableWidth > 0 && availableHeight > 0 {
				m.liveMatchesList.SetSize(availableWidth, availableHeight)
			}
		} else if m.currentView == viewStats {
			leftWidth := m.width * 40 / 100
			if leftWidth < 30 {
				leftWidth = 30
			}
			h, v := 2, 2
			titleHeight := 3
			spinnerHeight := 3 // Reserved space at top for spinner
			availableWidth := leftWidth - h*2
			availableHeight := m.height - v*2 - titleHeight - spinnerHeight
			if availableWidth > 0 && availableHeight > 0 {
				m.statsMatchesList.SetSize(availableWidth, availableHeight)
			}
		}
		return m, nil
	case spinner.TickMsg:
		if m.loading || m.mainViewLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	case liveUpdateMsg:
		if msg.update != "" {
			m.liveUpdates = append(m.liveUpdates, msg.update)
		}
		// Continue polling if match is live
		if m.polling && m.matchDetails != nil && m.matchDetails.Status == api.MatchStatusLive {
			cmds = append(cmds, pollMatchDetails(m.fotmobClient, m.parser, m.matchDetails.ID, m.lastEvents, m.useMockData))
		} else {
			m.loading = false
			m.polling = false
		}
		return m, tea.Batch(cmds...)
	case matchDetailsMsg:
		if msg.details != nil {
			m.matchDetails = msg.details
			m.liveViewLoading = false
			m.statsViewLoading = false

			// Only handle live updates and polling for live matches view
			if m.currentView == viewLiveMatches {
				// If this is the first load (lastEvents is empty), parse all events
				// Otherwise, only parse new events
				var eventsToParse []api.MatchEvent
				if len(m.lastEvents) == 0 {
					// First load: parse all existing events
					eventsToParse = msg.details.Events
				} else {
					// Subsequent loads: only parse new events
					eventsToParse = m.parser.NewEvents(m.lastEvents, msg.details.Events)
				}

				if len(eventsToParse) > 0 {
					// Parse events into updates
					updates := m.parser.ParseEvents(eventsToParse, msg.details.HomeTeam, msg.details.AwayTeam)
					m.liveUpdates = append(m.liveUpdates, updates...)
				}
				m.lastEvents = msg.details.Events

				// Continue polling if match is live
				if msg.details.Status == api.MatchStatusLive {
					m.polling = true
					m.loading = true
					cmds = append(cmds, pollMatchDetails(m.fotmobClient, m.parser, msg.details.ID, m.lastEvents, m.useMockData))
				} else {
					m.loading = false
					m.polling = false
				}
			} else {
				// For stats view, just set loading to false
				m.loading = false
			}
		} else {
			m.loading = false
			m.liveViewLoading = false
			m.statsViewLoading = false
		}
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.currentView != viewMain {
				m.currentView = viewMain
				m.selected = 0
				m.matchDetails = nil
				m.liveUpdates = []string{}
				m.lastEvents = []api.MatchEvent{}
				m.loading = false
				m.polling = false
				m.matches = []ui.MatchDisplay{}
				return m, nil
			}
		}

		// Handle view-specific key events
		switch m.currentView {
		case viewMain:
			return m.handleMainViewKeys(msg)
		case viewLiveMatches:
			// Delegate to list component
			var listCmd tea.Cmd
			m.liveMatchesList, listCmd = m.liveMatchesList.Update(msg)
			// Get selected index from list
			if selectedItem := m.liveMatchesList.SelectedItem(); selectedItem != nil {
				if item, ok := selectedItem.(ui.MatchListItem); ok {
					// Find match index
					for i, match := range m.matches {
						if match.ID == item.Match.ID {
							if i != m.selected {
								m.selected = i
								return m.loadMatchDetails(m.matches[m.selected].ID)
							}
							break
						}
					}
				}
			}
			return m, listCmd
		case viewStats:
			// Handle date range selector navigation first (left/right keys)
			if msg.String() == "h" || msg.String() == "left" || msg.String() == "l" || msg.String() == "right" {
				updatedM, statsCmd := m.handleStatsViewKeys(msg)
				return updatedM, statsCmd
			}
			// Delegate other keys to list component
			var listCmd tea.Cmd
			m.statsMatchesList, listCmd = m.statsMatchesList.Update(msg)
			// Get selected index from list
			if selectedItem := m.statsMatchesList.SelectedItem(); selectedItem != nil {
				if item, ok := selectedItem.(ui.MatchListItem); ok {
					// Find match index
					for i, match := range m.matches {
						if match.ID == item.Match.ID {
							if i != m.selected {
								m.selected = i
								return m.loadStatsMatchDetails(m.matches[m.selected].ID)
							}
							break
						}
					}
				}
			}
			return m, listCmd
		}
	case liveMatchesMsg:
		// Debug: Check if we got matches
		if len(msg.matches) == 0 {
			// No matches found, but stop loading
			m.liveViewLoading = false
			m.loading = false
			return m, nil
		}
		// Keep loading state true until match details are loaded
		// Convert to display format
		displayMatches := make([]ui.MatchDisplay, 0, len(msg.matches))
		for _, match := range msg.matches {
			displayMatches = append(displayMatches, ui.MatchDisplay{
				Match: match,
			})
		}

		m.matches = displayMatches
		m.selected = 0
		m.loading = false
		// Keep liveViewLoading true to show spinner while loading match details
		// Re-initialize spinner to ensure it's animating
		cmds = append(cmds, m.randomSpinner.Init())

		// Update list with items
		items := ui.ToMatchListItems(displayMatches)
		m.liveMatchesList.SetItems(items)

		// Set list size based on current window dimensions
		// Account for spinner height at top (3 lines reserved)
		// Use default size if window size not set yet
		spinnerHeight := 3
		leftWidth := m.width * 35 / 100
		if leftWidth < 25 {
			leftWidth = 25
		}
		if m.width == 0 {
			leftWidth = 40 // Default width if window size not set
		}
		// Approximate frame size (border + padding)
		frameWidth := 4
		frameHeight := 6
		titleHeight := 3
		availableWidth := leftWidth - frameWidth
		availableHeight := m.height - frameHeight - titleHeight - spinnerHeight
		if m.height == 0 {
			availableHeight = 20 // Default height if window size not set
		}
		if availableWidth > 0 && availableHeight > 0 {
			m.liveMatchesList.SetSize(availableWidth, availableHeight)
		}

		if len(displayMatches) > 0 {
			m.liveMatchesList.Select(0)
		}

		// Load details for first match if available
		// This will set liveViewLoading = true again and initialize spinner
		if len(m.matches) > 0 {
			var loadCmd tea.Cmd
			var updatedModel tea.Model
			updatedModel, loadCmd = m.loadMatchDetails(m.matches[0].ID)
			if updatedM, ok := updatedModel.(model); ok {
				m = updatedM
			}
			cmds = append(cmds, loadCmd)
			return m, tea.Batch(cmds...)
		}

		// If no matches to load details for, stop loading
		m.liveViewLoading = false
		return m, tea.Batch(cmds...)
	case finishedMatchesMsg:
		m.statsViewLoading = false
		// Convert to display format
		displayMatches := make([]ui.MatchDisplay, 0, len(msg.matches))
		for _, match := range msg.matches {
			displayMatches = append(displayMatches, ui.MatchDisplay{
				Match: match,
			})
		}

		m.matches = displayMatches
		m.selected = 0
		m.loading = false

		// Update list with items
		m.statsMatchesList.SetItems(ui.ToMatchListItems(displayMatches))
		if len(displayMatches) > 0 {
			m.statsMatchesList.Select(0)
		}

		// Load details for first match if available
		if len(m.matches) > 0 {
			return m.loadStatsMatchDetails(m.matches[0].ID)
		}

		return m, nil
	case ui.TickMsg:
		// Handle random spinner tick for all views
		if m.mainViewLoading || m.liveViewLoading || m.statsViewLoading {
			var cmd tea.Cmd
			var model tea.Model
			model, cmd = m.randomSpinner.Update(msg)
			if spinner, ok := model.(*ui.RandomCharSpinner); ok {
				m.randomSpinner = spinner
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)
	case mainViewCheckMsg:
		// Check completed, navigate to selected view
		m.mainViewLoading = false

		// Clear any previous view state
		m.matches = []ui.MatchDisplay{}
		m.matchDetails = nil
		m.liveUpdates = []string{}
		m.lastEvents = []api.MatchEvent{}
		m.polling = false
		m.selected = 0

		if msg.selection == 0 {
			// Stats view
			m.statsViewLoading = true
			if m.footballDataClient == nil {
				// No API key configured, show empty state
				m.currentView = viewStats
				m.loading = false
				m.statsViewLoading = false
				return m, nil
			}
			m.currentView = viewStats
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, m.randomSpinner.Init(), fetchFinishedMatches(m.footballDataClient, m.useMockData, m.statsDateRange))
		} else if msg.selection == 1 {
			// Live Matches view
			m.currentView = viewLiveMatches
			m.loading = true
			m.liveViewLoading = true
			return m, tea.Batch(m.spinner.Tick, m.randomSpinner.Init(), fetchLiveMatches(m.fotmobClient, m.useMockData))
		}
		return m, nil
	default:
		// Handle RandomCharSpinner TickMsg (from random_spinner.go)
		// This catches ui.TickMsg if the case above doesn't match for some reason
		if _, ok := msg.(ui.TickMsg); ok {
			if m.mainViewLoading || m.liveViewLoading || m.statsViewLoading {
				var cmd tea.Cmd
				var model tea.Model
				model, cmd = m.randomSpinner.Update(msg)
				if spinner, ok := model.(*ui.RandomCharSpinner); ok {
					m.randomSpinner = spinner
				}
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return m, tea.Batch(cmds...)
			}
		}
		return m, nil
	}
	return m, nil
}

// handleMainViewKeys processes keyboard input for the main menu view.
// Handles navigation (up/down) and selection (enter) to switch between views.
func (m model) handleMainViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selected < 1 && !m.mainViewLoading {
			m.selected++
		}
		return m, nil
	case "k", "up":
		if m.selected > 0 && !m.mainViewLoading {
			m.selected--
		}
		return m, nil
	case "enter":
		if m.mainViewLoading {
			return m, nil // Ignore enter while loading
		}
		if m.selected == 0 {
			// Stats - start loading check
			m.mainViewLoading = true
			return m, tea.Batch(m.spinner.Tick, performMainViewCheck(0))
		} else if m.selected == 1 {
			// Live Matches - start loading check
			m.mainViewLoading = true
			return m, tea.Batch(m.spinner.Tick, performMainViewCheck(1))
		}
		return m, nil
	}
	return m, nil
}

// handleLiveMatchesKeys processes keyboard input for the live matches view.
// Handles navigation between matches and loading match details on selection.
// Note: This function is currently unused as list component handles navigation directly.
func (m model) handleLiveMatchesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selected < len(m.matches)-1 {
			m.selected++
			// Load details for newly selected match
			if m.selected < len(m.matches) {
				return m.loadMatchDetails(m.matches[m.selected].ID)
			}
		}
		return m, nil
	case "k", "up":
		if m.selected > 0 {
			m.selected--
			// Load details for newly selected match
			if m.selected >= 0 && m.selected < len(m.matches) {
				return m.loadMatchDetails(m.matches[m.selected].ID)
			}
		}
		return m, nil
	}
	return m, nil
}

// handleStatsViewKeys processes keyboard input for the stats view.
// Handles date range navigation (left/right) to change the time period for finished matches.
func (m model) handleStatsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "left":
		// Navigate date range selector left (cycle backwards: 1 -> 7 -> 3 -> 1)
		if m.statsDateRange == 1 {
			m.statsDateRange = 7
		} else if m.statsDateRange == 3 {
			m.statsDateRange = 1
		} else if m.statsDateRange == 7 {
			m.statsDateRange = 3
		}
		// Reload matches with new date range
		m.statsViewLoading = true
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.randomSpinner.Init(), fetchFinishedMatches(m.footballDataClient, m.useMockData, m.statsDateRange))
	case "l", "right":
		// Navigate date range selector right (cycle forwards: 1 -> 3 -> 7 -> 1)
		if m.statsDateRange == 1 {
			m.statsDateRange = 3
		} else if m.statsDateRange == 3 {
			m.statsDateRange = 7
		} else if m.statsDateRange == 7 {
			m.statsDateRange = 1
		}
		// Reload matches with new date range
		m.statsViewLoading = true
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.randomSpinner.Init(), fetchFinishedMatches(m.footballDataClient, m.useMockData, m.statsDateRange))
	}
	return m, nil
}

// loadMatchDetails loads match details for the live matches view and starts live updates polling.
// Resets live updates and event history before fetching new details.
func (m model) loadMatchDetails(matchID int) (tea.Model, tea.Cmd) {
	m.liveUpdates = []string{}
	m.lastEvents = []api.MatchEvent{}
	m.loading = true
	m.liveViewLoading = true
	return m, tea.Batch(m.spinner.Tick, m.randomSpinner.Init(), fetchMatchDetails(m.fotmobClient, matchID, m.useMockData))
}

// loadStatsMatchDetails loads match details for the stats view.
// Fetches detailed statistics for a finished match.
func (m model) loadStatsMatchDetails(matchID int) (tea.Model, tea.Cmd) {
	m.loading = true
	m.statsViewLoading = true
	return m, tea.Batch(m.spinner.Tick, m.randomSpinner.Init(), fetchStatsMatchDetails(m.footballDataClient, matchID, m.useMockData))
}

func (m model) View() string {
	switch m.currentView {
	case viewMain:
		return ui.RenderMainMenu(m.width, m.height, m.selected, m.spinner, m.randomSpinner, m.mainViewLoading)
	case viewLiveMatches:
		// Ensure list size is set before rendering (in case window size changed or wasn't set)
		if m.width > 0 && m.height > 0 {
			leftWidth := m.width * 35 / 100
			if leftWidth < 25 {
				leftWidth = 25
			}
			h, v := 2, 2
			titleHeight := 3
			spinnerHeight := 3
			availableWidth := leftWidth - h*2
			availableHeight := m.height - v*2 - titleHeight - spinnerHeight
			if availableWidth > 0 && availableHeight > 0 {
				m.liveMatchesList.SetSize(availableWidth, availableHeight)
			}
		}
		return ui.RenderMultiPanelViewWithList(m.width, m.height, m.liveMatchesList, m.matchDetails, m.liveUpdates, m.spinner, m.loading, m.randomSpinner, m.liveViewLoading)
	case viewStats:
		// Ensure list size is set before rendering (in case window size changed or wasn't set)
		if m.width > 0 && m.height > 0 {
			leftWidth := m.width * 40 / 100
			if leftWidth < 30 {
				leftWidth = 30
			}
			h, v := 2, 2
			titleHeight := 3
			spinnerHeight := 3
			availableWidth := leftWidth - h*2
			availableHeight := m.height - v*2 - titleHeight - spinnerHeight
			if availableWidth > 0 && availableHeight > 0 {
				m.statsMatchesList.SetSize(availableWidth, availableHeight)
			}
		}
		return ui.RenderStatsViewWithList(m.width, m.height, m.statsMatchesList, m.matchDetails, m.randomSpinner, m.statsViewLoading, m.statsDateRange, m.apiKeyMissing)
	default:
		return ui.RenderMainMenu(m.width, m.height, m.selected, m.spinner, m.randomSpinner, m.mainViewLoading)
	}
}

// liveUpdateMsg is a message containing a live update string.
type liveUpdateMsg struct {
	update string
}

// matchDetailsMsg is a message containing match details.
type matchDetailsMsg struct {
	details *api.MatchDetails
}

// fetchLiveMatches fetches live matches from the API.
// If useMockData is true, always uses mock data.
// If useMockData is false, uses real API data (no fallback to mock data).
func fetchLiveMatches(client *fotmob.Client, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		// Use mock data if flag is set
		if useMockData {
			matches := data.MockLiveMatches()
			return liveMatchesMsg{matches: matches}
		}

		// If client is not available and not using mock data, return empty
		if client == nil {
			return liveMatchesMsg{matches: []api.Match{}}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		matches, err := client.LiveMatches(ctx)
		if err != nil {
			// Return empty on error when not using mock data
			return liveMatchesMsg{matches: []api.Match{}}
		}

		// Return actual API results (even if empty)
		return liveMatchesMsg{matches: matches}
	}
}

// liveMatchesMsg is a message containing live matches.
type liveMatchesMsg struct {
	matches []api.Match
}

// fetchMatchDetails fetches match details from the API.
// If useMockData is true, always uses mock data.
// If useMockData is false, uses real API data (no fallback to mock data).
func fetchMatchDetails(client *fotmob.Client, matchID int, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		// Use mock data if flag is set
		if useMockData {
			details, _ := data.MockMatchDetails(matchID)
			return matchDetailsMsg{details: details}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		details, err := client.MatchDetails(ctx, matchID)
		if err != nil {
			// Return nil on error when not using mock data
			return matchDetailsMsg{details: nil}
		}

		return matchDetailsMsg{details: details}
	}
}

// pollMatchDetails polls match details every 90 seconds for live updates.
// Conservative interval to avoid rate limiting (90 seconds = 1.5 minutes).
// If useMockData is true, always uses mock data.
func pollMatchDetails(client *fotmob.Client, parser *fotmob.LiveUpdateParser, matchID int, lastEvents []api.MatchEvent, useMockData bool) tea.Cmd {
	return tea.Tick(90*time.Second, func(t time.Time) tea.Msg {
		// Use mock data if flag is set
		if useMockData {
			details, _ := data.MockMatchDetails(matchID)
			return matchDetailsMsg{details: details}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		details, err := client.MatchDetails(ctx, matchID)
		if err != nil {
			return matchDetailsMsg{details: nil}
		}

		// Return match details - new events will be detected in the Update handler
		return matchDetailsMsg{details: details}
	})
}

// finishedMatchesMsg is a message containing finished matches.
type finishedMatchesMsg struct {
	matches []api.Match
}

// fetchFinishedMatches fetches finished matches from the API-Sports.io API.
// If useMockData is true, always uses mock data.
// If useMockData is false, uses real API data (no fallback to mock data).
// days specifies how many days to fetch (1, 3, or 7).
func fetchFinishedMatches(client *footballdata.Client, useMockData bool, days int) tea.Cmd {
	return func() tea.Msg {
		// Use mock data if flag is set
		if useMockData {
			matches := data.MockFinishedMatches()
			return finishedMatchesMsg{matches: matches}
		}

		// If client is not available and not using mock data, return empty
		if client == nil {
			return finishedMatchesMsg{matches: []api.Match{}}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Fetch matches from last N days
		matches, err := client.RecentFinishedMatches(ctx, days)
		if err != nil {
			// Return empty on error when not using mock data
			return finishedMatchesMsg{matches: []api.Match{}}
		}

		// Return actual API results
		return finishedMatchesMsg{matches: matches}
	}
}

// fetchStatsMatchDetails fetches match details from the API-Sports.io API.
// If useMockData is true, always uses mock data.
// If useMockData is false, uses real API data (no fallback to mock data).
func fetchStatsMatchDetails(client *footballdata.Client, matchID int, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		// Use mock data if flag is set
		if useMockData {
			details, _ := data.MockFinishedMatchDetails(matchID)
			return matchDetailsMsg{details: details}
		}

		// If client is not available and not using mock data, return nil
		if client == nil {
			return matchDetailsMsg{details: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		details, err := client.MatchDetails(ctx, matchID)
		if err != nil {
			// Return nil on error when not using mock data
			return matchDetailsMsg{details: nil}
		}

		return matchDetailsMsg{details: details}
	}
}
