// Package app implements the main application model and view navigation logic.
package app

import (
	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/fotmob"
	"github.com/0xjuanma/golazo/internal/ui"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// view represents the current application view.
type view int

const (
	viewMain view = iota
	viewLiveMatches
	viewStats
)

// model holds the application state.
// Fields are organized by concern: display, data, UI components, and configuration.
type model struct {
	// Display dimensions
	width  int
	height int

	// View state
	currentView view
	selected    int

	// Match data
	matches           []ui.MatchDisplay
	upcomingMatches   []ui.MatchDisplay // Upcoming matches for 1-day stats view
	matchDetails      *api.MatchDetails
	matchDetailsCache map[int]*api.MatchDetails // Cache to avoid repeated API calls
	liveUpdates       []string
	lastEvents        []api.MatchEvent

	// UI components
	spinner          spinner.Model
	randomSpinner    *ui.RandomCharSpinner
	statsViewSpinner *ui.RandomCharSpinner // Separate spinner for stats view

	// List components
	liveMatchesList     list.Model
	statsMatchesList    list.Model
	upcomingMatchesList list.Model

	// Loading states
	loading          bool
	mainViewLoading  bool
	liveViewLoading  bool
	statsViewLoading bool
	polling          bool
	pendingSelection int // Tracks which view is being preloaded (-1 = none, 0 = stats, 1 = live)

	// Configuration
	useMockData    bool
	statsDateRange int // 1 or 3 days (default: 1)

	// API clients
	fotmobClient *fotmob.Client
	parser       *fotmob.LiveUpdateParser
}

// New creates a new application model with default values.
// useMockData determines whether to use mock data instead of real API data.
func New(useMockData bool) model {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = ui.SpinnerStyle()

	// Initialize random character spinners
	randomSpinner := ui.NewRandomCharSpinner()
	randomSpinner.SetWidth(30)

	statsViewSpinner := ui.NewRandomCharSpinner()
	statsViewSpinner.SetWidth(30)

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

	upcomingList := list.New([]list.Item{}, delegate, 0, 0)
	upcomingList.Title = "Upcoming Matches"
	upcomingList.SetShowStatusBar(false)
	upcomingList.SetFilteringEnabled(false)

	return model{
		currentView:         viewMain,
		matchDetailsCache:   make(map[int]*api.MatchDetails),
		useMockData:         useMockData,
		fotmobClient:        fotmob.NewClient(),
		parser:              fotmob.NewLiveUpdateParser(),
		spinner:             s,
		randomSpinner:       randomSpinner,
		statsViewSpinner:    statsViewSpinner,
		liveMatchesList:     liveList,
		statsMatchesList:    statsList,
		upcomingMatchesList: upcomingList,
		statsDateRange:      1,
		pendingSelection:    -1, // No pending selection
	}
}

// Init initializes the application.
func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, ui.SpinnerTick())
}
